-- Tenant setup wizard fields
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS setup_step INT DEFAULT 0;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS is_setup_complete BOOLEAN DEFAULT FALSE;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS org_details JSONB DEFAULT '{}'::jsonb;

-- Operator TOTP and WhatsApp fields
ALTER TABLE operators ADD COLUMN IF NOT EXISTS whatsapp_number TEXT UNIQUE;
ALTER TABLE operators ADD COLUMN IF NOT EXISTS totp_secret_encrypted TEXT;
ALTER TABLE operators ADD COLUMN IF NOT EXISTS totp_verified_at TIMESTAMPTZ;
ALTER TABLE operators ADD COLUMN IF NOT EXISTS totp_setup_required BOOLEAN DEFAULT FALSE;
ALTER TABLE operators ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ;
ALTER TABLE operators ALTER COLUMN password_hash DROP NOT NULL;
ALTER TABLE operators ALTER COLUMN email DROP NOT NULL;

-- Index for lookup by tenant and whatsapp_number / email
CREATE INDEX IF NOT EXISTS idx_operators_tenant_whatsapp ON operators(tenant_id, whatsapp_number) WHERE whatsapp_number IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_operators_tenant_email ON operators(tenant_id, email) WHERE email IS NOT NULL;

-- TOTP Backup Codes (single use)
CREATE TABLE IF NOT EXISTS totp_backup_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    code_hash TEXT NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_totp_backup_codes_operator ON totp_backup_codes(operator_id);
CREATE INDEX IF NOT EXISTS idx_totp_backup_codes_unused ON totp_backup_codes(operator_id, used_at) WHERE used_at IS NULL;

-- TOTP Recovery Tokens
CREATE TABLE IF NOT EXISTS totp_recovery_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_totp_recovery_tokens_token_hash ON totp_recovery_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_totp_recovery_tokens_operator ON totp_recovery_tokens(operator_id);

-- Email Verification Tokens
CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    operator_id UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_token_hash ON email_verification_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_operator ON email_verification_tokens(operator_id);

-- Invitations
CREATE TABLE IF NOT EXISTS invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'operator',
    channel TEXT NOT NULL DEFAULT 'whatsapp',
    recipient TEXT NOT NULL,
    whatsapp_number TEXT,
    email TEXT,
    token_hash TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    created_by UUID REFERENCES operators(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_invitations_tenant ON invitations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_invitations_token_hash ON invitations(token_hash);

-- WhatsApp Invitation Delivery Tracking
CREATE TABLE IF NOT EXISTS whatsapp_invitation_delivery (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invitation_id UUID NOT NULL REFERENCES invitations(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'sent',
    message_id TEXT,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    delivered_at TIMESTAMPTZ,
    error_message TEXT
);
CREATE INDEX IF NOT EXISTS idx_whatsapp_invitation_delivery_invitation ON whatsapp_invitation_delivery(invitation_id);

-- Recovery & Security Audit Log
CREATE TABLE IF NOT EXISTS recovery_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    operator_id UUID REFERENCES operators(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    details JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_recovery_audit_log_tenant ON recovery_audit_log(tenant_id);
CREATE INDEX IF NOT EXISTS idx_recovery_audit_log_operator ON recovery_audit_log(operator_id);
