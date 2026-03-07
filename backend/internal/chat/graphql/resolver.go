package graphql

import (
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
)

// Resolver is the root resolver with injected dependencies.
// Each resolver delegates to existing domain services — no new business logic.
type Resolver struct {
	SubscriptionRepo      repository.SubscriptionRepository
	TransactionRepo       repository.TransactionRepository
	SnapshotRepo          repository.DailyMetricsSnapshotRepository
	AppRepo               repository.AppRepository
	PartnerAccountRepo    repository.PartnerAccountRepository
	SubscriptionEventRepo repository.SubscriptionEventRepository
	RiskEngine            *service.RiskEngine
	MetricsEngine         *service.MetricsEngine
}
