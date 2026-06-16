package server

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/MD-Mushfiqur123/lychee/api"
)

// StructuredHandler handles structured output generation requests.
func (s *Server) StructuredHandler(c *gin.Context) {
	var req api.StructuredRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}

	if req.Prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}

	opts := StructuredOpts{
		Model:      req.Model,
		Prompt:     req.Prompt,
		Schema:     req.Schema,
		MaxRetries: req.MaxRetries,
		Options:    req.Options,
		TimeoutSec: req.TimeoutSec,
	}

	if req.Stream {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")

		opts.OnEvent = func(event api.StructuredEvent) {
			data, err := json.Marshal(event)
			if err != nil {
				return
			}
			_, _ = c.Writer.Write([]byte("event: " + event.Event + "\n"))
			_, _ = c.Writer.Write(append(append([]byte("data: "), data...), '\n', '\n'))
			c.Writer.Flush()
		}
	}

	res, err := s.generateStructured(c.Request.Context(), opts)
	if err != nil {
		if req.Stream {
			errEvent := api.StructuredEvent{
				Event: "attempt_fail",
				Error: err.Error(),
			}
			data, _ := json.Marshal(errEvent)
			_, _ = c.Writer.Write([]byte("event: " + errEvent.Event + "\n"))
			_, _ = c.Writer.Write(append(append([]byte("data: "), data...), '\n', '\n'))
			c.Writer.Flush()
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !req.Stream {
		resp := api.StructuredResponse{
			Output:   res.Output,
			Valid:    res.Valid,
			Attempts: res.Attempts,
			Errors:   res.Errors,
		}
		c.JSON(http.StatusOK, resp)
	}
}
