package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lychee/lychee/api"
	"github.com/lychee/lychee/llm"
	chainmodel "github.com/lychee/lychee/model"
	"github.com/lychee/lychee/template"
	"github.com/lychee/lychee/types/model"
)

// ComposeHandler handles sequentially executing multiple models.
func (s *Server) ComposeHandler(c *gin.Context) {
	var req api.ComposeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Steps) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one step is required"})
		return
	}

	runStep := func(ctx context.Context, modelName string, prompt string, options map[string]any, onChunk func(string)) (string, error) {
		modelRef, err := parseAndValidateModelRef(modelName)
		if err != nil {
			return "", err
		}
		name := modelRef.Name
		name, err = getExistingName(name)
		if err != nil {
			return "", err
		}
		m, err := GetModel(name.String())
		if err != nil {
			return "", err
		}

		caps := []model.Capability{model.CapabilityCompletion}
		r, m, opts, err := s.scheduleRunner(ctx, name.String(), caps, options, nil, nil)
		if err != nil {
			return "", err
		}

		var values template.Values
		userMsg := api.Message{Role: "user", Content: prompt}
		values.Messages = append(m.Messages, userMsg)
		if m.System != "" {
			values.Messages = append([]api.Message{{Role: "system", Content: m.System}}, values.Messages...)
		}

		genTruncate := !m.IsMLX()
		renderedPrompt, media, err := chatPrompt(ctx, m, r.Tokenize, optionsForPrompt(opts, r), values.Messages, []api.Tool{}, nil, genTruncate)
		if err != nil {
			renderedPrompt = prompt
		}
		leadingBOS := leadingBOSForModel(m)

		var responseSB strings.Builder
		err = r.Completion(ctx, llm.CompletionRequest{
			Prompt:     renderedPrompt,
			Media:      media,
			Options:    opts,
			LeadingBOS: leadingBOS,
		}, func(cr llm.CompletionResponse) {
			responseSB.WriteString(cr.Content)
			if onChunk != nil {
				onChunk(cr.Content)
			}
		})
		if err != nil {
			return "", err
		}

		return responseSB.String(), nil
	}

	if req.Stream {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")

		onEvent := func(event chainmodel.ComposeEvent) {
			jsonData, err := json.Marshal(event)
			if err != nil {
				return
			}
			c.SSEvent("message", string(jsonData))
			c.Writer.Flush()
		}

		_, err := chainmodel.ExecuteChain(c.Request.Context(), &req, runStep, onEvent)
		if err != nil {
			errEvent := chainmodel.ComposeEvent{
				Event: "error",
				Text:  err.Error(),
			}
			jsonData, _ := json.Marshal(errEvent)
			c.SSEvent("message", string(jsonData))
			c.Writer.Flush()
		}
		return
	}

	resp, err := chainmodel.ExecuteChain(c.Request.Context(), &req, runStep, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
