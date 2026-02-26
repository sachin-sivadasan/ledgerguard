package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
)

// TransactionFetcher interface for fetching transactions from external API
type TransactionFetcher interface {
	FetchTransactions(ctx context.Context, accessToken string, appID uuid.UUID, from, to time.Time) ([]*entity.Transaction, error)
}

// Decryptor interface for decrypting tokens
type Decryptor interface {
	Decrypt(ciphertext []byte) ([]byte, error)
}

// SyncResult contains the result of a sync operation
type SyncResult struct {
	AppID            uuid.UUID
	AppName          string
	TransactionCount int
	SyncedAt         time.Time
	Error            error
}

// SyncService handles synchronization of transactions from Partner API
type SyncService struct {
	fetcher     TransactionFetcher
	txRepo      repository.TransactionRepository
	appRepo     repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
	decryptor   Decryptor
}

func NewSyncService(
	fetcher TransactionFetcher,
	txRepo repository.TransactionRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
	decryptor Decryptor,
) *SyncService {
	return &SyncService{
		fetcher:     fetcher,
		txRepo:      txRepo,
		appRepo:     appRepo,
		partnerRepo: partnerRepo,
		decryptor:   decryptor,
	}
}

// SyncApp synchronizes transactions for a single app
func (s *SyncService) SyncApp(ctx context.Context, appID uuid.UUID) (*SyncResult, error) {
	// Get app
	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to find app: %w", err)
	}

	// Get partner account for the app
	partnerAccount, err := s.getPartnerAccountForApp(ctx, app.PartnerAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner account: %w", err)
	}

	// Decrypt access token
	accessToken, err := s.decryptor.Decrypt(partnerAccount.EncryptedAccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	// Calculate 12-month window
	now := time.Now().UTC()
	from := now.AddDate(-1, 0, 0) // 12 months ago
	to := now

	// Fetch transactions from Partner API
	transactions, err := s.fetcher.FetchTransactions(ctx, string(accessToken), appID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	// Store transactions (upsert for idempotency)
	if len(transactions) > 0 {
		if err := s.txRepo.UpsertBatch(ctx, transactions); err != nil {
			return nil, fmt.Errorf("failed to store transactions: %w", err)
		}
	}

	return &SyncResult{
		AppID:            appID,
		AppName:          app.Name,
		TransactionCount: len(transactions),
		SyncedAt:         now,
	}, nil
}

// SyncAllApps synchronizes transactions for all apps of a partner account
func (s *SyncService) SyncAllApps(ctx context.Context, partnerAccountID uuid.UUID) ([]*SyncResult, error) {
	// Get all apps for the partner account
	apps, err := s.appRepo.FindByPartnerAccountID(ctx, partnerAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to find apps: %w", err)
	}

	var results []*SyncResult

	for _, app := range apps {
		if !app.TrackingEnabled {
			continue
		}

		result, err := s.SyncApp(ctx, app.ID)
		if err != nil {
			results = append(results, &SyncResult{
				AppID:    app.ID,
				AppName:  app.Name,
				SyncedAt: time.Now().UTC(),
				Error:    err,
			})
			continue
		}

		results = append(results, result)
	}

	return results, nil
}

func (s *SyncService) getPartnerAccountForApp(ctx context.Context, partnerAccountID uuid.UUID) (*entity.PartnerAccount, error) {
	return s.partnerRepo.FindByID(ctx, partnerAccountID)
}
