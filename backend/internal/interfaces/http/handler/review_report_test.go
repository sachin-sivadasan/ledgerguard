package handler

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// mockAppReviewRepo implements repository.AppReviewRepository for reviews-report tests.
type mockAppReviewRepo struct {
	reviews []*entity.AppReview
	err     error
}

func (m *mockAppReviewRepo) UpsertBatch(ctx context.Context, reviews []*entity.AppReview) error {
	return nil
}
func (m *mockAppReviewRepo) FindByAppID(ctx context.Context, appID uuid.UUID, limit, offset int) ([]*entity.AppReview, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.reviews, nil
}
func (m *mockAppReviewRepo) FindAllByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.AppReview, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.reviews, nil
}
func (m *mockAppReviewRepo) CountByAppID(ctx context.Context, appID uuid.UUID) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	return len(m.reviews), nil
}

// review builds an AppReview with the given rating and review date.
func review(appID uuid.UUID, rating int, body string, date time.Time) *entity.AppReview {
	return &entity.AppReview{
		ID:         uuid.New(),
		AppID:      appID,
		Author:     "Acme Store",
		Rating:     rating,
		Body:       body,
		ReviewDate: date,
		Location:   "United Kingdom",
		TimeUsing:  "4 months using the app",
		Source:     "shopify_app_store",
	}
}

func reviewsFixture(reviews []*entity.AppReview, repoErr error) (uuid.UUID, *entity.PartnerAccount, *ReviewHandler) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewReviewHandler(
		&mockAppReviewRepo{reviews: reviews, err: repoErr},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
		nil,
	)
	return appID, pa, h
}

func doReviewsReport(t *testing.T, h *ReviewHandler, appID uuid.UUID, pa *entity.PartnerAccount, query string) *httptest.ResponseRecorder {
	t.Helper()
	url := "/api/v1/apps/" + appID.String() + "/reports/reviews"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	rec := httptest.NewRecorder()
	h.GetReviewsReport(rec, req)
	return rec
}

func decodeReviewsReport(t *testing.T, rec *httptest.ResponseRecorder) reviewsReport {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp reviewsReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}

// TestReviewsReport_AvgRating verifies avgRating is the mean of all ratings.
func TestReviewsReport_AvgRating(t *testing.T) {
	appID := uuid.New()
	d := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	reviews := []*entity.AppReview{
		review(appID, 5, "great", d),
		review(appID, 4, "good", d),
		review(appID, 3, "meh", d),
	}
	aid, pa, h := reviewsFixture(reviews, nil)
	resp := decodeReviewsReport(t, doReviewsReport(t, h, aid, pa, ""))
	// (5+4+3)/3 = 4.0
	if resp.AvgRating != 4.0 {
		t.Errorf("avgRating: expected 4.0, got %v", resp.AvgRating)
	}
	if resp.TotalReviews != 3 {
		t.Errorf("totalReviews: expected 3, got %d", resp.TotalReviews)
	}
}

// TestReviewsReport_AvgRatingPrecision verifies avgRating is a raw (unrounded) mean,
// pinning the API contract against a future "helpful" rounding/truncation.
func TestReviewsReport_AvgRatingPrecision(t *testing.T) {
	appID := uuid.New()
	d := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	// (5+4+4)/3 = 4.3333… — non-terminating; must not be rounded/truncated.
	reviews := []*entity.AppReview{
		review(appID, 5, "a", d), review(appID, 4, "b", d), review(appID, 4, "c", d),
	}
	aid, pa, h := reviewsFixture(reviews, nil)
	resp := decodeReviewsReport(t, doReviewsReport(t, h, aid, pa, ""))
	want := 13.0 / 3.0
	if diff := resp.AvgRating - want; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("avgRating: expected ~%.10f (raw mean), got %v", want, resp.AvgRating)
	}
}

