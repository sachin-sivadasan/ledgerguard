package external

import (
	"context"
	"log"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
)

// NoopEmailSender is the default EmailSender: it logs what would be sent instead of
// delivering. Swap for a real provider (SMTP/SendGrid) by implementing service.EmailSender
// and wiring it in main — no call sites change.
type NoopEmailSender struct{}

func NewNoopEmailSender() *NoopEmailSender {
	return &NoopEmailSender{}
}

func (s *NoopEmailSender) Send(_ context.Context, msg service.EmailMessage) error {
	log.Printf("email(noop): would send to=%s subject=%q (no provider configured)", msg.To, msg.Subject)
	return nil
}

var _ service.EmailSender = (*NoopEmailSender)(nil)
