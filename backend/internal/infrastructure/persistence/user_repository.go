package persistence

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) FindByFirebaseUID(ctx context.Context, firebaseUID string) (*entity.User, error) {
	query := `
		SELECT id, firebase_uid, email, role, plan_tier, created_at
		FROM users
		WHERE firebase_uid = $1
	`

	var user entity.User
	var role string
	var planTier string

	err := r.pool.QueryRow(ctx, query, firebaseUID).Scan(
		&user.ID,
		&user.FirebaseUID,
		&user.Email,
		&role,
		&planTier,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, middleware.ErrUserNotFound
		}
		return nil, err
	}

	user.Role = valueobject.Role(role)
	user.PlanTier = valueobject.PlanTier(planTier)

	return &user, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, firebase_uid, email, role, plan_tier, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.FirebaseUID,
		user.Email,
		user.Role.String(),
		user.PlanTier.String(),
		user.CreatedAt,
	)

	return err
}
