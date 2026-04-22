package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

var ErrSyncJobNotFound = errors.New("sync job not found")

type PostgresSyncJobRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresSyncJobRepository(pool *pgxpool.Pool) *PostgresSyncJobRepository {
	return &PostgresSyncJobRepository{pool: pool}
}

func (r *PostgresSyncJobRepository) Create(ctx context.Context, job *entity.SyncJob) error {
	query := `
		INSERT INTO sync_jobs (
			id, app_id, user_id, partner_account_id, job_type, parent_job_id,
			status, priority, total_items, completed_items, entity_type,
			error_message, worker_id, started_at, completed_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`
	_, err := r.pool.Exec(ctx, query,
		job.ID, job.AppID, job.UserID, job.PartnerAccountID,
		job.JobType, job.ParentJobID,
		string(job.Status), job.Priority,
		job.TotalItems, job.CompletedItems, job.EntityType,
		job.ErrorMessage, job.WorkerID,
		job.StartedAt, job.CompletedAt,
		job.CreatedAt, job.UpdatedAt,
	)
	return err
}

func (r *PostgresSyncJobRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.SyncJob, error) {
	query := `
		SELECT id, app_id, user_id, partner_account_id, job_type, parent_job_id,
		       status, priority, total_items, completed_items, entity_type,
		       error_message, worker_id, started_at, completed_at, created_at, updated_at
		FROM sync_jobs WHERE id = $1
	`
	job, err := r.scanJob(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSyncJobNotFound
		}
		return nil, err
	}
	return job, nil
}

func (r *PostgresSyncJobRepository) FindByStatus(ctx context.Context, status entity.SyncJobStatus) ([]*entity.SyncJob, error) {
	query := `
		SELECT id, app_id, user_id, partner_account_id, job_type, parent_job_id,
		       status, priority, total_items, completed_items, entity_type,
		       error_message, worker_id, started_at, completed_at, created_at, updated_at
		FROM sync_jobs WHERE status = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, string(status))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanJobs(rows)
}

func (r *PostgresSyncJobRepository) FindActiveByAppIDAndType(ctx context.Context, appID uuid.UUID, jobType string) (*entity.SyncJob, error) {
	query := `
		SELECT id, app_id, user_id, partner_account_id, job_type, parent_job_id,
		       status, priority, total_items, completed_items, entity_type,
		       error_message, worker_id, started_at, completed_at, created_at, updated_at
		FROM sync_jobs
		WHERE app_id = $1 AND job_type = $2 AND status IN ('pending', 'processing')
		ORDER BY created_at DESC
		LIMIT 1
	`
	job, err := r.scanJob(r.pool.QueryRow(ctx, query, appID, jobType))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No active job — not an error
		}
		return nil, err
	}
	return job, nil
}

func (r *PostgresSyncJobRepository) FindByParentJobID(ctx context.Context, parentJobID uuid.UUID) ([]*entity.SyncJob, error) {
	query := `
		SELECT id, app_id, user_id, partner_account_id, job_type, parent_job_id,
		       status, priority, total_items, completed_items, entity_type,
		       error_message, worker_id, started_at, completed_at, created_at, updated_at
		FROM sync_jobs WHERE parent_job_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, parentJobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanJobs(rows)
}

func (r *PostgresSyncJobRepository) ListByAppID(ctx context.Context, appID uuid.UUID, status string, jobType string, limit, offset int) ([]*entity.SyncJob, int, error) {
	// Build dynamic WHERE clause
	args := []interface{}{appID}
	where := "WHERE app_id = $1"
	argIdx := 2

	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if jobType != "" {
		where += fmt.Sprintf(" AND job_type = $%d", argIdx)
		args = append(args, jobType)
		argIdx++
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM sync_jobs " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch page
	query := `
		SELECT id, app_id, user_id, partner_account_id, job_type, parent_job_id,
		       status, priority, total_items, completed_items, entity_type,
		       error_message, worker_id, started_at, completed_at, created_at, updated_at
		FROM sync_jobs ` + where + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	jobs, err := r.scanJobs(rows)
	if err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

func (r *PostgresSyncJobRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.SyncJobStatus) error {
	query := `UPDATE sync_jobs SET status = $2, updated_at = $3 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, string(status), time.Now().UTC())
	return err
}

func (r *PostgresSyncJobRepository) UpdateProgress(ctx context.Context, id uuid.UUID, totalItems, completedItems int) error {
	query := `UPDATE sync_jobs SET total_items = $2, completed_items = $3, updated_at = $4 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, totalItems, completedItems, time.Now().UTC())
	return err
}

func (r *PostgresSyncJobRepository) MarkStarted(ctx context.Context, id uuid.UUID, workerID string) error {
	now := time.Now().UTC()
	query := `UPDATE sync_jobs SET status = 'processing', worker_id = $2, started_at = $3, updated_at = $3 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, workerID, now)
	return err
}

func (r *PostgresSyncJobRepository) MarkCompleted(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	query := `UPDATE sync_jobs SET status = 'completed', completed_at = $2, updated_at = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, now)
	return err
}

func (r *PostgresSyncJobRepository) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	now := time.Now().UTC()
	query := `UPDATE sync_jobs SET status = 'failed', error_message = $2, completed_at = $3, updated_at = $3 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, errMsg, now)
	return err
}

func (r *PostgresSyncJobRepository) scanJob(row pgx.Row) (*entity.SyncJob, error) {
	var job entity.SyncJob
	var status string
	err := row.Scan(
		&job.ID, &job.AppID, &job.UserID, &job.PartnerAccountID,
		&job.JobType, &job.ParentJobID,
		&status, &job.Priority,
		&job.TotalItems, &job.CompletedItems, &job.EntityType,
		&job.ErrorMessage, &job.WorkerID,
		&job.StartedAt, &job.CompletedAt,
		&job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	job.Status = entity.SyncJobStatus(status)
	return &job, nil
}

func (r *PostgresSyncJobRepository) scanJobs(rows pgx.Rows) ([]*entity.SyncJob, error) {
	var jobs []*entity.SyncJob
	for rows.Next() {
		var job entity.SyncJob
		var status string
		err := rows.Scan(
			&job.ID, &job.AppID, &job.UserID, &job.PartnerAccountID,
			&job.JobType, &job.ParentJobID,
			&status, &job.Priority,
			&job.TotalItems, &job.CompletedItems, &job.EntityType,
			&job.ErrorMessage, &job.WorkerID,
			&job.StartedAt, &job.CompletedAt,
			&job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		job.Status = entity.SyncJobStatus(status)
		jobs = append(jobs, &job)
	}
	return jobs, rows.Err()
}
