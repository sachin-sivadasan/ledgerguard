package chat

import "context"

// Module is the plugin interface for chat feature modules.
// Each module is self-contained: it defines its tools, handles execution,
// and provides prompt fragments for the AI system prompt.
//
// Zero changes to core when adding a new module — register it and it
// auto-appears in tool lists, system prompts, and @module autocomplete.
type Module interface {
	// Name returns the unique module identifier (e.g. "risk", "subscriptions").
	// Used as prefix in tool names: "risk__get_risk_summary".
	Name() string

	// Description returns a human-readable description shown in @module autocomplete.
	Description() string

	// PromptFragment returns AI instructions appended to the system prompt
	// when this module is active. Should describe what the tools do and
	// how to interpret results.
	PromptFragment() string

	// Tools returns the tool definitions for this module.
	// Tool names must NOT include the module prefix — the registry adds it.
	Tools() []ToolDefinition

	// ExecuteTool runs a tool by name with the given arguments.
	// The call.Name is the local name (without module prefix).
	ExecuteTool(ctx context.Context, call ToolCall) ToolResult
}
