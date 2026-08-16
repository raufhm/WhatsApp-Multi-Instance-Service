-- Contact-scoped activities for the CRM surface.
-- conversation_id becomes nullable so a follow-up can target a contact directly.

ALTER TABLE activities ALTER COLUMN conversation_id DROP NOT NULL;

ALTER TABLE activities ADD COLUMN IF NOT EXISTS contact_id UUID REFERENCES contacts(id) ON DELETE CASCADE;

-- Backfill contact_id from the owning conversation where it is known.
UPDATE activities a
SET contact_id = c.contact_id
FROM conversations c
WHERE a.conversation_id IS NOT NULL
  AND a.conversation_id = c.id
  AND a.contact_id IS NULL;

CREATE INDEX IF NOT EXISTS activities_contact_idx ON activities (contact_id, created_at DESC);