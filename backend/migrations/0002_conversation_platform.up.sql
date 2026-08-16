CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    key_hash TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL DEFAULT '',
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS whatsapp_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    host_id TEXT NOT NULL,
    provider TEXT NOT NULL DEFAULT 'whatsmeow',
    display_name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, provider, host_id)
);

CREATE TABLE IF NOT EXISTS contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    normalized_address TEXT NOT NULL,
    provider_address TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, normalized_address)
);

CREATE SEQUENCE IF NOT EXISTS conversation_ticket_number_seq;
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES whatsapp_accounts(id) ON DELETE RESTRICT,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE RESTRICT,
    ticket_number BIGINT NOT NULL UNIQUE DEFAULT nextval('conversation_ticket_number_seq'),
    status TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN','BOT_ACTIVE','WAITING','HANDED_OFF','CLOSED')),
    bot_state TEXT NOT NULL DEFAULT 'ACTIVE',
    started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_activity_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMPTZ,
    handoff_at TIMESTAMPTZ,
    closure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS conversations_inbox_idx ON conversations (tenant_id, status, last_activity_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS conversations_active_contact_idx
    ON conversations (tenant_id, account_id, contact_id) WHERE status <> 'CLOSED';

CREATE TABLE IF NOT EXISTS conversation_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    actor TEXT NOT NULL CHECK (actor IN ('CONTACT','BOT','OPERATOR','SYSTEM')),
    provider TEXT NOT NULL DEFAULT 'whatsmeow',
    provider_message_id TEXT NOT NULL,
    direction TEXT NOT NULL CHECK (direction IN ('INCOMING','OUTGOING')),
    content TEXT NOT NULL DEFAULT '',
    message_type TEXT NOT NULL DEFAULT 'TEXT',
    media_url TEXT,
    status TEXT NOT NULL DEFAULT 'PENDING',
    provider_timestamp TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, provider, provider_message_id)
);
CREATE INDEX IF NOT EXISTS conversation_messages_timeline_idx ON conversation_messages (conversation_id, provider_timestamp, created_at);

CREATE TABLE IF NOT EXISTS bot_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL UNIQUE REFERENCES conversations(id) ON DELETE CASCADE,
    rule_version TEXT NOT NULL DEFAULT 'default',
    state JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    summary TEXT NOT NULL,
    next_action TEXT NOT NULL DEFAULT '',
    priority TEXT NOT NULL DEFAULT 'NORMAL',
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','ACKNOWLEDGED','DISMISSED')),
    due_at TIMESTAMPTZ,
    acknowledged_by TEXT,
    acknowledged_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS activities_queue_idx ON activities (tenant_id, status, due_at);
CREATE UNIQUE INDEX IF NOT EXISTS activities_conversation_type_pending_idx
    ON activities (conversation_id, type) WHERE status = 'PENDING';