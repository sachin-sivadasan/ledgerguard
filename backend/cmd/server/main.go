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

	"github.com/sachin-sivadasan/ledgerguard/internal/application/scheduler"
	appservice "github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	domainservice "github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/cache"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/config"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/external"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/persistence"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/handler"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/router"
	"github.com/sachin-sivadasan/ledgerguard/pkg/crypto"
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

	// Initialize repositories
	var userRepo *persistence.PostgresUserRepository
	var partnerRepo *persistence.PostgresPartnerAccountRepository
	var appRepo *persistence.PostgresAppRepository
	var txRepo *persistence.PostgresTransactionRepository
	var subscriptionRepo *persistence.PostgresSubscriptionRepository

	if db != nil {
		userRepo = persistence.NewPostgresUserRepository(db.Pool)
		partnerRepo = persistence.NewPostgresPartnerAccountRepository(db.Pool)
		appRepo = persistence.NewPostgresAppRepository(db.Pool)
		txRepo = persistence.NewPostgresTransactionRepository(db.Pool)
		subscriptionRepo = persistence.NewPostgresSubscriptionRepository(db.Pool)
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

	var appHandler *handler.AppHandler
	if partnerRepo != nil && appRepo != nil && encryptor != nil {
		// Note: PartnerClient would need to be initialized for full functionality
		appHandler = handler.NewAppHandler(nil, partnerRepo, appRepo, encryptor)
		log.Println("App handler initialized")
	}

	// Initialize sync service and handler
	var syncService *appservice.SyncService
	var syncHandler *handler.SyncHandler
	var syncScheduler *scheduler.SyncScheduler

	if txRepo != nil && appRepo != nil && partnerRepo != nil && encryptor != nil && subscriptionRepo != nil {
		// Initialize ledger service for rebuilding after sync
		ledgerService := domainservice.NewLedgerService(txRepo, subscriptionRepo)

		// Initialize sync service (no external fetcher in dev - would need Shopify Partner client)
		syncService = appservice.NewSyncService(
			nil, // TransactionFetcher - would be ShopifyPartnerClient in production
			txRepo,
			appRepo,
			partnerRepo,
			encryptor,
			ledgerService,
		)

		syncHandler = handler.NewSyncHandler(syncService, partnerRepo, appRepo)
		log.Println("Sync handler initialized")

		// Initialize and start scheduler
		syncScheduler = scheduler.NewSyncScheduler(syncService, partnerRepo)
		syncScheduler.Start(ctx)
		log.Println("Sync scheduler started (12-hour interval)")
	}

	// Initialize auth middleware
	var authMW func(http.Handler) http.Handler
	if firebaseAuth != nil && userRepo != nil {
		authMiddleware := middleware.NewAuthMiddleware(firebaseAuth, userRepo)
		authMW = authMiddleware.Authenticate
		log.Println("Auth middleware initialized")
	}

	// Initialize admin middleware (requires ADMIN or OWNER role)
	adminMW := middleware.RequireRoles(valueobject.RoleAdmin, valueobject.RoleOwner)

	// Build router config
	routerCfg := router.Config{
		HealthHandler:      healthHandler,
		OAuthHandler:       oauthHandler,
		ManualTokenHandler: manualTokenHandler,
		AppHandler:         appHandler,
		SyncHandler:        syncHandler,
		AuthMW:             authMW,
		AdminMW:            adminMW,
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

	// Stop scheduler gracefully
	if syncScheduler != nil {
		syncScheduler.Stop()
		log.Println("Sync scheduler stopped")
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Println("Server stopped")
	return nil
}
