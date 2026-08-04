package entity

import (
	"time"

	"github.com/google/uuid"
)

type AppReview struct {
	ID             uuid.UUID
	AppID          uuid.UUID
	SourceReviewID string // hash(author+date+body) for dedup
	Author         string
	Rating         int // 1-5
	Body           string
	ReviewDate     time.Time
	Location       string // e.g. "United Kingdom"
	TimeUsing      string // e.g. "4 months using the app"
	Source         string // e.g. "shopify_app_store"
	ScrapedAt      time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
