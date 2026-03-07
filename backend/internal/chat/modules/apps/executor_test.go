package apps

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
)

// mockGQLHandler returns canned GraphQL responses.
type mockGQLHandler struct {
	response string
}

func (h *mockGQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(h.response))
}

func TestListApps(t *testing.T) {
	handler := &mockGQLHandler{
		response: `{"data":{"apps":[{"id":"1","name":"Test App","partnerAppId":"gid://partners/App/123","trackingEnabled":true,"installCount":42}]}}`,
	}
	gqlExec := chat.NewGraphQLExecutor(handler)
	exec := &executor{gql: gqlExec}

	result := exec.execute(context.Background(), chat.ToolCall{
		ID:        "tc-1",
		Name:      "list_apps",
		Arguments: "{}",
	})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", result.Content)
	}

	var data struct {
		Apps []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"apps"`
	}
	if err := json.Unmarshal([]byte(result.Content), &data); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}
	if len(data.Apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(data.Apps))
	}
	if data.Apps[0].Name != "Test App" {
		t.Errorf("expected 'Test App', got %q", data.Apps[0].Name)
	}
}

func TestGetAppDetail(t *testing.T) {
	handler := &mockGQLHandler{
		response: `{"data":{"app":{"id":"1","name":"Test App","partnerAppId":"gid://partners/App/123","trackingEnabled":true,"installCount":42}}}`,
	}
	gqlExec := chat.NewGraphQLExecutor(handler)
	exec := &executor{gql: gqlExec}

	result := exec.execute(context.Background(), chat.ToolCall{
		ID:        "tc-2",
		Name:      "get_app_detail",
		Arguments: `{"app_id":"1"}`,
	})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", result.Content)
	}

	if !strings.Contains(result.Content, "Test App") {
		t.Errorf("expected result to contain 'Test App', got %s", result.Content)
	}
}

func TestUnknownTool(t *testing.T) {
	exec := &executor{gql: nil}
	result := exec.execute(context.Background(), chat.ToolCall{
		ID:   "tc-3",
		Name: "unknown",
	})

	if !result.IsError {
		t.Error("expected error for unknown tool")
	}
}
