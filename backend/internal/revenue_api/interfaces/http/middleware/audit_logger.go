package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/domain/repository"
)

// AuditLogger is middleware that logs API requests
type AuditLogger struct {
	repo repository.AuditLogRepository
}

// NewAuditLogger creates a new AuditLogger middleware
func NewAuditLogger(repo repository.AuditLogRepository) *AuditLogger {
	return &AuditLogger{repo: repo}
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Middleware returns the HTTP middleware handler
func (m *AuditLogger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get the API key from context
		apiKey := APIKeyFromContext(r.Context())
		if apiKey == nil {
			// No API key, skip logging
			next.ServeHTTP(w, r)
			return
		}

		// Wrap the response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Serve the request
		next.ServeHTTP(wrapped, r)

		// Calculate response time
		duration := time.Since(start)

		// Build request params (sanitized - no sensitive data)
		params := m.sanitizeParams(r)

		// Get client IP
		clientIP := m.getClientIP(r)

		// Create audit log entry
		auditLog := entity.NewAuditLog(
			apiKey.ID,
			r.URL.Path,
			r.Method,
			params,
			wrapped.statusCode,
			int(duration.Milliseconds()),
			clientIP,
			r.UserAgent(),
		)

		// Log asynchronously to not block the response
		m.repo.CreateAsync(auditLog)
	})
}

// sanitizeParams extracts and sanitizes request parameters
func (m *AuditLogger) sanitizeParams(r *http.Request) map[string]interface{} {
	params := make(map[string]interface{})

	// Add query parameters (exclude sensitive ones)
	for key, values := range r.URL.Query() {
		if !m.isSensitiveParam(key) && len(values) > 0 {
			if len(values) == 1 {
				params[key] = values[0]
			} else {
				params[key] = values
			}
		}
	}

	// For POST/PUT requests, try to capture non-sensitive body info
	if r.Method == "POST" || r.Method == "PUT" {
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			// Read body
			body, err := io.ReadAll(r.Body)
			if err == nil && len(body) > 0 {
				// Restore body for downstream handlers
				r.Body = io.NopCloser(bytes.NewBuffer(body))

				// Log body length, not content (for privacy)
				params["body_length"] = len(body)
			}
		}
	}

	return params
}

// isSensitiveParam checks if a parameter name is sensitive
func (m *AuditLogger) isSensitiveParam(name string) bool {
	sensitive := []string{
		"password", "secret", "token", "api_key", "apikey",
		"authorization", "auth", "credential", "key",
	}

	nameLower := strings.ToLower(name)
	for _, s := range sensitive {
		if strings.Contains(nameLower, s) {
			return true
		}
	}
	return false
}

// getClientIP extracts the client IP from the request
func (m *AuditLogger) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxied requests)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the list
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	// Strip port if present
	addr := r.RemoteAddr
	if colonIndex := strings.LastIndex(addr, ":"); colonIndex != -1 {
		// Check if this looks like IPv6
		if strings.Count(addr, ":") > 1 {
			// IPv6 - might be in [::1]:port format
			if addr[0] == '[' {
				bracketIndex := strings.Index(addr, "]")
				if bracketIndex != -1 {
					return addr[1:bracketIndex]
				}
			}
			return addr
		}
		return addr[:colonIndex]
	}

	return addr
}
