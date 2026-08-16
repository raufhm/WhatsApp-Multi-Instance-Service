package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/raufhm/whatsapp-testing/domain"
	"github.com/raufhm/whatsapp-testing/internal/totp"
)

var (
	ErrTokenNotFound       = errors.New("token not found or expired")
	ErrInvitationNotFound  = errors.New("invitation not found or expired")
	ErrInvalidBackupCode   = errors.New("invalid backup code")
	ErrUnauthorizedAdmin   = errors.New("admin permission required")
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrInvalidTOTPCode     = errors.New("invalid totp code")
	ErrDuplicateWhatsapp   = errors.New("whatsapp number already registered")
)

// isUniqueViolation reports whether err is a Postgres unique constraint
// violation, optionally limited to a specific constraint name.
func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if pgErr.Code != "23505" {
		return false
	}
	return constraint == "" || pgErr.ConstraintName == constraint
}

// scanOperatorRow scans standard operator columns into domain.Operator along with passwordHash and totpSecretEncrypted.
func scanOperatorRow(scanner interface {
	Scan(dest ...any) error
}) (domain.Operator, string, string, error) {
	var op domain.Operator
	var email, whatsappNumber, passHash, totpSecret sql.NullString
	err := scanner.Scan(
		&op.ID,
		&op.TenantID,
		&email,
		&whatsappNumber,
		&op.Name,
		&passHash,
		&op.Role,
		&op.IsActive,
		&totpSecret,
		&op.TotpVerifiedAt,
		&op.TotpSetupRequired,
		&op.EmailVerifiedAt,
		&op.LastLoginAt,
		&op.CreatedAt,
		&op.UpdatedAt,
	)
	if err != nil {
		return domain.Operator{}, "", "", err
	}
	op.Email = email.String
	op.WhatsappNumber = whatsappNumber.String
	return op, passHash.String, totpSecret.String, nil
}

// FindOperatorByIdentifier finds an operator by email OR whatsapp_number in the tenant.
func (p *PostgresStore) FindOperatorByIdentifier(tenantID uuid.UUID, identifier string) (domain.Operator, string, string, error) {
	row := p.db.QueryRow(
		`SELECT id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		        totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		        last_login_at, created_at, updated_at
		 FROM operators
		 WHERE tenant_id = $1 AND (email = $2 OR whatsapp_number = $2)`,
		tenantID, identifier,
	)
	op, passHash, totpSecret, err := scanOperatorRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Operator{}, "", "", ErrOperatorNotFound
		}
		return domain.Operator{}, "", "", err
	}
	return op, passHash, totpSecret, nil
}

// FindOperatorByEmail returns an operator by tenant + email.
func (p *PostgresStore) FindOperatorByEmail(tenantID uuid.UUID, email string) (domain.Operator, string, error) {
	row := p.db.QueryRow(
		`SELECT id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		        totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		        last_login_at, created_at, updated_at
		 FROM operators WHERE tenant_id=$1 AND email=$2`, tenantID, email)
	op, passHash, _, err := scanOperatorRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Operator{}, "", ErrOperatorNotFound
		}
		return domain.Operator{}, "", err
	}
	return op, passHash, nil
}

// GetOperatorByID returns an operator by tenant + operator id.
func (p *PostgresStore) GetOperatorByID(tenantID, operatorID uuid.UUID) (domain.Operator, error) {
	row := p.db.QueryRow(
		`SELECT id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		        totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		        last_login_at, created_at, updated_at
		 FROM operators WHERE tenant_id=$1 AND id=$2`, tenantID, operatorID)
	op, _, _, err := scanOperatorRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Operator{}, ErrOperatorNotFound
		}
		return domain.Operator{}, err
	}
	return op, nil
}

// GetOperatorByIDWithSecret returns an operator along with encrypted TOTP secret.
func (p *PostgresStore) GetOperatorByIDWithSecret(tenantID, operatorID uuid.UUID) (domain.Operator, string, error) {
	row := p.db.QueryRow(
		`SELECT id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		        totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		        last_login_at, created_at, updated_at
		 FROM operators WHERE tenant_id=$1 AND id=$2`, tenantID, operatorID)
	op, _, secret, err := scanOperatorRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Operator{}, "", ErrOperatorNotFound
		}
		return domain.Operator{}, "", err
	}
	return op, secret, nil
}

