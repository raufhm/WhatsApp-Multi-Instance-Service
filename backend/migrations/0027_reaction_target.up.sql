ALTER TABLE conversation_messages
    ADD COLUMN IF NOT EXISTS reaction_target TEXT;

CREATE INDEX IF NOT EXISTS conversation_messages_reaction_target_idx
    ON conversation_messages (conversation_id, reaction_target)
    WHERE reaction_target IS NOT NULL;
