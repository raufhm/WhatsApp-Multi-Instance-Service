DROP TABLE IF EXISTS whatsmeow_instance_events;
DROP TABLE IF EXISTS whatsmeow_status_events;

ALTER TABLE whatsmeow_instances
    DROP COLUMN IF EXISTS last_connected_at,
    DROP COLUMN IF EXISTS last_disconnected_at;