CREATE TABLE contact_field_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    key VARCHAR(64) NOT NULL,
    label VARCHAR(128) NOT NULL,
    field_type VARCHAR(32) NOT NULL CHECK (field_type IN ('text','number','date','select','checkbox')),
    options JSONB DEFAULT '[]'::jsonb,
    is_required BOOLEAN NOT NULL DEFAULT false,
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, key)
);

CREATE INDEX idx_contact_field_definitions_tenant ON contact_field_definitions(tenant_id, is_active, sort_order);
