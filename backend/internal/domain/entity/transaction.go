package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// EarningsStatus represents the availability status of earnings
type EarningsStatus string

const (
	EarningsStatusPending   EarningsStatus = "PENDING"   // Earnings not yet available
	EarningsStatusAvailable EarningsStatus = "AVAILABLE" // Earnings ready for payout
	EarningsStatusPaidOut   EarningsStatus = "PAID_OUT"  // Earnings have been disbursed
)

type Transaction struct {
	ID              uuid.UUID
	AppID           uuid.UUID
	ShopifyGID      string // Unique Shopify transaction GID
	MyshopifyDomain string
	ShopName        string // Human-readable shop name from Shopify
	ChargeType      valueobject.ChargeType
	GrossAmountCents   int64 // What the merchant paid (from Shopify Partner API)
	ShopifyFeeCents    int64 // Revenue share deducted (0%, 15%, or 20%)
	ProcessingFeeCents int64 // Processing fee (2.9%)
	TaxOnFeesCents     int64 // Tax on Shopify's fees
	NetAmountCents     int64 // What the developer receives
	Currency           string
	TransactionDate    time.Time
	CreatedAt          time.Time
	// Earnings tracking
	CreatedDate     time.Time      // When the charge was created in Shopify
	AvailableDate   time.Time      // When earnings become available for payout
	EarningsStatus  EarningsStatus // PENDING, AVAILABLE, or PAID_OUT
}

// AmountCents returns the net amount for revenue calculations (backwards compatible)
func (t *Transaction) AmountCents() int64 {
	return t.NetAmountCents
}

// TransactionFees contains the fee breakdown for a transaction
type TransactionFees struct {
	ShopifyFeeCents    int64 // Revenue share (0%, 15%, or 20%)
	ProcessingFeeCents int64 // Processing fee (2.9%)
	TaxOnFeesCents     int64 // Tax on fees
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

// NewTransactionWithFees creates a transaction with full fee breakdown from Shopify Partner API
func NewTransactionWithFees(
	appID uuid.UUID,
	shopifyGID string,
	myshopifyDomain string,
	shopName string,
	chargeType valueobject.ChargeType,
	grossAmountCents int64,
	fees TransactionFees,
	netAmountCents int64,
	currency string,
	transactionDate time.Time,
) *Transaction {
	return &Transaction{
		ID:                 uuid.New(),
		AppID:              appID,
		ShopifyGID:         shopifyGID,
		MyshopifyDomain:    myshopifyDomain,
		ShopName:           shopName,
		ChargeType:         chargeType,
		GrossAmountCents:   grossAmountCents,
		ShopifyFeeCents:    fees.ShopifyFeeCents,
		ProcessingFeeCents: fees.ProcessingFeeCents,
		TaxOnFeesCents:     fees.TaxOnFeesCents,
		NetAmountCents:     netAmountCents,
		Currency:           currency,
		TransactionDate:    transactionDate,
		CreatedAt:          time.Now(),
	}
}

// TotalFeesCents returns the total fees deducted from gross amount
func (t *Transaction) TotalFeesCents() int64 {
	return t.ShopifyFeeCents + t.ProcessingFeeCents + t.TaxOnFeesCents
}

// HasFeeBreakdown returns true if the transaction has detailed fee breakdown
func (t *Transaction) HasFeeBreakdown() bool {
	return t.GrossAmountCents > 0 && (t.ShopifyFeeCents > 0 || t.ProcessingFeeCents > 0)
}

// IsPending returns true if earnings are not yet available
func (t *Transaction) IsPending() bool {
	return t.EarningsStatus == EarningsStatusPending
}

// IsAvailable returns true if earnings are available for payout
func (t *Transaction) IsAvailable() bool {
	return t.EarningsStatus == EarningsStatusAvailable
}

// IsPaidOut returns true if earnings have been disbursed
func (t *Transaction) IsPaidOut() bool {
	return t.EarningsStatus == EarningsStatusPaidOut
}

// SetEarningsTracking sets the earnings tracking fields
func (t *Transaction) SetEarningsTracking(createdDate, availableDate time.Time, status EarningsStatus) {
	t.CreatedDate = createdDate
	t.AvailableDate = availableDate
	t.EarningsStatus = status
}

// UpdateEarningsStatus updates status based on current time
func (t *Transaction) UpdateEarningsStatus(now time.Time) {
	if t.EarningsStatus == EarningsStatusPaidOut {
		return // Don't change if already paid out
	}
	if now.Before(t.AvailableDate) {
		t.EarningsStatus = EarningsStatusPending
	} else {
		t.EarningsStatus = EarningsStatusAvailable
	}
}
