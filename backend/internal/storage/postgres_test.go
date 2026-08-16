package storage

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/raufhm/whatsapp-testing/domain"
)

type conversationScanStub struct {
	closureReason any
}

func (s conversationScanStub) Scan(destinations ...any) error {
	if len(destinations) != 16 {
		return &scanArgumentError{count: len(destinations)}
	}
	closure, ok := destinations[11].(*sql.NullString)
	if !ok {
		return &scanArgumentError{message: "closure reason was not scanned as sql.NullString"}
	}
	switch value := s.closureReason.(type) {
	case nil:
		closure.Valid = false
		closure.String = ""
	case string:
		closure.Valid = true
		closure.String = value
	default:
		return &scanArgumentError{message: "unsupported closure reason"}
	}
	return nil
}

type scanArgumentError struct {
	count   int
	message string
}

func (e *scanArgumentError) Error() string {
	if e.message != "" {
		return e.message
	}
	return "unexpected scan argument count"
}

func TestScanConversationAllowsNullClosureReason(t *testing.T) {
	var conversation domain.Conversation
	if err := scanConversation(conversationScanStub{}, &conversation); err != nil {
		t.Fatalf("scanConversation returned error for NULL closure reason: %v", err)
	}
	if conversation.ClosureReason != "" {
		t.Fatalf("ClosureReason = %q, want empty string", conversation.ClosureReason)
	}
}

func TestScanConversationMapsClosureReason(t *testing.T) {
	var conversation domain.Conversation
	if err := scanConversation(conversationScanStub{closureReason: "customer requested handoff"}, &conversation); err != nil {
		t.Fatalf("scanConversation returned error: %v", err)
	}
	if conversation.ClosureReason != "customer requested handoff" {
		t.Fatalf("ClosureReason = %q, want mapped value", conversation.ClosureReason)
	}
}

func newProjectionStore(t *testing.T) (*PostgresStore, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	return &PostgresStore{db: db}, mock, func() { _ = db.Close() }
}

func projectionIDs() (uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	return uuid.New(), uuid.New(), uuid.New(), uuid.New()
}

func expectProjectedMessage(mock sqlmock.Sqlmock, meta domain.MessageMetadata, tenantID, accountID, contactID, conversationID uuid.UUID) {
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, id FROM whatsapp_accounts WHERE provider='whatsmeow' AND host_id=$1")).WithArgs(meta.HostID).WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "id"}).AddRow(tenantID.String(), accountID.String()))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO contacts (tenant_id, normalized_address, provider_address)")).WithArgs(tenantID, "15551234567@s.whatsapp.net", "15551234567").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(contactID.String()))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO conversations (tenant_id, account_id, contact_id, started_at, last_activity_at)")).WithArgs(tenantID, accountID, contactID, meta.Timestamp).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(conversationID.String()))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO conversation_messages (tenant_id, conversation_id, actor, provider, provider_message_id")).WithArgs(tenantID, conversationID, domain.ActorContact, meta.WhatsappID, meta.Direction, meta.Content, meta.Type, nil, meta.Status, meta.Timestamp).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversations SET last_activity_at=$1, updated_at=CURRENT_TIMESTAMP WHERE id=$2")).WithArgs(meta.Timestamp, conversationID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
}

