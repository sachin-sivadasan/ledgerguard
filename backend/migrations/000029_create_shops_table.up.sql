CREATE TABLE shops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    myshopify_domain VARCHAR(255) UNIQUE NOT NULL,
    shopify_shop_gid VARCHAR(255),
    shop_name VARCHAR(255),
    logo_url TEXT,
    square_logo_url TEXT,
    cover_image_url TEXT,
    primary_domain VARCHAR(255),
    country_code VARCHAR(10),
    currency_code VARCHAR(3),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_shops_domain ON shops(myshopify_domain);
