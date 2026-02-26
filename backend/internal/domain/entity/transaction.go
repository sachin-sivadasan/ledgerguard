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
	ChargeType      valueobject.ChargeType
	AmountCents     int64
	Currency        string
	TransactionDate time.Time
	CreatedAt       time.Time
}

func NewTransaction(
	appID uuid.UUID,
	shopifyGID string,
	myshopifyDomain string,
	chargeType valueobject.ChargeType,
	amountCents int64,
	currency string,
	transactionDate time.Time,
) *Transaction {
	return &Transaction{
		ID:              uuid.New(),
		AppID:           appID,
		ShopifyGID:      shopifyGID,
		MyshopifyDomain: myshopifyDomain,
		ChargeType:      chargeType,
		AmountCents:     amountCents,
		Currency:        currency,
		TransactionDate: transactionDate,
		CreatedAt:       time.Now(),
	}
}
