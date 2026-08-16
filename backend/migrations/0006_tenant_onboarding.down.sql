DROP TABLE IF EXISTS recovery_audit_log;
DROP TABLE IF EXISTS whatsapp_invitation_delivery;
DROP TABLE IF EXISTS invitations;
DROP TABLE IF EXISTS email_verification_tokens;
DROP TABLE IF EXISTS totp_recovery_tokens;
DROP TABLE IF EXISTS totp_backup_codes;

DROP INDEX IF EXISTS idx_operators_tenant_whatsapp;
DROP INDEX IF EXISTS idx_operators_tenant_email;

ALTER TABLE operators DROP COLUMN IF EXISTS whatsapp_number;
ALTER TABLE operators DROP COLUMN IF EXISTS totp_secret_encrypted;
ALTER TABLE operators DROP COLUMN IF EXISTS totp_verified_at;
ALTER TABLE operators DROP COLUMN IF EXISTS totp_setup_required;
ALTER TABLE operators DROP COLUMN IF EXISTS email_verified_at;

ALTER TABLE tenants DROP COLUMN IF EXISTS setup_step;
ALTER TABLE tenants DROP COLUMN IF EXISTS is_setup_complete;
ALTER TABLE tenants DROP COLUMN IF EXISTS org_details;
