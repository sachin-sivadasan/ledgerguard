package observability

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

// HealthStatus represents the health check response
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version,omitempty"`
	Checks    map[string]string `json:"checks,omitempty"`
	System    *SystemInfo       `json:"system,omitempty"`
}

// SystemInfo contains system-level information
type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutines"`
	NumCPU       int    `json:"num_cpus"`
	MemAllocMB   uint64 `json:"mem_alloc_mb"`
	MemTotalMB   uint64 `json:"mem_total_mb"`
}

// HealthChecker is a function that checks a component's health
type HealthChecker func() error

// HealthHandler handles health check endpoints
type HealthHandler struct {
	version  string
	checkers map[string]HealthChecker
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		version:  version,
		checkers: make(map[string]HealthChecker),
	}
}

// RegisterChecker registers a health checker for a component
func (h *HealthHandler) RegisterChecker(name string, checker HealthChecker) {
	h.checkers[name] = checker
}

// Liveness returns a simple liveness check handler
// Used by Kubernetes to determine if the container should be restarted
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// Readiness returns a readiness check handler
// Used by Kubernetes to determine if the container should receive traffic
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	status := HealthStatus{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   h.version,
		Checks:    make(map[string]string),
	}

	allHealthy := true
	for name, checker := range h.checkers {
		if err := checker(); err != nil {
			status.Checks[name] = err.Error()
			allHealthy = false
		} else {
			status.Checks[name] = "healthy"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if allHealthy {
		status.Status = "ok"
		w.WriteHeader(http.StatusOK)
	} else {
		status.Status = "degraded"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(status)
}

// Health returns a detailed health check handler
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	status := HealthStatus{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   h.version,
		Checks:    make(map[string]string),
		System: &SystemInfo{
			GoVersion:    runtime.Version(),
			NumGoroutine: runtime.NumGoroutine(),
			NumCPU:       runtime.NumCPU(),
			MemAllocMB:   memStats.Alloc / 1024 / 1024,
			MemTotalMB:   memStats.TotalAlloc / 1024 / 1024,
		},
	}

	allHealthy := true
	for name, checker := range h.checkers {
		if err := checker(); err != nil {
			status.Checks[name] = err.Error()
			allHealthy = false
		} else {
			status.Checks[name] = "healthy"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if allHealthy {
		status.Status = "ok"
		w.WriteHeader(http.StatusOK)
	} else {
		status.Status = "degraded"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(status)
}
