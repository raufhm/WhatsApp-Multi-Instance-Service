-- Operator accounts and dashboard sessions for the operator dashboard.
-- Operators are tenant-scoped and provide password-based auth for the UI.

CREATE TABLE IF NOT EXISTS operators (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email      TEXT NOT NULL,
    name       TEXT NOT NULL DEFAULT '',
    password_hash TEXT NOT NULL,
    role       TEXT NOT NULL DEFAULT 'operator' CHECK (role IN ('admin','operator','viewer')),
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, email)
);

CREATE TABLE IF NOT EXISTS sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sessions_operator ON sessions(operator_id);
CREATE INDEX IF NOT EXISTS idx_sessions_tenant ON sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
