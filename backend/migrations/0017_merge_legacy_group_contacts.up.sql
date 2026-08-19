-- Merge legacy group contacts ending in @s.whatsapp.net into canonical @g.us contacts
BEGIN;

-- 1. Re-link dependent tables (conversations, activities, deal_stage_history) from legacy @s.whatsapp.net group contacts to their canonical @g.us contacts
UPDATE conversations c
SET contact_id = canonical.id,
    is_group = true
FROM contacts legacy
JOIN contacts canonical
  ON canonical.tenant_id = legacy.tenant_id
 AND canonical.provider_address = legacy.provider_address
 AND canonical.is_group = true
 AND canonical.normalized_address LIKE '%@g.us'
WHERE c.contact_id = legacy.id
  AND legacy.is_group = true
  AND legacy.normalized_address LIKE '%@s.whatsapp.net';

UPDATE activities a
SET contact_id = canonical.id
FROM contacts legacy
JOIN contacts canonical
  ON canonical.tenant_id = legacy.tenant_id
 AND canonical.provider_address = legacy.provider_address
 AND canonical.is_group = true
 AND canonical.normalized_address LIKE '%@g.us'
WHERE a.contact_id = legacy.id
  AND legacy.is_group = true
  AND legacy.normalized_address LIKE '%@s.whatsapp.net';

UPDATE deal_stage_history d
SET contact_id = canonical.id
FROM contacts legacy
JOIN contacts canonical
  ON canonical.tenant_id = legacy.tenant_id
 AND canonical.provider_address = legacy.provider_address
 AND canonical.is_group = true
 AND canonical.normalized_address LIKE '%@g.us'
WHERE d.contact_id = legacy.id
  AND legacy.is_group = true
  AND legacy.normalized_address LIKE '%@s.whatsapp.net';

-- 2. Delete duplicate legacy @s.whatsapp.net group contact records where canonical @g.us exists
DELETE FROM contacts legacy
USING contacts canonical
WHERE legacy.tenant_id = canonical.tenant_id
  AND legacy.provider_address = canonical.provider_address
  AND legacy.is_group = true
  AND canonical.is_group = true
  AND legacy.normalized_address LIKE '%@s.whatsapp.net'
  AND canonical.normalized_address LIKE '%@g.us';

-- 3. For any remaining group contacts still using @s.whatsapp.net (with no canonical @g.us peer), migrate normalized_address to @g.us
UPDATE contacts
SET normalized_address = REGEXP_REPLACE(normalized_address, '@s\.whatsapp\.net$', '@g.us'),
    updated_at = CURRENT_TIMESTAMP
WHERE is_group = true
  AND normalized_address LIKE '%@s.whatsapp.net';

COMMIT;
