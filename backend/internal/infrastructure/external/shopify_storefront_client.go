package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

const storefrontAPIVersion = "2026-01"

// ShopifyStorefrontClient fetches public brand data from Shopify Storefront API
type ShopifyStorefrontClient struct {
	httpClient *http.Client
	baseURL    string // empty = real Shopify, set = mock/proxy server
}

func NewShopifyStorefrontClient(baseURL string) *ShopifyStorefrontClient {
	return &ShopifyStorefrontClient{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		baseURL: baseURL,
	}
}

// storefrontBrandQuery is the GraphQL query for shop brand data
const storefrontBrandQuery = `{
  shop {
    id
    name
    brand {
      logo { image { url } }
      squareLogo { image { url } }
      coverImage { image { url } }
    }
    primaryDomain { host }
    paymentSettings { countryCode currencyCode }
  }
}`

type storefrontResponse struct {
	Data struct {
		Shop struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Brand struct {
				Logo struct {
					Image *struct {
						URL string `json:"url"`
					} `json:"image"`
				} `json:"logo"`
				SquareLogo struct {
					Image *struct {
						URL string `json:"url"`
					} `json:"image"`
				} `json:"squareLogo"`
				CoverImage struct {
					Image *struct {
						URL string `json:"url"`
					} `json:"image"`
				} `json:"coverImage"`
			} `json:"brand"`
			PrimaryDomain struct {
				Host string `json:"host"`
			} `json:"primaryDomain"`
			PaymentSettings struct {
				CountryCode  string `json:"countryCode"`
				CurrencyCode string `json:"currencyCode"`
			} `json:"paymentSettings"`
		} `json:"shop"`
	} `json:"data"`
}

// FetchBrand fetches brand data for a shop from the public Storefront API.
// Returns a Shop entity populated with brand info. On failure, returns an empty Shop (no logo).
func (c *ShopifyStorefrontClient) FetchBrand(ctx context.Context, myshopifyDomain string) (*entity.Shop, error) {
	var url string
	if c.baseURL != "" {
		url = fmt.Sprintf("%s/%s/graphql.json", c.baseURL, myshopifyDomain)
	} else {
		url = fmt.Sprintf("https://%s/api/%s/graphql.json", myshopifyDomain, storefrontAPIVersion)
	}

	body, err := json.Marshal(map[string]string{
		"query": storefrontBrandQuery,
	})
	if err != nil {
		return c.emptyShop(myshopifyDomain), fmt.Errorf("marshal query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return c.emptyShop(myshopifyDomain), fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return c.emptyShop(myshopifyDomain), fmt.Errorf("storefront request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.emptyShop(myshopifyDomain), fmt.Errorf("storefront returned %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.emptyShop(myshopifyDomain), fmt.Errorf("read response: %w", err)
	}

	var result storefrontResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return c.emptyShop(myshopifyDomain), fmt.Errorf("unmarshal response: %w", err)
	}

	now := time.Now().UTC()
	shop := &entity.Shop{
		ID:              uuid.New(),
		MyshopifyDomain: myshopifyDomain,
		ShopifyShopGID:  result.Data.Shop.ID,
		ShopName:        result.Data.Shop.Name,
		PrimaryDomain:   result.Data.Shop.PrimaryDomain.Host,
		CountryCode:     result.Data.Shop.PaymentSettings.CountryCode,
		CurrencyCode:    result.Data.Shop.PaymentSettings.CurrencyCode,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if result.Data.Shop.Brand.Logo.Image != nil {
		shop.LogoURL = result.Data.Shop.Brand.Logo.Image.URL
	}
	if result.Data.Shop.Brand.SquareLogo.Image != nil {
		shop.SquareLogoURL = result.Data.Shop.Brand.SquareLogo.Image.URL
	}
	if result.Data.Shop.Brand.CoverImage.Image != nil {
		shop.CoverImageURL = result.Data.Shop.Brand.CoverImage.Image.URL
	}

	return shop, nil
}

func (c *ShopifyStorefrontClient) emptyShop(domain string) *entity.Shop {
	now := time.Now().UTC()
	return &entity.Shop{
		ID:              uuid.New(),
		MyshopifyDomain: domain,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
