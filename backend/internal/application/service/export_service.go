package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
)

// ExportFormat represents the output format for exports
type ExportFormat string

const (
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatJSON ExportFormat = "json"
)

// ExportType represents the type of data being exported
type ExportType string

const (
	ExportTypeTransactions  ExportType = "transactions"
	ExportTypeSubscriptions ExportType = "subscriptions"
	ExportTypeMetrics       ExportType = "metrics"
)

// ExportResult contains the exported data and metadata
type ExportResult struct {
	Data        []byte
	ContentType string
	Filename    string
	RecordCount int
}

// ExportService handles data export operations
type ExportService struct {
	transactionRepo  repository.TransactionRepository
	subscriptionRepo repository.SubscriptionRepository
	metricsRepo      repository.DailyMetricsSnapshotRepository
}

// NewExportService creates a new export service
func NewExportService(
	transactionRepo repository.TransactionRepository,
	subscriptionRepo repository.SubscriptionRepository,
	metricsRepo repository.DailyMetricsSnapshotRepository,
) *ExportService {
	return &ExportService{
		transactionRepo:  transactionRepo,
		subscriptionRepo: subscriptionRepo,
		metricsRepo:      metricsRepo,
	}
}

// ExportTransactions exports transactions for an app within a date range
func (s *ExportService) ExportTransactions(
	ctx context.Context,
	appID uuid.UUID,
	from, to time.Time,
	format ExportFormat,
) (*ExportResult, error) {
	transactions, err := s.transactionRepo.FindByAppID(ctx, appID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	var data []byte
	var contentType string
	filename := fmt.Sprintf("transactions_%s_%s_to_%s",
		appID.String()[:8],
		from.Format("2006-01-02"),
		to.Format("2006-01-02"),
	)

	switch format {
	case ExportFormatCSV:
		data, err = s.transactionsToCSV(transactions)
		contentType = "text/csv"
		filename += ".csv"
	case ExportFormatJSON:
		data, err = s.transactionsToJSON(transactions)
		contentType = "application/json"
		filename += ".json"
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}

	if err != nil {
		return nil, err
	}

	return &ExportResult{
		Data:        data,
		ContentType: contentType,
		Filename:    filename,
		RecordCount: len(transactions),
	}, nil
}

// ExportSubscriptions exports subscriptions for an app
func (s *ExportService) ExportSubscriptions(
	ctx context.Context,
	appID uuid.UUID,
	format ExportFormat,
) (*ExportResult, error) {
	subscriptions, err := s.subscriptionRepo.FindByAppID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscriptions: %w", err)
	}

	var data []byte
	var contentType string
	filename := fmt.Sprintf("subscriptions_%s_%s",
		appID.String()[:8],
		time.Now().Format("2006-01-02"),
	)

	switch format {
	case ExportFormatCSV:
		data, err = s.subscriptionsToCSV(subscriptions)
		contentType = "text/csv"
		filename += ".csv"
	case ExportFormatJSON:
		data, err = s.subscriptionsToJSON(subscriptions)
		contentType = "application/json"
		filename += ".json"
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}

	if err != nil {
		return nil, err
	}

	return &ExportResult{
		Data:        data,
		ContentType: contentType,
		Filename:    filename,
		RecordCount: len(subscriptions),
	}, nil
}

// ExportMetrics exports daily metrics snapshots for an app
func (s *ExportService) ExportMetrics(
	ctx context.Context,
	appID uuid.UUID,
	from, to time.Time,
	format ExportFormat,
) (*ExportResult, error) {
	metrics, err := s.metricsRepo.FindByAppIDRange(ctx, appID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}

	var data []byte
	var contentType string
	filename := fmt.Sprintf("metrics_%s_%s_to_%s",
		appID.String()[:8],
		from.Format("2006-01-02"),
		to.Format("2006-01-02"),
	)

	switch format {
	case ExportFormatCSV:
		data, err = s.metricsToCSV(metrics)
		contentType = "text/csv"
		filename += ".csv"
	case ExportFormatJSON:
		data, err = s.metricsToJSON(metrics)
		contentType = "application/json"
		filename += ".json"
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}

	if err != nil {
		return nil, err
	}

	return &ExportResult{
		Data:        data,
		ContentType: contentType,
		Filename:    filename,
		RecordCount: len(metrics),
	}, nil
}

