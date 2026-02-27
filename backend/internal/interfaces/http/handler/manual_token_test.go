package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// Mock implementations for testing
type mockPartnerRepoForManual struct {
	account     *entity.PartnerAccount
	createErr   error
	findErr     error
	deleteErr   error
	createCalls int
	deleteCalls int
}

func (m *mockPartnerRepoForManual) Create(ctx context.Context, account *entity.PartnerAccount) error {
	m.createCalls++
	if m.createErr != nil {
		return m.createErr
	}
	m.account = account
	return nil
}

func (m *mockPartnerRepoForManual) FindByID(ctx context.Context, id uuid.UUID) (*entity.PartnerAccount, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.account, nil
}

func (m *mockPartnerRepoForManual) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.PartnerAccount, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.account, nil
}

func (m *mockPartnerRepoForManual) FindByPartnerID(ctx context.Context, partnerID string) (*entity.PartnerAccount, error) {
	return nil, nil
}

func (m *mockPartnerRepoForManual) Update(ctx context.Context, account *entity.PartnerAccount) error {
	return nil
}

func (m *mockPartnerRepoForManual) Delete(ctx context.Context, userID uuid.UUID) error {
	m.deleteCalls++
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.account = nil
	return nil
}

func (m *mockPartnerRepoForManual) GetAllIDs(ctx context.Context) ([]uuid.UUID, error) {
	if m.account != nil {
		return []uuid.UUID{m.account.ID}, nil
	}
	return []uuid.UUID{}, nil
}

type mockEncryptorForManual struct {
	encrypted []byte
	decrypted []byte
	encErr    error
	decErr    error
}

func (m *mockEncryptorForManual) Encrypt(plaintext []byte) ([]byte, error) {
	if m.encErr != nil {
		return nil, m.encErr
	}
	if m.encrypted != nil {
		return m.encrypted, nil
	}
	return []byte("encrypted:" + string(plaintext)), nil
}

func (m *mockEncryptorForManual) Decrypt(ciphertext []byte) ([]byte, error) {
	if m.decErr != nil {
		return nil, m.decErr
	}
	if m.decrypted != nil {
		return m.decrypted, nil
	}
	return []byte("shppa_1234567890abcdef"), nil
}

