-- Remove existing contacts that are WhatsApp Line Identifiers (LID)
-- instead of real mobile phone numbers. LIDs are 15-digit numeric IDs.
-- Only non-group personal contacts are affected.

WITH lid_contacts AS (
    SELECT id
    FROM contacts
    WHERE is_group = false
      AND (
          provider_address ~ '^[0-9]{15}(@s\.whatsapp\.net)?$'
          OR normalized_address ~ '^[0-9]{15}@s\.whatsapp\.net$'
      )
)
DELETE FROM deal_stage_history
WHERE contact_id IN (SELECT id FROM lid_contacts);

WITH lid_contacts AS (
    SELECT id
    FROM contacts
    WHERE is_group = false
      AND (
          provider_address ~ '^[0-9]{15}(@s\.whatsapp\.net)?$'
          OR normalized_address ~ '^[0-9]{15}@s\.whatsapp\.net$'
      )
)
DELETE FROM activities
WHERE contact_id IN (SELECT id FROM lid_contacts);

WITH lid_contacts AS (
    SELECT id
    FROM contacts
    WHERE is_group = false
      AND (
          provider_address ~ '^[0-9]{15}(@s\.whatsapp\.net)?$'
          OR normalized_address ~ '^[0-9]{15}@s\.whatsapp\.net$'
      )
)
DELETE FROM conversations
WHERE contact_id IN (SELECT id FROM lid_contacts);

WITH lid_contacts AS (
    SELECT id
    FROM contacts
    WHERE is_group = false
      AND (
          provider_address ~ '^[0-9]{15}(@s\.whatsapp\.net)?$'
          OR normalized_address ~ '^[0-9]{15}@s\.whatsapp\.net$'
      )
)
DELETE FROM contacts
WHERE id IN (SELECT id FROM lid_contacts);
