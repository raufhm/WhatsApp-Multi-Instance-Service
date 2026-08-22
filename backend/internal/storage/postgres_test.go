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
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO contacts (tenant_id, normalized_address, provider_address, is_group)")).WithArgs(tenantID, "15551234567@s.whatsapp.net", "15551234567", false).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(contactID.String()))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO conversations (tenant_id, account_id, contact_id, is_group, started_at, last_activity_at)")).WithArgs(tenantID, accountID, contactID, false, meta.Timestamp).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(conversationID.String()))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO conversation_messages (tenant_id, conversation_id, actor, operator_id, operator_name, provider, provider_message_id")).WithArgs(tenantID, conversationID, domain.ActorContact, nil, "", meta.WhatsappID, meta.Direction, meta.Content, meta.Type, nil, meta.Status, meta.Timestamp).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversations SET started_at=LEAST(started_at, $1), last_activity_at=GREATEST(last_activity_at, $1), updated_at=CURRENT_TIMESTAMP WHERE id=$2")).WithArgs(meta.Timestamp, conversationID).WillReturnResult(sqlmock.NewResult(1, 1))
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
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO contacts (tenant_id, normalized_address, provider_address, is_group)")).WithArgs(tenantID, "15551234567@s.whatsapp.net", "15551234567", false).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(contactID.String()))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO conversations (tenant_id, account_id, contact_id, is_group, started_at, last_activity_at)")).WithArgs(tenantID, accountID, contactID, false, now).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM conversations WHERE tenant_id=$1 AND account_id=$2 AND contact_id=$3 AND status <> 'CLOSED' FOR UPDATE")).WithArgs(tenantID, accountID, contactID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(conversationID.String()))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO conversation_messages (tenant_id, conversation_id, actor, operator_id, operator_name, provider, provider_message_id")).WithArgs(tenantID, conversationID, domain.ActorContact, nil, "", "m-2", domain.Incoming, "again", domain.Text, nil, domain.StatusPending, now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversations SET started_at=LEAST(started_at, $1), last_activity_at=GREATEST(last_activity_at, $1), updated_at=CURRENT_TIMESTAMP WHERE id=$2")).WithArgs(now, conversationID).WillReturnResult(sqlmock.NewResult(1, 1))
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
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO contacts (tenant_id, normalized_address, provider_address, is_group)")).WithArgs(tenantID, "15551234567@s.whatsapp.net", "15551234567", false).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(contactID.String()))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO conversations (tenant_id, account_id, contact_id, is_group, started_at, last_activity_at)")).WithArgs(tenantID, accountID, contactID, false, now).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM conversations WHERE tenant_id=$1 AND account_id=$2 AND contact_id=$3 AND status <> 'CLOSED' FOR UPDATE")).WithArgs(tenantID, accountID, contactID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(conversationID.String()))
	messageUpdate := `(?s)` + regexp.QuoteMeta("INSERT INTO conversation_messages (tenant_id, conversation_id, actor, operator_id, operator_name, provider, provider_message_id") + `.*` + regexp.QuoteMeta("ON CONFLICT (tenant_id, provider, provider_message_id) DO UPDATE SET status=CASE WHEN conversation_messages.status='READ' OR (conversation_messages.status='DELIVERED' AND EXCLUDED.status='SENT')")
	mock.ExpectExec(messageUpdate).WithArgs(tenantID, conversationID, domain.ActorContact, nil, "", "m-1", domain.Incoming, "hello", domain.Text, nil, domain.StatusSent, now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversations SET started_at=LEAST(started_at, $1), last_activity_at=GREATEST(last_activity_at, $1), updated_at=CURRENT_TIMESTAMP WHERE id=$2")).WithArgs(now, conversationID).WillReturnResult(sqlmock.NewResult(1, 1))
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

// ListUploadJobs must serialize to `[]` (not `null`) when the tenant has no
// jobs so dashboard pages can always iterate the response.
func TestListUploadJobsReturnsEmptySliceWhenNoRows(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	tenant := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT `+uploadJobColumns+` FROM upload_jobs WHERE tenant_id=$1 ORDER BY next_attempt_at DESC LIMIT $2 OFFSET $3`)).
		WithArgs(tenant, 50, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "message_id", "host_id", "object_key", "mime_type", "media_path", "status", "attempt_count", "next_attempt_at", "last_error", "media_url", "lease_until", "created_at", "updated_at"}))

	jobs, err := store.ListUploadJobs(tenant, "", 50, 0)
	if err != nil {
		t.Fatalf("ListUploadJobs: %v", err)
	}
	if jobs == nil {
		t.Fatalf("expected non-nil empty slice, got nil")
	}
	if len(jobs) != 0 {
		t.Fatalf("expected 0 jobs, got %d", len(jobs))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

// ListBotRuleSets must serialize to `[]` (not `null`) when the tenant has no
// rulesets so dashboard pages can always iterate the response.
func TestListBotRuleSetsReturnsEmptySliceWhenNoRows(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	tenant := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT ` + botRuleSetColumns + ` FROM bot_rule_sets WHERE tenant_id=$1 ORDER BY version DESC`)).
		WithArgs(tenant).
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "version", "rules", "is_active", "created_at", "updated_at"}))

	sets, err := store.ListBotRuleSets(tenant)
	if err != nil {
		t.Fatalf("ListBotRuleSets: %v", err)
	}
	if sets == nil {
		t.Fatalf("expected non-nil empty slice, got nil")
	}
	if len(sets) != 0 {
		t.Fatalf("expected 0 sets, got %d", len(sets))
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

func TestPostgresStore_ListConversationSummaries(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()

	tenantID := uuid.New()
	conversationID := uuid.New()
	now := time.Now()

	query := `(?s)` + regexp.QuoteMeta(`SELECT * FROM (`) +
		`.*` + regexp.QuoteMeta(`SELECT DISTINCT ON (c.contact_id)`) +
		`.*` + regexp.QuoteMeta(`FROM conversations c`) +
		`.*` + regexp.QuoteMeta(`LEFT JOIN contacts co`) +
		`.*` + regexp.QuoteMeta(`LEFT JOIN LATERAL`) +
		`.*` + regexp.QuoteMeta(`WHERE c.tenant_id = $1`) +
		`.*` + regexp.QuoteMeta(`ORDER BY c.contact_id, `) +
		`.*` + regexp.QuoteMeta(`c.last_activity_at DESC`) +
		`.*` + regexp.QuoteMeta(`) inbox ORDER BY last_activity_at DESC LIMIT $2 OFFSET $3`)

	mock.ExpectQuery(query).
		WithArgs(tenantID, 50, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "account_id", "contact_id", "ticket_number", "status",
			"bot_state", "started_at", "last_activity_at", "closed_at", "handoff_at",
			"closure_reason", "assignee", "merged_into_id", "created_at", "updated_at",
			"contact_name", "contact_number", "is_group", "last_message", "last_actor",
		}).
			AddRow(conversationID, tenantID, uuid.New(), uuid.New(), 42, "OPEN", "", now, now, nil, nil,
				"", "Ada", nil, now, now,
				"Ada Lovelace", "15551234567", false, "What is the status of my order?", "CONTACT"))

	summaries, err := store.ListConversationSummaries(tenantID, "", 50, 0)
	if err != nil {
		t.Fatalf("ListConversationSummaries: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	s := summaries[0]
	if s.ID != conversationID || s.TicketNumber != 42 || s.Assignee != "Ada" {
		t.Fatalf("unexpected conversation fields: %+v", s)
	}
	if s.ContactName != "Ada Lovelace" || s.ContactNumber != "15551234567" {
		t.Fatalf("unexpected contact identity: name=%q number=%q", s.ContactName, s.ContactNumber)
	}
	if s.LastMessage != "What is the status of my order?" || s.LastActor != domain.ActorContact {
		t.Fatalf("unexpected last message: content=%q actor=%q", s.LastMessage, s.LastActor)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPostgresStore_RecordInstanceEvent_EmptyDirection(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()

	tenantID := uuid.New()
	hostID := "6285121514764"
	now := time.Now()

	query := regexp.QuoteMeta(`INSERT INTO whatsmeow_instance_events (tenant_id, host_id, event_type, direction, payload)
		VALUES ($1,$2,$3,$4,$5) RETURNING id, host_id, event_type, COALESCE(direction, ''), COALESCE(payload, '{}'::jsonb), occurred_at`)

	mock.ExpectQuery(query).
		WithArgs(tenantID, hostID, domain.EventQueueDepth, nil, []byte(`{"connected":true,"queue_size":0}`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "host_id", "event_type", "direction", "payload", "occurred_at"}).
			AddRow(1, hostID, domain.EventQueueDepth, "", []byte(`{"connected":true,"queue_size":0}`), now))

	ev, err := store.RecordInstanceEvent(context.Background(), tenantID, hostID, domain.EventQueueDepth, "", map[string]any{
		"connected":  true,
		"queue_size": 0,
	})
	if err != nil {
		t.Fatalf("RecordInstanceEvent with empty direction returned error: %v", err)
	}
	if ev.ID != 1 || ev.HostID != hostID || ev.EventType != domain.EventQueueDepth || ev.Direction != "" {
		t.Fatalf("unexpected event: %+v", ev)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestProjectMessageGroupContactUsesGUsServer(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	tenantID, accountID, contactID, conversationID := projectionIDs()
	now := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)

	// A group message must normalize the address to @g.us so the group contact
	// is distinct from a personal contact that shares the same numeric prefix.
	meta := domain.MessageMetadata{
		WhatsappID: "gm-1", HostID: "host-1", Sender: "15551234567",
		Recipient: "120363", Content: "group hello", IsGroup: true,
		Direction: domain.Incoming, Type: domain.Text, Status: domain.StatusPending, Timestamp: now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, id FROM whatsapp_accounts WHERE provider='whatsmeow' AND host_id=$1")).WithArgs("host-1").WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "id"}).AddRow(tenantID.String(), accountID.String()))
	// Group address normalizes to @g.us
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO contacts (tenant_id, normalized_address, provider_address, is_group)")).WithArgs(tenantID, "120363@g.us", "120363", true).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(contactID.String()))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO conversations (tenant_id, account_id, contact_id, is_group, started_at, last_activity_at)")).WithArgs(tenantID, accountID, contactID, true, now).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(conversationID.String()))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO conversation_messages (tenant_id, conversation_id, actor, operator_id, operator_name, provider, provider_message_id")).WithArgs(tenantID, conversationID, domain.ActorContact, nil, "", "gm-1", domain.Incoming, "group hello", domain.Text, nil, domain.StatusPending, now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversations SET started_at=LEAST(started_at, $1), last_activity_at=GREATEST(last_activity_at, $1), updated_at=CURRENT_TIMESTAMP WHERE id=$2")).WithArgs(now, conversationID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := store.ProjectMessage(meta); err != nil {
		t.Fatalf("ProjectMessage group: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("group projection expectations: %v", err)
	}
}

func TestProjectMessagePersonalAndGroupDistinctContacts(t *testing.T) {
	// Verify that a personal message and a group message that share a numeric
	// prefix produce distinct normalized addresses and thus distinct contacts.
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	tenantID, accountID, personalContactID, groupContactID := projectionIDs()
	_, _, _, personalConversationID := projectionIDs()
	_, _, _, groupConversationID := projectionIDs()
	now := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)

	// 1. Personal message
	personalMeta := domain.MessageMetadata{
		WhatsappID: "pm-1", HostID: "host-1", Sender: "120363",
		Content: "personal hello", IsGroup: false,
		Direction: domain.Incoming, Type: domain.Text, Status: domain.StatusPending, Timestamp: now,
	}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, id FROM whatsapp_accounts WHERE provider='whatsmeow' AND host_id=$1")).WithArgs("host-1").WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "id"}).AddRow(tenantID.String(), accountID.String()))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO contacts (tenant_id, normalized_address, provider_address, is_group)")).WithArgs(tenantID, "120363@s.whatsapp.net", "120363", false).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(personalContactID.String()))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO conversations (tenant_id, account_id, contact_id, is_group, started_at, last_activity_at)")).WithArgs(tenantID, accountID, personalContactID, false, now).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(personalConversationID.String()))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO conversation_messages (tenant_id, conversation_id, actor, operator_id, operator_name, provider, provider_message_id")).WithArgs(tenantID, personalConversationID, domain.ActorContact, nil, "", "pm-1", domain.Incoming, "personal hello", domain.Text, nil, domain.StatusPending, now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversations SET started_at=LEAST(started_at, $1), last_activity_at=GREATEST(last_activity_at, $1), updated_at=CURRENT_TIMESTAMP WHERE id=$2")).WithArgs(now, personalConversationID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// 2. Group message with the same numeric prefix
	groupMeta := domain.MessageMetadata{
		WhatsappID: "gm-1", HostID: "host-1", Sender: "user-in-group",
		Recipient: "120363", Content: "group hello", IsGroup: true,
		Direction: domain.Incoming, Type: domain.Text, Status: domain.StatusPending, Timestamp: now,
	}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, id FROM whatsapp_accounts WHERE provider='whatsmeow' AND host_id=$1")).WithArgs("host-1").WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "id"}).AddRow(tenantID.String(), accountID.String()))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO contacts (tenant_id, normalized_address, provider_address, is_group)")).WithArgs(tenantID, "120363@g.us", "120363", true).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(groupContactID.String()))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO conversations (tenant_id, account_id, contact_id, is_group, started_at, last_activity_at)")).WithArgs(tenantID, accountID, groupContactID, true, now).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(groupConversationID.String()))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO conversation_messages (tenant_id, conversation_id, actor, operator_id, operator_name, provider, provider_message_id")).WithArgs(tenantID, groupConversationID, domain.ActorContact, nil, "", "gm-1", domain.Incoming, "group hello", domain.Text, nil, domain.StatusPending, now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversations SET started_at=LEAST(started_at, $1), last_activity_at=GREATEST(last_activity_at, $1), updated_at=CURRENT_TIMESTAMP WHERE id=$2")).WithArgs(now, groupConversationID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := store.ProjectMessage(personalMeta); err != nil {
		t.Fatalf("ProjectMessage personal: %v", err)
	}
	if err := store.ProjectMessage(groupMeta); err != nil {
		t.Fatalf("ProjectMessage group: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("personal+group expectations: %v", err)
	}
}

