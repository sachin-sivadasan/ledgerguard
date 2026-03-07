package external

import (
	"encoding/json"
	"testing"

	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
	openai "github.com/sashabaranov/go-openai"
)

func TestConvertMessages_SystemPrompt(t *testing.T) {
	msgs := convertMessages("You are helpful.", []chat.ChatMessage{
		{Role: "user", Content: "Hello"},
	})

	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Role != openai.ChatMessageRoleSystem {
		t.Errorf("expected system role, got %s", msgs[0].Role)
	}
	if msgs[0].Content != "You are helpful." {
		t.Errorf("expected system prompt content, got %s", msgs[0].Content)
	}
	if msgs[1].Role != "user" {
		t.Errorf("expected user role, got %s", msgs[1].Role)
	}
}

func TestConvertMessages_ToolCalls(t *testing.T) {
	msgs := convertMessages("", []chat.ChatMessage{
		{
			Role:    "assistant",
			Content: "",
			ToolCalls: []chat.ToolCall{
				{ID: "tc-1", Name: "risk__get_summary", Arguments: `{"app_id":"abc"}`},
			},
		},
	})

	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if len(msgs[0].ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(msgs[0].ToolCalls))
	}
	if msgs[0].ToolCalls[0].Function.Name != "risk__get_summary" {
		t.Errorf("expected tool name risk__get_summary, got %s", msgs[0].ToolCalls[0].Function.Name)
	}
}

func TestConvertMessages_ToolResult(t *testing.T) {
	msgs := convertMessages("", []chat.ChatMessage{
		{
			Role: "tool",
			ToolResult: &chat.ToolResult{
				ToolCallID: "tc-1",
				Content:    `{"data": "value"}`,
			},
		},
	})

	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Role != openai.ChatMessageRoleTool {
		t.Errorf("expected tool role, got %s", msgs[0].Role)
	}
	if msgs[0].ToolCallID != "tc-1" {
		t.Errorf("expected tool call ID tc-1, got %s", msgs[0].ToolCallID)
	}
}

func TestConvertTools(t *testing.T) {
	tools := convertTools([]chat.ToolDefinition{
		{
			Name:        "risk__get_summary",
			Description: "Get risk summary",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"app_id":{"type":"string"}},"required":["app_id"]}`),
		},
	})

	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}
	if tools[0].Type != openai.ToolTypeFunction {
		t.Errorf("expected function type, got %s", tools[0].Type)
	}
	if tools[0].Function.Name != "risk__get_summary" {
		t.Errorf("expected name risk__get_summary, got %s", tools[0].Function.Name)
	}
}

func TestConvertTools_NilParameters(t *testing.T) {
	tools := convertTools([]chat.ToolDefinition{
		{Name: "test", Description: "test tool"},
	})

	if tools[0].Function.Parameters == nil {
		t.Error("expected default parameters for nil input")
	}
}

func TestNewOpenAIClient_DefaultModel(t *testing.T) {
	client := NewOpenAIClient("test-key", "")
	if client.model != "gpt-4o" {
		t.Errorf("expected default model gpt-4o, got %s", client.model)
	}
}

func TestNewOpenAIClient_CustomModel(t *testing.T) {
	client := NewOpenAIClient("test-key", "gpt-4o-mini")
	if client.model != "gpt-4o-mini" {
		t.Errorf("expected model gpt-4o-mini, got %s", client.model)
	}
}
