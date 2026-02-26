package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

type mockTokenVerifier struct {
	claims *service.TokenClaims
	err    error
}

func (m *mockTokenVerifier) VerifyIDToken(ctx context.Context, idToken string) (*service.TokenClaims, error) {
	return m.claims, m.err
}

type mockUserRepository struct {
	user      *entity.User
	findErr   error
	createErr error
	created   *entity.User
}

func (m *mockUserRepository) FindByFirebaseUID(ctx context.Context, firebaseUID string) (*entity.User, error) {
	return m.user, m.findErr
}

func (m *mockUserRepository) Create(ctx context.Context, user *entity.User) error {
	m.created = user
	return m.createErr
}

func TestAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	verifier := &mockTokenVerifier{}
	userRepo := &mockUserRepository{}
	middleware := NewAuthMiddleware(verifier, userRepo)

	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthMiddleware_InvalidAuthorizationFormat(t *testing.T) {
	verifier := &mockTokenVerifier{}
	userRepo := &mockUserRepository{}
	middleware := NewAuthMiddleware(verifier, userRepo)

	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	verifier := &mockTokenVerifier{
		err: errors.New("invalid token"),
	}
	userRepo := &mockUserRepository{}
	middleware := NewAuthMiddleware(verifier, userRepo)

	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthMiddleware_ExistingUser(t *testing.T) {
	existingUser := &entity.User{
		FirebaseUID: "firebase-123",
		Email:       "test@example.com",
		Role:        valueobject.RoleOwner,
	}

	verifier := &mockTokenVerifier{
		claims: &service.TokenClaims{
			UID:   "firebase-123",
			Email: "test@example.com",
		},
	}
	userRepo := &mockUserRepository{
		user: existingUser,
	}
	middleware := NewAuthMiddleware(verifier, userRepo)

	var ctxUser *entity.User
	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxUser = UserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if ctxUser == nil {
		t.Fatal("expected user in context")
	}

	if ctxUser.FirebaseUID != "firebase-123" {
		t.Errorf("expected FirebaseUID 'firebase-123', got '%s'", ctxUser.FirebaseUID)
	}
}

func TestAuthMiddleware_NewUser_AutoCreate(t *testing.T) {
	verifier := &mockTokenVerifier{
		claims: &service.TokenClaims{
			UID:   "new-firebase-456",
			Email: "newuser@example.com",
		},
	}
	userRepo := &mockUserRepository{
		findErr: ErrUserNotFound,
	}
	middleware := NewAuthMiddleware(verifier, userRepo)

	var ctxUser *entity.User
	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxUser = UserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if userRepo.created == nil {
		t.Fatal("expected user to be created")
	}

	if userRepo.created.FirebaseUID != "new-firebase-456" {
		t.Errorf("expected FirebaseUID 'new-firebase-456', got '%s'", userRepo.created.FirebaseUID)
	}

	if userRepo.created.Email != "newuser@example.com" {
		t.Errorf("expected Email 'newuser@example.com', got '%s'", userRepo.created.Email)
	}

	if userRepo.created.Role != valueobject.RoleOwner {
		t.Errorf("expected Role OWNER, got '%s'", userRepo.created.Role)
	}

	if ctxUser == nil {
		t.Fatal("expected user in context")
	}
}

func TestAuthMiddleware_CreateUserError(t *testing.T) {
	verifier := &mockTokenVerifier{
		claims: &service.TokenClaims{
			UID:   "new-firebase-789",
			Email: "error@example.com",
		},
	}
	userRepo := &mockUserRepository{
		findErr:   ErrUserNotFound,
		createErr: errors.New("database error"),
	}
	middleware := NewAuthMiddleware(verifier, userRepo)

	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}
