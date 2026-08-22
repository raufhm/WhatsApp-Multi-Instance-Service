-- Older projections could incorrectly store the group JID as the sender.
-- Clear those values so a history re-sync can replace them with the participant.
UPDATE conversation_messages m
SET sender_address = NULL
FROM conversations c
JOIN contacts co ON co.id = c.contact_id AND co.tenant_id = c.tenant_id
WHERE m.conversation_id = c.id
  AND m.tenant_id = c.tenant_id
  AND co.is_group = TRUE
  AND m.sender_address IN (co.provider_address, co.normalized_address);
