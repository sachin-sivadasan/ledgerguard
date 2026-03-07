package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

// GraphQLExecutor executes GraphQL queries against the internal gqlgen schema.
// Thread-safe: the underlying HTTP handler is stateless.
type GraphQLExecutor struct {
	handler http.Handler
}

// NewGraphQLExecutor wraps a gqlgen HTTP handler for programmatic execution.
func NewGraphQLExecutor(handler http.Handler) *GraphQLExecutor {
	return &GraphQLExecutor{handler: handler}
}

type gqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []gqlError      `json:"errors,omitempty"`
}

type gqlError struct {
	Message string `json:"message"`
}

// Execute runs a GraphQL query and returns the raw JSON data.
// The context is forwarded to resolvers (carries auth info, etc.).
func (e *GraphQLExecutor) Execute(ctx context.Context, query string, vars map[string]any) (json.RawMessage, error) {
	body, err := json.Marshal(gqlRequest{Query: query, Variables: vars})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/graphql", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	e.handler.ServeHTTP(rec, req)

	respBody, err := io.ReadAll(rec.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if rec.Code != http.StatusOK {
		return nil, fmt.Errorf("graphql: HTTP %d: %s", rec.Code, string(respBody))
	}

	var resp gqlResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("graphql: %s", resp.Errors[0].Message)
	}

	return resp.Data, nil
}
