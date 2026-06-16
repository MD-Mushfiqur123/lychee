package server

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/MD-Mushfiqur123/lychee/api"
)

func TestMemoryStore(t *testing.T) {
	tempDir := t.TempDir()
	store := NewMemoryStore(tempDir)

	convID := "test-conv-123"
	conv := &Conversation{
		ID:    convID,
		Model: "gemma3",
		Messages: []api.Message{
			{Role: "user", Content: "Hello world!"},
			{Role: "assistant", Content: "Hi there! How can I help you today?"},
		},
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now(),
	}

	t.Run("save and load roundtrip", func(t *testing.T) {
		err := store.Save(conv)
		if err != nil {
			t.Fatalf("failed to save conversation: %v", err)
		}

		loaded, err := store.Load(convID)
		if err != nil {
			t.Fatalf("failed to load conversation: %v", err)
		}

		if loaded.ID != conv.ID {
			t.Errorf("expected ID %q, got %q", conv.ID, loaded.ID)
		}
		if loaded.Model != conv.Model {
			t.Errorf("expected Model %q, got %q", conv.Model, loaded.Model)
		}
		if len(loaded.Messages) != len(conv.Messages) {
			t.Fatalf("expected %d messages, got %d", len(conv.Messages), len(loaded.Messages))
		}
		if loaded.Messages[0].Content != conv.Messages[0].Content {
			t.Errorf("expected first message content %q, got %q", conv.Messages[0].Content, loaded.Messages[0].Content)
		}
	})

	t.Run("list conversations", func(t *testing.T) {
		list, total, err := store.List(50, 0)
		if err != nil {
			t.Fatalf("failed to list conversations: %v", err)
		}

		if len(list) != 1 {
			t.Fatalf("expected 1 conversation summary, got %d", len(list))
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}

		summary := list[0]
		if summary.ID != convID {
			t.Errorf("expected ID %q, got %q", convID, summary.ID)
		}
		if summary.Title != "Hello world!" {
			t.Errorf("expected Title %q, got %q", "Hello world!", summary.Title)
		}
		if summary.Messages != 2 {
			t.Errorf("expected Messages count 2, got %d", summary.Messages)
		}
	})

	t.Run("append message", func(t *testing.T) {
		newMsg := api.Message{Role: "user", Content: "One more question."}
		err := store.AppendMessage(convID, newMsg)
		if err != nil {
			t.Fatalf("failed to append message: %v", err)
		}

		loaded, err := store.Load(convID)
		if err != nil {
			t.Fatalf("failed to load: %v", err)
		}

		if len(loaded.Messages) != 3 {
			t.Fatalf("expected 3 messages, got %d", len(loaded.Messages))
		}
		if loaded.Messages[2].Content != "One more question." {
			t.Errorf("expected last message content 'One more question.', got %q", loaded.Messages[2].Content)
		}
	})

	t.Run("concurrent access safety", func(t *testing.T) {
		var wg sync.WaitGroup
		concurrency := 10

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(num int) {
				defer wg.Done()
				_ = store.AppendMessage(convID, api.Message{Role: "user", Content: "Concurrent message"})
				_, _, _ = store.List(50, 0)
			}(i)
		}

		wg.Wait()

		loaded, err := store.Load(convID)
		if err != nil {
			t.Fatalf("failed to load: %v", err)
		}
		expectedLen := 3 + concurrency
		if len(loaded.Messages) != expectedLen {
			t.Errorf("expected %d messages, got %d", expectedLen, len(loaded.Messages))
		}
	})

	t.Run("delete conversation", func(t *testing.T) {
		err := store.Delete(convID)
		if err != nil {
			t.Fatalf("failed to delete conversation: %v", err)
		}

		_, err = store.Load(convID)
		if err == nil {
			t.Error("expected error when loading deleted conversation, got nil")
		}

		list, _, err := store.List(50, 0)
		if err != nil {
			t.Fatalf("failed to list: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("expected 0 summaries, got %d", len(list))
		}
	})

	t.Run("load non-existent ID", func(t *testing.T) {
		_, err := store.Load("does-not-exist")
		if err == nil {
			t.Error("expected error loading non-existent conversation, got nil")
		}
	})

	t.Run("pagination", func(t *testing.T) {
		store2 := NewMemoryStore(t.TempDir())
		for i := 0; i < 5; i++ {
			_ = store2.Save(&Conversation{
				ID:    fmt.Sprintf("conv-%d", i),
				Model: "test",
				Messages: []api.Message{
					{Role: "user", Content: fmt.Sprintf("msg %d", i)},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now().Add(time.Duration(i) * time.Minute),
			})
		}

		list, total, err := store2.List(2, 0)
		if err != nil {
			t.Fatal(err)
		}
		if total != 5 {
			t.Errorf("expected total 5, got %d", total)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 results, got %d", len(list))
		}

		list2, total2, _ := store2.List(2, 3)
		if total2 != 5 {
			t.Errorf("expected total 5, got %d", total2)
		}
		if len(list2) != 2 {
			t.Errorf("expected 2 results at offset 3, got %d", len(list2))
		}

		list3, _, _ := store2.List(10, 10)
		if len(list3) != 0 {
			t.Errorf("expected 0 results at offset 10, got %d", len(list3))
		}
	})

	t.Run("title auto-update", func(t *testing.T) {
		store3 := NewMemoryStore(t.TempDir())
		conv3 := &Conversation{
			ID:    "title-conv",
			Model: "test",
			Messages: []api.Message{
				{Role: "user", Content: "First message"},
				{Role: "assistant", Content: "Hello!"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = store3.Save(conv3)

		list, _, _ := store3.List(10, 0)
		if list[0].Title != "First message" {
			t.Errorf("expected Title to be 'First message', got %q", list[0].Title)
		}

		conv3.Messages = append(conv3.Messages,
			api.Message{Role: "user", Content: "Third message"},
		)
		_ = store3.Save(conv3)
		list, _, _ = store3.List(10, 0)
		if list[0].Title != "First message" {
			t.Errorf("expected Title to still be 'First message', got %q", list[0].Title)
		}

		conv3.Messages = append(conv3.Messages,
			api.Message{Role: "assistant", Content: "Response"},
			api.Message{Role: "user", Content: "Fifth message"},
		)
		_ = store3.Save(conv3)
		list, _, _ = store3.List(10, 0)
		if list[0].Title != "Fifth message" {
			t.Errorf("expected Title to auto-update to latest user message, got %q", list[0].Title)
		}
	})

	t.Run("search conversations", func(t *testing.T) {
		store4 := NewMemoryStore(t.TempDir())
		_ = store4.Save(&Conversation{
			ID:    "quantum-conv",
			Model: "test",
			Messages: []api.Message{
				{Role: "user", Content: "Tell me about quantum physics"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		_ = store4.Save(&Conversation{
			ID:    "cooking-conv",
			Model: "test",
			Messages: []api.Message{
				{Role: "user", Content: "How to cook pasta"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})

		results, total, err := store4.Search("quantum", 50, 0)
		if err != nil {
			t.Fatal(err)
		}
		if total != 1 {
			t.Errorf("expected 1 match, got %d", total)
		}
		if results[0].ID != "quantum-conv" {
			t.Errorf("expected quantum-conv, got %s", results[0].ID)
		}
	})
}
