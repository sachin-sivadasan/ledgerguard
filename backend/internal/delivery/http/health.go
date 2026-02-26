package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type DBChecker interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	db DBChecker
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func NewHealthHandler(db DBChecker) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:   "ok",
		Database: "not configured",
	}
	statusCode := http.StatusOK

	if h.db != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := h.db.Ping(ctx); err != nil {
			resp.Status = "degraded"
			resp.Database = "disconnected"
			statusCode = http.StatusServiceUnavailable
		} else {
			resp.Database = "connected"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}
