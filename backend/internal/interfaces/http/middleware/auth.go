package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
)

var ErrUserNotFound = errors.New("user not found")

type contextKey string

const userContextKey contextKey = "user"

// OrgProvisioner creates a default organization (with the creator as OWNER) for a
// newly-created user, so every user has an org and org-scoped routes work. Satisfied
// by application/service.OrgService.
type OrgProvisioner interface {
	CreateOrganization(ctx context.Context, name string, creatorID uuid.UUID) (*entity.Organization, error)
}

type AuthMiddleware struct {
	tokenVerifier  service.AuthTokenVerifier
	userRepo       repository.UserRepository
	tracker        service.EventTracker
	orgProvisioner OrgProvisioner
}

func NewAuthMiddleware(tokenVerifier service.AuthTokenVerifier, userRepo repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		tokenVerifier: tokenVerifier,
		userRepo:      userRepo,
	}
}

// SetTracker sets the event tracker for lifecycle events.
func (m *AuthMiddleware) SetTracker(t service.EventTracker) {
	m.tracker = t
}

// SetOrgProvisioner sets the provisioner used to create a default org for new users.
func (m *AuthMiddleware) SetOrgProvisioner(p OrgProvisioner) {
	m.orgProvisioner = p
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
					log.Printf("auth: failed to create user for UID %s: %v", claims.UID, err)
					writeError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
					return
				}
				if m.tracker != nil {
					m.tracker.Track(r.Context(), user.ID.String(), "user_signup", service.EventProperties{
						"email": user.Email,
						"role":  string(user.Role),
					})
					m.tracker.SetUserProperties(r.Context(), user.ID.String(), service.EventProperties{
						"$email":    user.Email,
						"role":      string(user.Role),
						"plan_tier": string(user.PlanTier),
					})
				}
				// Provision a default org + OWNER membership so the new user can use
				// org-scoped routes immediately. A failure here leaves the user without
				// an org (recoverable via the backfill migration), so log loudly rather
				// than block login.
				if m.orgProvisioner != nil {
					if _, perr := m.orgProvisioner.CreateOrganization(r.Context(), defaultOrgName(user.Email), user.ID); perr != nil {
						// This runs once (new-user branch only), so a failure leaves the user
						// org-less until the backfill (000042) is re-applied for them. Log with
						// identity so it's actionable without a table scan.
						log.Printf("auth: failed to provision default org for user id=%s email=%s uid=%s: %v", user.ID, user.Email, claims.UID, perr)
					}
				}
			} else {
				log.Printf("auth: DB error looking up user for UID %s: %v", claims.UID, err)
				writeError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
				return
			}
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// defaultOrgName builds a default org name from the email local-part (text before
// "@"), e.g. "alice@shop.com" → "alice's Organization"; falls back to
// "My Organization" when there's no local-part. Kept in sync with the SQL backfill
// (migration 000042), which derives the same name via split_part(email,'@',1).
func defaultOrgName(email string) string {
	if i := strings.IndexByte(email, '@'); i > 0 {
		return email[:i] + "'s Organization"
	}
	return "My Organization"
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
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
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
