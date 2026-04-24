package external

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ScrapedReview represents a review scraped from the Shopify App Store
type ScrapedReview struct {
	Author    string
	Rating    int
	Body      string
	Date      time.Time
	Location  string
	TimeUsing string
}

// SourceReviewID returns a deterministic dedup key for this review
func (r ScrapedReview) SourceReviewID() string {
	bodyPrefix := r.Body
	if len(bodyPrefix) > 100 {
		bodyPrefix = bodyPrefix[:100]
	}
	data := r.Author + r.Date.Format("2006-01-02") + bodyPrefix
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash[:16])
}

// ShopifyAppStoreClient scrapes reviews from the Shopify App Store
type ShopifyAppStoreClient struct {
	httpClient *http.Client
}

// NewShopifyAppStoreClient creates a new scraper client
func NewShopifyAppStoreClient() *ShopifyAppStoreClient {
	return &ShopifyAppStoreClient{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// ScrapeReviews fetches reviews from the Shopify App Store for the given app slug.
// maxPages controls how many pages to scrape (each page has ~10 reviews).
func (c *ShopifyAppStoreClient) ScrapeReviews(ctx context.Context, slug string, maxPages int) ([]ScrapedReview, error) {
	if slug == "" {
		return nil, fmt.Errorf("empty app store slug")
	}
	if maxPages <= 0 {
		maxPages = 5
	}

	var allReviews []ScrapedReview

	for page := 1; page <= maxPages; page++ {
		url := fmt.Sprintf("https://apps.shopify.com/%s/reviews?page=%d", slug, page)

		reviews, hasMore, err := c.scrapePage(ctx, url)
		if err != nil {
			log.Printf("WARNING: review scrape page %d failed for %s: %v", page, slug, err)
			break // stop on first error, return what we have
		}

		allReviews = append(allReviews, reviews...)

		if !hasMore {
			break
		}

		// Rate limit: 1 second delay between page fetches
		if page < maxPages {
			select {
			case <-ctx.Done():
				return allReviews, ctx.Err()
			case <-time.After(1 * time.Second):
			}
		}
	}

	return allReviews, nil
}

func (c *ShopifyAppStoreClient) scrapePage(ctx context.Context, url string) ([]ScrapedReview, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return nil, false, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("parse HTML: %w", err)
	}

	var reviews []ScrapedReview

	doc.Find("div[data-merchant-review]").Each(func(_ int, s *goquery.Selection) {
		review := c.parseReview(s)
		if review.Author != "" || review.Body != "" {
			reviews = append(reviews, review)
		}
	})

	// Check if there's a next page
	hasNext := doc.Find(`a[rel="next"][aria-label="Go to Next Page"]`).Length() > 0

	return reviews, hasNext, nil
}

func (c *ShopifyAppStoreClient) parseReview(s *goquery.Selection) ScrapedReview {
	var review ScrapedReview

	// Rating: from aria-label like "5 out of 5 stars"
	s.Find(`div[aria-label$="stars"][role="img"]`).Each(func(_ int, star *goquery.Selection) {
		if label, exists := star.Attr("aria-label"); exists {
			review.Rating = parseRating(label)
		}
	})

	// Date: text in .tw-text-body-xs.tw-text-fg-tertiary within the review content area
	s.Find(".tw-order-2 .tw-text-body-xs.tw-text-fg-tertiary").Each(func(_ int, dateEl *goquery.Selection) {
		dateText := strings.TrimSpace(dateEl.Text())
		review.Date = parseReviewDate(dateText)
	})

	// Body: concatenate all paragraphs within data-truncate-content-copy
	var bodyParts []string
	s.Find("div[data-truncate-content-copy] p.tw-break-words").Each(func(_ int, p *goquery.Selection) {
		text := strings.TrimSpace(p.Text())
		if text != "" {
			bodyParts = append(bodyParts, text)
		}
	})
	review.Body = strings.Join(bodyParts, "\n\n")

	// Author: from span[title] inside .tw-text-heading-xs
	s.Find(".tw-text-heading-xs span[title]").Each(func(_ int, span *goquery.Selection) {
		if title, exists := span.Attr("title"); exists {
			review.Author = strings.TrimSpace(title)
		}
	})

	// Location and TimeUsing: child divs in the sidebar (.tw-order-1)
	sidebar := s.Find(".tw-order-1")
	// Skip the first child (author name section) - Location and TimeUsing are plain divs after it
	sidebar.Children().Each(func(i int, child *goquery.Selection) {
		// Skip the author heading div
		if child.HasClass("tw-text-heading-xs") {
			return
		}
		text := strings.TrimSpace(child.Text())
		if text == "" {
			return
		}
		if strings.Contains(strings.ToLower(text), "using the app") || strings.Contains(strings.ToLower(text), "using app") {
			review.TimeUsing = text
		} else if review.Location == "" {
			review.Location = text
		}
	})

	return review
}

// parseRating extracts rating from "N out of 5 stars"
func parseRating(label string) int {
	// "5 out of 5 stars" -> 5
	parts := strings.Fields(label)
	if len(parts) >= 1 {
		n, err := strconv.Atoi(parts[0])
		if err == nil && n >= 1 && n <= 5 {
			return n
		}
	}
	return 0
}

// parseReviewDate parses dates like "February 26, 2026"
func parseReviewDate(s string) time.Time {
	s = strings.TrimSpace(s)
	layouts := []string{
		"January 2, 2006",
		"Jan 2, 2006",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
