package chat

import (
	"context"
	"testing"
)

func TestRegistry_Register(t *testing.T) {
	reg := NewRegistry()
	reg.Register(newStubModule("risk", "get_risk_summary"))
	reg.Register(newStubModule("metrics", "get_latest"))

	modules := reg.Modules()
	if len(modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(modules))
	}
	if modules[0].Name() != "risk" {
		t.Errorf("expected first module 'risk', got '%s'", modules[0].Name())
	}
}

func TestRegistry_ListAllTools_PrefixesNames(t *testing.T) {
	reg := NewRegistry()
	reg.Register(newStubModule("risk", "get_risk_summary", "list_at_risk"))
	reg.Register(newStubModule("metrics", "get_latest"))

	tools := reg.ListAllTools()
	if len(tools) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(tools))
	}

	expected := map[string]bool{
		"risk__get_risk_summary": false,
		"risk__list_at_risk":     false,
		"metrics__get_latest":    false,
	}
	for _, tool := range tools {
		if _, ok := expected[tool.Name]; !ok {
			t.Errorf("unexpected tool name: %s", tool.Name)
		}
		expected[tool.Name] = true
	}
	for name, found := range expected {
		if !found {
			t.Errorf("missing expected tool: %s", name)
		}
	}
}

func TestRegistry_ListModuleTools(t *testing.T) {
	reg := NewRegistry()
	reg.Register(newStubModule("risk", "get_risk_summary", "list_at_risk"))
	reg.Register(newStubModule("metrics", "get_latest"))

	tools := reg.ListModuleTools("risk")
	if len(tools) != 2 {
		t.Fatalf("expected 2 risk tools, got %d", len(tools))
	}
	if tools[0].Name != "risk__get_risk_summary" {
		t.Errorf("expected risk__get_risk_summary, got %s", tools[0].Name)
	}

	// Unknown module returns nil
	unknown := reg.ListModuleTools("nonexistent")
	if unknown != nil {
		t.Errorf("expected nil for unknown module, got %v", unknown)
	}
}

func TestRegistry_RouteToolCall(t *testing.T) {
	riskMod := newStubModule("risk", "get_risk_summary")
	metricsMod := newStubModule("metrics", "get_latest")

	reg := NewRegistry()
	reg.Register(riskMod)
	reg.Register(metricsMod)

	call := ToolCall{
		ID:        "call-1",
		Name:      "risk__get_risk_summary",
		Arguments: `{"app_id": "abc"}`,
	}

	result := reg.RouteToolCall(context.Background(), call)
	if result.IsError {
		t.Fatalf("expected no error, got: %s", result.Content)
	}
	if result.ToolCallID != "call-1" {
		t.Errorf("expected toolCallID 'call-1', got '%s'", result.ToolCallID)
	}

	// Verify the local call stripped the prefix
	if riskMod.lastCall == nil {
		t.Fatal("expected risk module to receive the call")
	}
	if riskMod.lastCall.Name != "get_risk_summary" {
		t.Errorf("expected local name 'get_risk_summary', got '%s'", riskMod.lastCall.Name)
	}
	if riskMod.lastCall.Arguments != `{"app_id": "abc"}` {
		t.Errorf("expected arguments preserved, got '%s'", riskMod.lastCall.Arguments)
	}
}

func TestRegistry_RouteToolCall_UnknownTool(t *testing.T) {
	reg := NewRegistry()
	reg.Register(newStubModule("risk", "get_risk_summary"))

	call := ToolCall{
		ID:   "call-2",
		Name: "unknown__tool",
	}

	result := reg.RouteToolCall(context.Background(), call)
	if !result.IsError {
		t.Error("expected error for unknown tool")
	}
}

func TestRegistry_BuildSystemPrompt_AllModules(t *testing.T) {
	reg := NewRegistry()
	reg.Register(newStubModule("risk", "get_risk_summary"))
	reg.Register(newStubModule("metrics", "get_latest"))

	prompt := reg.BuildSystemPrompt("")
	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}
	if !contains(prompt, "risk") || !contains(prompt, "metrics") {
		t.Error("expected prompt to contain both module fragments")
	}
}

func TestRegistry_BuildSystemPrompt_Scoped(t *testing.T) {
	reg := NewRegistry()
	reg.Register(newStubModule("risk", "get_risk_summary"))
	reg.Register(newStubModule("metrics", "get_latest"))

	prompt := reg.BuildSystemPrompt("risk")
	if !contains(prompt, "risk") {
		t.Error("expected prompt to contain risk fragment")
	}
	if contains(prompt, "metrics") {
		t.Error("expected prompt to NOT contain metrics fragment when scoped to risk")
	}
}
