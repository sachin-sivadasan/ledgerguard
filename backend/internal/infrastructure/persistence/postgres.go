package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDB struct {
	Pool *pgxpool.Pool
}

func NewPostgresDB(ctx context.Context, dsn string) (*PostgresDB, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &PostgresDB{Pool: pool}, nil
}

func (db *PostgresDB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

func (db *PostgresDB) Close() {
	db.Pool.Close()
}

// MigrationStatus reports the current schema version and dirty flag from the
// golang-migrate `schema_migrations` table. It returns an error if the table
// does not exist (migrations were never applied) — used by the health check to
// surface a half-migrated / broken schema instead of falsely reporting healthy.
func (db *PostgresDB) MigrationStatus(ctx context.Context) (version int64, dirty bool, err error) {
	err = db.Pool.QueryRow(ctx, "SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&version, &dirty)
	return version, dirty, err
}
