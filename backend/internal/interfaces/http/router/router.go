package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/handler"
)

type Config struct {
	HealthHandler      *handler.HealthHandler
	OAuthHandler       *handler.OAuthHandler
	ManualTokenHandler *handler.ManualTokenHandler
	AppHandler         *handler.AppHandler
	SyncHandler        *handler.SyncHandler
	AuthMW             func(next http.Handler) http.Handler
	AdminMW            func(next http.Handler) http.Handler // RequireRoles(ADMIN)
}

func New(cfg Config) *chi.Mux {
	r := chi.NewRouter()

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "https://*.ledgerguard.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Public routes
	r.Get("/health", cfg.HealthHandler.Health)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Shopify integration routes
		r.Route("/integrations/shopify", func(r chi.Router) {
			// OAuth routes
			if cfg.OAuthHandler != nil {
				// StartOAuth requires auth (user must be logged in)
				r.With(cfg.AuthMW).Get("/oauth", cfg.OAuthHandler.StartOAuth)
				// Callback is public (receives redirect from Shopify)
				r.Get("/callback", cfg.OAuthHandler.Callback)
			}

			// Manual token routes (ADMIN only)
			if cfg.ManualTokenHandler != nil && cfg.AuthMW != nil && cfg.AdminMW != nil {
				r.With(cfg.AuthMW, cfg.AdminMW).Post("/token", cfg.ManualTokenHandler.AddToken)
				r.With(cfg.AuthMW, cfg.AdminMW).Get("/token", cfg.ManualTokenHandler.GetToken)
				r.With(cfg.AuthMW, cfg.AdminMW).Delete("/token", cfg.ManualTokenHandler.RevokeToken)
			}
		})

		// App routes (requires auth)
		if cfg.AppHandler != nil && cfg.AuthMW != nil {
			r.Route("/apps", func(r chi.Router) {
				r.Use(cfg.AuthMW)
				r.Get("/available", cfg.AppHandler.GetAvailableApps)
				r.Post("/select", cfg.AppHandler.SelectApp)
				r.Get("/", cfg.AppHandler.ListApps)
			})
		}

		// Sync routes (requires auth)
		if cfg.SyncHandler != nil && cfg.AuthMW != nil {
			r.Route("/sync", func(r chi.Router) {
				r.Use(cfg.AuthMW)
				r.Post("/", cfg.SyncHandler.SyncAllApps)
				r.Post("/{appID}", cfg.SyncHandler.SyncApp)
			})
		}
	})

	return r
}
