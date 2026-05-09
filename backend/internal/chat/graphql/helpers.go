package graphql

import (
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

func mapEntitySubscription(sub *entity.Subscription, now time.Time) *AppSubscription {
	s := &AppSubscription{
		ID:              sub.ID.String(),
		Domain:          sub.MyshopifyDomain,
		ShopName:        sub.ShopName,
		Status:          AppSubscriptionStatus(sub.Status),
		RiskState:       RiskState(sub.RiskState),
		MrrCents:        int(sub.MRRCents()),
		Currency:        sub.Currency,
		BillingInterval: BillingInterval(sub.BillingInterval),
	}

	if sub.PlanName != "" {
		plan := sub.PlanName
		s.Plan = &plan
	}

	if sub.LastRecurringChargeDate != nil {
		s.LastPaymentDate = sub.LastRecurringChargeDate
	}

	if sub.ExpectedNextChargeDate != nil {
		s.ExpectedNextCharge = sub.ExpectedNextChargeDate
	}

	if sub.ExpectedNextChargeDate != nil && now.After(*sub.ExpectedNextChargeDate) {
		days := int(now.Sub(*sub.ExpectedNextChargeDate).Hours() / 24)
		s.DaysPastDue = &days
	}

	return s
}

func mapEntityTransaction(tx *entity.Transaction) *Transaction {
	t := &Transaction{
		ID:               tx.ID.String(),
		Domain:           tx.MyshopifyDomain,
		ShopName:         tx.ShopName,
		ChargeType:       ChargeType(tx.ChargeType),
		GrossAmountCents: int(tx.GrossAmountCents),
		NetAmountCents:   int(tx.AmountCents()),
		Currency:         tx.Currency,
		TransactionDate:  tx.CreatedDate,
	}
	if tx.SubscriptionGID != "" {
		t.SubscriptionGid = &tx.SubscriptionGID
	}
	if tx.BillingInterval != "" {
		bi := BillingInterval(tx.BillingInterval)
		t.BillingInterval = &bi
	}
	return t
}

func derefInt(p *int, def int) int {
	if p != nil {
		return *p
	}
	return def
}
