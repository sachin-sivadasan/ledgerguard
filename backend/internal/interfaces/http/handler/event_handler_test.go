package handler

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

type mockEventRepo struct {
	events  []*entity.AppEvent
	findErr error
}

func (m *mockEventRepo) UpsertBatch(ctx context.Context, events []*entity.AppEvent) error {
	return nil
}
func (m *mockEventRepo) FindByAppAndShop(ctx context.Context, appID uuid.UUID, shopGID string) ([]*entity.AppEvent, error) {
	return nil, nil
}
func (m *mockEventRepo) FindByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.AppEvent, error) {
	return m.events, m.findErr
}
func (m *mockEventRepo) FindByAppIDPaginated(ctx context.Context, appID uuid.UUID, filters repository.EventFilters) (*repository.EventPage, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}

	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	total := len(m.events)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	return &repository.EventPage{
		Events:     m.events[start:end],
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	}, nil
}

func makeEventTestData(appID uuid.UUID, count int) []*entity.AppEvent {
	events := make([]*entity.AppEvent, count)
	for i := 0; i < count; i++ {
		events[i] = &entity.AppEvent{
			ID:             uuid.New(),
			AppID:          appID,
			ShopifyShopGID: "store.myshopify.com",
			EventType:      "RELATIONSHIP_INSTALLED",
			OccurredAt:     time.Now().Add(-time.Duration(i) * time.Hour),
			CreatedAt:      time.Now(),
		}
	}
	return events
}

func TestEventHandler_List_DefaultPage(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	appID := uuid.New()
	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccount.ID,
		Name:             "Test App",
		PartnerAppID:     "gid://partners/App/12345",
	}

	events := makeEventTestData(appID, 50)
	eventRepo := &mockEventRepo{events: events}
	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}

	handler := NewEventHandler(eventRepo, partnerRepo, appRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+appID.String()+"/events", nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)

	eventList := resp["events"].([]interface{})
	if len(eventList) != 20 {
		t.Errorf("expected 20 events (default pageSize), got %d", len(eventList))
	}
	if resp["total"].(float64) != 50 {
		t.Errorf("expected total=50, got %v", resp["total"])
	}
	if resp["page"].(float64) != 1 {
		t.Errorf("expected page=1, got %v", resp["page"])
	}
	if resp["totalPages"].(float64) != 3 {
		t.Errorf("expected totalPages=3, got %v", resp["totalPages"])
	}
}

func TestEventHandler_List_CustomPage(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	appID := uuid.New()
	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccount.ID,
		Name:             "Test App",
		PartnerAppID:     "gid://partners/App/12345",
	}

	events := makeEventTestData(appID, 30)
	eventRepo := &mockEventRepo{events: events}
	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}

	handler := NewEventHandler(eventRepo, partnerRepo, appRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+appID.String()+"/events?page=2&pageSize=10", nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)

	eventList := resp["events"].([]interface{})
	if len(eventList) != 10 {
		t.Errorf("expected 10 events, got %d", len(eventList))
	}
	if resp["page"].(float64) != 2 {
		t.Errorf("expected page=2, got %v", resp["page"])
	}
	if resp["totalPages"].(float64) != 3 {
		t.Errorf("expected totalPages=3, got %v", resp["totalPages"])
	}

	// Check event type mapping works
	first := eventList[0].(map[string]interface{})
	if first["type"] != "APP_INSTALL" {
		t.Errorf("expected mapped type APP_INSTALL, got %v", first["type"])
	}
}
