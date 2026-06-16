package server

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/MD-Mushfiqur123/lychee/api"
)

// ListConversationsHandler returns a list of all conversations in the store.
func (s *Server) ListConversationsHandler(c *gin.Context) {
	if s.memoryStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "memory store not initialized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}

	var summaries []ConversationSummary
	var total int
	var err error

	query := c.Query("q")
	if query != "" {
		summaries, total, err = s.memoryStore.Search(query, limit, offset)
	} else {
		summaries, total, err = s.memoryStore.List(limit, offset)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if summaries == nil {
		summaries = []ConversationSummary{}
	}

	c.Header("X-Total-Count", strconv.Itoa(total))
	c.JSON(http.StatusOK, summaries)
}

// GetConversationHandler loads a specific conversation by ID.
func (s *Server) GetConversationHandler(c *gin.Context) {
	if s.memoryStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "memory store not initialized"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	conv, err := s.memoryStore.Load(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	c.JSON(http.StatusOK, conv)
}

// CreateConversationHandler creates a new conversation from scratch.
func (s *Server) CreateConversationHandler(c *gin.Context) {
	if s.memoryStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "memory store not initialized"})
		return
	}

	var req api.ConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conv := &Conversation{
		ID:        uuid.New().String(),
		Model:     req.Model,
		Messages:  req.Messages,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if conv.Messages == nil {
		conv.Messages = []api.Message{}
	}

	if err := s.memoryStore.Save(conv); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, conv)
}

// DeleteConversationHandler deletes a conversation.
func (s *Server) DeleteConversationHandler(c *gin.Context) {
	if s.memoryStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "memory store not initialized"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := s.memoryStore.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ExportConversationHandler exports a conversation by ID as JSON.
func (s *Server) ExportConversationHandler(c *gin.Context) {
	if s.memoryStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "memory store not initialized"})
		return
	}
	id := c.Param("id")
	data, err := s.memoryStore.Export(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=conversation-%s.json", id))
	c.Data(http.StatusOK, "application/json", data)
}

// ImportConversationHandler imports a conversation from the request body.
func (s *Server) ImportConversationHandler(c *gin.Context) {
	if s.memoryStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "memory store not initialized"})
		return
	}
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	conv, err := s.memoryStore.Import(bodyBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, conv)
}
