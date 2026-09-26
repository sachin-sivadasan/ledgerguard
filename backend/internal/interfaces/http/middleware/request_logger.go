package middleware

import (
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/logging"
)

// RequestLogger attaches a request-scoped logger carrying the chi request_id to the
// request context, so any downstream handler/service that logs via
// logging.FromContext(ctx) gets the correlation id automatically — the field that lets
// you reconstruct one request's whole story in the logs. Use AFTER chi's RequestID.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := logging.FromContext(r.Context()).With(zap.String("request_id", chimw.GetReqID(r.Context())))
		next.ServeHTTP(w, r.WithContext(logging.ContextWithLogger(r.Context(), l)))
	})
}
