-- Deal pipeline stages (tenant-configurable)
CREATE TABLE deal_stages (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    key         VARCHAR(64) NOT NULL,
    label       VARCHAR(128) NOT NULL,
    color       VARCHAR(32) NOT NULL DEFAULT 'gray',
    icon        VARCHAR(64) NOT NULL DEFAULT '',
    sort_order  INT NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    is_won      BOOLEAN NOT NULL DEFAULT false,
    is_lost     BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, key)
);

-- Deal stage transition audit log
CREATE TABLE deal_stage_history (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    contact_id   UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    from_stage   VARCHAR(64),
    to_stage     VARCHAR(64) NOT NULL,
    note         TEXT DEFAULT '',
    moved_by     UUID REFERENCES operators(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_deal_stage_history_contact
    ON deal_stage_history(tenant_id, contact_id, created_at DESC);

-- Add deal_stage_key to contacts
ALTER TABLE contacts ADD COLUMN deal_stage_key VARCHAR(64) DEFAULT NULL;
