-- Seed default deal stages for all existing tenants that don't have any yet.
-- This is idempotent: ON CONFLICT (tenant_id, key) DO NOTHING prevents duplicates.

INSERT INTO deal_stages (tenant_id, key, label, color, icon, sort_order, is_won, is_lost)
SELECT t.id, 'NEW_LEAD', 'New Lead', '#94a3b8', 'user-plus', 1, false, false
FROM (SELECT DISTINCT tenant_id AS id FROM whatsapp_accounts) t
ON CONFLICT (tenant_id, key) DO NOTHING;

INSERT INTO deal_stages (tenant_id, key, label, color, icon, sort_order, is_won, is_lost)
SELECT t.id, 'APPOINTMENT_SCHEDULED', 'Appointment Scheduled', '#60a5fa', 'calendar', 2, false, false
FROM (SELECT DISTINCT tenant_id AS id FROM whatsapp_accounts) t
ON CONFLICT (tenant_id, key) DO NOTHING;

INSERT INTO deal_stages (tenant_id, key, label, color, icon, sort_order, is_won, is_lost)
SELECT t.id, 'HOT_LEAD', 'Hot Lead', '#f97316', 'flame', 3, false, false
FROM (SELECT DISTINCT tenant_id AS id FROM whatsapp_accounts) t
ON CONFLICT (tenant_id, key) DO NOTHING;

INSERT INTO deal_stages (tenant_id, key, label, color, icon, sort_order, is_won, is_lost)
SELECT t.id, 'COLD_LEAD', 'Cold Lead', '#64748b', 'snowflake', 4, false, false
FROM (SELECT DISTINCT tenant_id AS id FROM whatsapp_accounts) t
ON CONFLICT (tenant_id, key) DO NOTHING;

INSERT INTO deal_stages (tenant_id, key, label, color, icon, sort_order, is_won, is_lost)
SELECT t.id, 'IN_PROGRESS', 'In Progress', '#a78bfa', 'spinner', 5, false, false
FROM (SELECT DISTINCT tenant_id AS id FROM whatsapp_accounts) t
ON CONFLICT (tenant_id, key) DO NOTHING;

INSERT INTO deal_stages (tenant_id, key, label, color, icon, sort_order, is_won, is_lost)
SELECT t.id, 'CLOSED_WON', 'Closed Won', '#22c55e', 'trophy', 6, true, false
FROM (SELECT DISTINCT tenant_id AS id FROM whatsapp_accounts) t
ON CONFLICT (tenant_id, key) DO NOTHING;

INSERT INTO deal_stages (tenant_id, key, label, color, icon, sort_order, is_won, is_lost)
SELECT t.id, 'CLOSED_LOST', 'Closed Lost', '#ef4444', 'x-circle', 7, false, true
FROM (SELECT DISTINCT tenant_id AS id FROM whatsapp_accounts) t
ON CONFLICT (tenant_id, key) DO NOTHING;
