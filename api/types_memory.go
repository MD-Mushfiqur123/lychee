package api

type ConversationRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ListConversationsResponse struct {
	Conversations []interface{} `json:"conversations"` // Should map to server.ConversationSummary
	Total         int           `json:"total"`
	HasMore       bool          `json:"has_more"`
}