// transactionsToCSV converts transactions to CSV format
func (s *ExportService) transactionsToCSV(transactions []*entity.Transaction) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"id",
		"shopify_gid",
		"shop_domain",
		"shop_name",
		"charge_type",
		"gross_amount",
		"shopify_fee",
		"processing_fee",
		"net_amount",
		"currency",
		"transaction_date",
		"earnings_status",
		"available_date",
	}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, tx := range transactions {
		row := []string{
			tx.ID.String(),
			tx.ShopifyGID,
			tx.MyshopifyDomain,
			tx.ShopName,
			string(tx.ChargeType),
			fmt.Sprintf("%.2f", float64(tx.GrossAmountCents)/100),
			fmt.Sprintf("%.2f", float64(tx.ShopifyFeeCents)/100),
			fmt.Sprintf("%.2f", float64(tx.ProcessingFeeCents)/100),
			fmt.Sprintf("%.2f", float64(tx.NetAmountCents)/100),
			tx.Currency,
			tx.TransactionDate.Format(time.RFC3339),
			string(tx.EarningsStatus),
			tx.AvailableDate.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// transactionExportRow represents a transaction for JSON export
type transactionExportRow struct {
	ID              string  `json:"id"`
	ShopifyGID      string  `json:"shopify_gid"`
	ShopDomain      string  `json:"shop_domain"`
	ShopName        string  `json:"shop_name"`
	ChargeType      string  `json:"charge_type"`
	GrossAmount     float64 `json:"gross_amount"`
	ShopifyFee      float64 `json:"shopify_fee"`
	ProcessingFee   float64 `json:"processing_fee"`
	NetAmount       float64 `json:"net_amount"`
	Currency        string  `json:"currency"`
	TransactionDate string  `json:"transaction_date"`
	EarningsStatus  string  `json:"earnings_status"`
	AvailableDate   string  `json:"available_date"`
}

// transactionsToJSON converts transactions to JSON format
func (s *ExportService) transactionsToJSON(transactions []*entity.Transaction) ([]byte, error) {
	rows := make([]transactionExportRow, len(transactions))
	for i, tx := range transactions {
		rows[i] = transactionExportRow{
			ID:              tx.ID.String(),
			ShopifyGID:      tx.ShopifyGID,
			ShopDomain:      tx.MyshopifyDomain,
			ShopName:        tx.ShopName,
			ChargeType:      string(tx.ChargeType),
			GrossAmount:     float64(tx.GrossAmountCents) / 100,
			ShopifyFee:      float64(tx.ShopifyFeeCents) / 100,
			ProcessingFee:   float64(tx.ProcessingFeeCents) / 100,
			NetAmount:       float64(tx.NetAmountCents) / 100,
			Currency:        tx.Currency,
			TransactionDate: tx.TransactionDate.Format(time.RFC3339),
			EarningsStatus:  string(tx.EarningsStatus),
			AvailableDate:   tx.AvailableDate.Format(time.RFC3339),
		}
	}

	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return data, nil
}

// subscriptionsToCSV converts subscriptions to CSV format
func (s *ExportService) subscriptionsToCSV(subscriptions []*entity.Subscription) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"id",
		"shopify_gid",
		"shop_domain",
		"shop_name",
		"plan_name",
		"base_price",
		"currency",
		"billing_interval",
		"status",
		"risk_state",
		"mrr",
		"last_charge_date",
		"next_charge_date",
		"created_at",
	}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, sub := range subscriptions {
		lastCharge := ""
		if sub.LastRecurringChargeDate != nil {
			lastCharge = sub.LastRecurringChargeDate.Format(time.RFC3339)
		}
		nextCharge := ""
		if sub.ExpectedNextChargeDate != nil {
			nextCharge = sub.ExpectedNextChargeDate.Format(time.RFC3339)
		}

		row := []string{
			sub.ID.String(),
			sub.ShopifyGID,
			sub.MyshopifyDomain,
			sub.ShopName,
			sub.PlanName,
			fmt.Sprintf("%.2f", float64(sub.BasePriceCents)/100),
			sub.Currency,
			string(sub.BillingInterval),
			sub.Status,
			string(sub.RiskState),
			fmt.Sprintf("%.2f", float64(sub.MRRCents())/100),
			lastCharge,
			nextCharge,
			sub.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// subscriptionExportRow represents a subscription for JSON export
type subscriptionExportRow struct {
	ID              string   `json:"id"`
	ShopifyGID      string   `json:"shopify_gid"`
	ShopDomain      string   `json:"shop_domain"`
	ShopName        string   `json:"shop_name"`
	PlanName        string   `json:"plan_name"`
	BasePrice       float64  `json:"base_price"`
	Currency        string   `json:"currency"`
	BillingInterval string   `json:"billing_interval"`
	Status          string   `json:"status"`
	RiskState       string   `json:"risk_state"`
	MRR             float64  `json:"mrr"`
	LastChargeDate  *string  `json:"last_charge_date"`
	NextChargeDate  *string  `json:"next_charge_date"`
	CreatedAt       string   `json:"created_at"`
}

// subscriptionsToJSON converts subscriptions to JSON format
func (s *ExportService) subscriptionsToJSON(subscriptions []*entity.Subscription) ([]byte, error) {
	rows := make([]subscriptionExportRow, len(subscriptions))
	for i, sub := range subscriptions {
		var lastCharge, nextCharge *string
		if sub.LastRecurringChargeDate != nil {
			s := sub.LastRecurringChargeDate.Format(time.RFC3339)
			lastCharge = &s
		}
		if sub.ExpectedNextChargeDate != nil {
			s := sub.ExpectedNextChargeDate.Format(time.RFC3339)
			nextCharge = &s
		}

		rows[i] = subscriptionExportRow{
			ID:              sub.ID.String(),
			ShopifyGID:      sub.ShopifyGID,
			ShopDomain:      sub.MyshopifyDomain,
			ShopName:        sub.ShopName,
			PlanName:        sub.PlanName,
			BasePrice:       float64(sub.BasePriceCents) / 100,
			Currency:        sub.Currency,
			BillingInterval: string(sub.BillingInterval),
			Status:          sub.Status,
			RiskState:       string(sub.RiskState),
			MRR:             float64(sub.MRRCents()) / 100,
			LastChargeDate:  lastCharge,
			NextChargeDate:  nextCharge,
			CreatedAt:       sub.CreatedAt.Format(time.RFC3339),
		}
	}

	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return data, nil
}

// metricsToCSV converts daily metrics to CSV format
func (s *ExportService) metricsToCSV(metrics []*entity.DailyMetricsSnapshot) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"date",
		"mrr",
		"total_revenue",
		"usage_revenue",
		"revenue_at_risk",
		"renewal_rate",
		"safe_count",
		"one_cycle_missed",
		"two_cycle_missed",
		"churned_count",
	}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, m := range metrics {
		row := []string{
			m.Date.Format("2006-01-02"),
			fmt.Sprintf("%.2f", float64(m.ActiveMRRCents)/100),
			fmt.Sprintf("%.2f", float64(m.TotalRevenueCents)/100),
			fmt.Sprintf("%.2f", float64(m.UsageRevenueCents)/100),
			fmt.Sprintf("%.2f", float64(m.RevenueAtRiskCents)/100),
			fmt.Sprintf("%.2f", m.RenewalSuccessRate),
			fmt.Sprintf("%d", m.SafeCount),
			fmt.Sprintf("%d", m.OneCycleMissedCount),
			fmt.Sprintf("%d", m.TwoCyclesMissedCount),
			fmt.Sprintf("%d", m.ChurnedCount),
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// metricsExportRow represents a daily metrics snapshot for JSON export
type metricsExportRow struct {
	Date             string  `json:"date"`
	MRR              float64 `json:"mrr"`
	TotalRevenue     float64 `json:"total_revenue"`
	UsageRevenue     float64 `json:"usage_revenue"`
	RevenueAtRisk    float64 `json:"revenue_at_risk"`
	RenewalRate      float64 `json:"renewal_rate"`
	SafeCount        int     `json:"safe_count"`
	OneCycleMissed   int     `json:"one_cycle_missed"`
	TwoCycleMissed   int     `json:"two_cycle_missed"`
	ChurnedCount     int     `json:"churned_count"`
}

// metricsToJSON converts daily metrics to JSON format
func (s *ExportService) metricsToJSON(metrics []*entity.DailyMetricsSnapshot) ([]byte, error) {
	rows := make([]metricsExportRow, len(metrics))
	for i, m := range metrics {
		rows[i] = metricsExportRow{
			Date:           m.Date.Format("2006-01-02"),
			MRR:            float64(m.ActiveMRRCents) / 100,
			TotalRevenue:   float64(m.TotalRevenueCents) / 100,
			UsageRevenue:   float64(m.UsageRevenueCents) / 100,
			RevenueAtRisk:  float64(m.RevenueAtRiskCents) / 100,
			RenewalRate:    m.RenewalSuccessRate,
			SafeCount:      m.SafeCount,
			OneCycleMissed: m.OneCycleMissedCount,
			TwoCycleMissed: m.TwoCyclesMissedCount,
			ChurnedCount:   m.ChurnedCount,
		}
	}

	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return data, nil
}
