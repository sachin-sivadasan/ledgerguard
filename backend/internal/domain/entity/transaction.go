package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

type Transaction struct {
	ID              uuid.UUID
	AppID           uuid.UUID
	ShopifyGID      string // Unique Shopify transaction GID
	MyshopifyDomain string
	ShopName        string // Human-readable shop name from Shopify
	ChargeType      valueobject.ChargeType
	GrossAmountCents int64  // Subscription price (what customer pays)
	NetAmountCents   int64  // Revenue (what you receive after Shopify's cut)
	Currency        string
	TransactionDate time.Time
	CreatedAt       time.Time
}

// AmountCents returns the net amount for revenue calculations (backwards compatible)
func (t *Transaction) AmountCents() int64 {
	return t.NetAmountCents
}

func NewTransaction(
	appID uuid.UUID,
	shopifyGID string,
	myshopifyDomain string,
	shopName string,
	chargeType valueobject.ChargeType,
	grossAmountCents int64,
	netAmountCents int64,
	currency string,
	transactionDate time.Time,
) *Transaction {
	return &Transaction{
		ID:               uuid.New(),
		AppID:            appID,
		ShopifyGID:       shopifyGID,
		MyshopifyDomain:  myshopifyDomain,
		ShopName:         shopName,
		ChargeType:       chargeType,
		GrossAmountCents: grossAmountCents,
		NetAmountCents:   netAmountCents,
		Currency:         currency,
		TransactionDate:  transactionDate,
		CreatedAt:        time.Now(),
	}
}
