package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/handler"
)

type Config struct {
	HealthHandler *handler.HealthHandler
	OAuthHandler  *handler.OAuthHandler
	AuthMW        func(next http.Handler) http.Handler
}

func New(cfg Config) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Public routes
	r.Get("/health", cfg.HealthHandler.Health)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// OAuth routes (public - no auth required for redirect)
		r.Route("/integrations/shopify", func(r chi.Router) {
			if cfg.OAuthHandler != nil {
				// StartOAuth requires auth (user must be logged in)
				r.With(cfg.AuthMW).Get("/oauth", cfg.OAuthHandler.StartOAuth)
				// Callback is public (receives redirect from Shopify)
				r.Get("/callback", cfg.OAuthHandler.Callback)
			}
		})
	})

	return r
}
