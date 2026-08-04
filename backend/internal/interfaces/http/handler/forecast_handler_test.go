package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// mockSnapshotRepoForForecast implements DailyMetricsSnapshotRepository
type mockSnapshotRepoForForecast struct {
	snapshots []*entity.DailyMetricsSnapshot
	rangeErr  error
}

func (m *mockSnapshotRepoForForecast) Upsert(ctx context.Context, s *entity.DailyMetricsSnapshot) error {
	return nil
}
func (m *mockSnapshotRepoForForecast) UpsertBatch(ctx context.Context, snapshots []*entity.DailyMetricsSnapshot) error {
	return nil
}
func (m *mockSnapshotRepoForForecast) FindByAppIDAndDate(ctx context.Context, appID uuid.UUID, date time.Time) (*entity.DailyMetricsSnapshot, error) {
	return nil, nil
}
func (m *mockSnapshotRepoForForecast) FindByAppIDRange(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.DailyMetricsSnapshot, error) {
	if m.rangeErr != nil {
		return nil, m.rangeErr
	}
	return m.snapshots, nil
}
func (m *mockSnapshotRepoForForecast) FindLatestByAppID(ctx context.Context, appID uuid.UUID) (*entity.DailyMetricsSnapshot, error) {
	if len(m.snapshots) == 0 {
		return nil, nil
	}
	return m.snapshots[len(m.snapshots)-1], nil
}

// mockPartnerRepoForForecast implements PartnerAccountRepository
type mockPartnerRepoForForecast struct {
	account *entity.PartnerAccount
}

func (m *mockPartnerRepoForForecast) Create(ctx context.Context, a *entity.PartnerAccount) error {
	return nil
}
func (m *mockPartnerRepoForForecast) FindByID(ctx context.Context, id uuid.UUID) (*entity.PartnerAccount, error) {
	return m.account, nil
}
func (m *mockPartnerRepoForForecast) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.PartnerAccount, error) {
	return m.account, nil
}
func (m *mockPartnerRepoForForecast) FindByOrgID(ctx context.Context, orgID uuid.UUID) (*entity.PartnerAccount, error) {
	return m.account, nil
}
func (m *mockPartnerRepoForForecast) FindByPartnerID(ctx context.Context, partnerID string) (*entity.PartnerAccount, error) {
	return m.account, nil
}
func (m *mockPartnerRepoForForecast) Update(ctx context.Context, a *entity.PartnerAccount) error {
	return nil
}
func (m *mockPartnerRepoForForecast) Delete(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *mockPartnerRepoForForecast) GetAllIDs(ctx context.Context) ([]uuid.UUID, error) {
	return nil, nil
}

// mockAppRepoForForecast implements AppRepository
type mockAppRepoForForecast struct {
	app *entity.App
}

func (m *mockAppRepoForForecast) Create(ctx context.Context, app *entity.App) error { return nil }
func (m *mockAppRepoForForecast) FindByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	return m.app, nil
}
func (m *mockAppRepoForForecast) FindByPartnerAccountID(ctx context.Context, partnerAccountID uuid.UUID) ([]*entity.App, error) {
	return []*entity.App{m.app}, nil
}
func (m *mockAppRepoForForecast) FindByPartnerAppID(ctx context.Context, partnerAccountID uuid.UUID, partnerAppID string) (*entity.App, error) {
	return m.app, nil
}
func (m *mockAppRepoForForecast) Update(ctx context.Context, app *entity.App) error { return nil }
func (m *mockAppRepoForForecast) Delete(ctx context.Context, id uuid.UUID) error    { return nil }
func (m *mockAppRepoForForecast) FindAllByPartnerAppID(ctx context.Context, partnerAppID string) ([]*entity.App, error) {
	return []*entity.App{m.app}, nil
}

func TestForecastHandler_InsufficientData(t *testing.T) {
	userID := uuid.New()
	partnerID := uuid.New()
	appID := uuid.New()
	app := &entity.App{ID: appID, PartnerAccountID: partnerID, Name: "Test App"}

	// Only 30 snapshots (need 90)
	snapshots := make([]*entity.DailyMetricsSnapshot, 30)
	for i := range snapshots {
		snapshots[i] = &entity.DailyMetricsSnapshot{
			ID: uuid.New(), AppID: appID,
			Date: time.Now().AddDate(0, 0, -30+i), ActiveMRRCents: 500000 + int64(i)*1000,
		}
	}

	h := NewForecastHandler(
		&mockSnapshotRepoForForecast{snapshots: snapshots},
		&mockAppRepoForForecast{app: app},
		&mockPartnerRepoForForecast{account: &entity.PartnerAccount{ID: partnerID, UserID: userID}},
	)

	req := httptest.NewRequest("GET", "/api/v1/apps/"+appID.String()+"/forecast", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("appID", appID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(middleware.SetUserContext(req.Context(), &entity.User{ID: userID}))

	w := httptest.NewRecorder()
	h.GetForecast(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d: %s", w.Code, w.Body.String())
	}
}

func TestForecastHandler_LinearModel(t *testing.T) {
	userID := uuid.New()
	partnerID := uuid.New()
	appID := uuid.New()
	app := &entity.App{ID: appID, PartnerAccountID: partnerID, Name: "Test App"}

	snapshots := make([]*entity.DailyMetricsSnapshot, 120)
	for i := range snapshots {
		snapshots[i] = &entity.DailyMetricsSnapshot{
			ID: uuid.New(), AppID: appID,
			Date: time.Now().AddDate(0, 0, -120+i), ActiveMRRCents: 500000 + int64(i)*1000,
		}
	}

	h := NewForecastHandler(
		&mockSnapshotRepoForForecast{snapshots: snapshots},
		&mockAppRepoForForecast{app: app},
		&mockPartnerRepoForForecast{account: &entity.PartnerAccount{ID: partnerID, UserID: userID}},
	)

	req := httptest.NewRequest("GET", "/api/v1/apps/"+appID.String()+"/forecast?model=linear&months=6", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("appID", appID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(middleware.SetUserContext(req.Context(), &entity.User{ID: userID}))

	w := httptest.NewRecorder()
	h.GetForecast(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp["model"] != "linear" {
		t.Errorf("expected model 'linear', got %v", resp["model"])
	}
	forecast := resp["forecast"].([]interface{})
	if len(forecast) != 6 {
		t.Errorf("expected 6 forecast points, got %d", len(forecast))
	}
}

func TestForecastHandler_InvalidModel(t *testing.T) {
	userID := uuid.New()
	partnerID := uuid.New()
	appID := uuid.New()
	app := &entity.App{ID: appID, PartnerAccountID: partnerID, Name: "Test App"}

	h := NewForecastHandler(
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForForecast{app: app},
		&mockPartnerRepoForForecast{account: &entity.PartnerAccount{ID: partnerID, UserID: userID}},
	)

	req := httptest.NewRequest("GET", "/api/v1/apps/"+appID.String()+"/forecast?model=invalid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("appID", appID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(middleware.SetUserContext(req.Context(), &entity.User{ID: userID}))

	w := httptest.NewRecorder()
	h.GetForecast(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func (m *mockAppRepoForForecast) UpdateInstallCount(ctx context.Context, appID uuid.UUID, count int) error {
	return nil
}
