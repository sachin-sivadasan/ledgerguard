package external

import (
	"context"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
)

// NoopTracker is a no-op implementation of EventTracker for dev/test
// when no analytics token is configured.
type NoopTracker struct{}

func NewNoopTracker() *NoopTracker {
	return &NoopTracker{}
}

func (t *NoopTracker) Track(_ context.Context, _ string, _ string, _ service.EventProperties) {}

func (t *NoopTracker) SetUserProperties(_ context.Context, _ string, _ service.EventProperties) {}

var _ service.EventTracker = (*NoopTracker)(nil)
