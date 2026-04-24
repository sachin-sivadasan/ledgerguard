package entity

import (
	"time"

	"github.com/google/uuid"
)

// Shop represents a Shopify store with brand information
type Shop struct {
	ID              uuid.UUID
	MyshopifyDomain string
	ShopifyShopGID  string
	ShopName        string
	LogoURL         string
	SquareLogoURL   string
	CoverImageURL   string
	PrimaryDomain   string
	CountryCode     string
	CurrencyCode    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
