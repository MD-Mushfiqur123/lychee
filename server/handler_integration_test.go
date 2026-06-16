package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/MD-Mushfiqur123/lychee/llm"
)

func TestHTTPIntegration(t *testing.T) {
	t.Setenv("LYCHEE_MODELS", t.TempDir())
	gin.SetMode(gin.TestMode)

	mock := mockRunner{
		CompletionFn: func(ctx context.Context, req llm.CompletionRequest, fn func(llm.CompletionResponse)) error {
			fn(llm.CompletionResponse{
				Content: `{"name":"Bob","age":42}`,
				Done:    true,
			})
			return nil
		},
	}

	s := newServerWithMockRunner(t, &mock)
	createMinimalGGUFModel(t, s, "test-model", nil, "", nil)

	r := gin.New()
	r.POST("/api/conversations", s.CreateConversationHandler)
	r.GET("/api/conversations", s.ListConversationsHandler)
	r.GET("/api/conversations/:id", s.GetConversationHandler)
	r.DELETE("/api/conversations/:id", s.DeleteConversationHandler)
	r.GET("/api/conversations/:id/export", s.ExportConversationHandler)
	r.POST("/api/conversations/import", s.ImportConversationHandler)
	r.POST("/api/routes", s.CreateRouteHandler)
	r.GET("/api/routes", s.ListRoutesHandler)
	r.DELETE("/api/routes/:name", s.DeleteRouteHandler)
	r.GET("/api/routes/:name/status", s.RouteStatusHandler)
	r.POST("/api/structured", s.StructuredHandler)
	r.POST("/api/aliases", s.SetAliasHandler)
	r.GET("/api/aliases", s.ListAliasesHandler)
	r.DELETE("/api/aliases/:name", s.DeleteAliasHandler)
	r.POST("/api/chat", s.ChatHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	client := &http.Client{}

	t.Run("conversations CRUD", func(t *testing.T) {
		// 1. Create conversation
		reqBody := api.ConversationRequest{
			Model: "test-model",
			Messages: []api.Message{
				{Role: "user", Content: "Hello!"},
			},
		}
		jsonBytes, _ := json.Marshal(reqBody)
		resp, err := client.Post(ts.URL+"/api/conversations", "application/json", bytes.NewReader(jsonBytes))
		if err != nil {
			t.Fatalf("failed to POST: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
		}

		var conv Conversation
		_ = json.NewDecoder(resp.Body).Decode(&conv)
		resp.Body.Close()

		if conv.ID == "" {
			t.Fatal("expected non-empty conversation ID")
		}

		// 2. List conversations
		resp, err = client.Get(ts.URL + "/api/conversations?limit=10")
		if err != nil {
			t.Fatalf("failed to GET list: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}
		var list []ConversationSummary
		_ = json.NewDecoder(resp.Body).Decode(&list)
		resp.Body.Close()

		if len(list) != 1 {
			t.Fatalf("expected 1 conversation, got %d", len(list))
		}

		// 3. Get single conversation
		resp, err = client.Get(ts.URL + "/api/conversations/" + conv.ID)
		if err != nil {
			t.Fatalf("failed to GET single: %v", err)
		}
		var loaded Conversation
		_ = json.NewDecoder(resp.Body).Decode(&loaded)
		resp.Body.Close()

		if loaded.ID != conv.ID {
			t.Errorf("expected loaded ID %s, got %s", conv.ID, loaded.ID)
		}

		// 4. Delete conversation
		reqDel, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/conversations/"+conv.ID, nil)
		resp, err = client.Do(reqDel)
		if err != nil {
			t.Fatalf("failed to DELETE: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK on delete, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	t.Run("routes CRUD and status", func(t *testing.T) {
		// 1. Create route
		routeReq := api.RouteRequest{
			Name: "fast",
			Endpoints: []api.ModelEndpoint{
				{Host: "http://localhost:11434", Model: "test-model"},
			},
			Strategy: "round_robin",
		}
		jsonBytes, _ := json.Marshal(routeReq)
		resp, err := client.Post(ts.URL+"/api/routes", "application/json", bytes.NewReader(jsonBytes))
		if err != nil {
			t.Fatalf("failed to create route: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}
		resp.Body.Close()

		// 2. List routes
		resp, err = client.Get(ts.URL + "/api/routes")
		if err != nil {
			t.Fatalf("failed to get routes: %v", err)
		}
		var routesList []api.ModelRoute
		_ = json.NewDecoder(resp.Body).Decode(&routesList)
		resp.Body.Close()

		if len(routesList) != 1 {
			t.Fatalf("expected 1 route, got %d", len(routesList))
		}

		// 3. Get route status
		resp, err = client.Get(ts.URL + "/api/routes/fast/status")
		if err != nil {
			t.Fatalf("failed to get route status: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK on status, got %d", resp.StatusCode)
		}
		var status api.RouteStatusResponse
		_ = json.NewDecoder(resp.Body).Decode(&status)
		resp.Body.Close()

		if status.Name != "fast" {
			t.Errorf("expected route status name 'fast', got %s", status.Name)
		}

		// 4. Delete route
		reqDel, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/routes/fast", nil)
		resp, _ = client.Do(reqDel)
		resp.Body.Close()
	})

	t.Run("structured output", func(t *testing.T) {
		reqBody := api.StructuredRequest{
			Model:  "test-model",
			Prompt: "make profile",
			Schema: json.RawMessage(`{"type":"object"}`),
		}
		jsonBytes, _ := json.Marshal(reqBody)
		resp, err := client.Post(ts.URL+"/api/structured", "application/json", bytes.NewReader(jsonBytes))
		if err != nil {
			t.Fatalf("failed structured call: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}
		var strResp api.StructuredResponse
		_ = json.NewDecoder(resp.Body).Decode(&strResp)
		resp.Body.Close()

		if !strResp.Valid {
			t.Error("expected valid structured response")
		}
	})

	t.Run("conversations export and import", func(t *testing.T) {
		// 1. Create a conversation
		reqBody := api.ConversationRequest{
			Model: "test-model",
			Messages: []api.Message{
				{Role: "user", Content: "Hello context export import!"},
			},
		}
		jsonBytes, _ := json.Marshal(reqBody)
		resp, err := client.Post(ts.URL+"/api/conversations", "application/json", bytes.NewReader(jsonBytes))
		if err != nil {
			t.Fatalf("failed to create conversation: %v", err)
		}
		var conv Conversation
		_ = json.NewDecoder(resp.Body).Decode(&conv)
		resp.Body.Close()

		// 2. Export conversation
		resp, err = client.Get(ts.URL + "/api/conversations/" + conv.ID + "/export")
		if err != nil {
			t.Fatalf("failed to export: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK on export, got %d", resp.StatusCode)
		}
		var exportData map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&exportData)
		resp.Body.Close()

		if exportData["id"] != conv.ID {
			t.Errorf("expected exported id %s, got %v", conv.ID, exportData["id"])
		}

		// 3. Delete conversation
		reqDel, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/conversations/"+conv.ID, nil)
		resp, _ = client.Do(reqDel)
		resp.Body.Close()

		// 4. Import conversation
		exportBytes, _ := json.Marshal(exportData)
		resp, err = client.Post(ts.URL+"/api/conversations/import", "application/json", bytes.NewReader(exportBytes))
		if err != nil {
			t.Fatalf("failed to import: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK on import, got %d", resp.StatusCode)
		}
		var imported Conversation
		_ = json.NewDecoder(resp.Body).Decode(&imported)
		resp.Body.Close()

		if imported.ID != conv.ID {
			t.Errorf("expected imported ID %s, got %s", conv.ID, imported.ID)
		}

		// Verify it is back in the list
		resp, err = client.Get(ts.URL + "/api/conversations?limit=10")
		var list []ConversationSummary
		_ = json.NewDecoder(resp.Body).Decode(&list)
		resp.Body.Close()

		found := false
		for _, s := range list {
			if s.ID == conv.ID {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected imported conversation to be listed")
		}
	})

	t.Run("model aliases CRUD and resolution", func(t *testing.T) {
		// 1. Set model alias
		aliasReq := AliasRequest{
			Name:   "my-gemma-alias",
			Target: "test-model",
		}
		jsonBytes, _ := json.Marshal(aliasReq)
		resp, err := client.Post(ts.URL+"/api/aliases", "application/json", bytes.NewReader(jsonBytes))
		if err != nil {
			t.Fatalf("failed to set alias: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}
		resp.Body.Close()

		// 2. List aliases
		resp, err = client.Get(ts.URL + "/api/aliases")
		if err != nil {
			t.Fatalf("failed to get aliases: %v", err)
		}
		var aliases map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&aliases)
		resp.Body.Close()

		if target, ok := aliases["my-gemma-alias"]; !ok || target != "test-model" {
			t.Fatalf("expected alias 'my-gemma-alias' to target 'test-model', got %s", target)
		}

		// 3. Request with alias name (verify resolution)
		chatReq := api.ChatRequest{
			Model: "my-gemma-alias",
			Messages: []api.Message{
				{Role: "user", Content: "Hello!"},
			},
			DebugRenderOnly: true,
		}
		jsonBytes, _ = json.Marshal(chatReq)
		resp, err = client.Post(ts.URL+"/api/chat", "application/json", bytes.NewReader(jsonBytes))
		if err != nil {
			t.Fatalf("failed chat request: %v", err)
		}
		var chatResp api.ChatResponse
		_ = json.NewDecoder(resp.Body).Decode(&chatResp)
		resp.Body.Close()

		// The model inside the response should be resolved to the target "test-model"
		if chatResp.Model != "test-model" {
			t.Errorf("expected model resolved to 'test-model', got '%s'", chatResp.Model)
		}

		// 4. Delete alias
		reqDel, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/aliases/my-gemma-alias", nil)
		resp, err = client.Do(reqDel)
		if err != nil {
			t.Fatalf("failed delete alias: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK on delete alias, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	})
}
