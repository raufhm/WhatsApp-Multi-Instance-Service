CREATE TABLE IF NOT EXISTS upload_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    message_id TEXT NOT NULL DEFAULT '',
    host_id TEXT NOT NULL DEFAULT '',
    object_key TEXT NOT NULL UNIQUE,
    mime_type TEXT NOT NULL DEFAULT '',
    media_path TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','PROCESSING','COMPLETED','FAILED')),
    attempt_count INT NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_error TEXT,
    media_url TEXT,
    lease_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Due index: lets the worker claim pending jobs in next-attempt order and
-- reclaim PROCESSING jobs whose lease has expired (recovery after restart).
CREATE INDEX IF NOT EXISTS upload_jobs_due_idx ON upload_jobs (status, next_attempt_at)
    WHERE status IN ('PENDING','PROCESSING');

CREATE INDEX IF NOT EXISTS upload_jobs_host_idx ON upload_jobs (host_id, created_at);
