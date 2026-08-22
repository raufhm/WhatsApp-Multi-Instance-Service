DROP INDEX IF EXISTS conversation_messages_reaction_target_idx;
ALTER TABLE conversation_messages DROP COLUMN IF EXISTS reaction_target;
