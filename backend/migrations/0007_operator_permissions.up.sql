-- Operator permission enforcement audit trail.
-- The role CHECK constraint already exists on operators (added in
-- 0005_operator_dashboard.up.sql), so this migration only adds the table
-- used to record every permission check (allowed and denied).

CREATE TABLE IF NOT EXISTS operator_permission_checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    action TEXT NOT NULL,
    resource TEXT,
    resource_id UUID,
    allowed BOOLEAN NOT NULL,
    reason TEXT,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_permission_checks_operator ON operator_permission_checks(operator_id);
CREATE INDEX IF NOT EXISTS idx_permission_checks_action ON operator_permission_checks(action);
CREATE INDEX IF NOT EXISTS idx_permission_checks_created ON operator_permission_checks(created_at);

COMMENT ON TABLE operator_permission_checks IS
'Audit log for all permission checks, successful and denied';