func TestProjectMessageCreatesContactConversationAndTimeline(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	tenantID, accountID, contactID, conversationID := projectionIDs()
	now := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	meta := domain.MessageMetadata{WhatsappID: "m-1", HostID: "host-1", Sender: "15551234567", Content: "hello", Direction: domain.Incoming, Type: domain.Text, Status: domain.StatusPending, Timestamp: now}
	expectProjectedMessage(mock, meta, tenantID, accountID, contactID, conversationID)

	if err := store.ProjectMessage(meta); err != nil {
		t.Fatalf("ProjectMessage: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("projection expectations: %v", err)
	}
}

func TestProjectMessageReturningContactReusesActiveConversation(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	tenantID, accountID, contactID, conversationID := projectionIDs()
	now := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	meta := domain.MessageMetadata{WhatsappID: "m-2", HostID: "host-1", Sender: "15551234567", Content: "again", Direction: domain.Incoming, Type: domain.Text, Status: domain.StatusPending, Timestamp: now}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, id FROM whatsapp_accounts WHERE provider='whatsmeow' AND host_id=$1")).WithArgs("host-1").WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "id"}).AddRow(tenantID.String(), accountID.String()))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO contacts (tenant_id, normalized_address, provider_address)")).WithArgs(tenantID, "15551234567@s.whatsapp.net", "15551234567").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(contactID.String()))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO conversations (tenant_id, account_id, contact_id, started_at, last_activity_at)")).WithArgs(tenantID, accountID, contactID, now).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM conversations WHERE tenant_id=$1 AND account_id=$2 AND contact_id=$3 AND status <> 'CLOSED' FOR UPDATE")).WithArgs(tenantID, accountID, contactID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(conversationID.String()))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO conversation_messages (tenant_id, conversation_id, actor, provider, provider_message_id")).WithArgs(tenantID, conversationID, domain.ActorContact, "m-2", domain.Incoming, "again", domain.Text, nil, domain.StatusPending, now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversations SET last_activity_at=$1, updated_at=CURRENT_TIMESTAMP WHERE id=$2")).WithArgs(now, conversationID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := store.ProjectMessage(meta); err != nil {
		t.Fatalf("ProjectMessage returning contact: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("returning contact expectations: %v", err)
	}
}

func TestProjectMessageDuplicateIsIdempotentAndProtectsReadStatus(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	tenantID, accountID, contactID, conversationID := projectionIDs()
	now := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	meta := domain.MessageMetadata{WhatsappID: "m-1", HostID: "host-1", Sender: "15551234567", Content: "hello", Direction: domain.Incoming, Type: domain.Text, Status: domain.StatusSent, Timestamp: now}
	first := domain.MessageMetadata{WhatsappID: "m-1", HostID: "host-1", Sender: "15551234567", Content: "hello", Direction: domain.Incoming, Type: domain.Text, Status: domain.StatusPending, Timestamp: now}
	expectProjectedMessage(mock, first, tenantID, accountID, contactID, conversationID)
	if err := store.ProjectMessage(first); err != nil {
		t.Fatalf("initial ProjectMessage: %v", err)
	}
	// A duplicate provider event takes the conflict-update path; the SQL itself
	// guarantees one timeline row and does not downgrade an already READ row.
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, id FROM whatsapp_accounts WHERE provider='whatsmeow' AND host_id=$1")).WithArgs("host-1").WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "id"}).AddRow(tenantID.String(), accountID.String()))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO contacts (tenant_id, normalized_address, provider_address)")).WithArgs(tenantID, "15551234567@s.whatsapp.net", "15551234567").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(contactID.String()))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO conversations (tenant_id, account_id, contact_id, started_at, last_activity_at)")).WithArgs(tenantID, accountID, contactID, now).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM conversations WHERE tenant_id=$1 AND account_id=$2 AND contact_id=$3 AND status <> 'CLOSED' FOR UPDATE")).WithArgs(tenantID, accountID, contactID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(conversationID.String()))
	messageUpdate := `(?s)` + regexp.QuoteMeta("INSERT INTO conversation_messages (tenant_id, conversation_id, actor, provider, provider_message_id") + `.*` + regexp.QuoteMeta("ON CONFLICT (tenant_id, provider, provider_message_id) DO UPDATE SET status=CASE WHEN conversation_messages.status='READ' OR (conversation_messages.status='DELIVERED' AND EXCLUDED.status='SENT')")
	mock.ExpectExec(messageUpdate).WithArgs(tenantID, conversationID, domain.ActorContact, "m-1", domain.Incoming, "hello", domain.Text, nil, domain.StatusSent, now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversations SET last_activity_at=$1, updated_at=CURRENT_TIMESTAMP WHERE id=$2")).WithArgs(now, conversationID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := store.ProjectMessage(meta); err != nil {
		t.Fatalf("duplicate ProjectMessage: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("duplicate expectations: %v", err)
	}
}

// --- Upload job repository ---

