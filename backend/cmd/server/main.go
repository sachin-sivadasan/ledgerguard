package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sachin-sivadasan/ledgerguard/internal/application/scheduler"
	appservice "github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	domainservice "github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/cache"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/config"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/external"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/persistence"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/queue"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/queue/processors"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/handler"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/router"
	"github.com/sachin-sivadasan/ledgerguard/pkg/crypto"
	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
	chatgraphql "github.com/sachin-sivadasan/ledgerguard/internal/chat/graphql"
	chatapps "github.com/sachin-sivadasan/ledgerguard/internal/chat/modules/apps"
	chatearnings "github.com/sachin-sivadasan/ledgerguard/internal/chat/modules/earnings"
	chatmetrics "github.com/sachin-sivadasan/ledgerguard/internal/chat/modules/metrics"
	chatrisk "github.com/sachin-sivadasan/ledgerguard/internal/chat/modules/risk"
	chatstorehealth "github.com/sachin-sivadasan/ledgerguard/internal/chat/modules/store_health"
	chatsubs "github.com/sachin-sivadasan/ledgerguard/internal/chat/modules/subscriptions"
	apikeyhandler "github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/interfaces/http/handler"
	apikeysvc "github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/application/service"
	apikeypersist "github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/infrastructure/persistence"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}

