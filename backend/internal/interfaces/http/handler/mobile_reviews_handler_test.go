package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/external"
)

func TestParseIosAppID(t *testing.T) {
	cases := map[string]string{
		"310633997": "310633997",
		"https://apps.apple.com/us/app/whatsapp-messenger/id310633997": "310633997",
		"":          "",
		"not-an-id": "",
	}
	for in, want := range cases {
		if got := parseIosAppID(in); got != want {
			t.Errorf("parseIosAppID(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParsePlayPackage(t *testing.T) {
	cases := map[string]string{
		"com.whatsapp": "com.whatsapp",
		"https://play.google.com/store/apps/details?id=com.zoko.app&hl=en": "com.zoko.app",
		"":         "",
		"whatsapp": "", // no dots → not a package
	}
	for in, want := range cases {
		if got := parsePlayPackage(in); got != want {
			t.Errorf("parsePlayPackage(%q) = %q, want %q", in, got, want)
		}
	}
}

type mockMobileLinksRepo struct {
	links *entity.MobileLinks
	saved *entity.MobileLinks
}

func (m *mockMobileLinksRepo) FindByAppID(_ context.Context, appID uuid.UUID) (*entity.MobileLinks, error) {
	if m.links != nil {
		return m.links, nil
	}
	return &entity.MobileLinks{AppID: appID}, nil
}
func (m *mockMobileLinksRepo) Upsert(_ context.Context, l *entity.MobileLinks) error {
	m.saved = l
	return nil
}

type mockStoreFetcher struct{}

func (mockStoreFetcher) AppleLookup(_ context.Context, id, _ string) (*external.MobileRatingSummary, error) {
	return &external.MobileRatingSummary{Store: "app_store", AppName: "WhatsApp", RatingValue: 4.69, RatingCount: 18368633}, nil
}
func (mockStoreFetcher) AppleReviews(_ context.Context, id, _ string) ([]external.MobileReview, error) {
	return []external.MobileReview{
		{Author: "a", Rating: 5, Title: "Love"}, {Author: "b", Rating: 1, Title: "Bad"}, {Author: "c", Rating: 3, Title: "Meh"},
	}, nil
}
func (mockStoreFetcher) GooglePlayListing(_ context.Context, pkg, _ string) (*external.MobileRatingSummary, error) {
	return &external.MobileRatingSummary{Store: "google_play", AppName: "WhatsApp", RatingValue: 4.63, RatingCount: 240297698}, nil
}

func newMobileHandler(app *entity.App, pa *entity.PartnerAccount, links *entity.MobileLinks) (*MobileReviewsHandler, *mockMobileLinksRepo) {
	repo := &mockMobileLinksRepo{links: links}
	h := NewMobileReviewsHandler(repo, mockStoreFetcher{}, &mockAppRepoForSub{app: app}, &mockPartnerRepoForSub{account: pa})
	return h, repo
}

func TestMobileReviews_BothStores(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "T"}
	h, _ := newMobileHandler(app, pa,
		&entity.MobileLinks{AppID: appID, IosAppID: "310633997", PlayPackage: "com.whatsapp"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+appID.String()+"/mobile/reviews", nil)
	req = withURLParam(req, "appID", appID.String())
	req = req.WithContext(contextWithUser(req.Context(), &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}))
	rec := httptest.NewRecorder()
	h.GetMobileReviews(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var resp mobileReviewsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.AppStore == nil || resp.AppStore.RatingCount != 18368633 || !resp.AppStore.ReviewsAvailable {
		t.Errorf("appStore = %+v", resp.AppStore)
	}
	// Apple sentiment over [5,1,3]: 1 positive, 1 neutral, 1 negative.
	if resp.AppStore.Positive != 1 || resp.AppStore.Neutral != 1 || resp.AppStore.Negative != 1 {
		t.Errorf("sentiment = %d/%d/%d, want 1/1/1", resp.AppStore.Positive, resp.AppStore.Neutral, resp.AppStore.Negative)
	}
	// Google is rating-only.
	if resp.GooglePlay == nil || resp.GooglePlay.ReviewsAvailable || resp.GooglePlay.RatingCount != 240297698 {
		t.Errorf("googlePlay = %+v", resp.GooglePlay)
	}
}

func TestMobileReviews_UnlinkedStoresOmitted(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "T"}
	h, _ := newMobileHandler(app, pa, nil) // no links

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+appID.String()+"/mobile/reviews", nil)
	req = withURLParam(req, "appID", appID.String())
	req = req.WithContext(contextWithUser(req.Context(), &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}))
	rec := httptest.NewRecorder()
	h.GetMobileReviews(rec, req)

	var resp mobileReviewsResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.AppStore != nil || resp.GooglePlay != nil {
		t.Errorf("expected both stores nil when unlinked, got %+v / %+v", resp.AppStore, resp.GooglePlay)
	}
}

type erroringFetcher struct{}

func (erroringFetcher) AppleLookup(context.Context, string, string) (*external.MobileRatingSummary, error) {
	return nil, errFetch
}
func (erroringFetcher) AppleReviews(context.Context, string, string) ([]external.MobileReview, error) {
	return nil, errFetch
}
func (erroringFetcher) GooglePlayListing(context.Context, string, string) (*external.MobileRatingSummary, error) {
	return nil, errFetch
}

var errFetch = &fetchError{}

type fetchError struct{}

func (*fetchError) Error() string { return "upstream down" }

// TestMobileReviews_PerStoreErrorIsolated: a store fetch failure fills that store's Error
// (still HTTP 200) rather than failing the whole response.
func TestMobileReviews_PerStoreErrorIsolated(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "T"}
	repo := &mockMobileLinksRepo{links: &entity.MobileLinks{AppID: appID, IosAppID: "310633997", PlayPackage: "com.whatsapp"}}
	h := NewMobileReviewsHandler(repo, erroringFetcher{}, &mockAppRepoForSub{app: app}, &mockPartnerRepoForSub{account: pa})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+appID.String()+"/mobile/reviews", nil)
	req = withURLParam(req, "appID", appID.String())
	req = req.WithContext(contextWithUser(req.Context(), &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}))
	rec := httptest.NewRecorder()
	h.GetMobileReviews(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, want 200 (errors isolated, not fatal)", rec.Code)
	}
	var resp mobileReviewsResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.AppStore == nil || resp.AppStore.Error == "" {
		t.Errorf("appStore.Error should be set, got %+v", resp.AppStore)
	}
	if resp.GooglePlay == nil || resp.GooglePlay.Error == "" {
		t.Errorf("googlePlay.Error should be set, got %+v", resp.GooglePlay)
	}
}

func TestMobileLinks_PutParsesUrls(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "T"}
	h, repo := newMobileHandler(app, pa, nil)

	body := `{"appStore":"https://apps.apple.com/us/app/x/id310633997","googlePlay":"com.whatsapp"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/apps/"+appID.String()+"/mobile/links", strings.NewReader(body))
	req = withURLParam(req, "appID", appID.String())
	req = req.WithContext(contextWithUser(req.Context(), &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}))
	rec := httptest.NewRecorder()
	h.PutMobileLinks(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	if repo.saved == nil || repo.saved.IosAppID != "310633997" || repo.saved.PlayPackage != "com.whatsapp" {
		t.Errorf("saved = %+v", repo.saved)
	}
}
