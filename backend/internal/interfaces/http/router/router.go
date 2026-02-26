package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/handler"
)

func New(healthHandler *handler.HealthHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/health", healthHandler.Health)

	return r
}
