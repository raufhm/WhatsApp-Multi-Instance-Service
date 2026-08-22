DROP INDEX IF EXISTS idx_conversation_messages_sender;
ALTER TABLE conversation_messages DROP COLUMN IF EXISTS sender_address;
