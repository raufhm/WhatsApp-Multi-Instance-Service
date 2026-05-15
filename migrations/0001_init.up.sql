CREATE TABLE IF NOT EXISTS public.messages (
    id SERIAL PRIMARY KEY,
    whatsapp_id TEXT UNIQUE,
    host_id TEXT,
    sender TEXT,
    recipient TEXT,
    content TEXT,
    is_group BOOLEAN,
    direction TEXT,
    msg_type TEXT,
    reaction_target TEXT,
    status TEXT DEFAULT 'SENT',
    timestamp TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS public.message_receipts (
    id SERIAL PRIMARY KEY,
    whatsapp_id TEXT,
    recipient_id TEXT,
    status TEXT,
    timestamp TIMESTAMP WITH TIME ZONE,
    UNIQUE(whatsapp_id, recipient_id, status)
);

CREATE TABLE IF NOT EXISTS public.groups (
    group_id TEXT PRIMARY KEY,
    name TEXT,
    description TEXT,
    owner_jid TEXT,
    participants JSONB,
    hosts JSONB,
    participant_count INT,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.instance_events (
    id SERIAL PRIMARY KEY,
    host_id TEXT,
    status TEXT,
    message TEXT,
    timestamp TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS public.instances (
    host_id TEXT PRIMARY KEY,
    status TEXT,
    is_connected BOOLEAN,
    last_seen TIMESTAMP WITH TIME ZONE
);
