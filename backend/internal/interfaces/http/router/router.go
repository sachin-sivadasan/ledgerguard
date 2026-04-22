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
	HealthHandler                   *handler.HealthHandler
	MeHandler                       *handler.MeHandler
	OAuthHandler                    *handler.OAuthHandler
	ManualTokenHandler              *handler.ManualTokenHandler
	IntegrationStatusHandler        *handler.IntegrationStatusHandler
	AppHandler                      *handler.AppHandler
	MetricsHandler                  *handler.MetricsHandler
	RevenueHandler                  *handler.RevenueHandler
	SyncHandler                     *handler.SyncHandler
	SubscriptionHandler             *handler.SubscriptionHandler
	StoreHealthHandler              *handler.StoreHealthHandler
	FeeHandler                      *handler.FeeHandler
	UserPreferencesHandler          *handler.UserPreferencesHandler
	NotificationPreferencesHandler  *handler.NotificationPreferencesHandler
	DeviceHandler                   *handler.DeviceHandler
	InsightHandler                  *handler.InsightHandler
	WebhookHandler                  *handler.WebhookHandler
	BillingHandler                  *handler.BillingHandler
	ReviewHandler                   *handler.ReviewHandler
	APIKeyHandler                   *apikeyhandler.APIKeyHandler
	QueueSyncHandler                *handler.QueueSyncHandler
	GraphQLHandler                  http.Handler       // Internal chat GraphQL endpoint
	ChatHandler                     http.HandlerFunc   // POST /api/v1/chat (SSE)
	ChatModulesHandler              http.HandlerFunc   // GET /api/v1/chat/modules
	AuthMW                          func(next http.Handler) http.Handler
	AdminMW                         func(next http.Handler) http.Handler // RequireRoles(ADMIN)
	InternalMW                      func(next http.Handler) http.Handler // Internal key authentication
}

