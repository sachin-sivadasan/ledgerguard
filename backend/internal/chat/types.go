package chat

import "encoding/json"

// ToolDefinition describes a tool that an AI model can call.
// Provider-agnostic: both OpenAI function calling and Claude tool use
// accept equivalent schemas.
type ToolDefinition struct {
	Name        string          `json:"name"`        // e.g. "risk__get_risk_summary"
	Description string          `json:"description"` // shown to the AI model
	Parameters  json.RawMessage `json:"parameters"`  // JSON Schema for arguments
}

// ToolCall represents a tool invocation requested by the AI model.
type ToolCall struct {
	ID        string `json:"id"`        // provider-assigned ID for result pairing
	Name      string `json:"name"`      // fully qualified: "module__tool_name"
	Arguments string `json:"arguments"` // JSON string of arguments
}

// ToolResult is the response from executing a tool call.
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"` // matches ToolCall.ID
	Content    string `json:"content"`      // JSON result or error message
	IsError    bool   `json:"is_error"`
}

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	Role       string      `json:"role"`                  // "system", "user", "assistant", "tool"
	Content    string      `json:"content"`               // text content
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`  // assistant requesting tool calls
	ToolResult *ToolResult `json:"tool_result,omitempty"` // tool response
}
