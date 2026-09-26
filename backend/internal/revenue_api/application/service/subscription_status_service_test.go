package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	domainentity "github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	reventity "github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/domain/entity"
)

// --- minimal mocks (only the methods the service uses are meaningful) ---

type mockStatusRepo struct {
	byGID map[string]*reventity.SubscriptionStatus
}

func (m *mockStatusRepo) GetByShopifyGID(_ context.Context, gid string) (*reventity.SubscriptionStatus, error) {
	if s, ok := m.byGID[gid]; ok {
		return s, nil
	}
	return nil, errors.New("not found")
}
func (m *mockStatusRepo) GetByShopifyGIDs(_ context.Context, gids []string) ([]*reventity.SubscriptionStatus, error) {
	var out []*reventity.SubscriptionStatus
	for _, g := range gids {
		if s, ok := m.byGID[g]; ok {
			out = append(out, s)
		}
	}
	return out, nil
}
func (m *mockStatusRepo) GetByDomain(_ context.Context, appID uuid.UUID, domain string) (*reventity.SubscriptionStatus, error) {
	for _, s := range m.byGID {
		if s.AppID == appID && s.MyshopifyDomain == domain {
			return s, nil
		}
	}
	return nil, errors.New("not found")
}
func (m *mockStatusRepo) Upsert(context.Context, *reventity.SubscriptionStatus) error      { return nil }
func (m *mockStatusRepo) UpsertBatch(context.Context, []*reventity.SubscriptionStatus) error { return nil }
func (m *mockStatusRepo) GetByDomains(context.Context, uuid.UUID, []string) ([]*reventity.SubscriptionStatus, error) {
	return nil, nil
}
func (m *mockStatusRepo) GetByAppID(context.Context, uuid.UUID) ([]*reventity.SubscriptionStatus, error) {
	return nil, nil
}
func (m *mockStatusRepo) GetByAppIDAndRiskState(context.Context, uuid.UUID, valueobject.RiskState) ([]*reventity.SubscriptionStatus, error) {
	return nil, nil
}
func (m *mockStatusRepo) DeleteByAppID(context.Context, uuid.UUID) error { return nil }

type mockAppRepo struct {
	byID      map[uuid.UUID]*domainentity.App
	byPartner map[uuid.UUID][]*domainentity.App
}

func (m *mockAppRepo) FindByID(_ context.Context, id uuid.UUID) (*domainentity.App, error) {
	if a, ok := m.byID[id]; ok {
		return a, nil
	}
	return nil, errors.New("app not found")
}
func (m *mockAppRepo) FindByPartnerAccountID(_ context.Context, paID uuid.UUID) ([]*domainentity.App, error) {
	return m.byPartner[paID], nil
}
func (m *mockAppRepo) Create(context.Context, *domainentity.App) error { return nil }
func (m *mockAppRepo) FindByPartnerAppID(context.Context, uuid.UUID, string) (*domainentity.App, error) {
	return nil, errors.New("not found")
}
func (m *mockAppRepo) FindAllByPartnerAppID(context.Context, string) ([]*domainentity.App, error) {
	return nil, nil
}
func (m *mockAppRepo) Update(context.Context, *domainentity.App) error                 { return nil }
func (m *mockAppRepo) UpdateInstallCount(context.Context, uuid.UUID, int) error        { return nil }
func (m *mockAppRepo) Delete(context.Context, uuid.UUID) error                         { return nil }

type mockPartnerRepo struct {
	byUser map[uuid.UUID]*domainentity.PartnerAccount
}

func (m *mockPartnerRepo) FindByUserID(_ context.Context, userID uuid.UUID) (*domainentity.PartnerAccount, error) {
	if p, ok := m.byUser[userID]; ok {
		return p, nil
	}
	return nil, errors.New("partner account not found")
}
func (m *mockPartnerRepo) Create(context.Context, *domainentity.PartnerAccount) error { return nil }
func (m *mockPartnerRepo) FindByID(context.Context, uuid.UUID) (*domainentity.PartnerAccount, error) {
	return nil, errors.New("not found")
}
func (m *mockPartnerRepo) FindByOrgID(context.Context, uuid.UUID) (*domainentity.PartnerAccount, error) {
	return nil, errors.New("not found")
}
func (m *mockPartnerRepo) FindByPartnerID(context.Context, string) (*domainentity.PartnerAccount, error) {
	return nil, errors.New("not found")
}
func (m *mockPartnerRepo) Update(context.Context, *domainentity.PartnerAccount) error { return nil }
func (m *mockPartnerRepo) Delete(context.Context, uuid.UUID) error                   { return nil }
func (m *mockPartnerRepo) GetAllIDs(context.Context) ([]uuid.UUID, error)            { return nil, nil }

