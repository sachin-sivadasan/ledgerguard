package handler

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// reviewDistributionBucket is the count of reviews at a single star rating.
type reviewDistributionBucket struct {
	Rating int `json:"rating"`
	Count  int `json:"count"`
}

// reviewSentimentBreakdown counts reviews by derived sentiment label.
type reviewSentimentBreakdown struct {
	Positive int `json:"positive"`
	Neutral  int `json:"neutral"`
	Negative int `json:"negative"`
}

// reviewSummary is a single review in the report's `recent` list. JSON keys are
// aligned with the shared Flutter model (text/date, not body/review_date).
type reviewSummary struct {
	ID        string `json:"id"`
	AppID     string `json:"app_id"`
	Author    string `json:"author"`
	Rating    int    `json:"rating"`
	Text      string `json:"text"`
	Date      string `json:"date"`
	Sentiment string `json:"sentiment"`
	Location  string `json:"location"`
	TimeUsing string `json:"time_using"`
}

// reviewsReport is the full JSON contract for the Reviews report.
type reviewsReport struct {
	AvgRating    float64                    `json:"avgRating"`
	TotalReviews int                        `json:"totalReviews"`
	Distribution []reviewDistributionBucket `json:"distribution"`
	Sentiment    reviewSentimentBreakdown   `json:"sentiment"`
	Recent       []reviewSummary            `json:"recent"`
}

// maxRecentReviews caps the number of reviews returned in the report's `recent` list.
const maxRecentReviews = 12

// GetReviewsReport returns the aggregate Reviews report for an app.
// GET /api/v1/apps/{appID}/reports/reviews?format=csv
func (h *ReviewHandler) GetReviewsReport(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	reviews, err := h.reviewRepo.FindAllByAppID(r.Context(), app.ID)
	if err != nil {
		writeReviewRepoError(w, "FindAllByAppID", err)
		return
	}

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeReviewsCSV(w, reviews)
		return
	}

	report := buildReviewsReport(reviews)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("reviews: encode report: %v", err)
	}
}

// writeReviewRepoError logs a repository failure and responds 503. The review repo
// has no not-found sentinel — every error is an infrastructure failure (ADR-042).
func writeReviewRepoError(w http.ResponseWriter, op string, err error) {
	log.Printf("reviews: repo error in %s: %v", op, err)
	writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
}

// buildReviewsReport aggregates the average rating, star distribution, sentiment
// breakdown and the most-recent reviews. Reviews are assumed pre-sorted by
// review date descending (as returned by FindAllByAppID).
func buildReviewsReport(reviews []*entity.AppReview) reviewsReport {
	distByRating := map[int]int{}
	sentiment := reviewSentimentBreakdown{}
	var ratingSum int

	for _, rev := range reviews {
		ratingSum += rev.Rating
		distByRating[rev.Rating]++
		switch deriveSentiment(rev.Rating) {
		case "positive":
			sentiment.Positive++
		case "neutral":
			sentiment.Neutral++
		default:
			sentiment.Negative++
		}
	}

	avgRating := 0.0
	if len(reviews) > 0 {
		avgRating = float64(ratingSum) / float64(len(reviews))
	}

	// Always emit all five star buckets, ordered 5 → 1, count 0 when absent.
	distribution := make([]reviewDistributionBucket, 0, 5)
	for rating := 5; rating >= 1; rating-- {
		distribution = append(distribution, reviewDistributionBucket{
			Rating: rating,
			Count:  distByRating[rating],
		})
	}

	recent := make([]reviewSummary, 0, maxRecentReviews)
	for i, rev := range reviews {
		if i >= maxRecentReviews {
			break
		}
		recent = append(recent, reviewSummary{
			ID:        rev.ID.String(),
			AppID:     rev.AppID.String(),
			Author:    rev.Author,
			Rating:    rev.Rating,
			Text:      rev.Body,
			Date:      rev.ReviewDate.Format(dateLayout),
			Sentiment: deriveSentiment(rev.Rating),
			Location:  rev.Location,
			TimeUsing: rev.TimeUsing,
		})
	}

	return reviewsReport{
		AvgRating:    avgRating,
		TotalReviews: len(reviews),
		Distribution: distribution,
		Sentiment:    sentiment,
		Recent:       recent,
	}
}

// writeReviewsCSV writes ALL reviews as a CSV attachment. Uses encoding/csv so
// free-text bodies with commas/quotes/newlines stay a single column.
func writeReviewsCSV(w http.ResponseWriter, reviews []*entity.AppReview) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="reviews.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"author", "rating", "reviewDate", "location", "timeUsing", "sentiment", "source", "body"})
	for _, rev := range reviews {
		_ = cw.Write([]string{
			rev.Author,
			strconv.Itoa(rev.Rating),
			rev.ReviewDate.Format(dateLayout),
			rev.Location,
			rev.TimeUsing,
			deriveSentiment(rev.Rating),
			rev.Source,
			rev.Body,
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("reviews: write CSV: %v", err)
	}
}
