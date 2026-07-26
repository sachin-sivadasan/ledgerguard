package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/external"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type ReviewHandler struct {
	reviewRepo repository.AppReviewRepository
	appRepo    repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
	scraper    *external.ShopifyAppStoreClient
}

func NewReviewHandler(
	reviewRepo repository.AppReviewRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
	scraper *external.ShopifyAppStoreClient,
) *ReviewHandler {
	return &ReviewHandler{
		reviewRepo:  reviewRepo,
		appRepo:     appRepo,
		partnerRepo: partnerRepo,
		scraper:     scraper,
	}
}

// List returns paginated reviews for an app
// GET /api/v1/apps/{appID}/reviews?page=1&per_page=20
func (h *ReviewHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	app, err := h.findApp(r)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "app not found")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	reviews, err := h.reviewRepo.FindByAppID(r.Context(), app.ID, perPage, offset)
	if err != nil {
		writeReviewRepoError(w, "FindByAppID", err)
		return
	}

	total, err := h.reviewRepo.CountByAppID(r.Context(), app.ID)
	if err != nil {
		writeReviewRepoError(w, "CountByAppID", err)
		return
	}

	reviewResponses := make([]map[string]any, len(reviews))
	for i, rev := range reviews {
		reviewResponses[i] = map[string]any{
			"id":          rev.ID,
			"author":      rev.Author,
			"rating":      rev.Rating,
			"body":        rev.Body,
			"review_date": rev.ReviewDate.Format("2006-01-02"),
			"location":    rev.Location,
			"time_using":  rev.TimeUsing,
			"sentiment":   deriveSentiment(rev.Rating),
			"source":      rev.Source,
			"scraped_at":  rev.ScrapedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"reviews":  reviewResponses,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

type scrapeRequest struct {
	MaxPages int `json:"max_pages"`
}

// Scrape triggers a review scrape for an app
// POST /api/v1/apps/{appID}/reviews/scrape
func (h *ReviewHandler) Scrape(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if h.scraper == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "review scraper not configured")
		return
	}

	app, err := h.findApp(r)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "app not found")
		return
	}

	if app.AppStoreSlug == "" {
		writeJSONError(w, http.StatusBadRequest, "app has no app_store_slug configured")
		return
	}

	var req scrapeRequest
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&req) // ignore decode errors, use defaults
	}
	if req.MaxPages <= 0 {
		req.MaxPages = 5
	}

	scraped, err := h.scraper.ScrapeReviews(r.Context(), app.AppStoreSlug, req.MaxPages)
	if err != nil {
		log.Printf("WARNING: scrape failed for app %s (slug: %s): %v", app.ID, app.AppStoreSlug, err)
		writeJSONError(w, http.StatusBadGateway, "scrape failed: "+err.Error())
		return
	}

	// Convert scraped reviews to domain entities
	now := time.Now().UTC()
	var reviews []*entity.AppReview
	for _, sr := range scraped {
		if sr.Rating == 0 || sr.Date.IsZero() {
			continue // skip malformed reviews
		}
		reviews = append(reviews, &entity.AppReview{
			ID:             uuid.New(),
			AppID:          app.ID,
			SourceReviewID: sr.SourceReviewID(),
			Author:         sr.Author,
			Rating:         sr.Rating,
			Body:           sr.Body,
			ReviewDate:     sr.Date,
			Location:       sr.Location,
			TimeUsing:      sr.TimeUsing,
			Source:         "shopify_app_store",
			ScrapedAt:      now,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}

	if len(reviews) > 0 {
		if err := h.reviewRepo.UpsertBatch(r.Context(), reviews); err != nil {
			log.Printf("WARNING: failed to upsert reviews for app %s: %v", app.ID, err)
			writeJSONError(w, http.StatusInternalServerError, "failed to store reviews")
			return
		}
	}

	total, _ := h.reviewRepo.CountByAppID(r.Context(), app.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message":     "Scrape completed",
		"new_reviews": len(reviews),
		"total_reviews": total,
	})
}

// findApp resolves the app from the URL parameter (UUID or partner app ID)
func (h *ReviewHandler) findApp(r *http.Request) (*entity.App, error) {
	appIDStr := chi.URLParam(r, "appID")

	appID, uuidErr := uuid.Parse(appIDStr)
	if uuidErr == nil {
		return h.appRepo.FindByID(r.Context(), appID)
	}

	// Not a UUID — try as partner app ID
	partnerAccount, lookupErr := resolvePartnerAccount(r, h.partnerRepo)
	if lookupErr != nil {
		return nil, errors.New(lookupErr.message)
	}
	partnerAppID := appIDStr
	if !strings.HasPrefix(partnerAppID, "gid://") {
		partnerAppID = "gid://partners/App/" + appIDStr
	}
	return h.appRepo.FindByPartnerAppID(r.Context(), partnerAccount.ID, partnerAppID)
}

// deriveSentiment returns a sentiment label from rating
func deriveSentiment(rating int) string {
	switch {
	case rating <= 2:
		return "negative"
	case rating == 3:
		return "neutral"
	default:
		return "positive"
	}
}
