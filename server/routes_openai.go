package server

import (
	"github.com/gin-gonic/gin"
	"github.com/MD-Mushfiqur123/lychee/middleware"
)

func (s *Server) registerOpenAIRoutes(r *gin.Engine) {
	// Inference (OpenAI compatibility)
	// TODO(cloud-stage-a): apply Modelfile overlay deltas for local models with cloud
	// parents on v1 request families while preserving this explicit :cloud passthrough.
	r.POST("/v1/chat/completions", s.withInferenceRequestLogging("/v1/chat/completions", cloudPassthroughMiddleware(cloudErrRemoteInferenceUnavailable), middleware.ChatMiddleware(), s.ChatHandler)...)
	r.POST("/v1/completions", s.withInferenceRequestLogging("/v1/completions", cloudPassthroughMiddleware(cloudErrRemoteInferenceUnavailable), middleware.CompletionsMiddleware(), s.GenerateHandler)...)
	r.POST("/v1/embeddings", cloudPassthroughMiddleware(cloudErrRemoteInferenceUnavailable), middleware.EmbeddingsMiddleware(), s.EmbedHandler)
	r.GET("/v1/models", middleware.ListMiddleware(), s.ListHandler)
	r.GET("/v1/models/:model", cloudModelPathPassthroughMiddleware(cloudErrRemoteModelDetailsUnavailable), middleware.RetrieveMiddleware(), s.ShowHandler)
	r.POST("/v1/responses", s.withInferenceRequestLogging("/v1/responses", cloudPassthroughMiddleware(cloudErrRemoteInferenceUnavailable), middleware.ResponsesMiddleware(), s.ChatHandler)...)
	
	// OpenAI-compatible image generation endpoints
	r.POST("/v1/images/generations", cloudPassthroughMiddleware(cloudErrRemoteInferenceUnavailable), middleware.ImageGenerationsMiddleware(), s.GenerateHandler)
	r.POST("/v1/images/edits", cloudPassthroughMiddleware(cloudErrRemoteInferenceUnavailable), middleware.ImageEditsMiddleware(), s.GenerateHandler)
	
	// OpenAI-compatible audio endpoint
	r.POST("/v1/audio/transcriptions", middleware.TranscriptionMiddleware(), s.ChatHandler)
}
