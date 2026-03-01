package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

type PostgresSubscriptionEventRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresSubscriptionEventRepository(pool *pgxpool.Pool) *PostgresSubscriptionEventRepository {
	return &PostgresSubscriptionEventRepository{pool: pool}
}

func (r *PostgresSubscriptionEventRepository) Create(ctx context.Context, event *entity.SubscriptionEvent) error {
	query := `
		INSERT INTO subscription_events (
			id, subscription_id, from_status, to_status,
			from_risk_state, to_risk_state, event_type, reason,
			occurred_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(ctx, query,
		event.ID,
		event.SubscriptionID,
		event.FromStatus,
		event.ToStatus,
		string(event.FromRiskState),
		string(event.ToRiskState),
		event.EventType,
		event.Reason,
		event.OccurredAt,
		event.CreatedAt,
	)

	return err
}

func (r *PostgresSubscriptionEventRepository) FindBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entity.SubscriptionEvent, error) {
	query := `
		SELECT id, subscription_id, from_status, to_status,
		       from_risk_state, to_risk_state, event_type, reason,
		       occurred_at, created_at
		FROM subscription_events
		WHERE subscription_id = $1
		ORDER BY occurred_at DESC
	`

	rows, err := r.pool.Query(ctx, query, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

func (r *PostgresSubscriptionEventRepository) FindByAppID(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.SubscriptionEvent, error) {
	query := `
		SELECT se.id, se.subscription_id, se.from_status, se.to_status,
		       se.from_risk_state, se.to_risk_state, se.event_type, se.reason,
		       se.occurred_at, se.created_at
		FROM subscription_events se
		JOIN subscriptions s ON se.subscription_id = s.id
		WHERE s.app_id = $1 AND se.occurred_at >= $2 AND se.occurred_at <= $3
		ORDER BY se.occurred_at DESC
	`

	rows, err := r.pool.Query(ctx, query, appID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

func (r *PostgresSubscriptionEventRepository) FindChurnEvents(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.SubscriptionEvent, error) {
	query := `
		SELECT se.id, se.subscription_id, se.from_status, se.to_status,
		       se.from_risk_state, se.to_risk_state, se.event_type, se.reason,
		       se.occurred_at, se.created_at
		FROM subscription_events se
		JOIN subscriptions s ON se.subscription_id = s.id
		WHERE s.app_id = $1
		  AND se.to_risk_state = 'CHURNED'
		  AND se.from_risk_state != 'CHURNED'
		  AND se.occurred_at >= $2 AND se.occurred_at <= $3
		ORDER BY se.occurred_at DESC
	`

	rows, err := r.pool.Query(ctx, query, appID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

func (r *PostgresSubscriptionEventRepository) CountByEventType(ctx context.Context, appID uuid.UUID, from, to time.Time) (map[string]int, error) {
	query := `
		SELECT se.event_type, COUNT(*) as count
		FROM subscription_events se
		JOIN subscriptions s ON se.subscription_id = s.id
		WHERE s.app_id = $1 AND se.occurred_at >= $2 AND se.occurred_at <= $3
		GROUP BY se.event_type
	`

	rows, err := r.pool.Query(ctx, query, appID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var eventType string
		var count int
		if err := rows.Scan(&eventType, &count); err != nil {
			return nil, err
		}
		counts[eventType] = count
	}

	return counts, rows.Err()
}

type eventRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func (r *PostgresSubscriptionEventRepository) scanEvents(rows eventRows) ([]*entity.SubscriptionEvent, error) {
	var events []*entity.SubscriptionEvent
	for rows.Next() {
		var event entity.SubscriptionEvent
		var fromRiskState, toRiskState string
		err := rows.Scan(
			&event.ID,
			&event.SubscriptionID,
			&event.FromStatus,
			&event.ToStatus,
			&fromRiskState,
			&toRiskState,
			&event.EventType,
			&event.Reason,
			&event.OccurredAt,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		event.FromRiskState = valueobject.ParseRiskState(fromRiskState)
		event.ToRiskState = valueobject.ParseRiskState(toRiskState)
		events = append(events, &event)
	}

	return events, rows.Err()
}