func TestPostgresStore_RecordInstanceEvent_WithDirection(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()

	tenantID := uuid.New()
	hostID := "6285121514764"
	now := time.Now()
	direction := "OUT"

	query := regexp.QuoteMeta(`INSERT INTO whatsmeow_instance_events (tenant_id, host_id, event_type, direction, payload)
		VALUES ($1,$2,$3,$4,$5) RETURNING id, host_id, event_type, COALESCE(direction, ''), COALESCE(payload, '{}'::jsonb), occurred_at`)

	mock.ExpectQuery(query).
		WithArgs(tenantID, hostID, domain.EventMessageOut, &direction, []byte(`{"message_id":"m1"}`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "host_id", "event_type", "direction", "payload", "occurred_at"}).
			AddRow(2, hostID, domain.EventMessageOut, "OUT", []byte(`{"message_id":"m1"}`), now))

	ev, err := store.RecordInstanceEvent(context.Background(), tenantID, hostID, domain.EventMessageOut, direction, map[string]any{
		"message_id": "m1",
	})
	if err != nil {
		t.Fatalf("RecordInstanceEvent with direction returned error: %v", err)
	}
	if ev.ID != 2 || ev.HostID != hostID || ev.EventType != domain.EventMessageOut || ev.Direction != "OUT" {
		t.Fatalf("unexpected event: %+v", ev)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestListContacts_HandlesNullDealStageKey(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()

	tenantID := uuid.New()
	contactID := uuid.New()
	now := time.Now().UTC()

	countQuery := regexp.QuoteMeta(`SELECT COUNT(*) FROM (
		SELECT DISTINCT ON (normalized_address, is_group) id
		FROM contacts
		WHERE tenant_id=$1) sub`)
	mock.ExpectQuery(countQuery).WithArgs(tenantID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	selectQuery := regexp.QuoteMeta(`SELECT * FROM (
		SELECT DISTINCT ON (normalized_address, is_group)
			id, tenant_id, normalized_address, provider_address, display_name, is_group, deal_stage_key, deal_stage_id, metadata, created_at, updated_at
		FROM contacts
		WHERE tenant_id=$1
		ORDER BY normalized_address, is_group, updated_at DESC
	) sub
	ORDER BY updated_at DESC LIMIT $2 OFFSET $3`)

	contactCols := []string{"id", "tenant_id", "normalized_address", "provider_address", "display_name", "is_group", "deal_stage_key", "deal_stage_id", "metadata", "created_at", "updated_at"}
	mock.ExpectQuery(selectQuery).WithArgs(tenantID, 20, 0).
		WillReturnRows(sqlmock.NewRows(contactCols).
			AddRow(contactID, tenantID, "15551234567@s.whatsapp.net", "15551234567", "John Doe", false, nil, nil, []byte(`{"email":"john@example.com"}`), now, now))

	contacts, total, err := store.ListContacts(tenantID, 20, 0, "")
	if err != nil {
		t.Fatalf("ListContacts returned unexpected error: %v", err)
	}
	if total != 1 || len(contacts) != 1 {
		t.Fatalf("expected 1 contact, got total=%d len=%d", total, len(contacts))
	}
	if contacts[0].DealStageKey != "" {
		t.Fatalf("expected empty deal stage key for null db value, got %q", contacts[0].DealStageKey)
	}
	if contacts[0].DisplayName != "John Doe" {
		t.Fatalf("expected display name John Doe, got %q", contacts[0].DisplayName)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestGetContact_HandlesNullDealStageKey(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()

	tenantID := uuid.New()
	contactID := uuid.New()
	now := time.Now().UTC()

	selectQuery := regexp.QuoteMeta(`SELECT id, tenant_id, normalized_address, provider_address, display_name, is_group, deal_stage_key, deal_stage_id, metadata, created_at, updated_at FROM contacts WHERE tenant_id=$1 AND id=$2`)
	contactCols := []string{"id", "tenant_id", "normalized_address", "provider_address", "display_name", "is_group", "deal_stage_key", "deal_stage_id", "metadata", "created_at", "updated_at"}
	mock.ExpectQuery(selectQuery).WithArgs(tenantID, contactID).
		WillReturnRows(sqlmock.NewRows(contactCols).
			AddRow(contactID, tenantID, "15551234567@s.whatsapp.net", "15551234567", "Jane Doe", false, nil, nil, []byte(`{}`), now, now))

	c, err := store.GetContact(tenantID, contactID)
	if err != nil {
		t.Fatalf("GetContact returned error: %v", err)
	}
	if c.ID != contactID || c.DealStageKey != "" {
		t.Fatalf("unexpected contact result: %+v", c)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUpsertContact_HandlesNullDealStageKey(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()

	tenantID := uuid.New()
	contactID := uuid.New()
	now := time.Now().UTC()

	insertQuery := regexp.QuoteMeta(`INSERT INTO contacts (tenant_id, normalized_address, provider_address, display_name, is_group, metadata)`)
	contactCols := []string{"id", "tenant_id", "normalized_address", "provider_address", "display_name", "is_group", "deal_stage_key", "deal_stage_id", "metadata", "created_at", "updated_at"}
	mock.ExpectQuery(insertQuery).WithArgs(tenantID, "15551234567@s.whatsapp.net", "15551234567", "New Contact", false, []byte(`{}`)).
		WillReturnRows(sqlmock.NewRows(contactCols).
			AddRow(contactID, tenantID, "15551234567@s.whatsapp.net", "15551234567", "New Contact", false, nil, nil, []byte(`{}`), now, now))

	c, err := store.UpsertContact(tenantID, domain.ContactUpsert{
		ProviderAddress: "15551234567",
		DisplayName:     "New Contact",
		IsGroup:         false,
	})
	if err != nil {
		t.Fatalf("UpsertContact returned error: %v", err)
	}
	if c.ID != contactID || c.DealStageKey != "" {
		t.Fatalf("unexpected contact result: %+v", c)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUpdateContact_HandlesNullDealStageKey(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()

	tenantID := uuid.New()
	contactID := uuid.New()
	now := time.Now().UTC()

	selectQuery := regexp.QuoteMeta(`SELECT id, tenant_id, normalized_address, provider_address, display_name, is_group, deal_stage_key, deal_stage_id, metadata, created_at, updated_at FROM contacts WHERE tenant_id=$1 AND id=$2`)
	contactCols := []string{"id", "tenant_id", "normalized_address", "provider_address", "display_name", "is_group", "deal_stage_key", "deal_stage_id", "metadata", "created_at", "updated_at"}
	mock.ExpectQuery(selectQuery).WithArgs(tenantID, contactID).
		WillReturnRows(sqlmock.NewRows(contactCols).
			AddRow(contactID, tenantID, "15551234567@s.whatsapp.net", "15551234567", "Old Name", false, nil, nil, []byte(`{}`), now, now))

	updateQuery := regexp.QuoteMeta(`UPDATE contacts SET display_name=$1, deal_stage_key=$2, deal_stage_id=$3, metadata=$4, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$5 AND id=$6`)
	mock.ExpectQuery(updateQuery).WithArgs("Updated Name", nil, nil, []byte(`{}`), tenantID, contactID).
		WillReturnRows(sqlmock.NewRows(contactCols).
			AddRow(contactID, tenantID, "15551234567@s.whatsapp.net", "15551234567", "Updated Name", false, nil, nil, []byte(`{}`), now, now))

	c, err := store.UpdateContact(tenantID, contactID, domain.ContactUpdateInput{
		DisplayName: "Updated Name",
	})
	if err != nil {
		t.Fatalf("UpdateContact returned error: %v", err)
	}
	if c.ID != contactID || c.DisplayName != "Updated Name" || c.DealStageKey != "" {
		t.Fatalf("unexpected contact result: %+v", c)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestGetConversationTimeline_HandlesNullOperatorName(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()

	tenantID := uuid.New()
	convID := uuid.New()
	msgID := uuid.New()
	now := time.Now().UTC()

	selectQuery := regexp.QuoteMeta(`SELECT id, tenant_id, conversation_id, actor, operator_id, COALESCE(operator_name, ''), provider, provider_message_id, direction, content, message_type, COALESCE(media_url,''), status, provider_timestamp, is_internal, created_at, updated_at FROM conversation_messages WHERE tenant_id=$1 AND conversation_id=$2 ORDER BY provider_timestamp, created_at LIMIT $3 OFFSET $4`)
	msgCols := []string{"id", "tenant_id", "conversation_id", "actor", "operator_id", "operator_name", "provider", "provider_message_id", "direction", "content", "message_type", "media_url", "status", "provider_timestamp", "is_internal", "created_at", "updated_at"}
	mock.ExpectQuery(selectQuery).WithArgs(tenantID, convID, 50, 0).
		WillReturnRows(sqlmock.NewRows(msgCols).
			AddRow(msgID, tenantID, convID, "CUSTOMER", nil, "", "whatsmeow", "msg-1", "INCOMING", "Hello there", "TEXT", "", "DELIVERED", now, false, now, now))

	msgs, err := store.GetConversationTimeline(tenantID, convID, 50, 0)
	if err != nil {
		t.Fatalf("GetConversationTimeline returned error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].OperatorName != "" || msgs[0].OperatorID != nil {
		t.Fatalf("unexpected operator fields: %+v", msgs[0])
	}
	if msgs[0].Content != "Hello there" {
		t.Fatalf("unexpected content: %q", msgs[0].Content)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
