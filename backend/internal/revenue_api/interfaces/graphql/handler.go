package graphql

import (
	"encoding/json"
	"net/http"

	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/interfaces/http/middleware"
)

// Handler handles GraphQL requests for the Revenue API
type Handler struct {
	resolver *Resolver
}

// NewHandler creates a new GraphQL handler
func NewHandler(resolver *Resolver) *Handler {
	return &Handler{resolver: resolver}
}

// GraphQLRequest represents an incoming GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   interface{}     `json:"data,omitempty"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

// ServeHTTP handles GraphQL requests
// POST /graphql
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	apiKey := middleware.APIKeyFromContext(r.Context())
	if apiKey == nil {
		writeGraphQLError(w, http.StatusUnauthorized, "API key required", "UNAUTHORIZED")
		return
	}

	// Only accept POST
	if r.Method != http.MethodPost {
		writeGraphQLError(w, http.StatusMethodNotAllowed, "only POST is allowed", "METHOD_NOT_ALLOWED")
		return
	}

	// Parse request
	var req GraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGraphQLError(w, http.StatusBadRequest, "invalid request body", "INVALID_REQUEST")
		return
	}

	if req.Query == "" {
		writeGraphQLError(w, http.StatusBadRequest, "query is required", "INVALID_REQUEST")
		return
	}

	// Execute query using our simple resolver
	// Note: For a full implementation, you would use gqlgen's generated executor
	// This is a simplified version that demonstrates the pattern
	result, err := h.executeQuery(r.Context(), req)
	if err != nil {
		writeGraphQLError(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GraphQLResponse{Data: result})
}

// executeQuery executes a GraphQL query
// This is a simplified implementation - a full implementation would use gqlgen
func (h *Handler) executeQuery(ctx interface{}, req GraphQLRequest) (interface{}, error) {
	// For now, return a helpful message
	// Full gqlgen integration would parse and execute the query
	return map[string]interface{}{
		"_info": "GraphQL endpoint active. Use gqlgen for full query parsing.",
		"_schema": map[string]interface{}{
			"queries": []string{
				"subscription(shopifyGid: ID!): SubscriptionStatus",
				"subscriptionByDomain(domain: String!): SubscriptionStatus",
				"subscriptions(shopifyGids: [ID!]!): SubscriptionBatchResult",
				"usage(shopifyGid: ID!): UsageStatus",
				"usages(shopifyGids: [ID!]!): UsageBatchResult",
			},
		},
	}, nil
}

func writeGraphQLError(w http.ResponseWriter, status int, message string, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(GraphQLResponse{
		Errors: []GraphQLError{{Message: message, Code: code}},
	})
}
