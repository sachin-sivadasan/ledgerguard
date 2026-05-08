package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

type PartnerAccount struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	OrgID                uuid.UUID
	IntegrationType      valueobject.IntegrationType
	PartnerID            string
	EncryptedAccessToken []byte
	CreatedAt            time.Time
}

func NewPartnerAccount(userID, orgID uuid.UUID, partnerID string, integrationType valueobject.IntegrationType, encryptedToken []byte) *PartnerAccount {
	return &PartnerAccount{
		ID:                   uuid.New(),
		UserID:               userID,
		OrgID:                orgID,
		IntegrationType:      integrationType,
		PartnerID:            partnerID,
		EncryptedAccessToken: encryptedToken,
		CreatedAt:            time.Now().UTC(),
	}
}
