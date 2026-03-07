package chat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockAIClient returns canned responses for testing.
type mockAIClient struct {
	responses []*ChatCompletionResponse
	callIndex int
}

func (m *mockAIClient) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if m.callIndex >= len(m.responses) {
		return &ChatCompletionResponse{Content: "fallback"}, nil
	}
	resp := m.responses[m.callIndex]
	m.callIndex++
	return resp, nil
}

func TestHandleChat_SimpleResponse(t *testing.T) {
	reg := NewRegistry()
	aiProviders := NewAIProviderRegistry("openai")
	aiProviders.Register("openai", &mockAIClient{
		responses: []*ChatCompletionResponse{
			{Content: "Your MRR is $1,500.00", Usage: TokenUsage{TotalTokens: 100}},
		},
	})

	handler := NewHandler(reg, aiProviders)

	body, _ := json.Marshal(ChatRequest{
		Messages: []ChatMessage{{Role: "user", Content: "What is my MRR?"}},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.HandleChat(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Parse SSE events
	events := parseSSEEvents(t, rec.Body.String())
	if len(events) != 1 {
		t.Fatalf("expected 1 SSE event, got %d", len(events))
	}

	if events[0].Type != "response" {
		t.Errorf("expected 'response' event, got '%s'", events[0].Type)
	}

	var respData responseEvent
	raw, _ := json.Marshal(events[0].Data)
	json.Unmarshal(raw, &respData)
	if respData.Message != "Your MRR is $1,500.00" {
		t.Errorf("expected message 'Your MRR is $1,500.00', got '%s'", respData.Message)
	}
}

func TestHandleChat_ToolCallLoop(t *testing.T) {
	// Set up a module
	stub := &stubModule{
		name:  "risk",
		tools: []ToolDefinition{{Name: "get_risk_summary", Description: "test"}},
		result: ToolResult{
			ToolCallID: "tc-1",
			Content:    `{"riskSummary":{"safe":40,"churned":2}}`,
		},
	}

	reg := NewRegistry()
	reg.Register(stub)

	aiProviders := NewAIProviderRegistry("openai")
	aiProviders.Register("openai", &mockAIClient{
		responses: []*ChatCompletionResponse{
			// First call: AI requests a tool call
			{
				ToolCalls: []ToolCallRequest{
					{ID: "tc-1", Name: "risk__get_risk_summary", Arguments: `{"app_id":"abc"}`},
				},
			},
			// Second call: AI responds with text after seeing tool result
			{Content: "You have 40 safe and 2 churned subscriptions."},
		},
	})

	handler := NewHandler(reg, aiProviders)

	body, _ := json.Marshal(ChatRequest{
		Messages: []ChatMessage{{Role: "user", Content: "Show risk summary"}},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.HandleChat(rec, req)

	events := parseSSEEvents(t, rec.Body.String())

	// Should have: tool_call, tool_result, response
	if len(events) < 3 {
		t.Fatalf("expected at least 3 events, got %d: %v", len(events), events)
	}

	if events[0].Type != "tool_call" {
		t.Errorf("expected first event 'tool_call', got '%s'", events[0].Type)
	}
	if events[1].Type != "tool_result" {
		t.Errorf("expected second event 'tool_result', got '%s'", events[1].Type)
	}
	if events[len(events)-1].Type != "response" {
		t.Errorf("expected last event 'response', got '%s'", events[len(events)-1].Type)
	}
}

func TestHandleChat_EmptyMessages(t *testing.T) {
	handler := NewHandler(NewRegistry(), NewAIProviderRegistry("openai"))

	body, _ := json.Marshal(ChatRequest{Messages: []ChatMessage{}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.HandleChat(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleChat_NoAIProvider(t *testing.T) {
	handler := NewHandler(NewRegistry(), NewAIProviderRegistry("openai"))

	body, _ := json.Marshal(ChatRequest{
		Messages: []ChatMessage{{Role: "user", Content: "hello"}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.HandleChat(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

func TestHandleListModules(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&stubModule{name: "risk", desc: "Risk analysis", tools: make([]ToolDefinition, 3)})
	reg.Register(&stubModule{name: "metrics", desc: "Revenue metrics", tools: make([]ToolDefinition, 2)})

	handler := NewHandler(reg, NewAIProviderRegistry("openai"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/modules", nil)
	rec := httptest.NewRecorder()

	handler.HandleListModules(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var modules []struct {
		Name      string `json:"name"`
		ToolCount int    `json:"tool_count"`
	}
	json.NewDecoder(rec.Body).Decode(&modules)

	if len(modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(modules))
	}
	if modules[0].ToolCount != 3 {
		t.Errorf("expected 3 tools for risk, got %d", modules[0].ToolCount)
	}
}

func TestExtractState_LatestWins(t *testing.T) {
	records := []toolRecord{
		{Module: "risk", Name: "risk__get_risk_summary", Result: `{"old":"data"}`, Error: false},
		{Module: "risk", Name: "risk__list_at_risk", Result: `{"new":"data"}`, Error: false},
	}

	state := extractState(records)
	if string(state["risk"]) != `{"new":"data"}` {
		t.Errorf("expected latest-wins, got %s", string(state["risk"]))
	}
}

func TestExtractState_SkipsErrors(t *testing.T) {
	records := []toolRecord{
		{Module: "risk", Name: "risk__get_risk_summary", Result: `{"good":"data"}`, Error: false},
		{Module: "risk", Name: "risk__list_at_risk", Result: `{"error":"oops"}`, Error: true},
	}

	state := extractState(records)
	if string(state["risk"]) != `{"good":"data"}` {
		t.Errorf("expected non-error result, got %s", string(state["risk"]))
	}
}

func TestGenerateSuggestions(t *testing.T) {
	records := []toolRecord{
		{Name: "risk__get_risk_summary", Error: false},
	}

	suggestions := generateSuggestions(records)
	if len(suggestions) == 0 {
		t.Error("expected suggestions for risk_summary")
	}
}

// --- Test helpers ---

func parseSSEEvents(t *testing.T, body string) []SSEEvent {
	t.Helper()
	var events []SSEEvent
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			var event SSEEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				t.Logf("skipping unparseable SSE data: %s", data)
				continue
			}
			events = append(events, event)
		}
	}
	return events
}