// ListOperators lists all operators within a tenant.
func (p *PostgresStore) ListOperators(tenantID uuid.UUID) ([]domain.Operator, error) {
	rows, err := p.db.Query(
		`SELECT id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		        totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		        last_login_at, created_at, updated_at
		 FROM operators WHERE tenant_id=$1 ORDER BY created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var operators []domain.Operator
	for rows.Next() {
		op, _, _, err := scanOperatorRow(rows)
		if err != nil {
			return nil, err
		}
		operators = append(operators, op)
	}
	return operators, rows.Err()
}

// SignupTenant creates a new tenant, an admin operator, and an email verification token.
func (p *PostgresStore) SignupTenant(tenantName, adminName, adminEmail, adminWhatsapp string) (domain.Tenant, domain.Operator, string, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return domain.Tenant{}, domain.Operator{}, "", err
	}
	defer tx.Rollback()

	var tenant domain.Tenant
	var orgDetailsJSON []byte
	err = tx.QueryRow(
		`INSERT INTO tenants (name, setup_step, is_setup_complete, org_details)
		 VALUES ($1, 0, false, '{}'::jsonb)
		 RETURNING id, name, setup_step, is_setup_complete, org_details, created_at, updated_at`,
		tenantName,
	).Scan(&tenant.ID, &tenant.Name, &tenant.SetupStep, &tenant.IsSetupComplete, &orgDetailsJSON, &tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		return domain.Tenant{}, domain.Operator{}, "", fmt.Errorf("insert tenant: %w", err)
	}
	if len(orgDetailsJSON) > 0 {
		_ = json.Unmarshal(orgDetailsJSON, &tenant.OrgDetails)
	}

	var emailVal, whatsappVal *string
	if adminEmail != "" {
		emailVal = &adminEmail
	}
	if adminWhatsapp != "" {
		whatsappVal = &adminWhatsapp
	}

	row := tx.QueryRow(
		`INSERT INTO operators (tenant_id, name, email, whatsapp_number, role, is_active, totp_setup_required)
		 VALUES ($1, $2, $3, $4, 'admin', true, true)
		 RETURNING id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		           totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		           last_login_at, created_at, updated_at`,
		tenant.ID, adminName, emailVal, whatsappVal,
	)
	op, _, _, err := scanOperatorRow(row)
	if err != nil {
		if isUniqueViolation(err, "operators_whatsapp_number_key") {
			return domain.Tenant{}, domain.Operator{}, "", ErrDuplicateWhatsapp
		}
		return domain.Tenant{}, domain.Operator{}, "", fmt.Errorf("insert operator: %w", err)
	}

	rawToken := uuid.New().String()
	tokenHash := totp.HashToken(rawToken)
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	_, err = tx.Exec(
		`INSERT INTO email_verification_tokens (tenant_id, operator_id, token_hash, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		tenant.ID, op.ID, tokenHash, expiresAt,
	)
	if err != nil {
		return domain.Tenant{}, domain.Operator{}, "", fmt.Errorf("insert email verification token: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return domain.Tenant{}, domain.Operator{}, "", err
	}
	return tenant, op, rawToken, nil
}

// VerifyEmailToken validates an email token, marks it used, marks operator email verified, and creates a TOTP setup token.
func (p *PostgresStore) VerifyEmailToken(rawToken string) (domain.Tenant, domain.Operator, string, error) {
	tokenHash := totp.HashToken(rawToken)

	tx, err := p.db.Begin()
	if err != nil {
		return domain.Tenant{}, domain.Operator{}, "", err
	}
	defer tx.Rollback()

	var tokenID, tenantID, operatorID uuid.UUID
	err = tx.QueryRow(
		`SELECT id, tenant_id, operator_id FROM email_verification_tokens
		 WHERE token_hash = $1 AND used_at IS NULL AND expires_at > CURRENT_TIMESTAMP`,
		tokenHash,
	).Scan(&tokenID, &tenantID, &operatorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Tenant{}, domain.Operator{}, "", ErrTokenNotFound
		}
		return domain.Tenant{}, domain.Operator{}, "", err
	}

	now := time.Now().UTC()
	_, err = tx.Exec(`UPDATE email_verification_tokens SET used_at = $1 WHERE id = $2`, now, tokenID)
	if err != nil {
		return domain.Tenant{}, domain.Operator{}, "", err
	}

	row := tx.QueryRow(
		`UPDATE operators SET email_verified_at = $1, updated_at = $1
		 WHERE id = $2 AND tenant_id = $3
		 RETURNING id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		           totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		           last_login_at, created_at, updated_at`,
		now, operatorID, tenantID,
	)
	op, _, _, err := scanOperatorRow(row)
	if err != nil {
		return domain.Tenant{}, domain.Operator{}, "", err
	}

	var tenant domain.Tenant
	var orgDetailsJSON []byte
	err = tx.QueryRow(
		`SELECT id, name, setup_step, is_setup_complete, org_details, created_at, updated_at
		 FROM tenants WHERE id = $1`, tenantID,
	).Scan(&tenant.ID, &tenant.Name, &tenant.SetupStep, &tenant.IsSetupComplete, &orgDetailsJSON, &tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		return domain.Tenant{}, domain.Operator{}, "", err
	}
	if len(orgDetailsJSON) > 0 {
		_ = json.Unmarshal(orgDetailsJSON, &tenant.OrgDetails)
	}

	setupToken := uuid.New().String()
	setupTokenHash := totp.HashToken(setupToken)
	expiresAt := now.Add(24 * time.Hour)

	_, err = tx.Exec(
		`INSERT INTO totp_recovery_tokens (operator_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		operatorID, setupTokenHash, expiresAt,
	)
	if err != nil {
		return domain.Tenant{}, domain.Operator{}, "", err
	}

	if err := tx.Commit(); err != nil {
		return domain.Tenant{}, domain.Operator{}, "", err
	}
	return tenant, op, setupToken, nil
}

// GetTOTPSetupInfo retrieves or creates the TOTP secret for an operator given a valid setup token.
func (p *PostgresStore) GetTOTPSetupInfo(rawSetupToken string) (domain.Operator, domain.Tenant, string, error) {
	tokenHash := totp.HashToken(rawSetupToken)

	var opID uuid.UUID
	err := p.db.QueryRow(
		`SELECT operator_id FROM totp_recovery_tokens
		 WHERE token_hash = $1 AND used_at IS NULL AND expires_at > CURRENT_TIMESTAMP`,
		tokenHash,
	).Scan(&opID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Operator{}, domain.Tenant{}, "", ErrTokenNotFound
		}
		return domain.Operator{}, domain.Tenant{}, "", err
	}

	row := p.db.QueryRow(
		`SELECT id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		        totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		        last_login_at, created_at, updated_at
		 FROM operators WHERE id = $1`, opID,
	)
	op, _, encryptedSecret, err := scanOperatorRow(row)
	if err != nil {
		return domain.Operator{}, domain.Tenant{}, "", err
	}

	var tenant domain.Tenant
	var orgDetailsJSON []byte
	err = p.db.QueryRow(
		`SELECT id, name, setup_step, is_setup_complete, org_details, created_at, updated_at
		 FROM tenants WHERE id = $1`, op.TenantID,
	).Scan(&tenant.ID, &tenant.Name, &tenant.SetupStep, &tenant.IsSetupComplete, &orgDetailsJSON, &tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		return domain.Operator{}, domain.Tenant{}, "", err
	}
	if len(orgDetailsJSON) > 0 {
		_ = json.Unmarshal(orgDetailsJSON, &tenant.OrgDetails)
	}

	var plainSecret string
	if encryptedSecret == "" || op.TotpSetupRequired {
		plainSecret, err = totp.GenerateSecret()
		if err != nil {
			return domain.Operator{}, domain.Tenant{}, "", err
		}
		encrypted, err := totp.EncryptSecret(plainSecret)
		if err != nil {
			return domain.Operator{}, domain.Tenant{}, "", err
		}
		_, err = p.db.Exec(`UPDATE operators SET totp_secret_encrypted = $1 WHERE id = $2`, encrypted, op.ID)
		if err != nil {
			return domain.Operator{}, domain.Tenant{}, "", err
		}
	} else {
		plainSecret, err = totp.DecryptSecret(encryptedSecret)
		if err != nil {
			return domain.Operator{}, domain.Tenant{}, "", err
		}
	}

	return op, tenant, plainSecret, nil
}

// VerifyTOTPSetup verifies the initial TOTP code during setup, generates backup codes, and creates an operator session.
func (p *PostgresStore) VerifyTOTPSetup(rawSetupToken, code string) (domain.Operator, []string, domain.Session, error) {
	tokenHash := totp.HashToken(rawSetupToken)

	tx, err := p.db.Begin()
	if err != nil {
		return domain.Operator{}, nil, domain.Session{}, err
	}
	defer tx.Rollback()

	var tokenID, opID uuid.UUID
	err = tx.QueryRow(
		`SELECT id, operator_id FROM totp_recovery_tokens
		 WHERE token_hash = $1 AND used_at IS NULL AND expires_at > CURRENT_TIMESTAMP`,
		tokenHash,
	).Scan(&tokenID, &opID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Operator{}, nil, domain.Session{}, ErrTokenNotFound
		}
		return domain.Operator{}, nil, domain.Session{}, err
	}

	row := tx.QueryRow(
		`SELECT id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		        totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		        last_login_at, created_at, updated_at
		 FROM operators WHERE id = $1`, opID,
	)
	op, _, encryptedSecret, err := scanOperatorRow(row)
	if err != nil {
		return domain.Operator{}, nil, domain.Session{}, err
	}

	if encryptedSecret == "" {
		return domain.Operator{}, nil, domain.Session{}, errors.New("no totp secret found for operator")
	}

	plainSecret, err := totp.DecryptSecret(encryptedSecret)
	if err != nil {
		return domain.Operator{}, nil, domain.Session{}, err
	}

	if !totp.VerifyCode(plainSecret, code, time.Now()) {
		return domain.Operator{}, nil, domain.Session{}, ErrInvalidTOTPCode
	}

	now := time.Now().UTC()
	// Mark setup token as used
	_, err = tx.Exec(`UPDATE totp_recovery_tokens SET used_at = $1 WHERE id = $2`, now, tokenID)
	if err != nil {
		return domain.Operator{}, nil, domain.Session{}, err
	}

	// Update operator
	row = tx.QueryRow(
		`UPDATE operators
		 SET totp_verified_at = $1, totp_setup_required = false, updated_at = $1, last_login_at = $1
		 WHERE id = $2
		 RETURNING id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		           totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		           last_login_at, created_at, updated_at`,
		now, op.ID,
	)
	op, _, _, err = scanOperatorRow(row)
	if err != nil {
		return domain.Operator{}, nil, domain.Session{}, err
	}

	// Revoke old backup codes
	_, _ = tx.Exec(`DELETE FROM totp_backup_codes WHERE operator_id = $1`, op.ID)

	// Generate 10 new backup codes
	backupCodes, err := totp.GenerateBackupCodes(10)
	if err != nil {
		return domain.Operator{}, nil, domain.Session{}, err
	}

	for _, bc := range backupCodes {
		hash, err := totp.HashBackupCode(bc)
		if err != nil {
			return domain.Operator{}, nil, domain.Session{}, err
		}
		_, err = tx.Exec(
			`INSERT INTO totp_backup_codes (operator_id, code_hash) VALUES ($1, $2)`,
			op.ID, hash,
		)
		if err != nil {
			return domain.Operator{}, nil, domain.Session{}, err
		}
	}

	// Create session
	expires := now.Add(8 * time.Hour)
	var session domain.Session
	err = tx.QueryRow(
		`INSERT INTO sessions (operator_id, tenant_id, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, operator_id, tenant_id, expires_at, created_at`,
		op.ID, op.TenantID, expires,
	).Scan(&session.ID, &session.OperatorID, &session.TenantID, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		return domain.Operator{}, nil, domain.Session{}, err
	}

	// Audit log
	_, _ = tx.Exec(
		`INSERT INTO recovery_audit_log (tenant_id, operator_id, action, details)
		 VALUES ($1, $2, 'TOTP_SETUP_COMPLETED', '{"method": "totp_setup"}'::jsonb)`,
		op.TenantID, op.ID,
	)

	if err := tx.Commit(); err != nil {
		return domain.Operator{}, nil, domain.Session{}, err
	}

	return op, backupCodes, session, nil
}

// CreateInvitation creates a new invitation and returns the raw token.
func (p *PostgresStore) CreateInvitation(tenantID uuid.UUID, createdBy *uuid.UUID, recipient, channel, role, whatsappNumber, email string) (domain.Invitation, string, error) {
	if role == "" {
		role = "operator"
	}
	if channel == "" {
		channel = "whatsapp"
	}
	rawToken := uuid.New().String()
	tokenHash := totp.HashToken(rawToken)
	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)

	var whatsappVal, emailVal *string
	if whatsappNumber != "" {
		whatsappVal = &whatsappNumber
	}
	if email != "" {
		emailVal = &email
	}

	tx, err := p.db.Begin()
	if err != nil {
		return domain.Invitation{}, "", err
	}
	defer tx.Rollback()

	var inv domain.Invitation
	var whatsappOut, emailOut sql.NullString
	err = tx.QueryRow(
		`INSERT INTO invitations (tenant_id, role, channel, recipient, whatsapp_number, email, token_hash, status, expires_at, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', $8, $9)
		 RETURNING id, tenant_id, role, channel, recipient, whatsapp_number, email, token_hash, status, expires_at, accepted_at, created_by, created_at`,
		tenantID, role, channel, recipient, whatsappVal, emailVal, tokenHash, expiresAt, createdBy,
	).Scan(
		&inv.ID, &inv.TenantID, &inv.Role, &inv.Channel, &inv.Recipient,
		&whatsappOut, &emailOut, &inv.TokenHash, &inv.Status, &inv.ExpiresAt,
		&inv.AcceptedAt, &inv.CreatedBy, &inv.CreatedAt,
	)
	if err != nil {
		return domain.Invitation{}, "", err
	}
	inv.WhatsappNumber = whatsappOut.String
	inv.Email = emailOut.String

	if channel == "whatsapp" {
		_, err = tx.Exec(
			`INSERT INTO whatsapp_invitation_delivery (invitation_id, status)
			 VALUES ($1, 'sent')`,
			inv.ID,
		)
		if err != nil {
			return domain.Invitation{}, "", err
		}
	}

	if err := tx.Commit(); err != nil {
		return domain.Invitation{}, "", err
	}
	return inv, rawToken, nil
}

// GetInvitationByToken retrieves an invitation and tenant by raw token.
func (p *PostgresStore) GetInvitationByToken(rawToken string) (domain.Invitation, domain.Tenant, error) {
	tokenHash := totp.HashToken(rawToken)

	var inv domain.Invitation
	var whatsappOut, emailOut sql.NullString
	err := p.db.QueryRow(
		`SELECT id, tenant_id, role, channel, recipient, whatsapp_number, email, token_hash, status, expires_at, accepted_at, created_by, created_at
		 FROM invitations
		 WHERE token_hash = $1 AND status = 'pending' AND expires_at > CURRENT_TIMESTAMP`,
		tokenHash,
	).Scan(
		&inv.ID, &inv.TenantID, &inv.Role, &inv.Channel, &inv.Recipient,
		&whatsappOut, &emailOut, &inv.TokenHash, &inv.Status, &inv.ExpiresAt,
		&inv.AcceptedAt, &inv.CreatedBy, &inv.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Invitation{}, domain.Tenant{}, ErrInvitationNotFound
		}
		return domain.Invitation{}, domain.Tenant{}, err
	}
	inv.WhatsappNumber = whatsappOut.String
	inv.Email = emailOut.String

	var tenant domain.Tenant
	var orgDetailsJSON []byte
	err = p.db.QueryRow(
		`SELECT id, name, setup_step, is_setup_complete, org_details, created_at, updated_at
		 FROM tenants WHERE id = $1`, inv.TenantID,
	).Scan(&tenant.ID, &tenant.Name, &tenant.SetupStep, &tenant.IsSetupComplete, &orgDetailsJSON, &tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		return domain.Invitation{}, domain.Tenant{}, err
	}
	if len(orgDetailsJSON) > 0 {
		_ = json.Unmarshal(orgDetailsJSON, &tenant.OrgDetails)
	}

	return inv, tenant, nil
}

// AcceptInvitationAndSignupOperator creates an operator from an invitation and returns a setup token.
func (p *PostgresStore) AcceptInvitationAndSignupOperator(rawToken, name, whatsappNumber, email string) (domain.Operator, string, error) {
	tokenHash := totp.HashToken(rawToken)

	tx, err := p.db.Begin()
	if err != nil {
		return domain.Operator{}, "", err
	}
	defer tx.Rollback()

	var inv domain.Invitation
	var whatsappOut, emailOut sql.NullString
	err = tx.QueryRow(
		`SELECT id, tenant_id, role, channel, recipient, whatsapp_number, email, token_hash, status, expires_at, accepted_at, created_by, created_at
		 FROM invitations
		 WHERE token_hash = $1 AND status = 'pending' AND expires_at > CURRENT_TIMESTAMP`,
		tokenHash,
	).Scan(
		&inv.ID, &inv.TenantID, &inv.Role, &inv.Channel, &inv.Recipient,
		&whatsappOut, &emailOut, &inv.TokenHash, &inv.Status, &inv.ExpiresAt,
		&inv.AcceptedAt, &inv.CreatedBy, &inv.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Operator{}, "", ErrInvitationNotFound
		}
		return domain.Operator{}, "", err
	}

	if whatsappNumber == "" && whatsappOut.Valid {
		whatsappNumber = whatsappOut.String
	}
	if email == "" && emailOut.Valid {
		email = emailOut.String
	}

	var whatsappVal, emailVal *string
	if whatsappNumber != "" {
		whatsappVal = &whatsappNumber
	}
	if email != "" {
		emailVal = &email
	}

	// Generate TOTP secret
	plainSecret, err := totp.GenerateSecret()
	if err != nil {
		return domain.Operator{}, "", err
	}
	encryptedSecret, err := totp.EncryptSecret(plainSecret)
	if err != nil {
		return domain.Operator{}, "", err
	}

	row := tx.QueryRow(
		`INSERT INTO operators (tenant_id, name, email, whatsapp_number, role, is_active, totp_secret_encrypted, totp_setup_required)
		 VALUES ($1, $2, $3, $4, $5, true, $6, true)
		 RETURNING id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		           totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		           last_login_at, created_at, updated_at`,
		inv.TenantID, name, emailVal, whatsappVal, inv.Role, encryptedSecret,
	)
	op, _, _, err := scanOperatorRow(row)
	if err != nil {
		if isUniqueViolation(err, "operators_whatsapp_number_key") {
			return domain.Operator{}, "", ErrDuplicateWhatsapp
		}
		return domain.Operator{}, "", err
	}

	now := time.Now().UTC()
	_, err = tx.Exec(
		`UPDATE invitations SET status = 'accepted', accepted_at = $1 WHERE id = $2`,
		now, inv.ID,
	)
	if err != nil {
		return domain.Operator{}, "", err
	}

	setupToken := uuid.New().String()
	setupTokenHash := totp.HashToken(setupToken)
	expiresAt := now.Add(24 * time.Hour)

	_, err = tx.Exec(
		`INSERT INTO totp_recovery_tokens (operator_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		op.ID, setupTokenHash, expiresAt,
	)
	if err != nil {
		return domain.Operator{}, "", err
	}

	if err := tx.Commit(); err != nil {
		return domain.Operator{}, "", err
	}
	return op, setupToken, nil
}

// ListInvitations lists invitations for a tenant.
func (p *PostgresStore) ListInvitations(tenantID uuid.UUID) ([]domain.Invitation, error) {
	rows, err := p.db.Query(
		`SELECT id, tenant_id, role, channel, recipient, whatsapp_number, email, token_hash, status, expires_at, accepted_at, created_by, created_at
		 FROM invitations WHERE tenant_id = $1 ORDER BY created_at DESC`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []domain.Invitation
	for rows.Next() {
		var inv domain.Invitation
		var whatsappOut, emailOut sql.NullString
		err := rows.Scan(
			&inv.ID, &inv.TenantID, &inv.Role, &inv.Channel, &inv.Recipient,
			&whatsappOut, &emailOut, &inv.TokenHash, &inv.Status, &inv.ExpiresAt,
			&inv.AcceptedAt, &inv.CreatedBy, &inv.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		inv.WhatsappNumber = whatsappOut.String
		inv.Email = emailOut.String
		invitations = append(invitations, inv)
	}
	return invitations, rows.Err()
}

// RevokeInvitation revokes a pending invitation.
func (p *PostgresStore) RevokeInvitation(tenantID, invitationID uuid.UUID) error {
	res, err := p.db.Exec(
		`UPDATE invitations SET status = 'revoked' WHERE tenant_id = $1 AND id = $2 AND status = 'pending'`,
		tenantID, invitationID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrInvitationNotFound
	}
	return nil
}

// VerifyBackupCodeAndLogin checks a backup code against unused codes, burns it, and creates a session.
func (p *PostgresStore) VerifyBackupCodeAndLogin(tenantID uuid.UUID, identifier, code string) (domain.Operator, domain.Session, int, error) {
	op, _, _, err := p.FindOperatorByIdentifier(tenantID, identifier)
	if err != nil {
		return domain.Operator{}, domain.Session{}, 0, err
	}
	if !op.IsActive {
		return domain.Operator{}, domain.Session{}, 0, errors.New("operator account is disabled")
	}

	tx, err := p.db.Begin()
	if err != nil {
		return domain.Operator{}, domain.Session{}, 0, err
	}
	defer tx.Rollback()

	rows, err := tx.Query(
		`SELECT id, code_hash FROM totp_backup_codes WHERE operator_id = $1 AND used_at IS NULL`,
		op.ID,
	)
	if err != nil {
		return domain.Operator{}, domain.Session{}, 0, err
	}
	defer rows.Close()

	var matchedID *uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		var codeHash string
		if err := rows.Scan(&id, &codeHash); err != nil {
			return domain.Operator{}, domain.Session{}, 0, err
		}
		if totp.VerifyBackupCode(code, codeHash) {
			matchedID = &id
			break
		}
	}
	if matchedID == nil {
		return domain.Operator{}, domain.Session{}, 0, ErrInvalidBackupCode
	}

	now := time.Now().UTC()
	_, err = tx.Exec(`UPDATE totp_backup_codes SET used_at = $1 WHERE id = $2`, now, *matchedID)
	if err != nil {
		return domain.Operator{}, domain.Session{}, 0, err
	}

	// Update last_login_at
	_, _ = tx.Exec(`UPDATE operators SET last_login_at = $1 WHERE id = $2`, now, op.ID)

	// Create session
	expires := now.Add(8 * time.Hour)
	var session domain.Session
	err = tx.QueryRow(
		`INSERT INTO sessions (operator_id, tenant_id, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, operator_id, tenant_id, expires_at, created_at`,
		op.ID, op.TenantID, expires,
	).Scan(&session.ID, &session.OperatorID, &session.TenantID, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		return domain.Operator{}, domain.Session{}, 0, err
	}

	// Count remaining unused backup codes
	var remaining int
	err = tx.QueryRow(
		`SELECT count(*) FROM totp_backup_codes WHERE operator_id = $1 AND used_at IS NULL`,
		op.ID,
	).Scan(&remaining)
	if err != nil {
		return domain.Operator{}, domain.Session{}, 0, err
	}

	// Audit log
	_, _ = tx.Exec(
		`INSERT INTO recovery_audit_log (tenant_id, operator_id, action, details)
		 VALUES ($1, $2, 'BACKUP_CODE_LOGIN', '{"method": "backup_code"}'::jsonb)`,
		op.TenantID, op.ID,
	)

	if err := tx.Commit(); err != nil {
		return domain.Operator{}, domain.Session{}, 0, err
	}
	return op, session, remaining, nil
}

// RegenerateBackupCodes replaces old backup codes with 10 new ones.
func (p *PostgresStore) RegenerateBackupCodes(operatorID uuid.UUID) ([]string, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var tenantID uuid.UUID
	err = tx.QueryRow(`SELECT tenant_id FROM operators WHERE id = $1`, operatorID).Scan(&tenantID)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`DELETE FROM totp_backup_codes WHERE operator_id = $1`, operatorID)
	if err != nil {
		return nil, err
	}

	codes, err := totp.GenerateBackupCodes(10)
	if err != nil {
		return nil, err
	}

	for _, c := range codes {
		hash, err := totp.HashBackupCode(c)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(
			`INSERT INTO totp_backup_codes (operator_id, code_hash) VALUES ($1, $2)`,
			operatorID, hash,
		)
		if err != nil {
			return nil, err
		}
	}

	_, _ = tx.Exec(
		`INSERT INTO recovery_audit_log (tenant_id, operator_id, action, details)
		 VALUES ($1, $2, 'BACKUP_CODES_REGENERATED', '{"count": 10}'::jsonb)`,
		tenantID, operatorID,
	)

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return codes, nil
}

// CountRemainingBackupCodes returns the count of unused backup codes for an operator.
func (p *PostgresStore) CountRemainingBackupCodes(operatorID uuid.UUID) (int, error) {
	var count int
	err := p.db.QueryRow(
		`SELECT count(*) FROM totp_backup_codes WHERE operator_id = $1 AND used_at IS NULL`,
		operatorID,
	).Scan(&count)
	return count, err
}

// ResetOperatorTOTPByAdmin resets an operator's TOTP configuration and generates a recovery token.
func (p *PostgresStore) ResetOperatorTOTPByAdmin(tenantID, adminID, targetOperatorID uuid.UUID) (string, error) {
	admin, err := p.GetOperatorByID(tenantID, adminID)
	if err != nil || admin.Role != "admin" {
		return "", ErrUnauthorizedAdmin
	}

	tx, err := p.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	res, err := tx.Exec(
		`UPDATE operators
		 SET totp_setup_required = true, totp_verified_at = NULL, totp_secret_encrypted = NULL, updated_at = $1
		 WHERE tenant_id = $2 AND id = $3`,
		now, tenantID, targetOperatorID,
	)
	if err != nil {
		return "", err
	}
	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return "", ErrOperatorNotFound
	}

	// Delete backup codes and active sessions for target operator
	_, _ = tx.Exec(`DELETE FROM totp_backup_codes WHERE operator_id = $1`, targetOperatorID)
	_, _ = tx.Exec(`DELETE FROM sessions WHERE operator_id = $1`, targetOperatorID)

	setupToken := uuid.New().String()
	setupTokenHash := totp.HashToken(setupToken)
	expiresAt := now.Add(24 * time.Hour)

	_, err = tx.Exec(
		`INSERT INTO totp_recovery_tokens (operator_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		targetOperatorID, setupTokenHash, expiresAt,
	)
	if err != nil {
		return "", err
	}

	_, _ = tx.Exec(
		`INSERT INTO recovery_audit_log (tenant_id, operator_id, action, details)
		 VALUES ($1, $2, 'ADMIN_TOTP_RESET', json_build_object('admin_id', $3::text)::jsonb)`,
		tenantID, targetOperatorID, adminID.String(),
	)

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return setupToken, nil
}

// RequestRecovery creates a recovery token for an operator.
func (p *PostgresStore) RequestRecovery(tenantID uuid.UUID, identifier string) (string, error) {
	op, _, _, err := p.FindOperatorByIdentifier(tenantID, identifier)
	if err != nil {
		return "", err
	}

	token := uuid.New().String()
	tokenHash := totp.HashToken(token)
	expiresAt := time.Now().UTC().Add(1 * time.Hour)

	_, err = p.db.Exec(
		`INSERT INTO totp_recovery_tokens (operator_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		op.ID, tokenHash, expiresAt,
	)
	if err != nil {
		return "", err
	}

	_ = p.LogRecoveryAudit(tenantID, &op.ID, "RECOVERY_REQUESTED", "", "", map[string]any{"identifier": identifier})
	return token, nil
}

// ValidateRecoveryToken checks if a recovery token is valid.
func (p *PostgresStore) ValidateRecoveryToken(rawToken string) (domain.RecoveryToken, domain.Operator, error) {
	tokenHash := totp.HashToken(rawToken)

	var rec domain.RecoveryToken
	err := p.db.QueryRow(
		`SELECT id, operator_id, token_hash, expires_at, used_at, created_at
		 FROM totp_recovery_tokens
		 WHERE token_hash = $1 AND used_at IS NULL AND expires_at > CURRENT_TIMESTAMP`,
		tokenHash,
	).Scan(&rec.ID, &rec.OperatorID, &rec.TokenHash, &rec.ExpiresAt, &rec.UsedAt, &rec.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.RecoveryToken{}, domain.Operator{}, ErrTokenNotFound
		}
		return domain.RecoveryToken{}, domain.Operator{}, err
	}

	row := p.db.QueryRow(
		`SELECT id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		        totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		        last_login_at, created_at, updated_at
		 FROM operators WHERE id = $1`, rec.OperatorID,
	)
	op, _, _, err := scanOperatorRow(row)
	if err != nil {
		return domain.RecoveryToken{}, domain.Operator{}, err
	}
	return rec, op, nil
}

// GetTenantSetupStatus gets the wizard setup status for a tenant.
func (p *PostgresStore) GetTenantSetupStatus(tenantID uuid.UUID) (domain.TenantSetupStatus, error) {
	var status domain.TenantSetupStatus
	var orgDetailsJSON []byte
	err := p.db.QueryRow(
		`SELECT setup_step, is_setup_complete, org_details FROM tenants WHERE id = $1`,
		tenantID,
	).Scan(&status.SetupStep, &status.IsSetupComplete, &orgDetailsJSON)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.TenantSetupStatus{}, ErrTenantNotFound
		}
		return domain.TenantSetupStatus{}, err
	}
	if len(orgDetailsJSON) > 0 {
		_ = json.Unmarshal(orgDetailsJSON, &status.OrgDetails)
	}
	return status, nil
}

// UpdateTenantSetup updates the setup step and org details for a tenant.
func (p *PostgresStore) UpdateTenantSetup(tenantID uuid.UUID, setupStep int, orgDetails map[string]any) (domain.TenantSetupStatus, error) {
	orgJSON, err := json.Marshal(orgDetails)
	if err != nil {
		return domain.TenantSetupStatus{}, err
	}

	var status domain.TenantSetupStatus
	var orgDetailsJSON []byte
	err = p.db.QueryRow(
		`UPDATE tenants
		 SET setup_step = $1, org_details = $2, updated_at = CURRENT_TIMESTAMP
		 WHERE id = $3
		 RETURNING setup_step, is_setup_complete, org_details`,
		setupStep, orgJSON, tenantID,
	).Scan(&status.SetupStep, &status.IsSetupComplete, &orgDetailsJSON)
	if err != nil {
		return domain.TenantSetupStatus{}, err
	}
	if len(orgDetailsJSON) > 0 {
		_ = json.Unmarshal(orgDetailsJSON, &status.OrgDetails)
	}
	return status, nil
}

// CompleteTenantSetup marks tenant onboarding as complete.
func (p *PostgresStore) CompleteTenantSetup(tenantID uuid.UUID) error {
	res, err := p.db.Exec(
		`UPDATE tenants SET is_setup_complete = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1`,
		tenantID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return ErrTenantNotFound
	}
	return nil
}

// LogRecoveryAudit logs a security/recovery audit entry.
func (p *PostgresStore) LogRecoveryAudit(tenantID uuid.UUID, operatorID *uuid.UUID, action, ip, userAgent string, details map[string]any) error {
	detailsJSON, _ := json.Marshal(details)
	_, err := p.db.Exec(
		`INSERT INTO recovery_audit_log (tenant_id, operator_id, action, ip_address, user_agent, details)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		tenantID, operatorID, action, ip, userAgent, detailsJSON,
	)
	return err
}

// GetTenantByID returns a tenant by ID.
func (p *PostgresStore) GetTenantByID(tenantID uuid.UUID) (domain.Tenant, error) {
	var tenant domain.Tenant
	var orgDetailsJSON []byte
	err := p.db.QueryRow(
		`SELECT id, name, setup_step, is_setup_complete, org_details, created_at, updated_at
		 FROM tenants WHERE id = $1`,
		tenantID,
	).Scan(&tenant.ID, &tenant.Name, &tenant.SetupStep, &tenant.IsSetupComplete, &orgDetailsJSON, &tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Tenant{}, ErrTenantNotFound
		}
		return domain.Tenant{}, err
	}
	if len(orgDetailsJSON) > 0 {
		_ = json.Unmarshal(orgDetailsJSON, &tenant.OrgDetails)
	}
	return tenant, nil
}

// TrackInvitationDelivery records the delivery attempt of an invitation via WhatsApp.
func (p *PostgresStore) TrackInvitationDelivery(invitationID uuid.UUID, status, messageID, errorMessage string) error {
	var errVal *string
	if errorMessage != "" {
		errVal = &errorMessage
	}
	var msgVal *string
	if messageID != "" {
		msgVal = &messageID
	}
	_, err := p.db.Exec(
		`INSERT INTO whatsapp_invitation_delivery (invitation_id, status, message_id, error_message)
		 VALUES ($1, $2, $3, $4)`,
		invitationID, status, msgVal, errVal,
	)
	return err
}
