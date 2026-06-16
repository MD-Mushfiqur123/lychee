// prompt_cache.go provides helper utilities for prompt prefix hashing.
// Note: Actual KV cache management and sharing is handled by the upstream
// Ollama scheduler (server/sched.go). These functions compute SHA256
// hashes of message prefixes to enable cache key matching.

package server


import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/MD-Mushfiqur123/lychee/api"
)

// HasCacheControlChat checks if the request has cache control enabled.
func HasCacheControlChat(req *api.ChatRequest) bool {
	if req.CacheControl != nil {
		return true
	}
	for _, msg := range req.Messages {
		if msg.CacheControl != nil {
			return true
		}
	}
	return false
}

// HasCacheControlGenerate checks if the generate request has cache control enabled.
func HasCacheControlGenerate(req *api.GenerateRequest) bool {
	return req.CacheControl != nil
}

// ComputePrefixHash calculates a hash of the prefix content.
func ComputePrefixHash(messages []api.Message) string {
	var sb strings.Builder
	for i, msg := range messages {
		if i == len(messages)-1 && msg.Role == "user" {
			// Skip the final user message to query cache of previous prefix
			break
		}
		sb.WriteString(fmt.Sprintf("%s:%s\n", msg.Role, msg.Content))
	}
	hash := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}
