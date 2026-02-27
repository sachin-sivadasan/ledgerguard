package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/interfaces/graphql"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/interfaces/http/handler"
	revenueMiddleware "github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/interfaces/http/middleware"
)

// Config holds all handlers and middleware for the Revenue API router
type Config struct {
	// Handlers
	APIKeyHandler             *handler.APIKeyHandler
	SubscriptionStatusHandler *handler.SubscriptionStatusHandler
	UsageStatusHandler        *handler.UsageStatusHandler
	GraphQLHandler            *graphql.Handler

	// Middleware
	APIKeyAuthMW  *revenueMiddleware.APIKeyAuth
	RateLimiterMW *revenueMiddleware.RateLimiter
	AuditLoggerMW *revenueMiddleware.AuditLogger

	// Firebase auth middleware for API key management
	FirebaseAuthMW func(next http.Handler) http.Handler
}

// New creates a new Revenue API router
func New(cfg Config) *chi.Mux {
	r := chi.NewRouter()

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Allow all origins for API access
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}))

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"revenue-api"}`))
	})

	// API v1 routes
	r.Route("/v1", func(r chi.Router) {
		// API Key management routes (requires Firebase auth - user must be logged in)
		if cfg.APIKeyHandler != nil && cfg.FirebaseAuthMW != nil {
			r.Route("/api-keys", func(r chi.Router) {
				r.Use(cfg.FirebaseAuthMW)
				r.Post("/", cfg.APIKeyHandler.Create)
				r.Get("/", cfg.APIKeyHandler.List)
				r.Delete("/{keyID}", cfg.APIKeyHandler.Revoke)
			})
		}

		// Public API routes (requires API key auth)
		apiKeyProtected := r.Group(nil)
		if cfg.APIKeyAuthMW != nil {
			apiKeyProtected.Use(cfg.APIKeyAuthMW.Middleware)
		}
		if cfg.RateLimiterMW != nil {
			apiKeyProtected.Use(cfg.RateLimiterMW.Middleware)
		}
		if cfg.AuditLoggerMW != nil {
			apiKeyProtected.Use(cfg.AuditLoggerMW.Middleware)
		}

		// REST endpoints
		if cfg.SubscriptionStatusHandler != nil {
			apiKeyProtected.Get("/subscriptions/{shopify_gid}", cfg.SubscriptionStatusHandler.GetByGID)
			apiKeyProtected.Get("/subscriptions/status", cfg.SubscriptionStatusHandler.GetByDomain) // ?domain=
			apiKeyProtected.Post("/subscriptions/batch", cfg.SubscriptionStatusHandler.GetBatch)
		}

		if cfg.UsageStatusHandler != nil {
			apiKeyProtected.Get("/usages/{shopify_gid}", cfg.UsageStatusHandler.GetByGID)
			apiKeyProtected.Post("/usages/batch", cfg.UsageStatusHandler.GetBatch)
		}

		// GraphQL endpoint
		if cfg.GraphQLHandler != nil {
			apiKeyProtected.Mount("/graphql", cfg.GraphQLHandler)
		}
	})

	return r
}
