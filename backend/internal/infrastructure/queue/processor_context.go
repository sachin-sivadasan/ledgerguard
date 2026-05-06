package queue

import (
	"context"
	"fmt"
	"strings"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
)

// Decryptor interface for decrypting access tokens
type Decryptor interface {
	Decrypt(ciphertext []byte) ([]byte, error)
}

// ProcessorContext contains the shared context prepared for every processor
type ProcessorContext struct {
	App            *entity.App
	PartnerAccount *entity.PartnerAccount
	AccessToken    string
	OrganizationID string
}

// PrepareProcessorContext performs the common preamble for all non-orchestrator processors:
// app lookup → partner lookup → decrypt token
func PrepareProcessorContext(
	ctx context.Context,
	payload *SyncJobPayload,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
	decryptor Decryptor,
) (*ProcessorContext, error) {
	// Look up the app
	app, err := appRepo.FindByID(ctx, payload.AppID)
	if err != nil {
		return nil, fmt.Errorf("failed to find app %s: %w", payload.AppID, err)
	}

	// Look up the partner account
	partnerAccount, err := partnerRepo.FindByID(ctx, payload.PartnerAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to find partner account %s: %w", payload.PartnerAccountID, err)
	}

	// Decrypt the access token
	tokenBytes, err := decryptor.Decrypt(partnerAccount.EncryptedAccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	// Normalize GID casing (Shopify requires "App" not "app")
	if strings.Contains(app.PartnerAppID, "gid://partners/app/") {
		app.PartnerAppID = strings.Replace(app.PartnerAppID, "gid://partners/app/", "gid://partners/App/", 1)
	}

	return &ProcessorContext{
		App:            app,
		PartnerAccount: partnerAccount,
		AccessToken:    string(tokenBytes),
		OrganizationID: partnerAccount.PartnerID,
	}, nil
}
