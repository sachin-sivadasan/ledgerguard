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

func TestOAuthHandler_StartOAuth(t *testing.T) {
	oauthService := &mockOAuthService{authURL: "https://partners.shopify.com/authorize"}
	handler := NewOAuthHandler(oauthService, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/oauth", nil)
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
}

func TestOAuthHandler_Callback_MissingCode(t *testing.T) {
	handler := NewOAuthHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/callback", nil)
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestOAuthHandler_Callback_NoUser(t *testing.T) {
	oauthService := &mockOAuthService{token: "test-token"}
	handler := NewOAuthHandler(oauthService, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/callback?code=test-code&state=test-state", nil)
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestOAuthHandler_Callback_Success(t *testing.T) {
	oauthService := &mockOAuthService{token: "test-access-token"}
	encryptor := &mockEncryptor{encrypted: []byte("encrypted-token")}
	repo := &mockPartnerAccountRepo{}

	handler := NewOAuthHandler(oauthService, encryptor, repo)

	user := &entity.User{
		ID:   uuid.New(),
		Role: valueobject.RoleOwner,
	}

	req := httptest.NewRequest(http.MethodGet, "/callback?code=test-code&state=test-state", nil)
	req = req.WithContext(contextWithUser(req.Context(), user))
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if repo.account == nil {
		t.Fatal("expected partner account to be created")
	}

	if repo.account.UserID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, repo.account.UserID)
	}

	if repo.account.IntegrationType != valueobject.IntegrationTypeOAuth {
		t.Errorf("expected integration type OAUTH, got %s", repo.account.IntegrationType)
	}
}
