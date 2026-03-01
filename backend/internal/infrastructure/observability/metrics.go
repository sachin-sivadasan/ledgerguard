package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the application
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestSize     *prometheus.SummaryVec
	HTTPResponseSize    *prometheus.SummaryVec

	// Sync metrics
	SyncOperationsTotal   *prometheus.CounterVec
	SyncDuration          *prometheus.HistogramVec
	SyncTransactionsCount *prometheus.CounterVec

	// Database metrics
	DBQueryDuration *prometheus.HistogramVec
	DBQueryTotal    *prometheus.CounterVec

	// Cache metrics
	CacheHits   *prometheus.CounterVec
	CacheMisses *prometheus.CounterVec

	// Business metrics
	ActiveSubscriptions *prometheus.GaugeVec
	MRRCents            *prometheus.GaugeVec
	RevenueAtRiskCents  *prometheus.GaugeVec
}

// NewMetrics creates a new Metrics instance with all registered metrics
func NewMetrics(namespace string) *Metrics {
	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path"},
		),
		HTTPRequestSize: promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: namespace,
				Name:      "http_request_size_bytes",
				Help:      "HTTP request size in bytes",
			},
			[]string{"method", "path"},
		),
		HTTPResponseSize: promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: namespace,
				Name:      "http_response_size_bytes",
				Help:      "HTTP response size in bytes",
			},
			[]string{"method", "path"},
		),

		// Sync metrics
		SyncOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "sync_operations_total",
				Help:      "Total number of sync operations",
			},
			[]string{"status"}, // "success", "error"
		),
		SyncDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "sync_duration_seconds",
				Help:      "Sync operation duration in seconds",
				Buckets:   []float64{1, 5, 10, 30, 60, 120, 300, 600},
			},
			[]string{"app_id"},
		),
		SyncTransactionsCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "sync_transactions_total",
				Help:      "Total number of transactions synced",
			},
			[]string{"app_id", "charge_type"},
		),

		// Database metrics
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "db_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5},
			},
			[]string{"operation", "table"},
		),
		DBQueryTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "db_queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),

		// Cache metrics
		CacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_hits_total",
				Help:      "Total number of cache hits",
			},
			[]string{"cache"},
		),
		CacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_misses_total",
				Help:      "Total number of cache misses",
			},
			[]string{"cache"},
		),

		// Business metrics
		ActiveSubscriptions: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "active_subscriptions",
				Help:      "Number of active subscriptions",
			},
			[]string{"app_id", "risk_state"},
		),
		MRRCents: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "mrr_cents",
				Help:      "Monthly Recurring Revenue in cents",
			},
			[]string{"app_id"},
		),
		RevenueAtRiskCents: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "revenue_at_risk_cents",
				Help:      "Revenue at risk in cents",
			},
			[]string{"app_id"},
		),
	}
}

// HTTPMiddleware creates a middleware that records HTTP metrics
func (m *Metrics) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code and size
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		// Get route pattern (use chi route context if available)
		path := r.URL.Path
		if rctx := chi.RouteContext(r.Context()); rctx != nil && rctx.RoutePattern() != "" {
			path = rctx.RoutePattern()
		}

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrapped.statusCode)

		m.HTTPRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
		m.HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
		m.HTTPResponseSize.WithLabelValues(r.Method, path).Observe(float64(wrapped.bytesWritten))

		if r.ContentLength > 0 {
			m.HTTPRequestSize.WithLabelValues(r.Method, path).Observe(float64(r.ContentLength))
		}
	})
}

// RecordSyncStart records the start of a sync operation
func (m *Metrics) RecordSyncStart(appID string) func(status string, transactionCounts map[string]int) {
	start := time.Now()
	return func(status string, transactionCounts map[string]int) {
		duration := time.Since(start).Seconds()
		m.SyncOperationsTotal.WithLabelValues(status).Inc()
		m.SyncDuration.WithLabelValues(appID).Observe(duration)

		for chargeType, count := range transactionCounts {
			m.SyncTransactionsCount.WithLabelValues(appID, chargeType).Add(float64(count))
		}
	}
}

// RecordDBQuery records a database query metric
func (m *Metrics) RecordDBQuery(operation, table string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}
	m.DBQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
	m.DBQueryTotal.WithLabelValues(operation, table, status).Inc()
}

// RecordCacheHit records a cache hit
func (m *Metrics) RecordCacheHit(cacheName string) {
	m.CacheHits.WithLabelValues(cacheName).Inc()
}

// RecordCacheMiss records a cache miss
func (m *Metrics) RecordCacheMiss(cacheName string) {
	m.CacheMisses.WithLabelValues(cacheName).Inc()
}

// UpdateBusinessMetrics updates the business metrics gauges
func (m *Metrics) UpdateBusinessMetrics(appID string, mrrCents, revenueAtRiskCents int64, subscriptionsByState map[string]int) {
	m.MRRCents.WithLabelValues(appID).Set(float64(mrrCents))
	m.RevenueAtRiskCents.WithLabelValues(appID).Set(float64(revenueAtRiskCents))

	for state, count := range subscriptionsByState {
		m.ActiveSubscriptions.WithLabelValues(appID, state).Set(float64(count))
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// Global metrics instance (can be overridden in tests)
var DefaultMetrics = NewMetrics("ledgerguard")
