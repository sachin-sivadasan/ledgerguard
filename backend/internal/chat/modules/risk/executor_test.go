package risk

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
)

// mockGraphQLHandler returns canned responses based on the query.
type mockGraphQLHandler struct {
	response string
	err      bool
}

func (h *mockGraphQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.err {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(h.response))
}

func newTestExecutor(response string) *executor {
	handler := &mockGraphQLHandler{response: response}
	return &executor{gql: chat.NewGraphQLExecutor(handler)}
}

func TestGetRiskSummary(t *testing.T) {
	response := `{"data":{"riskSummary":{"totalSubscriptions":50,"safe":40,"oneCycleMissed":5,"twoCyclesMissed":3,"churned":2,"revenueAtRiskCents":15000,"atRiskSubscriptions":[]}}}`
	exec := newTestExecutor(response)

	result := exec.execute(context.Background(), chat.ToolCall{
		ID:        "call-1",
		Name:      "get_risk_summary",
		Arguments: `{"app_id": "00000000-0000-0000-0000-000000000001"}`,
	})

	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(result.Content), &data); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	summary, ok := data["riskSummary"].(map[string]any)
	if !ok {
		t.Fatal("expected riskSummary in response")
	}
	if summary["totalSubscriptions"].(float64) != 50 {
		t.Errorf("expected 50 total, got %v", summary["totalSubscriptions"])
	}
}

func TestListAtRisk(t *testing.T) {
	response := `{"data":{"riskSummary":{"atRiskSubscriptions":[{"domain":"store1.myshopify.com","shopName":"Store 1","riskState":"ONE_CYCLE_MISSED","daysPastDue":35,"mrrCents":2999}]}}}`
	exec := newTestExecutor(response)

	result := exec.execute(context.Background(), chat.ToolCall{
		ID:        "call-2",
		Name:      "list_at_risk",
		Arguments: `{"app_id": "00000000-0000-0000-0000-000000000001"}`,
	})

	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(result.Content), &data); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}
}

func TestListAtRisk_FilterByState(t *testing.T) {
	response := `{"data":{"riskSummary":{"atRiskSubscriptions":[{"domain":"a.myshopify.com","riskState":"ONE_CYCLE_MISSED","mrrCents":2999},{"domain":"b.myshopify.com","riskState":"TWO_CYCLES_MISSED","mrrCents":4999}]}}}`
	exec := newTestExecutor(response)

	result := exec.execute(context.Background(), chat.ToolCall{
		ID:        "call-3",
		Name:      "list_at_risk",
		Arguments: `{"app_id": "00000000-0000-0000-0000-000000000001", "risk_state": "TWO_CYCLES_MISSED"}`,
	})

	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}

	var data struct {
		RiskSummary struct {
			AtRiskSubscriptions []struct {
				RiskState string `json:"riskState"`
			} `json:"atRiskSubscriptions"`
		} `json:"riskSummary"`
	}
	if err := json.Unmarshal([]byte(result.Content), &data); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if len(data.RiskSummary.AtRiskSubscriptions) != 1 {
		t.Fatalf("expected 1 filtered subscription, got %d", len(data.RiskSummary.AtRiskSubscriptions))
	}
	if data.RiskSummary.AtRiskSubscriptions[0].RiskState != "TWO_CYCLES_MISSED" {
		t.Errorf("expected TWO_CYCLES_MISSED, got %s", data.RiskSummary.AtRiskSubscriptions[0].RiskState)
	}
}

func TestGetRiskTimeline(t *testing.T) {
	response := `{"data":{"metricsTrend":[{"date":"2026-01-01T00:00:00Z","safeCount":40,"churnedCount":2,"revenueAtRiskCents":10000}]}}`
	exec := newTestExecutor(response)

	result := exec.execute(context.Background(), chat.ToolCall{
		ID:        "call-4",
		Name:      "get_risk_timeline",
		Arguments: `{"app_id": "00000000-0000-0000-0000-000000000001", "months": 3}`,
	})

	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}
}

func TestUnknownTool(t *testing.T) {
	exec := newTestExecutor(`{}`)

	result := exec.execute(context.Background(), chat.ToolCall{
		ID:   "call-5",
		Name: "nonexistent",
	})

	if !result.IsError {
		t.Error("expected error for unknown tool")
	}
}

func TestInvalidArguments(t *testing.T) {
	exec := newTestExecutor(`{}`)

	result := exec.execute(context.Background(), chat.ToolCall{
		ID:        "call-6",
		Name:      "get_risk_summary",
		Arguments: "not-json",
	})

	if !result.IsError {
		t.Error("expected error for invalid arguments")
	}
}

func TestModuleInterface(t *testing.T) {
	handler := &mockGraphQLHandler{response: `{"data":{}}`}
	gql := chat.NewGraphQLExecutor(handler)
	mod := New(gql)

	if mod.Name() != "risk" {
		t.Errorf("expected name 'risk', got '%s'", mod.Name())
	}
	if len(mod.Tools()) != 3 {
		t.Errorf("expected 3 tools, got %d", len(mod.Tools()))
	}
	if mod.PromptFragment() == "" {
		t.Error("expected non-empty prompt fragment")
	}
}