func TestManualTokenHandler_AddToken_Success(t *testing.T) {
	repo := &mockPartnerRepoForManual{
		findErr: errors.New("not found"), // No existing account
	}
	encryptor := &mockEncryptorForManual{}
	handler := NewManualTokenHandler(encryptor, repo)

	body := `{"token": "shppa_test_token_12345", "partner_id": "12345"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/shopify/token", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	user := &entity.User{
		ID:   uuid.New(),
		Role: valueobject.RoleAdmin,
	}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.AddToken(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if repo.createCalls != 1 {
		t.Errorf("expected 1 create call, got %d", repo.createCalls)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp["message"] != "Manual token added successfully" {
		t.Errorf("unexpected message: %v", resp["message"])
	}

	if resp["masked_token"] != "***...2345" {
		t.Errorf("expected masked token '***...2345', got %v", resp["masked_token"])
	}
}

func TestManualTokenHandler_AddToken_MissingToken(t *testing.T) {
	handler := NewManualTokenHandler(&mockEncryptorForManual{}, &mockPartnerRepoForManual{})

	body := `{"partner_id": "12345"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/shopify/token", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	user := &entity.User{ID: uuid.New(), Role: valueobject.RoleAdmin}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.AddToken(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestManualTokenHandler_AddToken_MissingPartnerID(t *testing.T) {
	handler := NewManualTokenHandler(&mockEncryptorForManual{}, &mockPartnerRepoForManual{})

	body := `{"token": "shppa_test_token"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/shopify/token", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	user := &entity.User{ID: uuid.New(), Role: valueobject.RoleAdmin}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.AddToken(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestManualTokenHandler_AddToken_NoUser(t *testing.T) {
	handler := NewManualTokenHandler(&mockEncryptorForManual{}, &mockPartnerRepoForManual{})

	body := `{"token": "shppa_test_token", "partner_id": "12345"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/shopify/token", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.AddToken(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestManualTokenHandler_AddToken_AccountExists_Updates(t *testing.T) {
	existingAccount := &entity.PartnerAccount{
		ID:              uuid.New(),
		UserID:          uuid.New(),
		PartnerID:       "existing_partner",
		IntegrationType: valueobject.IntegrationTypeManual,
		CreatedAt:       time.Now(),
	}

	repo := &mockPartnerRepoForManual{
		account: existingAccount,
	}
	encryptor := &mockEncryptorForManual{}
	handler := NewManualTokenHandler(encryptor, repo)

	body := `{"token": "shppa_new_token_12345", "partner_id": "12345"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/shopify/token", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	user := &entity.User{ID: existingAccount.UserID, Role: valueobject.RoleAdmin}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.AddToken(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestManualTokenHandler_GetToken_Success(t *testing.T) {
	existingAccount := &entity.PartnerAccount{
		ID:                   uuid.New(),
		UserID:               uuid.New(),
		PartnerID:            "12345",
		IntegrationType:      valueobject.IntegrationTypeManual,
		EncryptedAccessToken: []byte("encrypted_token"),
		CreatedAt:            time.Now(),
	}

	repo := &mockPartnerRepoForManual{account: existingAccount}
	encryptor := &mockEncryptorForManual{
		decrypted: []byte("shppa_1234567890abcdef"),
	}
	handler := NewManualTokenHandler(encryptor, repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/shopify/token", nil)

	user := &entity.User{ID: existingAccount.UserID, Role: valueobject.RoleAdmin}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.GetToken(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp["masked_token"] != "***...cdef" {
		t.Errorf("expected masked token '***...cdef', got %v", resp["masked_token"])
	}

	if resp["partner_id"] != "12345" {
		t.Errorf("expected partner_id '12345', got %v", resp["partner_id"])
	}

	if resp["integration_type"] != "MANUAL" {
		t.Errorf("expected integration_type 'MANUAL', got %v", resp["integration_type"])
	}
}

func TestManualTokenHandler_GetToken_NotFound(t *testing.T) {
	repo := &mockPartnerRepoForManual{findErr: errors.New("not found")}
	handler := NewManualTokenHandler(&mockEncryptorForManual{}, repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/shopify/token", nil)

	user := &entity.User{ID: uuid.New(), Role: valueobject.RoleAdmin}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.GetToken(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestManualTokenHandler_GetToken_NoUser(t *testing.T) {
	handler := NewManualTokenHandler(&mockEncryptorForManual{}, &mockPartnerRepoForManual{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/shopify/token", nil)

	rec := httptest.NewRecorder()
	handler.GetToken(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestManualTokenHandler_RevokeToken_Success(t *testing.T) {
	existingAccount := &entity.PartnerAccount{
		ID:     uuid.New(),
		UserID: uuid.New(),
	}

	repo := &mockPartnerRepoForManual{account: existingAccount}
	handler := NewManualTokenHandler(&mockEncryptorForManual{}, repo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/integrations/shopify/token", nil)

	user := &entity.User{ID: existingAccount.UserID, Role: valueobject.RoleAdmin}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.RevokeToken(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if repo.deleteCalls != 1 {
		t.Errorf("expected 1 delete call, got %d", repo.deleteCalls)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp["message"] != "Token revoked successfully" {
		t.Errorf("unexpected message: %v", resp["message"])
	}
}

func TestManualTokenHandler_RevokeToken_NotFound(t *testing.T) {
	repo := &mockPartnerRepoForManual{deleteErr: errors.New("not found")}
	handler := NewManualTokenHandler(&mockEncryptorForManual{}, repo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/integrations/shopify/token", nil)

	user := &entity.User{ID: uuid.New(), Role: valueobject.RoleAdmin}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.RevokeToken(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestManualTokenHandler_RevokeToken_NoUser(t *testing.T) {
	handler := NewManualTokenHandler(&mockEncryptorForManual{}, &mockPartnerRepoForManual{})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/integrations/shopify/token", nil)

	rec := httptest.NewRecorder()
	handler.RevokeToken(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"shppa_1234567890abcdef", "***...cdef"},
		{"short", "***...hort"},
		{"ab", "***...ab"},
		{"a", "***...a"},
		{"", "***..."},
	}

	for _, tt := range tests {
		result := maskToken(tt.input)
		if result != tt.expected {
			t.Errorf("maskToken(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