func TestCreateUploadJob(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	tenant := uuid.New()
	now := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	jobID := uuid.New()
	cols := []string{"id", "tenant_id", "message_id", "host_id", "object_key", "mime_type", "media_path", "status", "attempt_count", "next_attempt_at", "last_error", "media_url", "lease_until", "created_at", "updated_at"}
	rows := sqlmock.NewRows(cols).AddRow(jobID.String(), tenant.String(), "m-1", "host-1", "media/k", "image/jpeg", "/tmp/a.jpg", domain.UploadPending, 0, now, nil, nil, nil, now, now)
	expected := regexp.QuoteMeta("INSERT INTO upload_jobs (tenant_id, message_id, host_id, object_key, mime_type, media_path, status, next_attempt_at)")
	mock.ExpectQuery(expected).WithArgs(tenant, "m-1", "host-1", "media/k", "image/jpeg", "/tmp/a.jpg", domain.UploadPending, now).WillReturnRows(rows)

	job, err := store.CreateUploadJob(context.Background(), tenant, domain.UploadJob{MessageID: "m-1", HostID: "host-1", ObjectKey: "media/k", MimeType: "image/jpeg", MediaPath: "/tmp/a.jpg", Status: domain.UploadPending, NextAttemptAt: now})
	if err != nil {
		t.Fatalf("CreateUploadJob: %v", err)
	}
	if job.ID != jobID || job.Status != domain.UploadPending {
		t.Fatalf("unexpected job: %+v", job)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestClaimDueUploadJobsUsesSkipLockedAndReclaimsExpiredLeases(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	now := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	jobID := uuid.New()

	// The claim must (a) reclaim PROCESSING rows with expired leases (recovery
	// after a crash) and (b) use FOR UPDATE SKIP LOCKED so concurrent workers
	// never claim the same job.
	query := `(?s)` + regexp.QuoteMeta("WITH due AS (") + `.*` +
		regexp.QuoteMeta("status='PENDING' AND next_attempt_at <= $1") + `.*` +
		regexp.QuoteMeta("status='PROCESSING' AND lease_until < $1") + `.*` +
		regexp.QuoteMeta("FOR UPDATE SKIP LOCKED") + `.*` +
		regexp.QuoteMeta("UPDATE upload_jobs SET status='PROCESSING'")
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "message_id", "host_id", "object_key", "mime_type", "media_path", "status", "attempt_count", "next_attempt_at", "last_error", "media_url", "lease_until", "created_at", "updated_at"})
	rows.AddRow(jobID.String(), uuid.Nil.String(), "m-1", "host-1", "media/k", "image/jpeg", "/tmp/a.jpg", domain.UploadProcessing, 1, now, nil, nil, nil, now, now)
	mock.ExpectQuery(query).WithArgs(now, 5, sqlmock.AnyArg()).WillReturnRows(rows)

	jobs, err := store.ClaimDueUploadJobs(context.Background(), now, 5)
	if err != nil {
		t.Fatalf("ClaimDueUploadJobs: %v", err)
	}
	if len(jobs) != 1 || jobs[0].ID != jobID || jobs[0].Status != domain.UploadProcessing || jobs[0].AttemptCount != 1 {
		t.Fatalf("unexpected claimed jobs: %+v", jobs)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestMarkUploadTransitions(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	id := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE upload_jobs SET status='COMPLETED', media_url=$2")).WithArgs(id, "s3://b/media/k").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE upload_jobs SET status='PENDING', last_error=$2")).WithArgs(id, "timeout", sqlmock.AnyArg(), 2).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE upload_jobs SET status='FAILED', last_error=$2")).WithArgs(id, "NoSuchKey", 2).WillReturnResult(sqlmock.NewResult(1, 1))

	if err := store.MarkUploadCompleted(context.Background(), id, "s3://b/media/k"); err != nil {
		t.Fatalf("complete: %v", err)
	}
	next := time.Now().UTC().Add(time.Second)
	if err := store.MarkUploadRetryable(context.Background(), id, "timeout", next, 2); err != nil {
		t.Fatalf("retry: %v", err)
	}
	if err := store.MarkUploadFailed(context.Background(), id, "NoSuchKey", 2); err != nil {
		t.Fatalf("fail: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestProjectReceiptReconcilesApplicationMessageStatus(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversation_messages m SET status=$1, updated_at=CURRENT_TIMESTAMP")).WithArgs(domain.StatusRead, "host-1", "m-1").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := store.ProjectReceipt(domain.Receipt{WhatsappID: "m-1", Sender: "host-1", Status: domain.StatusRead}); err != nil {
		t.Fatalf("ProjectReceipt: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("receipt expectations: %v", err)
	}
}

func TestPostgresStore_RegisterAccount(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()

	tenantID := uuid.New()
	accountID := uuid.New()
	now := time.Now()

	query := `(?s)` + regexp.QuoteMeta("INSERT INTO whatsapp_accounts (tenant_id, host_id, provider, display_name)") + `.*` + regexp.QuoteMeta("ON CONFLICT (tenant_id, provider, host_id)") + `.*` + regexp.QuoteMeta("RETURNING id, tenant_id, host_id, provider, display_name, created_at, updated_at")

	mock.ExpectQuery(query).
		WithArgs(tenantID, "15551234567", "whatsmeow", "Main Support").
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "host_id", "provider", "display_name", "created_at", "updated_at"}).
			AddRow(accountID, tenantID, "15551234567", "whatsmeow", "Main Support", now, now))

	acc, err := store.RegisterAccount(tenantID, "15551234567", "Main Support", "")
	if err != nil {
		t.Fatalf("RegisterAccount: %v", err)
	}
	if acc.ID != accountID || acc.HostID != "15551234567" || acc.Provider != "whatsmeow" || acc.DisplayName != "Main Support" {
		t.Fatalf("unexpected account returned: %+v", acc)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
