package chat

import (
	"context"
	"encoding/json"
)

// stubModule is a minimal Module implementation for testing.
type stubModule struct {
	name     string
	desc     string
	prompt   string
	tools    []ToolDefinition
	lastCall *ToolCall
	result   ToolResult
}

func (m *stubModule) Name() string           { return m.name }
func (m *stubModule) Description() string    { return m.desc }
func (m *stubModule) PromptFragment() string { return m.prompt }
func (m *stubModule) Tools() []ToolDefinition { return m.tools }
func (m *stubModule) ExecuteTool(ctx context.Context, call ToolCall) ToolResult {
	m.lastCall = &call
	if m.result.Content != "" {
		result := m.result
		result.ToolCallID = call.ID
		return result
	}
	return ToolResult{
		ToolCallID: call.ID,
		Content:    `{"ok": true}`,
	}
}

func newStubModule(name string, toolNames ...string) *stubModule {
	tools := make([]ToolDefinition, len(toolNames))
	for i, tn := range toolNames {
		tools[i] = ToolDefinition{
			Name:        tn,
			Description: "Test tool " + tn,
			Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		}
	}
	return &stubModule{
		name:   name,
		desc:   "Test " + name + " module",
		prompt: "You can query " + name + " data.",
		tools:  tools,
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
