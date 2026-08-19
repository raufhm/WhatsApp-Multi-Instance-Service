-- 0018_add_tenant_slug.up.sql
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS slug TEXT;

-- Backfill slug for existing rows where slug is NULL or empty
UPDATE tenants
SET slug = LOWER(REGEXP_REPLACE(REGEXP_REPLACE(TRIM(name), '[^a-zA-Z0-9]+', '-', 'g'), '^-+|-+$', '', 'g'))
WHERE slug IS NULL OR slug = '';

-- If any slug ended up empty or null, fallback to 'tenant-' || substring(id::text, 1, 8)
UPDATE tenants
SET slug = 'tenant-' || SUBSTRING(id::text, 1, 8)
WHERE slug IS NULL OR slug = '';

-- Resolve duplicate slugs if any exist in legacy data by appending row number
WITH numbered AS (
  SELECT id, slug,
         ROW_NUMBER() OVER (PARTITION BY slug ORDER BY created_at ASC) as rn
  FROM tenants
)
UPDATE tenants t
SET slug = t.slug || '-' || numbered.rn
FROM numbered
WHERE t.id = numbered.id AND numbered.rn > 1;

ALTER TABLE tenants ALTER COLUMN slug SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);
