package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
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

func (m *mockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return m.user, m.findErr
}

func (m *mockUserRepository) FindByFirebaseUID(ctx context.Context, firebaseUID string) (*entity.User, error) {
	return m.user, m.findErr
}

func (m *mockUserRepository) Create(ctx context.Context, user *entity.User) error {
	m.created = user
	return m.createErr
}

func (m *mockUserRepository) Update(ctx context.Context, user *entity.User) error {
	return nil
}

type mockOrgProvisioner struct {
	called     bool
	gotName    string
	gotCreator uuid.UUID
	err        error
}

func (m *mockOrgProvisioner) CreateOrganization(ctx context.Context, name string, creatorID uuid.UUID) (*entity.Organization, error) {
	m.called = true
	m.gotName = name
	m.gotCreator = creatorID
	if m.err != nil {
		return nil, m.err
	}
	return &entity.Organization{ID: uuid.New(), Name: name}, nil
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

func TestAuthMiddleware_NewUser_ProvisionsOrg(t *testing.T) {
	verifier := &mockTokenVerifier{
		claims: &service.TokenClaims{UID: "new-uid", Email: "newuser@example.com"},
	}
	userRepo := &mockUserRepository{findErr: ErrUserNotFound}
	provisioner := &mockOrgProvisioner{}
	mw := NewAuthMiddleware(verifier, userRepo)
	mw.SetOrgProvisioner(provisioner)

	handler := mw.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !provisioner.called {
		t.Fatal("expected org to be provisioned for new user")
	}
	if provisioner.gotName != "newuser's Organization" {
		t.Errorf("org name: expected \"newuser's Organization\", got %q", provisioner.gotName)
	}
	if userRepo.created == nil || provisioner.gotCreator != userRepo.created.ID {
		t.Errorf("org creator: expected new user's ID %v, got %v", userRepo.created.ID, provisioner.gotCreator)
	}
}

func TestAuthMiddleware_ExistingUser_NoProvision(t *testing.T) {
	verifier := &mockTokenVerifier{
		claims: &service.TokenClaims{UID: "existing-uid", Email: "e@example.com"},
	}
	userRepo := &mockUserRepository{user: &entity.User{ID: uuid.New(), FirebaseUID: "existing-uid"}}
	provisioner := &mockOrgProvisioner{}
	mw := NewAuthMiddleware(verifier, userRepo)
	mw.SetOrgProvisioner(provisioner)

	handler := mw.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if provisioner.called {
		t.Error("expected NO org provisioning for an existing user")
	}
}

func TestAuthMiddleware_ProvisionError_DoesNotBlockLogin(t *testing.T) {
	verifier := &mockTokenVerifier{
		claims: &service.TokenClaims{UID: "new-uid", Email: "newuser@example.com"},
	}
	userRepo := &mockUserRepository{findErr: ErrUserNotFound}
	provisioner := &mockOrgProvisioner{err: errors.New("org create failed")}
	mw := NewAuthMiddleware(verifier, userRepo)
	mw.SetOrgProvisioner(provisioner)

	var ctxUser *entity.User
	handler := mw.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxUser = UserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// A provisioning failure must not block login — the user is still created and served.
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 despite org-provision failure, got %d", rec.Code)
	}
	if ctxUser == nil {
		t.Error("expected user in context despite org-provision failure")
	}
}

func TestDefaultOrgName(t *testing.T) {
	cases := map[string]string{
		"alice@shop.com": "alice's Organization",
		"bob@x.io":       "bob's Organization",
		"":               "My Organization",
		"@nolocal.com":   "My Organization",
		"noatsign":       "My Organization",
	}
	for email, want := range cases {
		if got := defaultOrgName(email); got != want {
			t.Errorf("defaultOrgName(%q): expected %q, got %q", email, want, got)
		}
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

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
}

func TestAuthMiddleware_DBErrorOnLookup(t *testing.T) {
	verifier := &mockTokenVerifier{
		claims: &service.TokenClaims{UID: "existing-uid", Email: "user@test.com"},
	}
	userRepo := &mockUserRepository{
		findErr: errors.New("connection refused"),
	}
	mw := NewAuthMiddleware(verifier, userRepo)

	handler := mw.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
}
