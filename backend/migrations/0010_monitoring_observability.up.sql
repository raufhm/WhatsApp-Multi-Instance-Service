-- Monitoring & observability: status transition history and tailable event log.
CREATE TABLE IF NOT EXISTS whatsmeow_status_events (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    UUID NOT NULL REFERENCES tenants(id),
    host_id      TEXT NOT NULL,
    status       TEXT NOT NULL,
    is_connected BOOLEAN NOT NULL,
    message      TEXT,
    occurred_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_status_events_tenant_host_time
    ON whatsmeow_status_events (tenant_id, host_id, occurred_at DESC);

CREATE TABLE IF NOT EXISTS whatsmeow_instance_events (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    host_id     TEXT NOT NULL,
    event_type  TEXT NOT NULL,
    direction   TEXT,
    payload     JSONB,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_instance_events_tenant_host_time
    ON whatsmeow_instance_events (tenant_id, host_id, occurred_at DESC);

ALTER TABLE whatsmeow_instances
    ADD COLUMN IF NOT EXISTS last_connected_at    TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_disconnected_at TIMESTAMPTZ;