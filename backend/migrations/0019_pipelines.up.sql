-- Pipelines management
CREATE TABLE pipelines (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        VARCHAR(128) NOT NULL,
    description TEXT DEFAULT '',
    is_default  BOOLEAN NOT NULL DEFAULT false,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, name)
);

CREATE INDEX idx_pipelines_tenant ON pipelines(tenant_id, is_active);

-- Seed default pipeline for all existing tenants
INSERT INTO pipelines (tenant_id, name, description, is_default, is_active)
SELECT t.id, 'Sales Pipeline', 'Standard customer sales and deal qualification pipeline', true, true
FROM tenants t
ON CONFLICT (tenant_id, name) DO NOTHING;

-- Associate deal_stages with pipelines
ALTER TABLE deal_stages ADD COLUMN pipeline_id UUID REFERENCES pipelines(id) ON DELETE RESTRICT;

-- Backfill pipeline_id on existing deal_stages with default pipeline
UPDATE deal_stages ds
SET pipeline_id = p.id
FROM pipelines p
WHERE ds.tenant_id = p.tenant_id AND p.is_default = true AND ds.pipeline_id IS NULL;

CREATE INDEX idx_deal_stages_pipeline ON deal_stages(pipeline_id, sort_order);

-- Add deal_stage_id to contacts
ALTER TABLE contacts ADD COLUMN deal_stage_id UUID REFERENCES deal_stages(id) ON DELETE SET NULL;

-- Backfill deal_stage_id on contacts matching tenant and deal_stage_key
UPDATE contacts c
SET deal_stage_id = ds.id
FROM deal_stages ds
WHERE c.tenant_id = ds.tenant_id AND c.deal_stage_key = ds.key AND c.deal_stage_id IS NULL;

CREATE INDEX idx_contacts_deal_stage_id ON contacts(deal_stage_id);
