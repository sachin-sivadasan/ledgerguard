package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
)

var ErrUserNotFound = errors.New("user not found")

type contextKey string

const userContextKey contextKey = "user"

type AuthMiddleware struct {
	tokenVerifier service.AuthTokenVerifier
	userRepo      repository.UserRepository
}

func NewAuthMiddleware(tokenVerifier service.AuthTokenVerifier, userRepo repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		tokenVerifier: tokenVerifier,
		userRepo:      userRepo,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := extractBearerToken(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "missing or invalid authorization header")
			return
		}

		claims, err := m.tokenVerifier.VerifyIDToken(r.Context(), token)
		if err != nil {
			log.Printf("Token verification failed: %v", err)
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		user, err := m.userRepo.FindByFirebaseUID(r.Context(), claims.UID)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				user = entity.NewUser(claims.UID, claims.Email)
				if err := m.userRepo.Create(r.Context(), user); err != nil {
					writeError(w, http.StatusInternalServerError, "failed to create user")
					return
				}
			} else {
				writeError(w, http.StatusInternalServerError, "failed to lookup user")
				return
			}
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserFromContext(ctx context.Context) *entity.User {
	user, ok := ctx.Value(userContextKey).(*entity.User)
	if !ok {
		return nil
	}
	return user
}

func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid authorization format")
	}

	return parts[1], nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    http.StatusText(status),
			"message": message,
		},
	})
}

// InternalKeyMiddleware validates internal API key for service-to-service calls
type InternalKeyMiddleware struct {
	internalKey string
}

// NewInternalKeyMiddleware creates a middleware that validates X-Internal-Key header
func NewInternalKeyMiddleware(internalKey string) *InternalKeyMiddleware {
	return &InternalKeyMiddleware{
		internalKey: internalKey,
	}
}

// Authenticate validates the internal key from X-Internal-Key header
func (m *InternalKeyMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.internalKey == "" {
			writeError(w, http.StatusServiceUnavailable, "internal key not configured")
			return
		}

		key := r.Header.Get("X-Internal-Key")
		if key == "" {
			writeError(w, http.StatusUnauthorized, "missing X-Internal-Key header")
			return
		}

		if key != m.internalKey {
			writeError(w, http.StatusUnauthorized, "invalid internal key")
			return
		}

		next.ServeHTTP(w, r)
	})
}
