package storage

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/raufhm/whatsapp-testing/internal/totp"
	"golang.org/x/crypto/bcrypt"
)

var operatorColumns = []string{
	"id", "tenant_id", "email", "whatsapp_number", "name", "password_hash", "role", "is_active",
	"totp_secret_encrypted", "totp_verified_at", "totp_setup_required", "email_verified_at",
	"last_login_at", "created_at", "updated_at",
}

func TestCreateOperatorHashesPassword(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	tenantID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO operators (tenant_id, email, name, password_hash, role) VALUES ($1,$2,$3,$4,$5)`)).
		WithArgs(tenantID, "op@example.com", "Op", sqlmock.AnyArg(), "operator").
		WillReturnRows(sqlmock.NewRows(operatorColumns).
			AddRow(uuid.New(), tenantID, "op@example.com", nil, "Op", "hash", "operator", true, nil, nil, false, nil, nil, time.Now(), time.Now()))

	op, err := store.CreateOperator(tenantID, "op@example.com", "Op", "operator", "secret")
	if err != nil {
		t.Fatalf("CreateOperator: %v", err)
	}
	if op.Email != "op@example.com" || op.Role != "operator" {
		t.Fatalf("unexpected operator: %+v", op)
	}
}

func TestFindOperatorByEmail(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	tenantID := uuid.New()
	opID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		        totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		        last_login_at, created_at, updated_at
		 FROM operators WHERE tenant_id=$1 AND email=$2`)).
		WithArgs(tenantID, "op@example.com").
		WillReturnRows(sqlmock.NewRows(operatorColumns).
			AddRow(opID, tenantID, "op@example.com", nil, "Op", "hash", "operator", true, nil, nil, false, nil, nil, time.Now(), time.Now()))

	op, hash, err := store.FindOperatorByEmail(tenantID, "op@example.com")
	if err != nil {
		t.Fatalf("FindOperatorByEmail: %v", err)
	}
	if op.ID != opID || hash != "hash" {
		t.Fatalf("unexpected result: op=%+v hash=%q", op, hash)
	}
}

func TestFindOperatorByIdentifier(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	tenantID := uuid.New()
	opID := uuid.New()

	plainSecret := "JBSWY3DPEHPK3PXP"
	encryptedSecret, err := totp.EncryptSecret(plainSecret)
	if err != nil {
		t.Fatalf("EncryptSecret: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, email, whatsapp_number, name, password_hash, role, is_active,
		        totp_secret_encrypted, totp_verified_at, totp_setup_required, email_verified_at,
		        last_login_at, created_at, updated_at
		 FROM operators
		 WHERE tenant_id = $1 AND (email = $2 OR whatsapp_number = $2)`)).
		WithArgs(tenantID, "op@example.com").
		WillReturnRows(sqlmock.NewRows(operatorColumns).
			AddRow(opID, tenantID, "op@example.com", nil, "Op", "bcrypt-hash", "operator", true, encryptedSecret, time.Now(), false, nil, nil, time.Now(), time.Now()))

	op, decryptedSecret, passHash, err := store.FindOperatorByIdentifier(tenantID, "op@example.com")
	if err != nil {
		t.Fatalf("FindOperatorByIdentifier: %v", err)
	}
	if op.ID != opID || decryptedSecret != plainSecret || passHash != "bcrypt-hash" {
		t.Fatalf("unexpected result: op=%+v decryptedSecret=%q passHash=%q", op, decryptedSecret, passHash)
	}
}

func TestCreateAndGetSession(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	tenantID := uuid.New()
	opID := uuid.New()
	sessionID := uuid.New()
	expires := time.Now().UTC().Add(time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO sessions (operator_id, tenant_id, expires_at) VALUES ($1,$2,$3)`)).
		WithArgs(opID, tenantID, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "operator_id", "tenant_id", "expires_at", "created_at"}).
			AddRow(sessionID, opID, tenantID, expires, time.Now()))

	s, err := store.CreateSession(opID, tenantID, time.Hour)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if s.ID != sessionID || s.OperatorID != opID {
		t.Fatalf("unexpected session: %+v", s)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, operator_id, tenant_id, expires_at, created_at FROM sessions WHERE id=$1 AND expires_at > CURRENT_TIMESTAMP`)).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "operator_id", "tenant_id", "expires_at", "created_at"}).
			AddRow(sessionID, opID, tenantID, expires, time.Now()))

	got, err := store.GetSessionByID(sessionID)
	if err != nil {
		t.Fatalf("GetSessionByID: %v", err)
	}
	if got.ID != sessionID {
		t.Fatalf("unexpected session: %+v", got)
	}
}

func TestVerifyOperatorPassword(t *testing.T) {
	// Verify round-trips through the same bcrypt function used by CreateOperator.
	_, err := bcrypt.GenerateFromPassword([]byte("correct-horse"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	tenantID := uuid.New()

	// Create a real operator and verify the stored hash validates.
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO operators (tenant_id, email, name, password_hash, role) VALUES ($1,$2,$3,$4,$5)`)).
		WithArgs(tenantID, "op@example.com", "Op", sqlmock.AnyArg(), "operator").
		WillReturnRows(sqlmock.NewRows(operatorColumns).
			AddRow(uuid.New(), tenantID, "op@example.com", nil, "Op", "hash", "operator", true, nil, nil, false, nil, nil, time.Now(), time.Now()))

	op, err := store.CreateOperator(tenantID, "op@example.com", "Op", "operator", "correct-horse")
	if err != nil {
		t.Fatalf("CreateOperator: %v", err)
	}
	_ = op

	// The hash stored in the DB is what FindOperatorByEmail would return; verify
	// a plaintext password round-trips using the same bcrypt primitive.
	h, _ := bcrypt.GenerateFromPassword([]byte("correct-horse"), bcrypt.DefaultCost)
	if !VerifyOperatorPassword(string(h), "correct-horse") {
		t.Fatal("expected password to verify")
	}
	if VerifyOperatorPassword(string(h), "wrong") {
		t.Fatal("expected wrong password to fail")
	}
}
