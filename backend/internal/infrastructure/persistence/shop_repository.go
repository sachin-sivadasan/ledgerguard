package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

var ErrShopNotFound = errors.New("shop not found")

type PostgresShopRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresShopRepository(pool *pgxpool.Pool) *PostgresShopRepository {
	return &PostgresShopRepository{pool: pool}
}

func (r *PostgresShopRepository) Upsert(ctx context.Context, shop *entity.Shop) error {
	query := `
		INSERT INTO shops (
			id, myshopify_domain, shopify_shop_gid, shop_name,
			logo_url, square_logo_url, cover_image_url,
			primary_domain, country_code, currency_code,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (myshopify_domain) DO UPDATE SET
			shopify_shop_gid = EXCLUDED.shopify_shop_gid,
			shop_name = EXCLUDED.shop_name,
			logo_url = EXCLUDED.logo_url,
			square_logo_url = EXCLUDED.square_logo_url,
			cover_image_url = EXCLUDED.cover_image_url,
			primary_domain = EXCLUDED.primary_domain,
			country_code = EXCLUDED.country_code,
			currency_code = EXCLUDED.currency_code,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		shop.ID,
		shop.MyshopifyDomain,
		shop.ShopifyShopGID,
		shop.ShopName,
		shop.LogoURL,
		shop.SquareLogoURL,
		shop.CoverImageURL,
		shop.PrimaryDomain,
		shop.CountryCode,
		shop.CurrencyCode,
		shop.CreatedAt,
		shop.UpdatedAt,
	)

	return err
}

func (r *PostgresShopRepository) FindByDomain(ctx context.Context, domain string) (*entity.Shop, error) {
	query := `
		SELECT id, myshopify_domain, shopify_shop_gid, shop_name,
			logo_url, square_logo_url, cover_image_url,
			primary_domain, country_code, currency_code,
			created_at, updated_at
		FROM shops
		WHERE myshopify_domain = $1
	`

	return r.scanShop(r.pool.QueryRow(ctx, query, domain))
}

func (r *PostgresShopRepository) FindByDomains(ctx context.Context, domains []string) (map[string]*entity.Shop, error) {
	if len(domains) == 0 {
		return make(map[string]*entity.Shop), nil
	}

	// Build placeholders
	placeholders := make([]string, len(domains))
	args := make([]interface{}, len(domains))
	for i, d := range domains {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = d
	}

	query := fmt.Sprintf(`
		SELECT id, myshopify_domain, shopify_shop_gid, shop_name,
			logo_url, square_logo_url, cover_image_url,
			primary_domain, country_code, currency_code,
			created_at, updated_at
		FROM shops
		WHERE myshopify_domain IN (%s)
	`, strings.Join(placeholders, ", "))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]*entity.Shop)
	for rows.Next() {
		shop, err := r.scanShopFromRow(rows)
		if err != nil {
			return nil, err
		}
		result[shop.MyshopifyDomain] = shop
	}

	return result, rows.Err()
}

func (r *PostgresShopRepository) scanShop(row pgx.Row) (*entity.Shop, error) {
	var shop entity.Shop
	var shopGID, shopName, logoURL, squareLogoURL, coverImageURL, primaryDomain, countryCode, currencyCode *string

	err := row.Scan(
		&shop.ID,
		&shop.MyshopifyDomain,
		&shopGID,
		&shopName,
		&logoURL,
		&squareLogoURL,
		&coverImageURL,
		&primaryDomain,
		&countryCode,
		&currencyCode,
		&shop.CreatedAt,
		&shop.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}

	if shopGID != nil {
		shop.ShopifyShopGID = *shopGID
	}
	if shopName != nil {
		shop.ShopName = *shopName
	}
	if logoURL != nil {
		shop.LogoURL = *logoURL
	}
	if squareLogoURL != nil {
		shop.SquareLogoURL = *squareLogoURL
	}
	if coverImageURL != nil {
		shop.CoverImageURL = *coverImageURL
	}
	if primaryDomain != nil {
		shop.PrimaryDomain = *primaryDomain
	}
	if countryCode != nil {
		shop.CountryCode = *countryCode
	}
	if currencyCode != nil {
		shop.CurrencyCode = *currencyCode
	}

	return &shop, nil
}

func (r *PostgresShopRepository) scanShopFromRow(rows pgx.Rows) (*entity.Shop, error) {
	var shop entity.Shop
	var shopGID, shopName, logoURL, squareLogoURL, coverImageURL, primaryDomain, countryCode, currencyCode *string

	err := rows.Scan(
		&shop.ID,
		&shop.MyshopifyDomain,
		&shopGID,
		&shopName,
		&logoURL,
		&squareLogoURL,
		&coverImageURL,
		&primaryDomain,
		&countryCode,
		&currencyCode,
		&shop.CreatedAt,
		&shop.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if shopGID != nil {
		shop.ShopifyShopGID = *shopGID
	}
	if shopName != nil {
		shop.ShopName = *shopName
	}
	if logoURL != nil {
		shop.LogoURL = *logoURL
	}
	if squareLogoURL != nil {
		shop.SquareLogoURL = *squareLogoURL
	}
	if coverImageURL != nil {
		shop.CoverImageURL = *coverImageURL
	}
	if primaryDomain != nil {
		shop.PrimaryDomain = *primaryDomain
	}
	if countryCode != nil {
		shop.CountryCode = *countryCode
	}
	if currencyCode != nil {
		shop.CurrencyCode = *currencyCode
	}

	return &shop, nil
}
