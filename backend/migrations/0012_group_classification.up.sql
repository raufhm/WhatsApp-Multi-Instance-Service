-- Backfill any contacts whose address shape clearly belongs to a WhatsApp group
-- so they are rendered as Group instead of Personal.
UPDATE contacts
SET is_group = true, updated_at = CURRENT_TIMESTAMP
WHERE is_group = false
  AND (
    provider_address ~ '^[0-9]+-[0-9]+(@g\.us)?$'
    OR normalized_address LIKE '%@g.us'
  );
