-- Migration: Add bot rules versioning, operator actions, internal notes, and audit logs.

-- 1. Create bot_rule_sets for versioning, ordering, enabling/disabling bot rules
CREATE TABLE IF NOT EXISTS bot_rule_sets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    version INT NOT NULL,
    rules JSONB NOT NULL, -- Array of Rule objects: [{name, pattern, response, match, terminal, handoff, enabled}]
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, version)
);

CREATE UNIQUE INDEX IF NOT EXISTS bot_rule_sets_active_idx 
ON bot_rule_sets (tenant_id) WHERE is_active = TRUE;

-- 2. Add columns to conversations for assignee and merge support
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS assignee TEXT DEFAULT NULL;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS merged_into_id UUID REFERENCES conversations(id) ON DELETE SET NULL DEFAULT NULL;

-- 3. Add is_internal flag to conversation_messages for internal notes
ALTER TABLE conversation_messages ADD COLUMN IF NOT EXISTS is_internal BOOLEAN NOT NULL DEFAULT FALSE;

-- 4. Create operator_audit_logs table to audit operator actions
CREATE TABLE IF NOT EXISTS operator_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    operator_id TEXT NOT NULL,
    action TEXT NOT NULL, -- 'ASSIGN', 'CLOSE', 'REOPEN', 'HANDOFF', 'MERGE', 'SPLIT', 'UPDATE_RULES'
    conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS operator_audit_logs_tenant_idx ON operator_audit_logs (tenant_id, created_at DESC);
