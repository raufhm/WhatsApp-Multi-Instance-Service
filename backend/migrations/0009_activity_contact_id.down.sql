DROP INDEX IF EXISTS activities_contact_idx;

ALTER TABLE activities ALTER COLUMN conversation_id SET NOT NULL;

ALTER TABLE activities DROP COLUMN IF EXISTS contact_id;