func New(cfg Config) *chi.Mux {
	r := chi.NewRouter()

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "https://*.ledgerguard.app", "https://ledgerguard-c7557.web.app", "https://ledgerguard-c7557.firebaseapp.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
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

	// Webhook routes (no auth - validated via HMAC)
	if cfg.WebhookHandler != nil {
		r.Route("/webhooks/shopify", func(r chi.Router) {
			r.Post("/", cfg.WebhookHandler.HandleWebhook)
			r.Post("/subscriptions", cfg.WebhookHandler.HandleSubscriptionUpdate)
			r.Post("/uninstalled", cfg.WebhookHandler.HandleAppUninstalled)
			r.Post("/billing-failure", cfg.WebhookHandler.HandleBillingFailure)
		})
	}

	// Razorpay webhook route (no auth - validated via HMAC)
	if cfg.BillingHandler != nil {
		r.Post("/webhooks/razorpay", cfg.BillingHandler.HandleWebhook)
	}

	// Internal chat GraphQL endpoint (requires auth)
	if cfg.GraphQLHandler != nil && cfg.AuthMW != nil {
		r.Route("/graphql", func(r chi.Router) {
			r.Use(cfg.AuthMW)
			r.Handle("/*", cfg.GraphQLHandler)
			r.Handle("/", cfg.GraphQLHandler)
		})
	}

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Me endpoint (current user profile)
		if cfg.MeHandler != nil && cfg.AuthMW != nil {
			r.With(cfg.AuthMW).Get("/me", cfg.MeHandler.GetMe)
		}

		// User preferences routes
		if cfg.UserPreferencesHandler != nil && cfg.AuthMW != nil {
			r.Route("/user/preferences", func(r chi.Router) {
				r.Use(cfg.AuthMW)
				r.Get("/dashboard", cfg.UserPreferencesHandler.GetDashboardPreferences)
				r.Put("/dashboard", cfg.UserPreferencesHandler.SaveDashboardPreferences)
				r.Get("/default-app", cfg.UserPreferencesHandler.GetDefaultApp)
				r.Put("/default-app", cfg.UserPreferencesHandler.SetDefaultApp)
			})
		}

		// Notification preferences routes (plural "users" to match frontend)
		if cfg.NotificationPreferencesHandler != nil && cfg.AuthMW != nil {
			r.With(cfg.AuthMW).Get("/users/notification-preferences", cfg.NotificationPreferencesHandler.GetNotificationPreferences)
			r.With(cfg.AuthMW).Put("/users/notification-preferences", cfg.NotificationPreferencesHandler.SaveNotificationPreferences)
		}

		// Device registration routes for push notifications
		if cfg.DeviceHandler != nil && cfg.AuthMW != nil {
			r.Route("/devices", func(r chi.Router) {
				r.Use(cfg.AuthMW)
				r.Post("/", cfg.DeviceHandler.RegisterDevice)
				r.Delete("/", cfg.DeviceHandler.UnregisterDevice)
			})
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

		// Aggregate metrics (across all apps)
		if cfg.MetricsHandler != nil && cfg.AuthMW != nil {
			r.With(cfg.AuthMW).Get("/metrics/aggregate", cfg.MetricsHandler.GetAggregateMetrics)
		}

		// App routes (requires auth)
		if cfg.AppHandler != nil && cfg.AuthMW != nil {
			r.Route("/apps", func(r chi.Router) {
				r.Use(cfg.AuthMW)
				r.Get("/available", cfg.AppHandler.GetAvailableApps)
				r.Post("/select", cfg.AppHandler.SelectApp)
				r.Get("/", cfg.AppHandler.ListApps)

				// App settings routes
				r.Patch("/{appID}/tier", cfg.AppHandler.UpdateAppTier)

				// Metrics routes (appID is numeric, backend adds gid://partners/App/ prefix)
				if cfg.MetricsHandler != nil {
					r.Get("/{appID}/metrics/latest", cfg.MetricsHandler.GetLatestMetrics)
					r.Get("/{appID}/metrics", cfg.MetricsHandler.GetMetricsByPeriod)
				}

				// Subscription routes
				if cfg.SubscriptionHandler != nil {
					r.Get("/{appID}/subscriptions/summary", cfg.SubscriptionHandler.Summary)
					r.Get("/{appID}/subscriptions/price-stats", cfg.SubscriptionHandler.PriceStats)
					r.Get("/{appID}/subscriptions", cfg.SubscriptionHandler.List)
					r.Get("/{appID}/subscriptions/{subscriptionID}", cfg.SubscriptionHandler.GetByID)
				}

				// Earnings timeline routes
				if cfg.RevenueHandler != nil {
					r.Get("/{appID}/earnings", cfg.RevenueHandler.GetEarnings)
					r.Get("/{appID}/earnings/status", cfg.RevenueHandler.GetEarningsStatus)
				}

				// Fee breakdown routes
				if cfg.FeeHandler != nil {
					r.Get("/{appID}/fees/summary", cfg.FeeHandler.GetFeeSummary)
					r.Get("/{appID}/fees/breakdown", cfg.FeeHandler.GetTierBreakdown)
				}

				// Store health routes
				if cfg.StoreHealthHandler != nil {
					r.Get("/{appID}/stores/{domain}/health", cfg.StoreHealthHandler.GetStoreHealth)
				}

				// Insights routes
				if cfg.InsightHandler != nil {
					r.Get("/{appID}/insights/daily", cfg.InsightHandler.GetDailyInsight)
				}

				// Install count routes
				r.Get("/{appID}/install-count", cfg.AppHandler.GetInstallCount)
				r.Post("/{appID}/refresh-install-count", cfg.AppHandler.RefreshInstallCount)

				// App store slug route
				r.Patch("/{appID}/store-slug", cfg.AppHandler.UpdateStoreSlug)

				// Review routes
				if cfg.ReviewHandler != nil {
					r.Get("/{appID}/reviews", cfg.ReviewHandler.List)
					r.Post("/{appID}/reviews/scrape", cfg.ReviewHandler.Scrape)
				}
			})
		}

		// Billing routes (requires auth)
		if cfg.BillingHandler != nil && cfg.AuthMW != nil {
			r.Route("/billing", func(r chi.Router) {
				r.Use(cfg.AuthMW)
				r.Post("/checkout", cfg.BillingHandler.CreateCheckout)
				r.Get("/status", cfg.BillingHandler.GetStatus)
			})
		}

		// Tiers route (public info)
		if cfg.FeeHandler != nil {
			r.Get("/tiers", cfg.FeeHandler.ListAvailableTiers)
		}

		// Sync routes (requires auth)
		if cfg.SyncHandler != nil && cfg.AuthMW != nil {
			r.Route("/sync", func(r chi.Router) {
				r.Use(cfg.AuthMW)
				r.Post("/", cfg.SyncHandler.SyncAllApps)
				r.Post("/{appID}", cfg.SyncHandler.SyncApp)

				// Queue-based sync routes (alongside existing sync — zero breakage)
				if cfg.QueueSyncHandler != nil {
					r.Post("/enqueue/{appID}", cfg.QueueSyncHandler.EnqueueSync)
					r.Get("/jobs", cfg.QueueSyncHandler.ListJobs)
					r.Get("/jobs/{jobID}", cfg.QueueSyncHandler.GetJobStatus)
					r.Get("/jobs/{jobID}/progress", cfg.QueueSyncHandler.GetJobProgress)
					r.Post("/jobs/{jobID}/cancel", cfg.QueueSyncHandler.CancelJob)
				}
			})
		}

		// Chat routes (requires auth, SSE streaming)
		if cfg.ChatHandler != nil && cfg.AuthMW != nil {
			r.Route("/chat", func(r chi.Router) {
				r.Use(cfg.AuthMW)
				r.Post("/", cfg.ChatHandler)
				if cfg.ChatModulesHandler != nil {
					r.Get("/modules", cfg.ChatModulesHandler)
				}
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

		// Internal routes (authenticated via X-Internal-Key header)
		// Used for service-to-service calls and internal testing
		if cfg.InternalMW != nil {
			r.Route("/internal", func(r chi.Router) {
				r.Use(cfg.InternalMW)

				// Refresh install count for all apps or a specific app
				if cfg.SyncHandler != nil {
					r.Post("/sync/transactions", cfg.SyncHandler.SyncAllApps)
					r.Post("/sync/transactions/{appID}", cfg.SyncHandler.SyncApp)
				}
			})
		}
	})

	return r
}