// --- test fixture: two orgs (A owns appA, B owns appB) sharing the status store ---

type isoFixture struct {
	svc   *SubscriptionStatusService
	userA uuid.UUID
	appA  uuid.UUID
	appB  uuid.UUID
}

func newIsoFixture() isoFixture {
	userA, paA, paB := uuid.New(), uuid.New(), uuid.New()
	appA := &domainentity.App{ID: uuid.New(), PartnerAccountID: paA}
	appB := &domainentity.App{ID: uuid.New(), PartnerAccountID: paB}

	statusRepo := &mockStatusRepo{byGID: map[string]*reventity.SubscriptionStatus{
		"gidA": {ID: uuid.New(), ShopifyGID: "gidA", AppID: appA.ID, MyshopifyDomain: "a.myshopify.com"},
		"gidB": {ID: uuid.New(), ShopifyGID: "gidB", AppID: appB.ID, MyshopifyDomain: "b.myshopify.com"},
	}}
	appRepo := &mockAppRepo{
		byID:      map[uuid.UUID]*domainentity.App{appA.ID: appA, appB.ID: appB},
		byPartner: map[uuid.UUID][]*domainentity.App{paA: {appA}, paB: {appB}},
	}
	partnerRepo := &mockPartnerRepo{byUser: map[uuid.UUID]*domainentity.PartnerAccount{
		userA: {ID: paA},
	}}

	return isoFixture{
		svc:   NewSubscriptionStatusService(statusRepo, appRepo, partnerRepo),
		userA: userA,
		appA:  appA.ID,
		appB:  appB.ID,
	}
}

// TestGetByShopifyGID_OwnApp_Succeeds: user A can read a subscription on their own app.
func TestGetByShopifyGID_OwnApp_Succeeds(t *testing.T) {
	f := newIsoFixture()
	got, err := f.svc.GetByShopifyGID(context.Background(), f.userA, "gidA")
	if err != nil {
		t.Fatalf("expected success reading own app, got %v", err)
	}
	if got.AppID != f.appA {
		t.Errorf("expected status for appA, got appID %v", got.AppID)
	}
}

// TestGetByShopifyGID_CrossOrg_Denied is the S6 isolation guard: user A must NOT be
// able to read a subscription that belongs to another org's app (appB).
func TestGetByShopifyGID_CrossOrg_Denied(t *testing.T) {
	f := newIsoFixture()
	got, err := f.svc.GetByShopifyGID(context.Background(), f.userA, "gidB")
	if !errors.Is(err, ErrAppAccessDenied) {
		t.Fatalf("expected ErrAppAccessDenied for cross-org access, got %v", err)
	}
	if got != nil {
		t.Error("expected no data leaked on cross-org access")
	}
}

// TestGetByShopifyGIDs_Batch_ExcludesOtherOrg: a batch mixing owned + other-org GIDs
// returns only the owned one; the other-org GID goes to not_found (never leaked).
func TestGetByShopifyGIDs_Batch_ExcludesOtherOrg(t *testing.T) {
	f := newIsoFixture()
	resp, err := f.svc.GetByShopifyGIDs(context.Background(), f.userA, []string{"gidA", "gidB"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Results) != 1 || resp.Results[0].SubscriptionID != "gidA" {
		t.Fatalf("expected only gidA in results, got %+v", resp.Results)
	}
	if len(resp.NotFound) != 1 || resp.NotFound[0] != "gidB" {
		t.Errorf("expected gidB in not_found (not leaked), got %+v", resp.NotFound)
	}
}

// TestGetByDomain_OnlyUsersApps: domain lookup only searches the caller's apps.
func TestGetByDomain_OnlyUsersApps(t *testing.T) {
	f := newIsoFixture()
	// Own app's domain resolves.
	if _, err := f.svc.GetByDomain(context.Background(), f.userA, "a.myshopify.com"); err != nil {
		t.Fatalf("expected to resolve own domain, got %v", err)
	}
	// Another org's domain is not found (their app isn't in the caller's app set).
	if _, err := f.svc.GetByDomain(context.Background(), f.userA, "b.myshopify.com"); !errors.Is(err, ErrSubscriptionNotFound) {
		t.Errorf("expected ErrSubscriptionNotFound for another org's domain, got %v", err)
	}
}
