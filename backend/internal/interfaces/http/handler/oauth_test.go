package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

type mockOAuthService struct {
	authURL string
	token   string
	err     error
}

func (m *mockOAuthService) GenerateAuthURL(state string) string {
	return m.authURL + "?state=" + state
}

func (m *mockOAuthService) ExchangeCodeForToken(ctx context.Context, code string) (string, error) {
	return m.token, m.err
}

type mockEncryptor struct {
	encrypted []byte
	err       error
}

func (m *mockEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	return m.encrypted, m.err
}

func (m *mockEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	return nil, nil
}

type mockPartnerAccountRepo struct {
	account   *entity.PartnerAccount
	createErr error
}

func (m *mockPartnerAccountRepo) Create(ctx context.Context, account *entity.PartnerAccount) error {
	m.account = account
	return m.createErr
}

func (m *mockPartnerAccountRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.PartnerAccount, error) {
	return nil, nil
}

func (m *mockPartnerAccountRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.PartnerAccount, error) {
	return nil, nil
}

func (m *mockPartnerAccountRepo) FindByPartnerID(ctx context.Context, partnerID string) (*entity.PartnerAccount, error) {
	return nil, nil
}

func (m *mockPartnerAccountRepo) Update(ctx context.Context, account *entity.PartnerAccount) error {
	return nil
}

func (m *mockPartnerAccountRepo) Delete(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (m *mockPartnerAccountRepo) GetAllIDs(ctx context.Context) ([]uuid.UUID, error) {
	return []uuid.UUID{}, nil
}

type mockUserRepo struct {
	user *entity.User
	err  error
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return m.user, m.err
}

func (m *mockUserRepo) FindByFirebaseUID(ctx context.Context, firebaseUID string) (*entity.User, error) {
	return m.user, m.err
}

func (m *mockUserRepo) Create(ctx context.Context, user *entity.User) error {
	return nil
}

type mockStateStore struct {
	storedState  string
	storedUserID uuid.UUID
	validState   bool
	returnUserID uuid.UUID
}

func (m *mockStateStore) Store(state string, userID uuid.UUID) {
	m.storedState = state
	m.storedUserID = userID
}

func (m *mockStateStore) Validate(state string) (uuid.UUID, bool) {
	if m.validState && state == m.storedState {
		return m.returnUserID, true
	}
	return uuid.Nil, false
}

func TestOAuthHandler_StartOAuth_NoUser(t *testing.T) {
	oauthService := &mockOAuthService{authURL: "https://partners.shopify.com/authorize"}
	stateStore := &mockStateStore{}
	handler := NewOAuthHandler(oauthService, nil, nil, nil, stateStore)

	req := httptest.NewRequest(http.MethodGet, "/oauth", nil)
	rec := httptest.NewRecorder()

	handler.StartOAuth(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestOAuthHandler_StartOAuth_Success(t *testing.T) {
	oauthService := &mockOAuthService{authURL: "https://partners.shopify.com/authorize"}
	stateStore := &mockStateStore{}

	handler := NewOAuthHandler(oauthService, nil, nil, nil, stateStore)

	user := &entity.User{
		ID:   uuid.New(),
		Role: valueobject.RoleOwner,
	}

	req := httptest.NewRequest(http.MethodGet, "/oauth", nil)
	req = req.WithContext(contextWithUser(req.Context(), user))
	rec := httptest.NewRecorder()

	handler.StartOAuth(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["url"] == "" {
		t.Error("expected url in response")
	}

	if resp["state"] == "" {
		t.Error("expected state in response")
	}

	// Verify state was stored with user ID
	if stateStore.storedState == "" {
		t.Error("expected state to be stored")
	}

	if stateStore.storedUserID != user.ID {
		t.Errorf("expected stored user ID %s, got %s", user.ID, stateStore.storedUserID)
	}
}

func TestOAuthHandler_Callback_MissingCode(t *testing.T) {
	handler := NewOAuthHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/callback", nil)
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestOAuthHandler_Callback_MissingState(t *testing.T) {
	handler := NewOAuthHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/callback?code=test-code", nil)
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestOAuthHandler_Callback_InvalidState(t *testing.T) {
	stateStore := &mockStateStore{validState: false}
	handler := NewOAuthHandler(nil, nil, nil, nil, stateStore)

	req := httptest.NewRequest(http.MethodGet, "/callback?code=test-code&state=invalid-state", nil)
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestOAuthHandler_Callback_Success(t *testing.T) {
	user := &entity.User{
		ID:   uuid.New(),
		Role: valueobject.RoleOwner,
	}

	oauthService := &mockOAuthService{token: "test-access-token"}
	encryptor := &mockEncryptor{encrypted: []byte("encrypted-token")}
	partnerRepo := &mockPartnerAccountRepo{}
	userRepo := &mockUserRepo{user: user}
	stateStore := &mockStateStore{
		storedState:  "valid-state",
		validState:   true,
		returnUserID: user.ID,
	}

	handler := NewOAuthHandler(oauthService, encryptor, partnerRepo, userRepo, stateStore)

	req := httptest.NewRequest(http.MethodGet, "/callback?code=test-code&state=valid-state", nil)
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d; body: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	if partnerRepo.account == nil {
		t.Fatal("expected partner account to be created")
	}

	if partnerRepo.account.UserID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, partnerRepo.account.UserID)
	}

	if partnerRepo.account.IntegrationType != valueobject.IntegrationTypeOAuth {
		t.Errorf("expected integration type OAUTH, got %s", partnerRepo.account.IntegrationType)
	}
}