// TestDeriveSentiment_Boundaries pins the sentiment label at each rating boundary.
func TestDeriveSentiment_Boundaries(t *testing.T) {
	cases := map[int]string{1: "negative", 2: "negative", 3: "neutral", 4: "positive", 5: "positive"}
	for rating, want := range cases {
		if got := deriveSentiment(rating); got != want {
			t.Errorf("deriveSentiment(%d): expected %q, got %q", rating, want, got)
		}
	}
}

// TestReviewsReport_OutOfRangeRatingExcluded verifies a malformed rating (outside
// 1–5) is excluded from avgRating, totalReviews, distribution and sentiment, so the
// bucket counts stay consistent with totalReviews.
func TestReviewsReport_OutOfRangeRatingExcluded(t *testing.T) {
	appID := uuid.New()
	d := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	reviews := []*entity.AppReview{
		review(appID, 5, "ok", d),
		review(appID, 0, "corrupt-low", d),  // out of range → excluded
		review(appID, 6, "corrupt-high", d), // out of range → excluded
		review(appID, 3, "ok", d),
	}
	aid, pa, h := reviewsFixture(reviews, nil)
	resp := decodeReviewsReport(t, doReviewsReport(t, h, aid, pa, ""))

	if resp.TotalReviews != 2 {
		t.Errorf("totalReviews: expected 2 (2 malformed excluded), got %d", resp.TotalReviews)
	}
	if resp.AvgRating != 4.0 { // (5+3)/2
		t.Errorf("avgRating: expected 4.0, got %v", resp.AvgRating)
	}
	// Bucket counts must sum to totalReviews.
	sum := 0
	for _, b := range resp.Distribution {
		sum += b.Count
	}
	if sum != resp.TotalReviews {
		t.Errorf("distribution sum %d != totalReviews %d (out-of-range leaked)", sum, resp.TotalReviews)
	}
	if resp.Sentiment.Positive+resp.Sentiment.Neutral+resp.Sentiment.Negative != 2 {
		t.Errorf("sentiment total: expected 2, got %+v", resp.Sentiment)
	}
}

// TestReviewsReport_CSVFieldOrder verifies each CSV column maps to the right field
// (guards against a silent column-order swap).
func TestReviewsReport_CSVFieldOrder(t *testing.T) {
	appID := uuid.New()
	d := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	rev := &entity.AppReview{
		ID: uuid.New(), AppID: appID, Author: "Distinct Author", Rating: 4,
		Body: "distinct body", ReviewDate: d, Location: "Canada",
		TimeUsing: "2 months using the app", Source: "shopify_app_store",
	}
	aid, pa, h := reviewsFixture([]*entity.AppReview{rev}, nil)
	rec := doReviewsReport(t, h, aid, pa, "format=csv")
	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected header + 1 row, got %d", len(records))
	}
	want := []string{"Distinct Author", "4", "2026-07-21", "Canada", "2 months using the app", "positive", "shopify_app_store", "distinct body"}
	for i, w := range want {
		if records[1][i] != w {
			t.Errorf("CSV col %d: expected %q, got %q", i, w, records[1][i])
		}
	}
}

