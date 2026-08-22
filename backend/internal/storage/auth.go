package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
	"golang.org/x/crypto/bcrypt"
)

// OperatorAuth is the interface for dashboard authentication, sessions,
// and tenant onboarding. The concrete *PostgresStore satisfies it.
type OperatorAuth interface {
	FindOperatorByEmail(tenantID uuid.UUID, email string) (domain.Operator, string, error)
	FindOperatorByIdentifier(tenantID uuid.UUID, identifier string) (op domain.Operator, plainTotpSecret string, passwordHash string, err error)
	GetOperatorByID(tenantID, operatorID uuid.UUID) (domain.Operator, error)
	GetOperatorByIDWithSecret(tenantID, operatorID uuid.UUID) (domain.Operator, string, error)
	ListOperators(tenantID uuid.UUID) ([]domain.Operator, error)
	CreateOperator(tenantID uuid.UUID, email, name, role, password string) (domain.Operator, error)
	CreateSession(operatorID, tenantID uuid.UUID, ttl time.Duration) (domain.Session, error)
	GetSessionByID(sessionID uuid.UUID) (domain.Session, error)
	DeleteSession(sessionID uuid.UUID) error
	TouchSession(sessionID uuid.UUID) error

	// Tenant onboarding & TOTP
	SignupTenant(tenantName, adminName, adminEmail, adminWhatsapp string) (domain.Tenant, domain.Operator, string, error)
	VerifyEmailToken(rawToken string) (domain.Tenant, domain.Operator, string, error)
	GetTOTPSetupInfo(rawSetupToken string) (domain.Operator, domain.Tenant, string, error)
	VerifyTOTPSetup(rawSetupToken, code string) (domain.Operator, []string, domain.Session, error)
	CreateInvitation(tenantID uuid.UUID, createdBy *uuid.UUID, recipient, channel, role, whatsappNumber, email string) (domain.Invitation, string, error)
	GetInvitationByToken(rawToken string) (domain.Invitation, domain.Tenant, error)
	AcceptInvitationAndSignupOperator(rawToken, name, whatsappNumber, email string) (domain.Operator, string, error)
	ListInvitations(tenantID uuid.UUID) ([]domain.Invitation, error)
	RevokeInvitation(tenantID, invitationID uuid.UUID) error
	VerifyBackupCodeAndLogin(tenantID uuid.UUID, identifier, code string) (domain.Operator, domain.Session, int, error)
	RegenerateBackupCodes(operatorID uuid.UUID) ([]string, error)
	CountRemainingBackupCodes(operatorID uuid.UUID) (int, error)
	ResetOperatorTOTPByAdmin(tenantID, adminID, targetOperatorID uuid.UUID) (string, error)
	RequestRecovery(tenantID uuid.UUID, identifier string) (string, error)
	ValidateRecoveryToken(rawToken string) (domain.RecoveryToken, domain.Operator, error)
	GetTenantSetupStatus(tenantID uuid.UUID) (domain.TenantSetupStatus, error)
	UpdateTenantSetup(tenantID uuid.UUID, name string, setupStep int, orgDetails map[string]any) (domain.TenantSetupStatus, error)
	CompleteTenantSetup(tenantID uuid.UUID) error
	GetTenantByID(tenantID uuid.UUID) (domain.Tenant, error)
	FindTenantBySlug(slug string) (domain.Tenant, error)
	TrackInvitationDelivery(invitationID uuid.UUID, status, messageID, errorMessage string) error
	LogRecoveryAudit(tenantID uuid.UUID, operatorID *uuid.UUID, action, ip, userAgent string, details map[string]any) error
}

// CreateOperator inserts a new operator with a bcrypt-hashed password.
func (p *PostgresStore) CreateOperator(tenantID uuid.UUID, email, name, role, password string) (domain.Operator, error) {
	role = domain.NormalizeRole(role)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.Operator{}, err
	}
	row := p.db.QueryRow(
		`INSERT INTO operators (tenant_id, email, name, password_hash, role) VALUES ($1,$2,$3,$4,$5)
		 RETURNING id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		           totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		           last_login_at, created_at, updated_at`,
		tenantID, email, name, string(hash), role)
	op, _, _, err := scanOperatorRow(row)
	return op, err
}

// CreateSession inserts a new session and returns it.
func (p *PostgresStore) CreateSession(operatorID, tenantID uuid.UUID, ttl time.Duration) (domain.Session, error) {
	expires := time.Now().UTC().Add(ttl)
	var s domain.Session
	err := p.db.QueryRow(
		`INSERT INTO sessions (operator_id, tenant_id, expires_at) VALUES ($1,$2,$3)
		 RETURNING id, operator_id, tenant_id, expires_at, created_at`, operatorID, tenantID, expires).Scan(
		&s.ID, &s.OperatorID, &s.TenantID, &s.ExpiresAt, &s.CreatedAt)
	return s, err
}

// GetSessionByID returns a session if it exists and has not expired.
func (p *PostgresStore) GetSessionByID(sessionID uuid.UUID) (domain.Session, error) {
	var s domain.Session
	err := p.db.QueryRow(
		`SELECT id, operator_id, tenant_id, expires_at, created_at FROM sessions WHERE id=$1 AND expires_at > CURRENT_TIMESTAMP`,
		sessionID).Scan(&s.ID, &s.OperatorID, &s.TenantID, &s.ExpiresAt, &s.CreatedAt)
	return s, err
}

// DeleteSession removes a session (logout).
func (p *PostgresStore) DeleteSession(sessionID uuid.UUID) error {
	_, err := p.db.Exec(`DELETE FROM sessions WHERE id=$1`, sessionID)
	return err
}

// TouchSession updates the last-used timestamp and prunes expired sessions.
func (p *PostgresStore) TouchSession(sessionID uuid.UUID) error {
	_, err := p.db.Exec(`UPDATE sessions SET last_used_at=CURRENT_TIMESTAMP WHERE id=$1`, sessionID)
	if err != nil {
		return err
	}
	_, _ = p.db.Exec(`DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP - INTERVAL '1 day'`)
	return nil
}

// VerifyOperatorPassword checks a plaintext password against a bcrypt hash.
func VerifyOperatorPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// ErrOperatorNotFound is returned when no operator matches the tenant+email.
var ErrOperatorNotFound = errors.New("operator not found")

var _ = sql.ErrNoRows
var _ = context.Background
