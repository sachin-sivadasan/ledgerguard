package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/external"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// mobileStoreFetcher is the public-store data source (see external.MobileStoreClient).
type mobileStoreFetcher interface {
	AppleLookup(ctx context.Context, iosAppID, country string) (*external.MobileRatingSummary, error)
	AppleReviews(ctx context.Context, iosAppID, country string) ([]external.MobileReview, error)
	GooglePlayListing(ctx context.Context, packageName, country string) (*external.MobileRatingSummary, error)
}

// MobileReviewsHandler serves the "Mobile Ratings & Reviews" screen: the developer's iOS/
// Android app ratings + reviews from the PUBLIC store endpoints (no credentials). Reviews
// are Apple-only — Google Play does not expose review text publicly.
type MobileReviewsHandler struct {
	linksRepo   repository.MobileLinksRepository
	store       mobileStoreFetcher
	appRepo     repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
}

func NewMobileReviewsHandler(
	linksRepo repository.MobileLinksRepository,
	store mobileStoreFetcher,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *MobileReviewsHandler {
	return &MobileReviewsHandler{linksRepo: linksRepo, store: store, appRepo: appRepo, partnerRepo: partnerRepo}
}

type storeBlock struct {
	Linked           bool                    `json:"linked"`
	AppName          string                  `json:"appName"`
	IconURL          string                  `json:"iconUrl"`
	RatingValue      float64                 `json:"ratingValue"`
	RatingCount      int64                   `json:"ratingCount"`
	StoreURL         string                  `json:"storeUrl"`
	ReviewsAvailable bool                    `json:"reviewsAvailable"`
	Reviews          []external.MobileReview `json:"reviews"`
	Positive         int                     `json:"positive"`
	Neutral          int                     `json:"neutral"`
	Negative         int                     `json:"negative"`
	Error            string                  `json:"error,omitempty"`
}

type mobileReviewsResponse struct {
	IosAppID    string      `json:"iosAppId"`
	PlayPackage string      `json:"playPackage"`
	AppStore    *storeBlock `json:"appStore"`
	GooglePlay  *storeBlock `json:"googlePlay"`
}

// GetMobileReviews returns the ratings + (Apple) reviews for an app's linked mobile apps.
// GET /api/v1/apps/{appID}/mobile/reviews
func (h *MobileReviewsHandler) GetMobileReviews(w http.ResponseWriter, r *http.Request) {
	if user := middleware.UserFromContext(r.Context()); user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	links, err := h.linksRepo.FindByAppID(r.Context(), app.ID)
	if err != nil {
		log.Printf("mobile-reviews: repo error: %v", err)
		writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
		return
	}

	resp := mobileReviewsResponse{IosAppID: links.IosAppID, PlayPackage: links.PlayPackage}
	if links.IosAppID != "" {
		resp.AppStore = h.appleBlock(r.Context(), links.IosAppID)
	}
	if links.PlayPackage != "" {
		resp.GooglePlay = h.googleBlock(r.Context(), links.PlayPackage)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("mobile-reviews: encode: %v", err)
	}
}

// appleBlock fetches the App Store rating + reviews; a fetch failure fills Error rather than
// failing the whole response (the other store may still render).
func (h *MobileReviewsHandler) appleBlock(ctx context.Context, iosAppID string) *storeBlock {
	b := &storeBlock{Linked: true, ReviewsAvailable: true}
	sum, err := h.store.AppleLookup(ctx, iosAppID, "us")
	if err != nil {
		log.Printf("mobile-reviews: apple lookup %s: %v", iosAppID, err)
		b.Error = "Couldn't reach the App Store."
		return b
	}
	b.AppName, b.IconURL, b.RatingValue, b.RatingCount, b.StoreURL =
		sum.AppName, sum.IconURL, sum.RatingValue, sum.RatingCount, sum.StoreURL

	reviews, err := h.store.AppleReviews(ctx, iosAppID, "us")
	if err != nil {
		log.Printf("mobile-reviews: apple reviews %s: %v", iosAppID, err)
		return b // keep the rating; reviews just absent
	}
	b.Reviews = reviews
	b.Positive, b.Neutral, b.Negative = sentimentFromRatings(reviews)
	return b
}

func (h *MobileReviewsHandler) googleBlock(ctx context.Context, pkg string) *storeBlock {
	// Google Play doesn't expose review text publicly, so this is rating-only by design.
	b := &storeBlock{Linked: true, ReviewsAvailable: false}
	sum, err := h.store.GooglePlayListing(ctx, pkg, "US")
	if err != nil {
		log.Printf("mobile-reviews: google listing %s: %v", pkg, err)
		b.Error = "Couldn't read the Google Play listing."
		return b
	}
	b.AppName, b.IconURL, b.RatingValue, b.RatingCount, b.StoreURL =
		sum.AppName, sum.IconURL, sum.RatingValue, sum.RatingCount, sum.StoreURL
	return b
}

// sentimentFromRatings buckets reviews by star rating: 4–5 positive, 3 neutral, 1–2 negative.
func sentimentFromRatings(reviews []external.MobileReview) (pos, neu, neg int) {
	for _, rv := range reviews {
		switch {
		case rv.Rating >= 4:
			pos++
		case rv.Rating == 3:
			neu++
		default:
			neg++
		}
	}
	return pos, neu, neg
}

var (
	iosIDFromURL  = regexp.MustCompile(`id(\d+)`)
	digitsOnly    = regexp.MustCompile(`^\d+$`)
	playIDFromURL = regexp.MustCompile(`[?&]id=([a-zA-Z0-9._]+)`)
	packageFormat = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*(\.[a-zA-Z][a-zA-Z0-9_]*)+$`)
)

// parseIosAppID accepts a raw numeric id or an apps.apple.com URL.
func parseIosAppID(in string) string {
	in = strings.TrimSpace(in)
	if in == "" {
		return ""
	}
	if digitsOnly.MatchString(in) {
		return in
	}
	if m := iosIDFromURL.FindStringSubmatch(in); m != nil {
		return m[1]
	}
	return ""
}

// parsePlayPackage accepts a raw package name or a play.google.com URL.
func parsePlayPackage(in string) string {
	in = strings.TrimSpace(in)
	if in == "" {
		return ""
	}
	if m := playIDFromURL.FindStringSubmatch(in); m != nil {
		return m[1]
	}
	if packageFormat.MatchString(in) {
		return in
	}
	return ""
}

// PutMobileLinks stores the app's mobile store links (accepts raw ids or full store URLs).
// PUT /api/v1/apps/{appID}/mobile/links   body: {"appStore":"...","googlePlay":"..."}
func (h *MobileReviewsHandler) PutMobileLinks(w http.ResponseWriter, r *http.Request) {
	if user := middleware.UserFromContext(r.Context()); user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	var body struct {
		AppStore   string `json:"appStore"`
		GooglePlay string `json:"googlePlay"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	iosAppID := parseIosAppID(body.AppStore)
	playPackage := parsePlayPackage(body.GooglePlay)
	if body.AppStore != "" && iosAppID == "" {
		writeJSONError(w, http.StatusBadRequest, "couldn't read an App Store id/URL")
		return
	}
	if body.GooglePlay != "" && playPackage == "" {
		writeJSONError(w, http.StatusBadRequest, "couldn't read a Google Play package/URL")
		return
	}

	if err := h.linksRepo.Upsert(r.Context(), &entity.MobileLinks{
		AppID: app.ID, IosAppID: iosAppID, PlayPackage: playPackage,
	}); err != nil {
		log.Printf("mobile-reviews: upsert links: %v", err)
		writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"iosAppId": iosAppID, "playPackage": playPackage})
}
