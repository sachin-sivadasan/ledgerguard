package service

import (
	"testing"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

func TestEarningsCalculator_CalculateAvailableDate(t *testing.T) {
	calc := NewEarningsCalculator()
	createdDate := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		chargeType  valueobject.ChargeType
		createdDate time.Time
		expected    time.Time
	}{
		{
			name:        "recurring charge - 7 day delay",
			chargeType:  valueobject.ChargeTypeRecurring,
			createdDate: createdDate,
			expected:    createdDate.AddDate(0, 0, 7),
		},
		{
			name:        "one-time charge - 7 day delay",
			chargeType:  valueobject.ChargeTypeOneTime,
			createdDate: createdDate,
			expected:    createdDate.AddDate(0, 0, 7),
		},
		{
			name:        "usage charge - 7 day delay",
			chargeType:  valueobject.ChargeTypeUsage,
			createdDate: createdDate,
			expected:    createdDate.AddDate(0, 0, 7),
		},
		{
			name:        "refund - immediate availability",
			chargeType:  valueobject.ChargeTypeRefund,
			createdDate: createdDate,
			expected:    createdDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateAvailableDate(tt.chargeType, tt.createdDate)
			if !result.Equal(tt.expected) {
				t.Errorf("CalculateAvailableDate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEarningsCalculator_DetermineEarningsStatus(t *testing.T) {
	calc := NewEarningsCalculator()
	availableDate := time.Date(2024, 1, 22, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		availableDate time.Time
		now           time.Time
		expected      entity.EarningsStatus
	}{
		{
			name:          "pending - before available date",
			availableDate: availableDate,
			now:           availableDate.AddDate(0, 0, -1), // 1 day before
			expected:      entity.EarningsStatusPending,
		},
		{
			name:          "available - on available date",
			availableDate: availableDate,
			now:           availableDate,
			expected:      entity.EarningsStatusAvailable,
		},
		{
			name:          "available - after available date",
			availableDate: availableDate,
			now:           availableDate.AddDate(0, 0, 5), // 5 days after
			expected:      entity.EarningsStatusAvailable,
		},
		{
			name:          "pending - hours before available",
			availableDate: availableDate,
			now:           availableDate.Add(-1 * time.Hour),
			expected:      entity.EarningsStatusPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.DetermineEarningsStatus(tt.availableDate, tt.now)
			if result != tt.expected {
				t.Errorf("DetermineEarningsStatus() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEarningsCalculator_ProcessTransaction(t *testing.T) {
	calc := NewEarningsCalculator()
	createdDate := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		chargeType     valueobject.ChargeType
		createdDate    time.Time
		now            time.Time
		expectedStatus entity.EarningsStatus
	}{
		{
			name:           "recurring - still pending",
			chargeType:     valueobject.ChargeTypeRecurring,
			createdDate:    createdDate,
			now:            createdDate.AddDate(0, 0, 3), // 3 days later
			expectedStatus: entity.EarningsStatusPending,
		},
		{
			name:           "recurring - now available",
			chargeType:     valueobject.ChargeTypeRecurring,
			createdDate:    createdDate,
			now:            createdDate.AddDate(0, 0, 10), // 10 days later
			expectedStatus: entity.EarningsStatusAvailable,
		},
		{
			name:           "one-time - now available",
			chargeType:     valueobject.ChargeTypeOneTime,
			createdDate:    createdDate,
			now:            createdDate.AddDate(0, 0, 7), // exactly 7 days
			expectedStatus: entity.EarningsStatusAvailable,
		},
		{
			name:           "refund - immediately available",
			chargeType:     valueobject.ChargeTypeRefund,
			createdDate:    createdDate,
			now:            createdDate, // same day
			expectedStatus: entity.EarningsStatusAvailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &entity.Transaction{
				ChargeType: tt.chargeType,
			}

			calc.ProcessTransaction(tx, tt.createdDate, tt.now)

			if tx.EarningsStatus != tt.expectedStatus {
				t.Errorf("ProcessTransaction() EarningsStatus = %v, want %v", tx.EarningsStatus, tt.expectedStatus)
			}

			if tx.CreatedDate != tt.createdDate {
				t.Errorf("ProcessTransaction() CreatedDate = %v, want %v", tx.CreatedDate, tt.createdDate)
			}

			expectedAvailable := calc.CalculateAvailableDate(tt.chargeType, tt.createdDate)
			if tx.AvailableDate != expectedAvailable {
				t.Errorf("ProcessTransaction() AvailableDate = %v, want %v", tx.AvailableDate, expectedAvailable)
			}
		})
	}
}

func TestEarningsCalculator_SummarizeEarnings(t *testing.T) {
	calc := NewEarningsCalculator()
	now := time.Date(2024, 1, 25, 10, 0, 0, 0, time.UTC)

	transactions := []*entity.Transaction{
		{
			NetAmountCents: 1000,
			EarningsStatus: entity.EarningsStatusPending,
			AvailableDate:  now.AddDate(0, 0, 5), // 5 days from now
		},
		{
			NetAmountCents: 2000,
			EarningsStatus: entity.EarningsStatusAvailable,
			AvailableDate:  now.AddDate(0, 0, -3), // 3 days ago
		},
		{
			NetAmountCents: 3000,
			EarningsStatus: entity.EarningsStatusAvailable,
			AvailableDate:  now.AddDate(0, 0, -10), // 10 days ago
		},
		{
			NetAmountCents: 500,
			EarningsStatus: entity.EarningsStatusPaidOut,
			AvailableDate:  now.AddDate(0, 0, -20), // 20 days ago
		},
	}

	summary := calc.SummarizeEarnings(transactions)

	if summary.PendingCents != 1000 {
		t.Errorf("PendingCents = %d, want 1000", summary.PendingCents)
	}
	if summary.AvailableCents != 5000 {
		t.Errorf("AvailableCents = %d, want 5000", summary.AvailableCents)
	}
	if summary.PaidOutCents != 500 {
		t.Errorf("PaidOutCents = %d, want 500", summary.PaidOutCents)
	}
	if summary.PendingCount != 1 {
		t.Errorf("PendingCount = %d, want 1", summary.PendingCount)
	}
	if summary.AvailableCount != 2 {
		t.Errorf("AvailableCount = %d, want 2", summary.AvailableCount)
	}
	if summary.PaidOutCount != 1 {
		t.Errorf("PaidOutCount = %d, want 1", summary.PaidOutCount)
	}
}
