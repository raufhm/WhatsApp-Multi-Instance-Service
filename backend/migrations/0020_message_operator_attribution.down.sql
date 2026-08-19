DROP INDEX IF EXISTS idx_conversation_messages_operator;
ALTER TABLE conversation_messages DROP COLUMN IF EXISTS operator_name;
ALTER TABLE conversation_messages DROP COLUMN IF EXISTS operator_id;
