package graphql

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

// NewHandler creates an HTTP handler for the internal chat GraphQL endpoint.
// GET requests serve the GraphQL Playground UI.
// POST requests execute GraphQL queries.
func NewHandler(resolver *Resolver) http.Handler {
	srv := handler.NewDefaultServer(NewExecutableSchema(Config{
		Resolvers: resolver,
	}))

	mux := http.NewServeMux()
	mux.Handle("GET /", playground.Handler("LedgerGuard Chat GraphQL", "/graphql"))
	mux.Handle("POST /", srv)

	return mux
}
