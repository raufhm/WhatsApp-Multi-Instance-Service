-- Rollback for operator permission enforcement audit trail.
-- The role CHECK constraint is owned by 0005_operator_dashboard and is left
-- in place on rollback.

DROP TABLE IF EXISTS operator_permission_checks;
