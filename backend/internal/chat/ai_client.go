package chat

import "context"

// AIClient abstracts the LLM provider — swap OpenAI for Claude without touching the chat handler.
type AIClient interface {
	// ChatCompletion sends messages + tools, returns a response that may contain
	// text content or tool call requests (or both).
	ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error)
}

// ChatCompletionRequest is the provider-agnostic request for a chat completion.
type ChatCompletionRequest struct {
	Model        string           `json:"model"`
	SystemPrompt string           `json:"system_prompt"`
	Messages     []ChatMessage    `json:"messages"`
	Tools        []ToolDefinition `json:"tools"`
}

// ChatCompletionResponse is the provider-agnostic response from a chat completion.
type ChatCompletionResponse struct {
	Content   string            `json:"content"`    // text response (empty if tool calls only)
	ToolCalls []ToolCallRequest `json:"tool_calls"` // tool calls to execute (empty if text response)
	Usage     TokenUsage        `json:"usage"`
}

// ToolCallRequest represents a tool call requested by the AI model.
type ToolCallRequest struct {
	ID        string `json:"id"`        // provider-assigned ID for result pairing
	Name      string `json:"name"`      // fully qualified tool name
	Arguments string `json:"arguments"` // JSON string of arguments
}

// TokenUsage tracks token consumption for monitoring/billing.
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
