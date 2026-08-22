-- WhatsApp Status updates use status@broadcast and are not customer chats.
-- Remove any historical status conversations created before broadcast filtering.
DELETE FROM conversations
WHERE contact_id IN (
    SELECT id FROM contacts
    WHERE normalized_address IN ('status@g.us', 'status@broadcast', 'status')
       OR provider_address IN ('status@g.us', 'status@broadcast', 'status')
       OR normalized_address LIKE '%@broadcast'
       OR provider_address LIKE '%@broadcast'
);

DELETE FROM contacts
WHERE normalized_address IN ('status@g.us', 'status@broadcast', 'status')
   OR provider_address IN ('status@g.us', 'status@broadcast', 'status')
   OR normalized_address LIKE '%@broadcast'
   OR provider_address LIKE '%@broadcast';
