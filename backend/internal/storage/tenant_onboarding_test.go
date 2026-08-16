package storage

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/raufhm/whatsapp-testing/internal/totp"
)

func TestSignupTenant(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()

	tenantID := uuid.New()
	opID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO tenants (name, setup_step, is_setup_complete, org_details)
		 VALUES ($1, 0, false, '{}'::jsonb)
		 RETURNING id, name, setup_step, is_setup_complete, org_details, created_at, updated_at`)).
		WithArgs("Acme Corp").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "setup_step", "is_setup_complete", "org_details", "created_at", "updated_at"}).
			AddRow(tenantID, "Acme Corp", 0, false, []byte("{}"), time.Now(), time.Now()))

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO operators (tenant_id, name, email, whatsapp_number, role, is_active, totp_setup_required)
		 VALUES ($1, $2, $3, $4, 'admin', true, true)
		 RETURNING id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		           totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		           last_login_at, created_at, updated_at`)).
		WithArgs(tenantID, "Alice Admin", "admin@acme.com", "+1234567890").
		WillReturnRows(sqlmock.NewRows(operatorColumns).
			AddRow(opID, tenantID, "admin@acme.com", "+1234567890", "Alice Admin", nil, "admin", true, nil, nil, true, nil, nil, time.Now(), time.Now()))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO email_verification_tokens (tenant_id, operator_id, token_hash, expires_at)
		 VALUES ($1, $2, $3, $4)`)).
		WithArgs(tenantID, opID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	tenant, op, rawToken, err := store.SignupTenant("Acme Corp", "Alice Admin", "admin@acme.com", "+1234567890")
	if err != nil {
		t.Fatalf("SignupTenant: %v", err)
	}
	if tenant.ID != tenantID || op.ID != opID || rawToken == "" {
		t.Fatalf("unexpected signup result: tenant=%+v op=%+v token=%s", tenant, op, rawToken)
	}
}

func TestVerifyEmailToken(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()

	tenantID := uuid.New()
	opID := uuid.New()
	tokenID := uuid.New()
	rawToken := "sample-email-token"
	tokenHash := totp.HashToken(rawToken)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, operator_id FROM email_verification_tokens
		 WHERE token_hash = $1 AND used_at IS NULL AND expires_at > CURRENT_TIMESTAMP`)).
		WithArgs(tokenHash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "operator_id"}).
			AddRow(tokenID, tenantID, opID))

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE email_verification_tokens SET used_at = $1 WHERE id = $2`)).
		WithArgs(sqlmock.AnyArg(), tokenID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE operators SET email_verified_at = $1, updated_at = $1
		 WHERE id = $2 AND tenant_id = $3
		 RETURNING id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		           totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		           last_login_at, created_at, updated_at`)).
		WithArgs(sqlmock.AnyArg(), opID, tenantID).
		WillReturnRows(sqlmock.NewRows(operatorColumns).
			AddRow(opID, tenantID, "admin@acme.com", "+1234567890", "Alice Admin", nil, "admin", true, nil, nil, true, time.Now(), nil, time.Now(), time.Now()))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, setup_step, is_setup_complete, org_details, created_at, updated_at
		 FROM tenants WHERE id = $1`)).
		WithArgs(tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "setup_step", "is_setup_complete", "org_details", "created_at", "updated_at"}).
			AddRow(tenantID, "Acme Corp", 0, false, []byte("{}"), time.Now(), time.Now()))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO totp_recovery_tokens (operator_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`)).
		WithArgs(opID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	tenant, op, setupToken, err := store.VerifyEmailToken(rawToken)
	if err != nil {
		t.Fatalf("VerifyEmailToken: %v", err)
	}
	if tenant.ID != tenantID || op.ID != opID || setupToken == "" {
		t.Fatalf("unexpected verify email result: tenant=%+v op=%+v setupToken=%s", tenant, op, setupToken)
	}
}

