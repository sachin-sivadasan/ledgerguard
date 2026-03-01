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
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/external"
)

// Mock implementations
type mockPartnerClient struct {
	apps         []external.PartnerApp
	err          error
	installCount int
	installErr   error
}

func (m *mockPartnerClient) FetchApps(ctx context.Context, organizationID, accessToken string) ([]external.PartnerApp, error) {
	return m.apps, m.err
}

func (m *mockPartnerClient) FetchInstallCount(ctx context.Context, organizationID, accessToken, partnerAppID string) (int, error) {
	return m.installCount, m.installErr
}

type mockAppRepo struct {
	apps       []*entity.App
	app        *entity.App
	createErr  error
	findErr    error
	findAllErr error
}

func (m *mockAppRepo) Create(ctx context.Context, app *entity.App) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.app = app
	return nil
}

func (m *mockAppRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	return m.app, m.findErr
}

func (m *mockAppRepo) FindByPartnerAccountID(ctx context.Context, partnerAccountID uuid.UUID) ([]*entity.App, error) {
	return m.apps, m.findAllErr
}

func (m *mockAppRepo) FindByPartnerAppID(ctx context.Context, partnerAccountID uuid.UUID, partnerAppID string) (*entity.App, error) {
	return m.app, m.findErr
}

func (m *mockAppRepo) Update(ctx context.Context, app *entity.App) error {
	return nil
}

func (m *mockAppRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockAppRepo) FindAllByPartnerAppID(ctx context.Context, partnerAppID string) ([]*entity.App, error) {
	if m.app != nil {
		return []*entity.App{m.app}, nil
	}
	return nil, m.findErr
}

type mockPartnerRepoForApp struct {
	account *entity.PartnerAccount
	findErr error
}

func (m *mockPartnerRepoForApp) Create(ctx context.Context, account *entity.PartnerAccount) error {
	return nil
}

func (m *mockPartnerRepoForApp) FindByID(ctx context.Context, id uuid.UUID) (*entity.PartnerAccount, error) {
	return m.account, m.findErr
}

func (m *mockPartnerRepoForApp) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.PartnerAccount, error) {
	return m.account, m.findErr
}

func (m *mockPartnerRepoForApp) FindByPartnerID(ctx context.Context, partnerID string) (*entity.PartnerAccount, error) {
	return nil, nil
}

func (m *mockPartnerRepoForApp) Update(ctx context.Context, account *entity.PartnerAccount) error {
	return nil
}

