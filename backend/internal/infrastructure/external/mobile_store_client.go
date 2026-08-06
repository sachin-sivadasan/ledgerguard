package external

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MobileStoreClient fetches PUBLIC app-store data (ratings + reviews) with no credentials:
//   - Apple: the official, no-auth iTunes Lookup API + RSS customer-reviews feed.
//   - Google Play: the app listing's JSON-LD block (rating + count only — Google does not
//     expose review text on the public listing).
type MobileStoreClient struct {
	http *http.Client
}

func NewMobileStoreClient() *MobileStoreClient {
	return &MobileStoreClient{http: &http.Client{Timeout: 12 * time.Second}}
}

// MobileRatingSummary is the top-line rating for one store.
type MobileRatingSummary struct {
	Store       string  `json:"store"` // "app_store" | "google_play"
	AppName     string  `json:"appName"`
	IconURL     string  `json:"iconUrl"`
	RatingValue float64 `json:"ratingValue"`
	RatingCount int64   `json:"ratingCount"`
	StoreURL    string  `json:"storeUrl"`
}

// MobileReview is a single public review (Apple only).
type MobileReview struct {
	Author  string `json:"author"`
	Rating  int    `json:"rating"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Version string `json:"version"`
}

func (c *MobileStoreClient) get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	// A desktop UA keeps Google Play from returning a stripped/blocked variant.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("store request %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 8<<20))
}

// AppleLookup returns the app's rating summary from the iTunes Lookup API.
func (c *MobileStoreClient) AppleLookup(ctx context.Context, iosAppID, country string) (*MobileRatingSummary, error) {
	if country == "" {
		country = "us"
	}
	body, err := c.get(ctx, fmt.Sprintf("https://itunes.apple.com/lookup?id=%s&country=%s", iosAppID, country))
	if err != nil {
		return nil, err
	}
	var lk struct {
		Results []struct {
			TrackName         string  `json:"trackName"`
			AverageUserRating float64 `json:"averageUserRating"`
			UserRatingCount   int64   `json:"userRatingCount"`
			ArtworkURL100     string  `json:"artworkUrl100"`
			TrackViewURL      string  `json:"trackViewUrl"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &lk); err != nil {
		return nil, fmt.Errorf("parse itunes lookup: %w", err)
	}
	if len(lk.Results) == 0 {
		return nil, fmt.Errorf("apple app %s not found", iosAppID)
	}
	r := lk.Results[0]
	return &MobileRatingSummary{
		Store:       "app_store",
		AppName:     r.TrackName,
		IconURL:     r.ArtworkURL100,
		RatingValue: r.AverageUserRating,
		RatingCount: r.UserRatingCount,
		StoreURL:    r.TrackViewURL,
	}, nil
}

// AppleReviews returns recent public reviews from the RSS customer-reviews feed. The feed's
// first entry is app metadata (no rating) and is skipped.
func (c *MobileStoreClient) AppleReviews(ctx context.Context, iosAppID, country string) ([]MobileReview, error) {
	if country == "" {
		country = "us"
	}
	body, err := c.get(ctx, fmt.Sprintf("https://itunes.apple.com/%s/rss/customerreviews/id=%s/sortBy=mostRecent/json", country, iosAppID))
	if err != nil {
		return nil, err
	}
	return parseAppleReviews(body)
}

type appleLabel struct {
	Label string `json:"label"`
}

type appleReviewEntry struct {
	Author struct {
		Name appleLabel `json:"name"`
	} `json:"author"`
	Rating  *appleLabel `json:"im:rating"`
	Title   appleLabel  `json:"title"`
	Content appleLabel  `json:"content"`
	Version *appleLabel `json:"im:version"`
}

func parseAppleReviews(body []byte) ([]MobileReview, error) {
	var feed struct {
		Feed struct {
			Entry json.RawMessage `json:"entry"`
		} `json:"feed"`
	}
	if err := json.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("parse apple reviews: %w", err)
	}
	if len(feed.Feed.Entry) == 0 {
		return []MobileReview{}, nil
	}
	// "entry" is an array for multiple reviews but a single object for one — handle both.
	var entries []appleReviewEntry
	if err := json.Unmarshal(feed.Feed.Entry, &entries); err != nil {
		var single appleReviewEntry
		if err2 := json.Unmarshal(feed.Feed.Entry, &single); err2 != nil {
			return nil, fmt.Errorf("parse apple review entries: %w", err)
		}
		entries = []appleReviewEntry{single}
	}
	reviews := make([]MobileReview, 0, len(entries))
	for _, e := range entries {
		if e.Rating == nil { // the app-metadata entry
			continue
		}
		rating, _ := strconv.Atoi(e.Rating.Label)
		version := ""
		if e.Version != nil {
			version = e.Version.Label
		}
		reviews = append(reviews, MobileReview{
			Author:  e.Author.Name.Label,
			Rating:  rating,
			Title:   e.Title.Label,
			Body:    e.Content.Label,
			Version: version,
		})
	}
	return reviews, nil
}

var ldJSONBlock = regexp.MustCompile(`(?s)<script type="application/ld\+json"[^>]*>(.*?)</script>`)

// GooglePlayListing returns the rating summary from the listing's JSON-LD block (rating +
// count only; Google does not expose review text publicly).
func (c *MobileStoreClient) GooglePlayListing(ctx context.Context, packageName, country string) (*MobileRatingSummary, error) {
	if country == "" {
		country = "US"
	}
	url := fmt.Sprintf("https://play.google.com/store/apps/details?id=%s&hl=en&gl=%s", packageName, country)
	body, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}
	summary, err := parseGooglePlayLD(body)
	if err != nil {
		return nil, err
	}
	summary.StoreURL = url
	return summary, nil
}

func parseGooglePlayLD(html []byte) (*MobileRatingSummary, error) {
	for _, m := range ldJSONBlock.FindAllSubmatch(html, -1) {
		var ld struct {
			Type            string `json:"@type"`
			Name            string `json:"name"`
			Image           string `json:"image"`
			AggregateRating *struct {
				RatingValue json.Number `json:"ratingValue"`
				RatingCount json.Number `json:"ratingCount"`
			} `json:"aggregateRating"`
		}
		if err := json.Unmarshal(m[1], &ld); err != nil {
			continue
		}
		if ld.AggregateRating == nil {
			continue
		}
		rv, _ := ld.AggregateRating.RatingValue.Float64()
		rc, _ := strconv.ParseInt(strings.TrimSpace(ld.AggregateRating.RatingCount.String()), 10, 64)
		return &MobileRatingSummary{
			Store:       "google_play",
			AppName:     ld.Name,
			IconURL:     ld.Image,
			RatingValue: rv,
			RatingCount: rc,
		}, nil
	}
	return nil, fmt.Errorf("google play rating not found in listing")
}
