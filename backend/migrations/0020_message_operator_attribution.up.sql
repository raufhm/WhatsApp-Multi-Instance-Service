-- Add operator attribution to conversation messages
ALTER TABLE conversation_messages
    ADD COLUMN IF NOT EXISTS operator_id UUID REFERENCES operators(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS operator_name VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_conversation_messages_operator
    ON conversation_messages(tenant_id, operator_id);
