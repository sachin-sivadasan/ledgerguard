package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpdelivery "github.com/sachin-sivadasan/ledgerguard/internal/delivery/http"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/config"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/database"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}

func run() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var db *database.DB
	db, err = database.NewPostgresDB(ctx, cfg.Database.DSN())
	if err != nil {
		log.Printf("WARNING: failed to connect to database: %v", err)
		log.Printf("Server will start without database connection")
		db = nil
	} else {
		defer db.Close()
		log.Println("Connected to PostgreSQL")
	}

	healthHandler := httpdelivery.NewHealthHandler(db)
	r := httpdelivery.NewRouter(healthHandler)

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

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Println("Server stopped")
	return nil
}
