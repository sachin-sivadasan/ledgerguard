package chat

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Registry manages chat modules and routes tool calls to the correct module.
type Registry struct {
	mu      sync.RWMutex
	modules []Module
}

// NewRegistry creates an empty module registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a module to the registry.
func (r *Registry) Register(m Module) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.modules = append(r.modules, m)
}

// Modules returns all registered modules.
func (r *Registry) Modules() []Module {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Module, len(r.modules))
	copy(result, r.modules)
	return result
}

// ListAllTools returns tool definitions from all modules with fully qualified
// names (module__tool_name).
func (r *Registry) ListAllTools() []ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tools []ToolDefinition
	for _, m := range r.modules {
		prefix := m.Name() + "__"
		for _, t := range m.Tools() {
			tools = append(tools, ToolDefinition{
				Name:        prefix + t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			})
		}
	}
	return tools
}

// ListModuleTools returns tool definitions for a specific module with
// fully qualified names.
func (r *Registry) ListModuleTools(moduleName string) []ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, m := range r.modules {
		if m.Name() == moduleName {
			prefix := m.Name() + "__"
			tools := make([]ToolDefinition, 0, len(m.Tools()))
			for _, t := range m.Tools() {
				tools = append(tools, ToolDefinition{
					Name:        prefix + t.Name,
					Description: t.Description,
					Parameters:  t.Parameters,
				})
			}
			return tools
		}
	}
	return nil
}

// RouteToolCall routes a tool call to the correct module based on the
// double-underscore prefix convention (e.g. "risk__get_risk_summary" → risk module).
func (r *Registry) RouteToolCall(ctx context.Context, call ToolCall) ToolResult {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, m := range r.modules {
		prefix := m.Name() + "__"
		if strings.HasPrefix(call.Name, prefix) {
			localCall := ToolCall{
				ID:        call.ID,
				Name:      strings.TrimPrefix(call.Name, prefix),
				Arguments: call.Arguments,
			}
			return m.ExecuteTool(ctx, localCall)
		}
	}

	return ToolResult{
		ToolCallID: call.ID,
		Content:    fmt.Sprintf(`{"error": "unknown tool: %s"}`, call.Name),
		IsError:    true,
	}
}

// BuildSystemPrompt assembles the system prompt from all module fragments.
// If scopedModule is non-empty, only that module's fragment is included.
func (r *Registry) BuildSystemPrompt(scopedModule string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var fragments []string
	for _, m := range r.modules {
		if scopedModule != "" && m.Name() != scopedModule {
			continue
		}
		if frag := m.PromptFragment(); frag != "" {
			fragments = append(fragments, frag)
		}
	}

	if len(fragments) == 0 {
		return ""
	}
	return strings.Join(fragments, "\n\n")
}
