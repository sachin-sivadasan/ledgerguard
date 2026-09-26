package service

import "context"

// EmailMessage is a single outbound email.
type EmailMessage struct {
	To      string
	Subject string
	Body    string // plain text
}

// EmailSender delivers transactional email (invitations, notifications). It's a seam:
// the default is a no-op logger, and a real provider (SMTP, SendGrid, …) can be dropped
// in later without changing call sites. Implementations should be safe to call
// best-effort — callers log-and-continue on error rather than failing the operation.
type EmailSender interface {
	Send(ctx context.Context, msg EmailMessage) error
}