func run() error {
	ctx := context.Background()

	// Parse command line flags
	configPath := flag.String("config", "", "Path to config file (yaml)")
	flag.Parse()

	// Allow CONFIG_PATH env var as fallback
	if *configPath == "" {
		*configPath = os.Getenv("CONFIG_PATH")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if *configPath != "" {
		log.Printf("Loaded config from: %s", *configPath)
	}

	// Initialize database connection
	var db *persistence.PostgresDB
	db, err = persistence.NewPostgresDB(ctx, cfg.Database.DSN())
	if err != nil {
		log.Printf("WARNING: failed to connect to database: %v", err)
		log.Printf("Server will start without database connection")
		db = nil
	} else {
		defer db.Close()
		log.Println("Connected to PostgreSQL")

		// Run database migrations
		if cfg.Database.MigrationsPath != "" {
			migrator, err := persistence.NewMigrator(cfg.Database.DSN(), cfg.Database.MigrationsPath)
			if err != nil {
				log.Printf("WARNING: failed to initialize migrator: %v", err)
			} else {
				if err := migrator.Up(); err != nil {
					log.Printf("WARNING: failed to run migrations: %v", err)
				} else {
					log.Println("Database migrations applied successfully")
				}
				migrator.Close()
			}
		}
	}

	// Initialize Firebase Auth (optional - will fail gracefully if not configured)
	var firebaseAuth *external.FirebaseAuthService
	firebaseAuth, err = external.NewFirebaseAuthService(ctx, cfg.Firebase.CredentialsFile)
	if err != nil {
		log.Printf("WARNING: Firebase Auth not configured: %v", err)
		log.Printf("Authentication will not work without Firebase configuration")
	} else {
		log.Println("Firebase Auth initialized")
	}

	// Initialize Firebase Messaging for push notifications (optional)
	var firebaseMessaging *external.FirebaseMessagingService
	if cfg.Firebase.CredentialsFile != "" {
		firebaseMessaging, err = external.NewFirebaseMessagingService(ctx, cfg.Firebase.CredentialsFile)
		if err != nil {
			log.Printf("WARNING: Firebase Messaging not configured: %v", err)
			log.Printf("Push notifications will not work")
		} else {
			log.Println("Firebase Messaging initialized")
		}
	}

	// Initialize encryption
	var encryptor *crypto.AESEncryptor
	if cfg.Encryption.MasterKey != "" {
		encryptor, err = crypto.NewAESEncryptor([]byte(cfg.Encryption.MasterKey))
		if err != nil {
			log.Printf("WARNING: Failed to initialize encryption: %v", err)
		} else {
			log.Println("Encryption initialized")
		}
	}

	// Initialize Redis client (optional — for queue-based sync)
	var redisClient *redis.Client
	if cfg.Queue.Enabled && cfg.Redis.Addr != "" {
		redisClient, err = queue.NewRedisClient(ctx, cfg.Redis)
		if err != nil {
			log.Printf("WARNING: Redis not available: %v", err)
			log.Printf("WARNING: Queue-based sync requires Redis — install and start Redis to enable")
			redisClient = nil
		}
	}

	// Initialize repositories
	var userRepo *persistence.PostgresUserRepository
	var partnerRepo *persistence.PostgresPartnerAccountRepository
	var appRepo *persistence.PostgresAppRepository
	var txRepo *persistence.PostgresTransactionRepository
	var subscriptionRepo *persistence.PostgresSubscriptionRepository
	var snapshotRepo *persistence.PostgresDailyMetricsSnapshotRepository
	var shopRepo *persistence.PostgresShopRepository
	var billingSubRepo *persistence.PostgresBillingSubscriptionRepository
	var reviewRepo *persistence.PostgresAppReviewRepository
	var syncJobRepo *persistence.PostgresSyncJobRepository
	var appEventRepo *persistence.PostgresAppEventRepository
	var adminRepo repository.AdminRepository

	if db != nil {
		userRepo = persistence.NewPostgresUserRepository(db.Pool)
		partnerRepo = persistence.NewPostgresPartnerAccountRepository(db.Pool)
		appRepo = persistence.NewPostgresAppRepository(db.Pool)
		txRepo = persistence.NewPostgresTransactionRepository(db.Pool)
		subscriptionRepo = persistence.NewPostgresSubscriptionRepository(db.Pool)
		snapshotRepo = persistence.NewPostgresDailyMetricsSnapshotRepository(db.Pool)
		shopRepo = persistence.NewPostgresShopRepository(db.Pool)
		billingSubRepo = persistence.NewPostgresBillingSubscriptionRepository(db.Pool)
		reviewRepo = persistence.NewPostgresAppReviewRepository(db.Pool)
		syncJobRepo = persistence.NewPostgresSyncJobRepository(db.Pool)
		appEventRepo = persistence.NewPostgresAppEventRepository(db.Pool)
		adminRepo = persistence.NewPostgresAdminRepository(db.Pool)
	}

	// Initialize event tracker (Mixpanel or Noop)
	var tracker domainservice.EventTracker
	if cfg.Mixpanel.Token != "" {
		tracker = external.NewMixpanelClient(cfg.Mixpanel.Token)
		log.Println("Mixpanel event tracker initialized")
	} else {
		tracker = external.NewNoopTracker()
		log.Println("Event tracker: noop (MIXPANEL_TOKEN not set)")
	}

	// Initialize OAuth state store (10 minute TTL)
	stateStore := cache.NewOAuthStateStore(10 * time.Minute)

	// Initialize Shopify OAuth service
	var oauthService *external.ShopifyOAuthService
	if cfg.Shopify.ClientID != "" {
		oauthService = external.NewShopifyOAuthService(
			cfg.Shopify.ClientID,
			cfg.Shopify.ClientSecret,
			cfg.Shopify.RedirectURI,
			cfg.Shopify.Scopes,
		)
		log.Println("Shopify OAuth initialized")
	}

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(db)
	meHandler := handler.NewMeHandler()

	var oauthHandler *handler.OAuthHandler
	if oauthService != nil && encryptor != nil && partnerRepo != nil && userRepo != nil {
		oauthHandler = handler.NewOAuthHandler(
			oauthService,
			encryptor,
			partnerRepo,
			userRepo,
			stateStore,
		)
		log.Println("OAuth handler initialized")
	}

	var manualTokenHandler *handler.ManualTokenHandler
	if encryptor != nil && partnerRepo != nil {
		manualTokenHandler = handler.NewManualTokenHandler(encryptor, partnerRepo)
		log.Println("Manual token handler initialized")
	}

	var integrationStatusHandler *handler.IntegrationStatusHandler
	if partnerRepo != nil {
		integrationStatusHandler = handler.NewIntegrationStatusHandler(partnerRepo)
		log.Println("Integration status handler initialized")
	}

	// Initialize Shopify Partner client for fetching apps with rate limiting
	partnerClient := external.NewShopifyPartnerClient(
		external.WithRequestsPerSecond(cfg.Shopify.RateLimitRPS),
	)

	var appHandler *handler.AppHandler
	if partnerRepo != nil && appRepo != nil && encryptor != nil {
		appHandler = handler.NewAppHandler(partnerClient, partnerRepo, appRepo, encryptor)
		appHandler.SetTracker(tracker)
		log.Println("App handler initialized with Partner client")
	}

	// Initialize metrics aggregation service and handler
	var metricsHandler *handler.MetricsHandler
	if snapshotRepo != nil && appRepo != nil && partnerRepo != nil && txRepo != nil {
		metricsEngine := domainservice.NewMetricsEngine()
		metricsAggregator := appservice.NewMetricsAggregationService(snapshotRepo, txRepo, metricsEngine)
		metricsHandler = handler.NewMetricsHandler(metricsAggregator, appRepo, partnerRepo)
		log.Println("Metrics handler initialized with aggregation service")
	} else {
		// Fallback to handler without aggregator (will use mock data)
		metricsHandler = handler.NewMetricsHandler(nil, appRepo, partnerRepo)
		log.Println("Metrics handler initialized (without aggregator)")
	}
	if metricsHandler != nil {
		metricsHandler.SetTracker(tracker)
	}

	// Initialize sync service and handler
	var syncService *appservice.SyncService
	var syncHandler *handler.SyncHandler
	var syncScheduler *scheduler.SyncScheduler
	var notificationScheduler *scheduler.NotificationScheduler

	if txRepo != nil && appRepo != nil && partnerRepo != nil && encryptor != nil && subscriptionRepo != nil {
		// Initialize ledger service for rebuilding after sync
		ledgerService := domainservice.NewLedgerService(txRepo, subscriptionRepo)
		if snapshotRepo != nil {
			ledgerService = ledgerService.WithSnapshotRepository(snapshotRepo)
		}

		// Initialize sync service with Shopify Partner client for live transaction fetching
		syncService = appservice.NewSyncService(
			partnerClient, // ShopifyPartnerClient implements TransactionFetcher
			txRepo,
			appRepo,
			partnerRepo,
			encryptor,
			ledgerService,
		)

		// Wire subscription repo for status enrichment and brand fetch
		syncService = syncService.WithSubscriptionRepo(subscriptionRepo)

		// Wire shop brand fetcher for logo sync
		if shopRepo != nil {
			storefrontClient := external.NewShopifyStorefrontClient()
			syncService = syncService.WithShopBrandFetcher(storefrontClient, shopRepo)
			log.Println("Shop brand fetcher initialized")
		}

		// Wire review scraper for app store review sync
		if reviewRepo != nil {
			syncAppStoreScraper := external.NewShopifyAppStoreClient()
			syncService = syncService.WithReviewScraper(syncAppStoreScraper, reviewRepo)
			log.Println("Review scraper initialized for sync")
		}

		syncHandler = handler.NewSyncHandler(syncService, partnerRepo, appRepo)
		log.Println("Sync handler initialized")

		// Initialize and start scheduler (skip if queue-based sync is enabled)
		if !cfg.Queue.Enabled {
			syncScheduler = scheduler.NewSyncScheduler(syncService, partnerRepo)
			syncScheduler.Start(ctx)
			log.Println("Sync scheduler started (12-hour interval)")
		} else if redisClient == nil {
			log.Println("WARNING: Sync scheduler skipped — queue enabled but Redis unavailable; no background sync will run")
		} else {
			log.Println("Sync scheduler skipped — queue-based sync handles scheduling")
		}
	}

	// Initialize subscription handler
	var subscriptionHandler *handler.SubscriptionHandler
	if subscriptionRepo != nil && partnerRepo != nil && appRepo != nil {
		subscriptionHandler = handler.NewSubscriptionHandler(subscriptionRepo, partnerRepo, appRepo)
		if shopRepo != nil {
			subscriptionHandler.SetShopRepo(shopRepo)
		}
		log.Println("Subscription handler initialized")
	}

	// Initialize store health handler
	var storeHealthHandler *handler.StoreHealthHandler
	if subscriptionRepo != nil && txRepo != nil && partnerRepo != nil && appRepo != nil {
		storeHealthHandler = handler.NewStoreHealthHandler(subscriptionRepo, txRepo, partnerRepo, appRepo)
		if shopRepo != nil {
			storeHealthHandler.SetShopRepo(shopRepo)
		}
		log.Println("Store health handler initialized")
	}

	// Initialize revenue (earnings timeline) handler
	var revenueHandler *handler.RevenueHandler
	if db != nil && partnerRepo != nil && appRepo != nil {
		revenueRepo := persistence.NewPostgresRevenueRepository(db.Pool)
		revenueSvc := appservice.NewRevenueMetricsService(revenueRepo)
		revenueHandler = handler.NewRevenueHandler(revenueSvc, partnerRepo, appRepo)
		log.Println("Revenue handler initialized")
	}

	// Initialize fee handler
	var feeHandler *handler.FeeHandler
	if appRepo != nil && partnerRepo != nil && txRepo != nil {
		feeService := domainservice.NewFeeVerificationService()
		feeHandler = handler.NewFeeHandler(appRepo, partnerRepo, txRepo, feeService)
		log.Println("Fee handler initialized")
	}

	// Initialize API key handler
	var apiKeyHandler *apikeyhandler.APIKeyHandler
	if db != nil {
		apiKeyRepo := apikeypersist.NewPostgresAPIKeyRepository(db.Pool)
		apiKeySvc := apikeysvc.NewAPIKeyService(apiKeyRepo)
		apiKeyHandler = apikeyhandler.NewAPIKeyHandler(apiKeySvc)
		log.Println("API key handler initialized")
	}

	// Initialize user preferences handler
	var userPreferencesHandler *handler.UserPreferencesHandler
	if db != nil {
		userPreferencesHandler = handler.NewUserPreferencesHandler(db.Pool)
		log.Println("User preferences handler initialized")
	}

	// Initialize notification preferences handler
	var notificationPreferencesHandler *handler.NotificationPreferencesHandler
	var notificationPrefsRepo *persistence.PostgresNotificationPreferencesRepository
	if db != nil {
		notificationPrefsRepo = persistence.NewPostgresNotificationPreferencesRepository(db.Pool)
		notificationPreferencesHandler = handler.NewNotificationPreferencesHandler(notificationPrefsRepo)
		log.Println("Notification preferences handler initialized")
	}

	// Initialize notification service and device handler
	var notificationService *appservice.NotificationService
	var deviceHandler *handler.DeviceHandler
	if db != nil && notificationPrefsRepo != nil {
		deviceTokenRepo := persistence.NewPostgresDeviceTokenRepository(db.Pool)

		// Create notification service with optional Firebase push provider
		var pushProvider appservice.PushNotificationProvider
		if firebaseMessaging != nil {
			pushProvider = firebaseMessaging
		}

		notificationService = appservice.NewNotificationService(
			deviceTokenRepo,
			notificationPrefsRepo,
			pushProvider,
		)

		// Add Slack notification support
		slackProvider := external.NewSlackNotificationProvider()
		notificationService = notificationService.WithSlackNotifier(slackProvider)

		deviceHandler = handler.NewDeviceHandler(notificationService)
		log.Println("Notification service and device handler initialized")

		// Initialize notification scheduler for daily summaries
		if snapshotRepo != nil && appRepo != nil && partnerRepo != nil {
			notificationScheduler = scheduler.NewNotificationScheduler(
				notificationService,
				notificationPrefsRepo,
				snapshotRepo,
				appRepo,
				partnerRepo,
			)
			notificationScheduler.Start(ctx)
			log.Println("Notification scheduler started (15-minute check interval)")
		}
	}

	// Initialize insight handler
	var insightHandler *handler.InsightHandler
	if db != nil && appRepo != nil && partnerRepo != nil {
		dailyInsightRepo := persistence.NewPostgresDailyInsightRepository(db.Pool)
		insightHandler = handler.NewInsightHandler(dailyInsightRepo, appRepo, partnerRepo)
		log.Println("Insight handler initialized")
	}

	// Initialize Razorpay billing (optional — gracefully skipped if not configured)
	var billingHandler *handler.BillingHandler
	if cfg.Razorpay.KeyID != "" && billingSubRepo != nil && userRepo != nil {
		razorpayClient := external.NewRazorpayClient(cfg.Razorpay.KeyID, cfg.Razorpay.KeySecret)
		billingService := appservice.NewBillingService(
			razorpayClient,
			billingSubRepo,
			userRepo,
			cfg.Razorpay.WebhookSecret,
			cfg.Razorpay.StarterPlanID,
			cfg.Razorpay.ProPlanID,
		)
		billingService.SetTracker(tracker)
		billingHandler = handler.NewBillingHandler(billingService)
		log.Println("Razorpay billing handler initialized")
	}

	// Initialize review handler with Shopify App Store scraper
	var reviewHandler *handler.ReviewHandler
	appStoreScraper := external.NewShopifyAppStoreClient()
	if reviewRepo != nil && appRepo != nil && partnerRepo != nil {
		reviewHandler = handler.NewReviewHandler(reviewRepo, appRepo, partnerRepo, appStoreScraper)
		log.Println("Review handler initialized")
	}

	// Initialize admin handler
	var adminHandler *handler.AdminHandler
	if adminRepo != nil {
		adminHandler = handler.NewAdminHandler(adminRepo)
		log.Println("Admin handler initialized")
	}

	// Initialize queue-based sync system (optional — requires Redis + Queue enabled)
	var queueSyncHandler *handler.QueueSyncHandler
	var queueSyncService *appservice.QueueSyncService
	var regularWorkerPool *queue.WorkerPool
	var fullSyncWorkerPool *queue.WorkerPool
	var recoveryService *queue.RecoveryService

	if cfg.Queue.Enabled && redisClient != nil && syncJobRepo != nil && appRepo != nil && partnerRepo != nil && encryptor != nil {
		lockManager := queue.NewLockManager(redisClient)
		progressTracker := queue.NewProgressTracker(redisClient, syncJobRepo, 2*time.Second, 30*time.Second)
		processorRegistry := queue.NewProcessorRegistry()

		// Register processors
		if txRepo != nil && subscriptionRepo != nil {
			ledgerServiceForQueue := domainservice.NewLedgerService(txRepo, subscriptionRepo)
			if snapshotRepo != nil {
				ledgerServiceForQueue = ledgerServiceForQueue.WithSnapshotRepository(snapshotRepo)
			}

			processorRegistry.Register(processors.NewTransactionProcessor(
				partnerClient, txRepo, appRepo, partnerRepo, encryptor, ledgerServiceForQueue,
				syncJobRepo, lockManager, progressTracker,
			))
			processorRegistry.Register(processors.NewSnapshotProcessor(
				txRepo, appRepo, partnerRepo, encryptor, ledgerServiceForQueue,
				syncJobRepo, lockManager, progressTracker,
			))
			processorRegistry.Register(processors.NewStatusProcessor(
				partnerClient, subscriptionRepo, appRepo, partnerRepo, encryptor,
				syncJobRepo, lockManager, progressTracker,
			))
			processorRegistry.Register(processors.NewStoreProcessor(
				external.NewShopifyStorefrontClient(), shopRepo, subscriptionRepo,
				appRepo, partnerRepo, encryptor,
				syncJobRepo, lockManager, progressTracker,
			))

			if appEventRepo != nil {
				processorRegistry.Register(processors.NewEventProcessor(
					partnerClient, appEventRepo, subscriptionRepo,
					appRepo, partnerRepo, encryptor,
					syncJobRepo, lockManager, progressTracker,
				))
			}
		}

		if reviewRepo != nil {
			processorRegistry.Register(processors.NewReviewProcessor(
				external.NewShopifyAppStoreClient(), reviewRepo, appRepo,
				syncJobRepo, lockManager, progressTracker,
			))
		}

		processorRegistry.Register(processors.NewFullSyncProcessor(
			syncJobRepo, redisClient, lockManager, progressTracker,
		))

		// Start worker pools
		regularWorkerPool = queue.NewWorkerPool(
			"regular", queue.RegularQueueKey, cfg.Queue.NumWorkers,
			redisClient, syncJobRepo, lockManager, progressTracker, processorRegistry,
		)
		regularWorkerPool.Start(ctx)

		fullSyncWorkerPool = queue.NewWorkerPool(
			"full", queue.FullSyncQueueKey, cfg.Queue.FullSyncWorkers,
			redisClient, syncJobRepo, lockManager, progressTracker, processorRegistry,
		)
		fullSyncWorkerPool.Start(ctx)

		// Start recovery
		recoveryService = queue.NewRecoveryService(syncJobRepo, redisClient, lockManager, 10*time.Minute)
		recoveryService.RecoverOnStartup(ctx)
		recoveryService.StartPeriodicRecovery(ctx)

		// Create queue sync service and handler
		queueSyncService = appservice.NewQueueSyncService(
			syncJobRepo, appRepo, partnerRepo, redisClient, lockManager, progressTracker,
		)
		queueSyncService.SetTracker(tracker)
		queueSyncHandler = handler.NewQueueSyncHandler(queueSyncService, partnerRepo, appRepo)
		log.Println("Queue-based sync system initialized")
	}

	// Initialize chat GraphQL handler and chat handler
	var graphqlHandler http.Handler
	var chatHandler *chat.Handler
	if subscriptionRepo != nil && txRepo != nil && snapshotRepo != nil && appRepo != nil && partnerRepo != nil {
		riskEngine := domainservice.NewRiskEngine()
		metricsEngine := domainservice.NewMetricsEngine()

		var subscriptionEventRepo *persistence.PostgresSubscriptionEventRepository
		subscriptionEventRepo = persistence.NewPostgresSubscriptionEventRepository(db.Pool)

		chatResolver := &chatgraphql.Resolver{
			SubscriptionRepo:      subscriptionRepo,
			TransactionRepo:       txRepo,
			SnapshotRepo:          snapshotRepo,
			AppRepo:               appRepo,
			PartnerAccountRepo:    partnerRepo,
			SubscriptionEventRepo: subscriptionEventRepo,
			RiskEngine:            riskEngine,
			MetricsEngine:         metricsEngine,
		}
		graphqlHandler = chatgraphql.NewHandler(chatResolver)
		log.Println("Chat GraphQL handler initialized")

		// Set up module registry and chat handler
		gqlExecutor := chat.NewGraphQLExecutor(graphqlHandler)
		moduleRegistry := chat.NewRegistry()
		moduleRegistry.Register(chatapps.New(gqlExecutor))
		moduleRegistry.Register(chatrisk.New(gqlExecutor))
		moduleRegistry.Register(chatsubs.New(gqlExecutor))
		moduleRegistry.Register(chatmetrics.New(gqlExecutor))
		moduleRegistry.Register(chatstorehealth.New(gqlExecutor))
		moduleRegistry.Register(chatearnings.New(gqlExecutor))

		// Set up AI provider
		aiProviders := chat.NewAIProviderRegistry("openai")
		if cfg.OpenAI.APIKey != "" {
			openaiClient := external.NewOpenAIClient(cfg.OpenAI.APIKey, cfg.OpenAI.Model)
			aiProviders.Register("openai", openaiClient)
			log.Println("OpenAI chat provider registered")
		}

		chatHandler = chat.NewHandler(moduleRegistry, aiProviders)
		log.Println("Chat handler initialized with", len(moduleRegistry.Modules()), "modules")
	}

	// Initialize auth middleware
	var authMW func(http.Handler) http.Handler
	if firebaseAuth != nil && userRepo != nil {
		authMiddleware := middleware.NewAuthMiddleware(firebaseAuth, userRepo)
		authMiddleware.SetTracker(tracker)
		authMW = authMiddleware.Authenticate
		log.Println("Auth middleware initialized")
	}

	// Initialize admin middleware (requires ADMIN or OWNER role)
	adminMW := middleware.RequireRoles(valueobject.RoleAdmin, valueobject.RoleOwner)

	// Initialize internal key middleware for service-to-service calls
	var internalMW func(http.Handler) http.Handler
	if cfg.Server.InternalKey != "" {
		internalKeyMiddleware := middleware.NewInternalKeyMiddleware(cfg.Server.InternalKey)
		internalMW = internalKeyMiddleware.Authenticate
		log.Println("Internal key middleware initialized")
	}

	// Wire auto-sync trigger on app selection
	if appHandler != nil {
		if queueSyncService != nil {
			appHandler.SetSyncTrigger(queueSyncService)
			log.Println("Auto-sync trigger wired (queue mode)")
		} else if syncService != nil {
			appHandler.SetSyncTrigger(syncService)
			log.Println("Auto-sync trigger wired (direct mode)")
		}
	}

	// Build router config
	routerCfg := router.Config{
		HealthHandler:                   healthHandler,
		MeHandler:                       meHandler,
		OAuthHandler:                    oauthHandler,
		ManualTokenHandler:              manualTokenHandler,
		IntegrationStatusHandler:        integrationStatusHandler,
		AppHandler:                      appHandler,
		MetricsHandler:                  metricsHandler,
		RevenueHandler:                  revenueHandler,
		FeeHandler:                      feeHandler,
		SyncHandler:                     syncHandler,
		SubscriptionHandler:             subscriptionHandler,
		StoreHealthHandler:              storeHealthHandler,
		UserPreferencesHandler:          userPreferencesHandler,
		NotificationPreferencesHandler:  notificationPreferencesHandler,
		DeviceHandler:                   deviceHandler,
		InsightHandler:                  insightHandler,
		BillingHandler:                  billingHandler,
		ReviewHandler:                   reviewHandler,
		QueueSyncHandler:                queueSyncHandler,
		AdminHandler:                    adminHandler,
		APIKeyHandler:                   apiKeyHandler,
		GraphQLHandler:                  graphqlHandler,
		AuthMW:                          authMW,
		AdminMW:                         adminMW,
		InternalMW:                      internalMW,
	}

	// Wire chat handler if available
	if chatHandler != nil {
		routerCfg.ChatHandler = chatHandler.HandleChat
		routerCfg.ChatModulesHandler = chatHandler.HandleListModules
	}

	r := router.New(routerCfg)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Stop queue worker pools gracefully (before server shutdown)
	if regularWorkerPool != nil {
		regularWorkerPool.Stop()
		log.Println("Regular worker pool stopped")
	}
	if fullSyncWorkerPool != nil {
		fullSyncWorkerPool.Stop()
		log.Println("Full sync worker pool stopped")
	}
	if recoveryService != nil {
		recoveryService.Stop()
		log.Println("Recovery service stopped")
	}

	// Stop schedulers gracefully
	if syncScheduler != nil {
		syncScheduler.Stop()
		log.Println("Sync scheduler stopped")
	}

	if notificationScheduler != nil {
		notificationScheduler.Stop()
		log.Println("Notification scheduler stopped")
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Println("Server stopped")
	return nil
}
