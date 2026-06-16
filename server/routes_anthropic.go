package server

import (
	"github.com/gin-gonic/gin"
	"github.com/MD-Mushfiqur123/lychee/middleware"
)

func (s *Server) registerAnthropicRoutes(r *gin.Engine) {
	// Inference (Anthropic compatibility)
	r.POST("/v1/messages", s.withInferenceRequestLogging("/v1/messages", cloudPassthroughMiddleware(cloudErrRemoteInferenceUnavailable), middleware.AnthropicMessagesMiddleware(), s.ChatHandler)...)
}