func TestVerifyTOTPSetup(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()

	tenantID := uuid.New()
	opID := uuid.New()
	tokenID := uuid.New()
	sessionID := uuid.New()
	rawSetupToken := "sample-setup-token"
	tokenHash := totp.HashToken(rawSetupToken)

	secret, err := totp.GenerateSecret()
	if err != nil {
		t.Fatal(err)
	}
	encryptedSecret, err := totp.EncryptSecret(secret)
	if err != nil {
		t.Fatal(err)
	}
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, operator_id FROM totp_recovery_tokens
		 WHERE token_hash = $1 AND used_at IS NULL AND expires_at > CURRENT_TIMESTAMP`)).
		WithArgs(tokenHash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "operator_id"}).AddRow(tokenID, opID))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		        totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		        last_login_at, created_at, updated_at
		 FROM operators WHERE id = $1`)).
		WithArgs(opID).
		WillReturnRows(sqlmock.NewRows(operatorColumns).
			AddRow(opID, tenantID, "admin@acme.com", "+1234567890", "Alice Admin", nil, "admin", true, encryptedSecret, nil, true, time.Now(), nil, time.Now(), time.Now()))

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE totp_recovery_tokens SET used_at = $1 WHERE id = $2`)).
		WithArgs(sqlmock.AnyArg(), tokenID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE operators
		 SET totp_verified_at = $1, totp_setup_required = false, updated_at = $1, last_login_at = $1
		 WHERE id = $2
		 RETURNING id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		           totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		           last_login_at, created_at, updated_at`)).
		WithArgs(sqlmock.AnyArg(), opID).
		WillReturnRows(sqlmock.NewRows(operatorColumns).
			AddRow(opID, tenantID, "admin@acme.com", "+1234567890", "Alice Admin", nil, "admin", true, encryptedSecret, time.Now(), false, time.Now(), time.Now(), time.Now(), time.Now()))

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM totp_backup_codes WHERE operator_id = $1`)).
		WithArgs(opID).
		WillReturnResult(sqlmock.NewResult(1, 0))

	for i := 0; i < 10; i++ {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO totp_backup_codes (operator_id, code_hash) VALUES ($1, $2)`)).
			WithArgs(opID, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO sessions (operator_id, tenant_id, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, operator_id, tenant_id, expires_at, created_at`)).
		WithArgs(opID, tenantID, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "operator_id", "tenant_id", "expires_at", "created_at"}).
			AddRow(sessionID, opID, tenantID, time.Now().Add(8*time.Hour), time.Now()))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO recovery_audit_log (tenant_id, operator_id, action, details)
		 VALUES ($1, $2, 'TOTP_SETUP_COMPLETED', '{"method": "totp_setup"}'::jsonb)`)).
		WithArgs(tenantID, opID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	op, backupCodes, session, err := store.VerifyTOTPSetup(rawSetupToken, code)
	if err != nil {
		t.Fatalf("VerifyTOTPSetup: %v", err)
	}
	if op.ID != opID || len(backupCodes) != 10 || session.ID != sessionID {
		t.Fatalf("unexpected VerifyTOTPSetup result: op=%+v codes=%d session=%+v", op, len(backupCodes), session)
	}
}

func TestInvitationFlow(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()

	tenantID := uuid.New()
	adminID := uuid.New()
	invID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO invitations (tenant_id, role, channel, recipient, whatsapp_number, email, token_hash, status, expires_at, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', $8, $9)
		 RETURNING id, tenant_id, role, channel, recipient, whatsapp_number, email, token_hash, status, expires_at, accepted_at, created_by, created_at`)).
		WithArgs(tenantID, "operator", "whatsapp", "+9876543210", "+9876543210", nil, sqlmock.AnyArg(), sqlmock.AnyArg(), adminID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "role", "channel", "recipient", "whatsapp_number", "email", "token_hash", "status", "expires_at", "accepted_at", "created_by", "created_at"}).
			AddRow(invID, tenantID, "operator", "whatsapp", "+9876543210", "+9876543210", nil, "hash", "pending", time.Now().Add(7*24*time.Hour), nil, adminID, time.Now()))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO whatsapp_invitation_delivery (invitation_id, status)
			 VALUES ($1, 'sent')`)).
		WithArgs(invID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	inv, token, err := store.CreateInvitation(tenantID, &adminID, "+9876543210", "whatsapp", "operator", "+9876543210", "")
	if err != nil {
		t.Fatalf("CreateInvitation: %v", err)
	}
	if inv.ID != invID || token == "" {
		t.Fatalf("unexpected inv: %+v token=%s", inv, token)
	}
}

func TestTenantSetupStatusAndComplete(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()

	tenantID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT setup_step, is_setup_complete, org_details FROM tenants WHERE id = $1`)).
		WithArgs(tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"setup_step", "is_setup_complete", "org_details"}).
			AddRow(1, false, []byte(`{"industry":"tech"}`)))

	status, err := store.GetTenantSetupStatus(tenantID)
	if err != nil {
		t.Fatalf("GetTenantSetupStatus: %v", err)
	}
	if status.SetupStep != 1 || status.IsSetupComplete || status.OrgDetails["industry"] != "tech" {
		t.Fatalf("unexpected status: %+v", status)
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenants SET is_setup_complete = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1`)).
		WithArgs(tenantID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := store.CompleteTenantSetup(tenantID); err != nil {
		t.Fatalf("CompleteTenantSetup: %v", err)
	}
}
