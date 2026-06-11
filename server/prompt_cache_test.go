package server

import (
	"testing"

	"github.com/lychee/lychee/api"
)

func TestPromptCache(t *testing.T) {
	t.Run("HasCacheControlChat", func(t *testing.T) {
		// Nil request
		reqEmpty := &api.ChatRequest{}
		if HasCacheControlChat(reqEmpty) {
			t.Errorf("expected false for empty request")
		}

		// Request cache control
		reqGlobal := &api.ChatRequest{
			CacheControl: &api.CacheControl{Type: "ephemeral"},
		}
		if !HasCacheControlChat(reqGlobal) {
			t.Errorf("expected true for global cache control request")
		}

		// Message level cache control
		reqMsg := &api.ChatRequest{
			Messages: []api.Message{
				{Role: "system", Content: "Hello"},
				{Role: "user", Content: "World", CacheControl: &api.CacheControl{Type: "ephemeral"}},
			},
		}
		if !HasCacheControlChat(reqMsg) {
			t.Errorf("expected true for message-level cache control")
		}
	})

	t.Run("HasCacheControlGenerate", func(t *testing.T) {
		reqEmpty := &api.GenerateRequest{}
		if HasCacheControlGenerate(reqEmpty) {
			t.Errorf("expected false for empty generate request")
		}

		reqSet := &api.GenerateRequest{
			CacheControl: &api.CacheControl{Type: "ephemeral"},
		}
		if !HasCacheControlGenerate(reqSet) {
			t.Errorf("expected true for cache control generate request")
		}
	})

	t.Run("ComputePrefixHash", func(t *testing.T) {
		msgs1 := []api.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello, model!"},
			{Role: "assistant", Content: "Hello, human!"},
			{Role: "user", Content: "What is 2+2?"}, // Final user message - should be skipped
		}

		msgs2 := []api.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello, model!"},
			{Role: "assistant", Content: "Hello, human!"},
			{Role: "user", Content: "What is 3+3?"}, // Different final user message - should yield same hash
		}

		hash1 := ComputePrefixHash(msgs1)
		hash2 := ComputePrefixHash(msgs2)

		if hash1 != hash2 {
			t.Errorf("expected identical hashes for identical prefixes, got: %s vs %s", hash1, hash2)
		}

		// Validate that helper skips final user message but not final assistant message
		msgsAssistantFinal := []api.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello, model!"},
			{Role: "assistant", Content: "Hello, human!"}, // Final message is assistant - should NOT be skipped
		}
		hash3 := ComputePrefixHash(msgsAssistantFinal)

		if hash1 == hash3 {
			t.Errorf("expected different hash since final assistant message is not skipped, got same: %s", hash1)
		}
	})
}
