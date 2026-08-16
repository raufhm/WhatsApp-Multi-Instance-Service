-- Migration Down: Remove bot rules versioning, operator actions, internal notes, and audit logs.

DROP TABLE IF EXISTS operator_audit_logs;
ALTER TABLE conversation_messages DROP COLUMN IF EXISTS is_internal;
ALTER TABLE conversations DROP COLUMN IF EXISTS merged_into_id;
ALTER TABLE conversations DROP COLUMN IF EXISTS assignee;
DROP TABLE IF EXISTS bot_rule_sets;