func (m *mockPartnerRepoForApp) Delete(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (m *mockPartnerRepoForApp) GetAllIDs(ctx context.Context) ([]uuid.UUID, error) {
	if m.account != nil {
		return []uuid.UUID{m.account.ID}, nil
	}
	return []uuid.UUID{}, nil
}

type mockDecryptor struct {
	decrypted []byte
	err       error
}

func (m *mockDecryptor) Encrypt(plaintext []byte) ([]byte, error) {
	return plaintext, nil
}

func (m *mockDecryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	return m.decrypted, m.err
}

func TestAppHandler_GetAvailableApps_Success(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{
		ID:                   uuid.New(),
		UserID:               uuid.New(),
		PartnerID:            "org123",
		EncryptedAccessToken: []byte("encrypted"),
	}

	partnerRepo := &mockPartnerRepoForApp{account: partnerAccount}
	decryptor := &mockDecryptor{decrypted: []byte("decrypted-token")}
	partnerClient := &mockPartnerClient{
		apps: []external.PartnerApp{
			{ID: "gid://partners/App/123", Name: "App One"},
			{ID: "gid://partners/App/456", Name: "App Two"},
		},
	}
	appRepo := &mockAppRepo{}

	handler := NewAppHandler(partnerClient, partnerRepo, appRepo, decryptor)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/available", nil)
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.GetAvailableApps(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)

	apps, ok := resp["apps"].([]interface{})
	if !ok {
		t.Fatal("expected apps array in response")
	}

	if len(apps) != 2 {
		t.Errorf("expected 2 apps, got %d", len(apps))
	}
}

func TestAppHandler_GetAvailableApps_NoPartnerClient(t *testing.T) {
	partnerRepo := &mockPartnerRepoForApp{}
	handler := NewAppHandler(nil, partnerRepo, nil, nil) // nil partnerClient

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/available", nil)
	user := &entity.User{ID: uuid.New(), Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.GetAvailableApps(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
}

func TestAppHandler_GetAvailableApps_NoPartnerAccount(t *testing.T) {
	partnerRepo := &mockPartnerRepoForApp{findErr: errors.New("not found")}
	partnerClient := &mockPartnerClient{} // Need to provide a mock client to pass nil check
	handler := NewAppHandler(partnerClient, partnerRepo, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/available", nil)
	user := &entity.User{ID: uuid.New(), Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.GetAvailableApps(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestAppHandler_GetAvailableApps_NoUser(t *testing.T) {
	handler := NewAppHandler(nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/available", nil)
	rec := httptest.NewRecorder()
	handler.GetAvailableApps(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAppHandler_SelectApp_Success(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		PartnerID: "org123",
	}

	partnerRepo := &mockPartnerRepoForApp{account: partnerAccount}
	appRepo := &mockAppRepo{findErr: errors.New("not found")} // App doesn't exist yet

	handler := NewAppHandler(nil, partnerRepo, appRepo, nil)

	body := `{"partner_app_id": "gid://partners/App/123", "name": "My App"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/apps/select", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.SelectApp(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if appRepo.app == nil {
		t.Fatal("expected app to be created")
	}

	if appRepo.app.PartnerAppID != "gid://partners/App/123" {
		t.Errorf("expected partner_app_id 'gid://partners/App/123', got %s", appRepo.app.PartnerAppID)
	}

	if appRepo.app.Name != "My App" {
		t.Errorf("expected name 'My App', got %s", appRepo.app.Name)
	}
}

func TestAppHandler_SelectApp_AlreadyExists(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{
		ID:     uuid.New(),
		UserID: uuid.New(),
	}

	existingApp := &entity.App{
		ID:               uuid.New(),
		PartnerAccountID: partnerAccount.ID,
		PartnerAppID:     "gid://partners/App/123",
		Name:             "Existing App",
	}

	partnerRepo := &mockPartnerRepoForApp{account: partnerAccount}
	appRepo := &mockAppRepo{app: existingApp} // App already exists

	handler := NewAppHandler(nil, partnerRepo, appRepo, nil)

	body := `{"partner_app_id": "gid://partners/App/123", "name": "My App"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/apps/select", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.SelectApp(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestAppHandler_SelectApp_MissingFields(t *testing.T) {
	partnerRepo := &mockPartnerRepoForApp{account: &entity.PartnerAccount{}}
	handler := NewAppHandler(nil, partnerRepo, nil, nil)

	body := `{"name": "My App"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/apps/select", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	user := &entity.User{ID: uuid.New(), Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.SelectApp(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestAppHandler_SelectApp_NoUser(t *testing.T) {
	handler := NewAppHandler(nil, nil, nil, nil)

	body := `{"partner_app_id": "123", "name": "My App"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/apps/select", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.SelectApp(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAppHandler_ListApps_Success(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{
		ID:     uuid.New(),
		UserID: uuid.New(),
	}

	apps := []*entity.App{
		{
			ID:               uuid.New(),
			PartnerAccountID: partnerAccount.ID,
			PartnerAppID:     "gid://partners/App/123",
			Name:             "App One",
			TrackingEnabled:  true,
			CreatedAt:        time.Now(),
		},
		{
			ID:               uuid.New(),
			PartnerAccountID: partnerAccount.ID,
			PartnerAppID:     "gid://partners/App/456",
			Name:             "App Two",
			TrackingEnabled:  true,
			CreatedAt:        time.Now(),
		},
	}

	partnerRepo := &mockPartnerRepoForApp{account: partnerAccount}
	appRepo := &mockAppRepo{apps: apps}

	handler := NewAppHandler(nil, partnerRepo, appRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps", nil)
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.ListApps(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)

	appsResp, ok := resp["apps"].([]interface{})
	if !ok {
		t.Fatal("expected apps array in response")
	}

	if len(appsResp) != 2 {
		t.Errorf("expected 2 apps, got %d", len(appsResp))
	}
}

func TestAppHandler_ListApps_NoPartnerAccount(t *testing.T) {
	partnerRepo := &mockPartnerRepoForApp{findErr: errors.New("not found")}
	handler := NewAppHandler(nil, partnerRepo, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps", nil)
	user := &entity.User{ID: uuid.New(), Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.ListApps(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestAppHandler_ListApps_NoUser(t *testing.T) {
	handler := NewAppHandler(nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps", nil)
	rec := httptest.NewRecorder()
	handler.ListApps(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}
