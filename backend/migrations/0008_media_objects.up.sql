CREATE TABLE IF NOT EXISTS media_objects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    object_key TEXT NOT NULL UNIQUE,
    mime_type TEXT NOT NULL DEFAULT '',
    size BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_media_objects_tenant_key ON media_objects(tenant_id, object_key);
