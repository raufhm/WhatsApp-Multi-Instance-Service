-- Preserve the WhatsApp participant who authored each group message.
-- Nullable for historical messages and non-group/system messages.
ALTER TABLE conversation_messages
    ADD COLUMN IF NOT EXISTS sender_address TEXT;

CREATE INDEX IF NOT EXISTS idx_conversation_messages_sender
    ON conversation_messages (tenant_id, sender_address)
    WHERE sender_address IS NOT NULL;
