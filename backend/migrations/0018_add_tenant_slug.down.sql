-- 0018_add_tenant_slug.down.sql
DROP INDEX IF EXISTS idx_tenants_slug;
ALTER TABLE tenants DROP COLUMN IF EXISTS slug;
