CREATE TABLE IF NOT EXISTS sync_jobs (
    id UUID PRIMARY KEY,
    app_id UUID NOT NULL,
    user_id UUID NOT NULL,
    partner_account_id UUID NOT NULL,
    job_type VARCHAR(50) NOT NULL,
    parent_job_id UUID REFERENCES sync_jobs(id),
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    priority INT NOT NULL DEFAULT 0,
    total_items INT NOT NULL DEFAULT 0,
    completed_items INT NOT NULL DEFAULT 0,
    entity_type VARCHAR(50) NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    worker_id VARCHAR(100) NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sync_jobs_app_id ON sync_jobs(app_id);
CREATE INDEX idx_sync_jobs_status ON sync_jobs(status);
CREATE INDEX idx_sync_jobs_parent_job_id ON sync_jobs(parent_job_id);
CREATE INDEX idx_sync_jobs_app_id_type_status ON sync_jobs(app_id, job_type, status);
CREATE INDEX idx_sync_jobs_created_at ON sync_jobs(created_at);
