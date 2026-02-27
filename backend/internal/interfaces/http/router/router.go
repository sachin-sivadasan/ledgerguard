package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/handler"
	apikeyhandler "github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/interfaces/http/handler"
)

type Config struct {
	HealthHandler            *handler.HealthHandler
	MeHandler                *handler.MeHandler
	OAuthHandler             *handler.OAuthHandler
	ManualTokenHandler       *handler.ManualTokenHandler
	IntegrationStatusHandler *handler.IntegrationStatusHandler
	AppHandler               *handler.AppHandler
	MetricsHandler           *handler.MetricsHandler
	SyncHandler              *handler.SyncHandler
	SubscriptionHandler      *handler.SubscriptionHandler
	APIKeyHandler            *apikeyhandler.APIKeyHandler
	AuthMW                   func(next http.Handler) http.Handler
	AdminMW                  func(next http.Handler) http.Handler // RequireRoles(ADMIN)
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
		// Me endpoint (current user profile)
		if cfg.MeHandler != nil && cfg.AuthMW != nil {
			r.With(cfg.AuthMW).Get("/me", cfg.MeHandler.GetMe)
		}

		// Shopify integration routes
		r.Route("/integrations/shopify", func(r chi.Router) {
			// Integration status (user accessible)
			if cfg.IntegrationStatusHandler != nil && cfg.AuthMW != nil {
				r.With(cfg.AuthMW).Get("/status", cfg.IntegrationStatusHandler.GetStatus)
			}

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

				// Metrics routes (appID is numeric, backend adds gid://partners/App/ prefix)
				if cfg.MetricsHandler != nil {
					r.Get("/{appID}/metrics/latest", cfg.MetricsHandler.GetLatestMetrics)
					r.Get("/{appID}/metrics", cfg.MetricsHandler.GetMetricsByPeriod)
				}

				// Subscription routes
				if cfg.SubscriptionHandler != nil {
					r.Get("/{appID}/subscriptions", cfg.SubscriptionHandler.List)
					r.Get("/{appID}/subscriptions/{subscriptionID}", cfg.SubscriptionHandler.GetByID)
				}
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

		// API key routes (requires auth)
		if cfg.APIKeyHandler != nil && cfg.AuthMW != nil {
			r.Route("/api-keys", func(r chi.Router) {
				r.Use(cfg.AuthMW)
				r.Get("/", cfg.APIKeyHandler.List)
				r.Post("/", cfg.APIKeyHandler.Create)
				r.Delete("/{id}", cfg.APIKeyHandler.Revoke)
			})
		}
	})

	return r
}