// TestReviewsReport_CSVRepoError503 verifies the CSV path also surfaces repo errors
// as 503 (not just the JSON path).
func TestReviewsReport_CSVRepoError503(t *testing.T) {
	aid, pa, h := reviewsFixture(nil, errors.New("db down"))
	rec := doReviewsReport(t, h, aid, pa, "format=csv")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 on CSV path, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestReviewsReport_Distribution verifies all five star buckets are present (ordered
// 5→1), including a rating with count 0.
func TestReviewsReport_Distribution(t *testing.T) {
	appID := uuid.New()
	d := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	// 5:2, 4:1, 3:0, 2:1, 1:0
	reviews := []*entity.AppReview{
		review(appID, 5, "a", d),
		review(appID, 5, "b", d),
		review(appID, 4, "c", d),
		review(appID, 2, "d", d),
	}
	aid, pa, h := reviewsFixture(reviews, nil)
	resp := decodeReviewsReport(t, doReviewsReport(t, h, aid, pa, ""))
	if len(resp.Distribution) != 5 {
		t.Fatalf("expected 5 distribution buckets, got %d", len(resp.Distribution))
	}
	want := []reviewDistributionBucket{
		{Rating: 5, Count: 2},
		{Rating: 4, Count: 1},
		{Rating: 3, Count: 0},
		{Rating: 2, Count: 1},
		{Rating: 1, Count: 0},
	}
	for i, w := range want {
		if resp.Distribution[i] != w {
			t.Errorf("distribution[%d]: expected %+v, got %+v", i, w, resp.Distribution[i])
		}
	}
}

// TestReviewsReport_Sentiment verifies the sentiment breakdown via deriveSentiment
// (<=2 negative, ==3 neutral, >=4 positive).
func TestReviewsReport_Sentiment(t *testing.T) {
	appID := uuid.New()
	d := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	reviews := []*entity.AppReview{
		review(appID, 5, "pos", d),
		review(appID, 4, "pos", d),
		review(appID, 3, "neu", d),
		review(appID, 2, "neg", d),
		review(appID, 1, "neg", d),
	}
	aid, pa, h := reviewsFixture(reviews, nil)
	resp := decodeReviewsReport(t, doReviewsReport(t, h, aid, pa, ""))
	if resp.Sentiment.Positive != 2 || resp.Sentiment.Neutral != 1 || resp.Sentiment.Negative != 2 {
		t.Errorf("sentiment: expected {pos:2,neu:1,neg:2}, got %+v", resp.Sentiment)
	}
}

// TestReviewsReport_RecentCappedAndSorted verifies recent is capped at 12 and stays
// in the (date-descending) order the repo returns.
func TestReviewsReport_RecentCappedAndSorted(t *testing.T) {
	appID := uuid.New()
	base := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	// 15 reviews, already sorted date-descending (newest first).
	var reviews []*entity.AppReview
	for i := 0; i < 15; i++ {
		reviews = append(reviews, review(appID, 5, "r", base.AddDate(0, 0, -i)))
	}
	aid, pa, h := reviewsFixture(reviews, nil)
	resp := decodeReviewsReport(t, doReviewsReport(t, h, aid, pa, ""))
	if len(resp.Recent) != 12 {
		t.Fatalf("expected recent capped at 12, got %d", len(resp.Recent))
	}
	if resp.TotalReviews != 15 {
		t.Errorf("totalReviews: expected 15, got %d", resp.TotalReviews)
	}
	// First entry is newest; order preserved descending.
	for i := 1; i < len(resp.Recent); i++ {
		if resp.Recent[i-1].Date < resp.Recent[i].Date {
			t.Errorf("recent not date-descending at %d: %s before %s", i, resp.Recent[i-1].Date, resp.Recent[i].Date)
		}
	}
	if resp.Recent[0].Date != "2026-07-21" {
		t.Errorf("recent[0] date: expected 2026-07-21, got %s", resp.Recent[0].Date)
	}
}

// TestReviewsReport_RecentFields verifies the recent[] JSON contract keys (text/date,
// not body/review_date) and derived sentiment.
func TestReviewsReport_RecentFields(t *testing.T) {
	appID := uuid.New()
	d := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	aid, pa, h := reviewsFixture([]*entity.AppReview{review(appID, 5, "Loved it", d)}, nil)
	rec := doReviewsReport(t, h, aid, pa, "")
	body := rec.Body.String()
	if !strings.Contains(body, `"text":"Loved it"`) {
		t.Errorf("expected recent to use `text` key, body: %s", body)
	}
	if !strings.Contains(body, `"date":"2026-07-21"`) {
		t.Errorf("expected recent to use `date` key, body: %s", body)
	}
	if strings.Contains(body, `"body"`) || strings.Contains(body, `"review_date"`) {
		t.Errorf("recent must not use body/review_date keys, body: %s", body)
	}
	resp := decodeReviewsReport(t, rec)
	if resp.Recent[0].Sentiment != "positive" {
		t.Errorf("expected positive sentiment, got %s", resp.Recent[0].Sentiment)
	}
}

// TestReviewsReport_Empty verifies the empty case: avgRating 0, totalReviews 0,
// all-zero distribution (all 5 buckets present) and recent serialized as [].
func TestReviewsReport_Empty(t *testing.T) {
	aid, pa, h := reviewsFixture(nil, nil)
	rec := doReviewsReport(t, h, aid, pa, "")
	resp := decodeReviewsReport(t, rec)
	if resp.AvgRating != 0 {
		t.Errorf("avgRating: expected 0, got %v", resp.AvgRating)
	}
	if resp.TotalReviews != 0 {
		t.Errorf("totalReviews: expected 0, got %d", resp.TotalReviews)
	}
	if len(resp.Distribution) != 5 {
		t.Fatalf("expected 5 distribution buckets, got %d", len(resp.Distribution))
	}
	for _, b := range resp.Distribution {
		if b.Count != 0 {
			t.Errorf("expected all-zero distribution, got %+v", b)
		}
	}
	if !strings.Contains(rec.Body.String(), `"recent":[]`) {
		t.Errorf("expected recent serialized as [], body: %s", rec.Body.String())
	}
}

// TestReviewsReport_RepoErrorReturns503 verifies review-repo failures surface as 503.
func TestReviewsReport_RepoErrorReturns503(t *testing.T) {
	aid, pa, h := reviewsFixture(nil, errors.New("db down"))
	rec := doReviewsReport(t, h, aid, pa, "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestReviewsReport_Unauthenticated verifies a nil user yields 401.
func TestReviewsReport_Unauthenticated(t *testing.T) {
	aid, _, h := reviewsFixture(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+aid.String()+"/reports/reviews", nil)
	req = withURLParam(req, "appID", aid.String())
	rec := httptest.NewRecorder()
	h.GetReviewsReport(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

// TestReviewsReport_CSVFormat verifies CSV output: header + one row per review (ALL
// reviews, not just recent).
func TestReviewsReport_CSVFormat(t *testing.T) {
	appID := uuid.New()
	d := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	var reviews []*entity.AppReview
	for i := 0; i < 15; i++ {
		reviews = append(reviews, review(appID, 5, "r", d))
	}
	aid, pa, h := reviewsFixture(reviews, nil)
	rec := doReviewsReport(t, h, aid, pa, "format=csv")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected Content-Type text/csv, got %q", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "reviews.csv") {
		t.Errorf("expected filename in Content-Disposition, got %q", cd)
	}
	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	// Header + all 15 reviews (not capped at 12).
	if len(records) != 16 {
		t.Fatalf("expected header + 15 rows, got %d", len(records))
	}
	wantHeader := []string{"author", "rating", "reviewDate", "location", "timeUsing", "sentiment", "source", "body"}
	for i, want := range wantHeader {
		if records[0][i] != want {
			t.Errorf("header[%d]: expected %q, got %q", i, want, records[0][i])
		}
	}
}

// TestReviewsReport_CSVEscaping verifies a review body with commas/quotes/newlines
// stays a single field.
func TestReviewsReport_CSVEscaping(t *testing.T) {
	appID := uuid.New()
	d := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	body := "Great app, really.\nBut \"pricey\", tbh."
	aid, pa, h := reviewsFixture([]*entity.AppReview{review(appID, 5, body, d)}, nil)
	rec := doReviewsReport(t, h, aid, pa, "format=csv")
	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected header + 1 row, got %d: %v", len(records), records)
	}
	if len(records[1]) != 8 {
		t.Fatalf("expected 8 columns, got %d: %v", len(records[1]), records[1])
	}
	if records[1][7] != body {
		t.Errorf("body column: expected %q, got %q", body, records[1][7])
	}
}
