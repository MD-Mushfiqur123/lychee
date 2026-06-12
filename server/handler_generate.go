package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lychee/lychee/api"
	"github.com/lychee/lychee/envconfig"
	"github.com/lychee/lychee/fs"
	"github.com/lychee/lychee/fs/ggml"
	internalcloud "github.com/lychee/lychee/internal/cloud"
	"github.com/lychee/lychee/llm"
	"github.com/lychee/lychee/model/parsers"
	"github.com/lychee/lychee/template"
	"github.com/lychee/lychee/thinking"
	"github.com/lychee/lychee/types/errtypes"
	"github.com/lychee/lychee/types/model"
)

func validateGenerateRequest(req *api.GenerateRequest) error {
	if req.TopLogprobs < 0 || req.TopLogprobs > 20 {
		return errors.New("top_logprobs must be between 0 and 20")
	}
	return nil
}

func (s *Server) GenerateHandler(c *gin.Context) {
	checkpointStart := time.Now()
	var req api.GenerateRequest
	if err := c.ShouldBindJSON(&req); errors.Is(err, io.EOF) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateGenerateRequest(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if s.modelAliases != nil {
		req.Model = s.modelAliases.Resolve(req.Model)
	}

	var releaseRoute func() = func() {}
	if s.modelRouter != nil {
		originalModel := req.Model
		endpoint, release, err := s.modelRouter.Resolve(req.Model)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if endpoint != nil {
			releaseRoute = release
			req.Model = endpoint.Model
			if endpoint.Host != "" {
				defer releaseRoute()
				remoteURL, err := url.Parse(endpoint.Host)
				if err != nil {
					s.modelRouter.RecordFailure(originalModel, endpoint.Host)
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				contentType := "application/x-ndjson"
				if req.Stream != nil && !*req.Stream {
					contentType = "application/json; charset=utf-8"
				}
				c.Header("Content-Type", contentType)

				fn := func(resp api.GenerateResponse) error {
					data, err := json.Marshal(resp)
					if err != nil {
						return err
					}
					if _, err = c.Writer.Write(append(data, '\n')); err != nil {
						return err
					}
					c.Writer.Flush()
					return nil
				}

				client := api.NewClient(remoteURL, http.DefaultClient)
				err = client.Generate(c, &req, fn)
				if err != nil {
					s.modelRouter.RecordFailure(originalModel, endpoint.Host)
					var apiError api.StatusError
					if errors.As(err, &apiError) {
						c.JSON(apiError.StatusCode, apiError)
						return
					}
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				s.modelRouter.RecordSuccess(originalModel, endpoint.Host)
				return
			}
		}
	}
	defer releaseRoute()

	modelRef, err := parseAndValidateModelRef(req.Model)
	if err != nil {
		writeModelRefParseError(c, err, http.StatusNotFound, fmt.Sprintf("model '%s' not found", req.Model))
		return
	}

	if modelRef.Source == modelSourceCloud {
		// TODO(drifkin): evaluate an `/api/*` passthrough for cloud where the
		// original body (modulo model name normalization) is sent to cloud.
		req.Model = modelRef.Base
		proxyCloudJSONRequest(c, req, cloudErrRemoteInferenceUnavailable)
		return
	}

	name := modelRef.Name

	// We cannot currently consolidate this into GetModel because all we'll
	// induce infinite recursion given the current code structure.
	name, err = getExistingName(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model '%s' not found", req.Model)})
		return
	}

	m, err := GetModel(name.String())
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model '%s' not found", req.Model)})
		case err.Error() == errtypes.InvalidModelNameErrMsg:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if modelRef.Source == modelSourceLocal && m.Config.RemoteHost != "" && m.Config.RemoteModel != "" {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model '%s' not found", req.Model)})
		return
	}

	if m.Config.RemoteHost != "" && m.Config.RemoteModel != "" {
		if disabled, _ := internalcloud.Status(); disabled {
			c.JSON(http.StatusForbidden, gin.H{"error": internalcloud.DisabledError(cloudErrRemoteInferenceUnavailable)})
			return
		}

		origModel := req.Model

		remoteURL, err := url.Parse(m.Config.RemoteHost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if !slices.Contains(envconfig.Remotes(), remoteURL.Hostname()) {
			slog.Info("remote model", "remotes", envconfig.Remotes(), "remoteURL", m.Config.RemoteHost, "hostname", remoteURL.Hostname())
			c.JSON(http.StatusBadRequest, gin.H{"error": "this server cannot run this remote model"})
			return
		}

		req.Model = m.Config.RemoteModel

		if req.Template == "" && m.Template.String() != "" {
			req.Template = m.Template.String()
		}

		if req.Options == nil {
			req.Options = map[string]any{}
		}

		for k, v := range m.Options {
			if _, ok := req.Options[k]; !ok {
				req.Options[k] = v
			}
		}

		// update the system prompt from the model if one isn't already specified
		if req.System == "" && m.System != "" {
			req.System = m.System
		}

		if len(m.Messages) > 0 {
			slog.Warn("embedded messages in the model not supported with '/api/generate'; try '/api/chat' instead")
		}

		contentType := "application/x-ndjson"
		if req.Stream != nil && !*req.Stream {
			contentType = "application/json; charset=utf-8"
		}
		c.Header("Content-Type", contentType)

		fn := func(resp api.GenerateResponse) error {
			resp.Model = origModel
			resp.RemoteModel = m.Config.RemoteModel
			resp.RemoteHost = m.Config.RemoteHost

			data, err := json.Marshal(resp)
			if err != nil {
				return err
			}

			if _, err = c.Writer.Write(append(data, '\n')); err != nil {
				return err
			}
			c.Writer.Flush()
			return nil
		}

		client := api.NewClient(remoteURL, http.DefaultClient)
		err = client.Generate(c, &req, fn)
		if err != nil {
			var authError api.AuthorizationError
			if errors.As(err, &authError) {
				sURL, sErr := signinURL()
				if sErr != nil {
					slog.Error(sErr.Error())
					c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting authorization details"})
					return
				}

				c.JSON(authError.StatusCode, gin.H{"error": "unauthorized", "signin_url": sURL})
				return
			}
			var apiError api.StatusError
			if errors.As(err, &apiError) {
				c.JSON(apiError.StatusCode, apiError)
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		return
	}

	// expire the runner if unload is requested (empty prompt, keep alive is 0)
	if req.Prompt == "" && req.KeepAlive != nil && req.KeepAlive.Duration == 0 {
		s.sched.expireRunner(m)

		c.JSON(http.StatusOK, api.GenerateResponse{
			Model:      req.Model,
			CreatedAt:  time.Now().UTC(),
			Response:   "",
			Done:       true,
			DoneReason: "unload",
		})
		return
	}

	// Handle image generation models
	if slices.Contains(m.Capabilities(), model.CapabilityImage) {
		s.handleImageGenerate(c, req, name.String(), checkpointStart)
		return
	}

	if req.Raw && (req.Template != "" || req.System != "" || len(req.Context) > 0) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "raw mode does not support template, system, or context"})
		return
	}

	var builtinParser parsers.Parser
	if shouldUseHarmony(m) {
		// harmony's Reasoning field only understands low/medium/high; map "max" to "high"
		if req.Think != nil {
			if s, ok := req.Think.Value.(string); ok && s == "max" {
				req.Think.Value = "high"
			}
		}
		if m.Config.Parser == "" {
			m.Config.Parser = "harmony"
		}
	}

	if !req.Raw && m.Config.Parser != "" {
		builtinParser = parsers.ParserForName(m.Config.Parser)
		if builtinParser != nil {
			// no tools or last message for generate endpoint
			builtinParser.Init(nil, nil, req.Think)
		}
	}

	caps := []model.Capability{model.CapabilityCompletion}
	if req.Suffix != "" {
		caps = append(caps, model.CapabilityInsert)
	}

	modelCaps := m.Capabilities()
	if slices.Contains(modelCaps, model.CapabilityThinking) {
		caps = append(caps, model.CapabilityThinking)
		if req.Think == nil {
			req.Think = &api.ThinkValue{Value: true}
		}
	} else {
		if req.Think != nil && req.Think.Bool() {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%q does not support thinking", req.Model)})
			return
		}
	}

	if HasCacheControlGenerate(&req) {
		trueVal := true
		req.Shift = &trueVal
	}

	r, m, opts, err := s.scheduleRunner(c.Request.Context(), name.String(), caps, req.Options, req.KeepAlive, req.Shift)
	if errors.Is(err, errCapabilityCompletion) {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%q does not support generate", req.Model)})
		return
	} else if err != nil {
		handleScheduleError(c, req.Model, err)
		return
	}

	checkpointLoaded := time.Now()

	// load the model
	if req.Prompt == "" {
		c.JSON(http.StatusOK, api.GenerateResponse{
			Model:      req.Model,
			CreatedAt:  time.Now().UTC(),
			Done:       true,
			DoneReason: "load",
		})
		return
	}

	if slices.Contains(m.Config.ModelFamilies, "mllama") && len(req.Images) > 1 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "this model only supports one image while more than one image requested"})
		return
	}

	media := make([]llm.MediaData, len(req.Images))
	for i := range req.Images {
		media[i] = llm.NewMediaData(i, req.Images[i])
	}

	prompt := req.Prompt
	var leadingBOS string
	if !req.Raw {
		tmpl := m.Template
		if req.Template != "" {
			tmpl, err = template.Parse(req.Template)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		var values template.Values
		if req.Suffix != "" {
			values.Prompt = prompt
			values.Suffix = req.Suffix
		} else {
			var msgs []api.Message
			if req.System != "" {
				msgs = append(msgs, api.Message{Role: "system", Content: req.System})
			} else if m.System != "" {
				msgs = append(msgs, api.Message{Role: "system", Content: m.System})
			}

			if req.Context == nil {
				msgs = append(msgs, m.Messages...)
			}

			userMsg := api.Message{Role: "user", Content: req.Prompt}
			for _, m := range media {
				userMsg.Images = append(userMsg.Images, m.Data)
			}
			values.Messages = append(msgs, userMsg)
		}

		values.Think = req.Think != nil && req.Think.Bool()
		values.ThinkLevel = ""
		if req.Think != nil {
			values.ThinkLevel = req.Think.String()
		}
		values.IsThinkSet = req.Think != nil

		var b bytes.Buffer
		if req.Context != nil {
			slog.Warn("the context field is deprecated and will be removed in a future version of Lychee")
			s, err := r.Detokenize(c.Request.Context(), req.Context)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			b.WriteString(s)
		}

		// check that we're in the `api/chat`-like flow, and if so, generate the
		// prompt the same way
		// TEMP(drifkin): we should really just detect the chat-like flow and call
		// the real chat handler, but doing this as a stopgap to get renderer
		// support for generate
		if values.Messages != nil && values.Suffix == "" && req.Template == "" {
			genTruncate := (req.Truncate == nil || *req.Truncate) && !m.IsMLX()
			prompt, media, err = chatPrompt(c.Request.Context(), m, r.Tokenize, optionsForPrompt(opts, r), values.Messages, []api.Tool{}, req.Think, genTruncate)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			// TEMP(drifkin): req.Context will be removed very soon, but we're temporarily supporting it in this flow here
			if req.Context != nil {
				b.WriteString(prompt)
				prompt = b.String()
			}
			leadingBOS = leadingBOSForModel(m)
		} else {
			// Direct template execution flow.
			if err := tmpl.Execute(&b, values); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			prompt = b.String()
		}
	}

	// If debug mode is enabled, return the rendered template instead of calling the model
	if req.DebugRenderOnly {
		c.JSON(http.StatusOK, api.GenerateResponse{
			Model:     req.Model,
			CreatedAt: time.Now().UTC(),
			DebugInfo: &api.DebugInfo{
				RenderedTemplate: prompt,
				ImageCount:       len(media),
			},
		})
		return
	}

	var thinkingState *thinking.Parser
	if builtinParser == nil {
		openingTag, closingTag := thinking.InferTags(m.Template.Template)
		if req.Think != nil && req.Think.Bool() && openingTag != "" && closingTag != "" {
			thinkingState = &thinking.Parser{
				OpeningTag: openingTag,
				ClosingTag: closingTag,
			}
			if strings.HasSuffix(strings.TrimSpace(prompt), openingTag) {
				thinkingState.AddContent(openingTag)
			}
		}
	}

	ch := make(chan any)
	go func() {
		// TODO (jmorganca): avoid building the response twice both here and below
		var sb strings.Builder
		defer close(ch)
		if err := r.Completion(c.Request.Context(), llm.CompletionRequest{
			Prompt:          prompt,
			Media:           media,
			Format:          req.Format,
			Options:         opts,
			Shift:           req.Shift == nil || *req.Shift,
			Truncate:        req.Truncate == nil || *req.Truncate,
			Logprobs:        req.Logprobs,
			TopLogprobs:     req.TopLogprobs,
			PreservedTokens: preservedTokensForCompletion(builtinParser),
			LeadingBOS:      leadingBOS,
		}, func(cr llm.CompletionResponse) {
			res := api.GenerateResponse{
				Model:     req.Model,
				CreatedAt: time.Now().UTC(),
				Response:  cr.Content,
				Done:      cr.Done,
				Metrics: api.Metrics{
					PromptEvalCount:    cr.PromptEvalCount,
					PromptEvalDuration: cr.PromptEvalDuration,
					EvalCount:          cr.EvalCount,
					EvalDuration:       cr.EvalDuration,
				},
				Logprobs: toAPILogprobs(cr.Logprobs),
			}

			if builtinParser != nil {
				content, thinking, toolCalls, err := builtinParser.Add(cr.Content, cr.Done)
				if err != nil {
					ch <- gin.H{"error": err.Error()}
					return
				}
				res.Response = content
				res.Thinking = thinking
				if cr.Done && len(toolCalls) > 0 {
					res.ToolCalls = toolCalls
				}
			} else if thinkingState != nil {
				thinking, content := thinkingState.AddContent(cr.Content)
				res.Thinking = thinking
				res.Response = content
			}

			if _, err := sb.WriteString(cr.Content); err != nil {
				ch <- gin.H{"error": err.Error()}
			}

			if cr.Done {
				res.DoneReason = cr.DoneReason.String()
				res.TotalDuration = time.Since(checkpointStart)
				res.LoadDuration = checkpointLoaded.Sub(checkpointStart)

				if !req.Raw {
					tokens, err := r.Tokenize(c.Request.Context(), prompt+sb.String())
					if err != nil {
						ch <- gin.H{"error": err.Error()}
						return
					}
					res.Context = tokens
				}
			}

			if builtinParser != nil {
				// Emit chunks that carry logprobs even if the parser is still buffering
				// visible content, otherwise generate logprobs disappear for models with
				// builtin thinking/tool parsers.
				if res.Response != "" || res.Thinking != "" || res.Done || len(res.ToolCalls) > 0 || len(res.Logprobs) > 0 {
					ch <- res
				}

				return
			}

			ch <- res
		}); err != nil {
			s.sched.expireRunnersForRuntimeOOM(m, err)
			var serr api.StatusError
			if errors.As(err, &serr) {
				ch <- gin.H{"error": serr.ErrorMessage, "status": serr.StatusCode}
			} else {
				ch <- gin.H{"error": err.Error()}
			}
		}
	}()

	if req.Stream != nil && !*req.Stream {
		var r api.GenerateResponse
		var allLogprobs []api.Logprob
		var sbThinking strings.Builder
		var sbContent strings.Builder
		for rr := range ch {
			switch t := rr.(type) {
			case api.GenerateResponse:
				sbThinking.WriteString(t.Thinking)
				sbContent.WriteString(t.Response)
				r = t
				// Accumulate logprobs from all chunks for non-streaming response
				if len(t.Logprobs) > 0 {
					allLogprobs = append(allLogprobs, t.Logprobs...)
				}
			case gin.H:
				msg, ok := t["error"].(string)
				if !ok {
					msg = "unexpected error format in response"
				}

				status, ok := t["status"].(int)
				if !ok {
					status = http.StatusInternalServerError
				}

				c.JSON(status, gin.H{"error": msg})
				return
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected response"})
				return
			}
		}

		r.Thinking = sbThinking.String()
		r.Response = sbContent.String()
		r.Logprobs = allLogprobs

		if req.ValidateOutput {
			if err := ValidateJSONSchema(r.Response, req.Format); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("schema validation failed: %v", err)})
				return
			}
		}

		c.JSON(http.StatusOK, r)
		return
	}

	streamResponse(c, ch)
}
