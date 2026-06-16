package server

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/MD-Mushfiqur123/lychee/types/model"
)

func (s *Server) EmbedHandler(c *gin.Context) {
	checkpointStart := time.Now()
	var req api.EmbedRequest
	err := c.ShouldBindJSON(&req)
	switch {
	case errors.Is(err, io.EOF):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	case err != nil:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	modelRef, err := parseAndValidateModelRef(req.Model)
	if err != nil {
		writeModelRefParseError(c, err, http.StatusNotFound, fmt.Sprintf("model '%s' not found", req.Model))
		return
	}

	if modelRef.Source == modelSourceCloud {
		req.Model = modelRef.Base
		proxyCloudJSONRequest(c, req, cloudErrRemoteInferenceUnavailable)
		return
	}

	var input []string

	switch i := req.Input.(type) {
	case string:
		if len(i) > 0 {
			input = append(input, i)
		}
	case []any:
		for _, v := range i {
			if _, ok := v.(string); !ok {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid input type"})
				return
			}
			input = append(input, v.(string))
		}
	default:
		if req.Input != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid input type"})
			return
		}
	}

	name, err := getExistingName(modelRef.Name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model '%s' not found", req.Model)})
		return
	}

	r, m, opts, err := s.scheduleRunner(c.Request.Context(), name.String(), []model.Capability{}, req.Options, req.KeepAlive, nil)
	if err != nil {
		handleScheduleError(c, req.Model, err)
		return
	}

	checkpointLoaded := time.Now()

	if len(input) == 0 {
		c.JSON(http.StatusOK, api.EmbedResponse{Model: req.Model, Embeddings: [][]float32{}})
		return
	}

	kvData, _, err := getModelData(m.ModelPath, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	adjustTokenLimit := func(tokens []int, limit int) int {
		if bos := kvData.Uint("tokenizer.ggml.bos_token_id"); len(tokens) > 0 && tokens[0] != int(bos) && kvData.Bool("add_bos_token", true) {
			limit--
		}
		if eos := kvData.Uint("tokenizer.ggml.eos_token_id"); len(tokens) > 0 && tokens[len(tokens)-1] != int(eos) && kvData.Bool("add_eos_token", true) {
			limit--
		}
		return limit
	}

	inputTokensAndContext := func(text string) ([]int, int, error) {
		tokens, err := r.Tokenize(ctx, text)
		if err != nil {
			return nil, 0, err
		}

		// TODO @nicolepardal: avoid reaching into kvData here; pass required tokenizer metadata via model/options instead
		ctxLen := int(kvData.ContextLength())
		if opts.NumCtx > 0 {
			ctxLen = min(opts.NumCtx, ctxLen)
		}

		return tokens, adjustTokenLimit(tokens, ctxLen), nil
	}

	truncateInputToLimit := func(text string, limit int) (string, bool, error) {
		tokens, ctxLen, err := inputTokensAndContext(text)
		if err != nil {
			return "", false, err
		}
		if limit > 0 {
			ctxLen = min(ctxLen, adjustTokenLimit(tokens, limit))
		}

		if ctxLen <= 0 {
			return "", false, fmt.Errorf("input after truncation exceeds maximum context length")
		}
		if len(tokens) <= ctxLen {
			return text, false, nil
		}

		truncated, err := r.Detokenize(ctx, tokens[:ctxLen])
		if err != nil {
			return "", false, err
		}
		return truncated, true, nil
	}

	truncateInput := func(text string) (string, bool, error) {
		return truncateInputToLimit(text, 0)
	}

	embedWithRetry := func(text string) ([]float32, int, error) {
		if req.Truncate != nil && !*req.Truncate {
			tokens, ctxLen, err := inputTokensAndContext(text)
			if err != nil {
				return nil, 0, err
			}
			if ctxLen <= 0 {
				return nil, 0, fmt.Errorf("input after truncation exceeds maximum context length")
			}
			if len(tokens) > ctxLen {
				return nil, 0, api.StatusError{
					StatusCode:   http.StatusBadRequest,
					ErrorMessage: "the input length exceeds the context length",
				}
			}
		} else {
			var err error
			text, _, err = truncateInput(text)
			if err != nil {
				return nil, 0, err
			}
		}

		emb, tokCount, err := r.Embedding(ctx, text)
		if err == nil {
			return emb, tokCount, nil
		}

		var serr api.StatusError
		if !errors.As(err, &serr) || serr.StatusCode != http.StatusBadRequest {
			return nil, 0, err
		}
		if req.Truncate != nil && !*req.Truncate {
			return nil, 0, err
		}

		truncated, ok, err := truncateInputToLimit(text, opts.NumBatch)
		if err != nil {
			return nil, 0, err
		}
		if !ok {
			return nil, 0, fmt.Errorf("input exceeds maximum context length and cannot be truncated further")
		}

		return r.Embedding(ctx, truncated)
	}

	var g errgroup.Group
	embeddings := make([][]float32, len(input))
	var totalTokens uint64
	for i, text := range input {
		g.Go(func() error {
			embedding, tokenCount, err := embedWithRetry(text)
			if err != nil {
				return err
			}
			// TODO: this first normalization should be done by the model
			embedding, err = normalize(embedding)
			if err != nil {
				return err
			}
			if req.Dimensions > 0 && req.Dimensions < len(embedding) {
				embedding, err = normalize(embedding[:req.Dimensions])
				if err != nil {
					return err
				}
			}
			embeddings[i] = embedding
			atomic.AddUint64(&totalTokens, uint64(tokenCount))
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		s.sched.expireRunnersForRuntimeOOM(m, err)
		var serr api.StatusError
		if errors.As(err, &serr) {
			c.AbortWithStatusJSON(serr.StatusCode, gin.H{
				"error": strings.TrimSpace(serr.ErrorMessage),
			})
			return
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": strings.TrimSpace(err.Error()),
		})
		return
	}

	resp := api.EmbedResponse{
		Model:           req.Model,
		Embeddings:      embeddings,
		TotalDuration:   time.Since(checkpointStart),
		LoadDuration:    checkpointLoaded.Sub(checkpointStart),
		PromptEvalCount: int(totalTokens),
	}
	c.JSON(http.StatusOK, resp)
}

func normalize(vec []float32) ([]float32, error) {
	var sum float32
	for _, v := range vec {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return nil, errors.New("embedding contains NaN or Inf values")
		}
		sum += v * v
	}

	norm := float32(1.0 / max(math.Sqrt(float64(sum)), 1e-12))
	for i := range vec {
		vec[i] *= norm
	}
	return vec, nil
}

func (s *Server) EmbeddingsHandler(c *gin.Context) {
	var req api.EmbeddingRequest
	if err := c.ShouldBindJSON(&req); errors.Is(err, io.EOF) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	modelRef, err := parseAndValidateModelRef(req.Model)
	if err != nil {
		writeModelRefParseError(c, err, http.StatusBadRequest, "model is required")
		return
	}

	if modelRef.Source == modelSourceCloud {
		req.Model = modelRef.Base
		proxyCloudJSONRequest(c, req, cloudErrRemoteInferenceUnavailable)
		return
	}

	name := modelRef.Name

	r, m, _, err := s.scheduleRunner(c.Request.Context(), name.String(), []model.Capability{}, req.Options, req.KeepAlive, nil)
	if err != nil {
		handleScheduleError(c, req.Model, err)
		return
	}

	// an empty request loads the model
	if req.Prompt == "" {
		c.JSON(http.StatusOK, api.EmbeddingResponse{Embedding: []float64{}})
		return
	}

	embedding, _, err := r.Embedding(c.Request.Context(), req.Prompt)
	if err != nil {
		s.sched.expireRunnersForRuntimeOOM(m, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": strings.TrimSpace(err.Error())})
		return
	}

	var e []float64
	for _, v := range embedding {
		e = append(e, float64(v))
	}

	resp := api.EmbeddingResponse{
		Embedding: e,
	}
	c.JSON(http.StatusOK, resp)
}
