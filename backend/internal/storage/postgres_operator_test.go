package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
)

func TestPostgresStore_BotRuleSets(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	tenantID := uuid.New()

	// 1. GetActiveBotRuleSet
	activeRuleSetID := uuid.New()
	rulesJSON := `[{"name": "welcome", "pattern": "hello", "match": "EXACT", "response": "Hi there!", "terminal": false, "handoff": false, "enabled": true}]`
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, version, rules, is_active, created_at, updated_at FROM bot_rule_sets WHERE tenant_id=$1 AND is_active=TRUE`)).
		WithArgs(tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "version", "rules", "is_active", "created_at", "updated_at"}).
			AddRow(activeRuleSetID, tenantID, 1, rulesJSON, true, time.Now(), time.Now()))

	rs, err := store.GetActiveBotRuleSet(tenantID)
	if err != nil {
		t.Fatalf("GetActiveBotRuleSet: %v", err)
	}
	if len(rs.Rules) != 1 || rs.Rules[0].Name != "welcome" {
		t.Fatalf("unexpected rules returned: %+v", rs)
	}

	// 2. SaveBotRuleSet
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(MAX(version), 0) FROM bot_rule_sets WHERE tenant_id=$1 FOR UPDATE`)).
		WithArgs(tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(1))
	newRules := []domain.BotRule{
		{Name: "bye", Pattern: "bye", Match: "EXACT", Response: "Goodbye!", Terminal: true, Enabled: true},
	}
	newRulesJSON, _ := json.Marshal(newRules)
	newRuleSetID := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO bot_rule_sets (tenant_id, version, rules, is_active) VALUES ($1,$2,$3,FALSE)`)).
		WithArgs(tenantID, 2, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "version", "rules", "is_active", "created_at", "updated_at"}).
			AddRow(newRuleSetID, tenantID, 2, newRulesJSON, false, time.Now(), time.Now()))
	mock.ExpectCommit()

	rs2, err := store.SaveBotRuleSet(tenantID, newRules)
	if err != nil {
		t.Fatalf("SaveBotRuleSet: %v", err)
	}
	if rs2.Version != 2 || len(rs2.Rules) != 1 || rs2.Rules[0].Name != "bye" {
		t.Fatalf("unexpected rules saved: %+v", rs2)
	}

	// 3. ListBotRuleSets
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, version, rules, is_active, created_at, updated_at FROM bot_rule_sets WHERE tenant_id=$1 ORDER BY version DESC`)).
		WithArgs(tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "version", "rules", "is_active", "created_at", "updated_at"}).
			AddRow(newRuleSetID, tenantID, 2, newRulesJSON, false, time.Now(), time.Now()).
			AddRow(activeRuleSetID, tenantID, 1, rulesJSON, true, time.Now(), time.Now()))

	list, err := store.ListBotRuleSets(tenantID)
	if err != nil {
		t.Fatalf("ListBotRuleSets: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 rule sets, got %d", len(list))
	}

	// 4. ActivateBotRuleSetVersion
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE bot_rule_sets SET is_active=FALSE, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$1 AND is_active=TRUE`)).
		WithArgs(tenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE bot_rule_sets SET is_active=TRUE, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$1 AND version=$2`)).
		WithArgs(tenantID, 2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "version", "rules", "is_active", "created_at", "updated_at"}).
			AddRow(newRuleSetID, tenantID, 2, newRulesJSON, true, time.Now(), time.Now()))
	mock.ExpectCommit()

	rsActive, err := store.ActivateBotRuleSetVersion(tenantID, 2)
	if err != nil {
		t.Fatalf("ActivateBotRuleSetVersion: %v", err)
	}
	if !rsActive.IsActive || rsActive.Version != 2 {
		t.Fatalf("activation failed: %+v", rsActive)
	}

	// 5. ActivateBotRuleSetVersion - not found
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE bot_rule_sets SET is_active=FALSE, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$1 AND is_active=TRUE`)).
		WithArgs(tenantID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE bot_rule_sets SET is_active=TRUE, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$1 AND version=$2`)).
		WithArgs(tenantID, 999).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err = store.ActivateBotRuleSetVersion(tenantID, 999)
	if err == nil {
		t.Fatalf("expected error for missing version")
	}
}

func TestPostgresStore_OperatorActions(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()
	tenantID := uuid.New()
	convID := uuid.New()
	accountID := uuid.New()
	contactID := uuid.New()

	convCols := []string{"id", "tenant_id", "account_id", "contact_id", "ticket_number", "status", "bot_state", "started_at", "last_activity_at", "closed_at", "handoff_at", "closure_reason", "assignee", "merged_into_id", "created_at", "updated_at"}
	convReturnCols := `id, tenant_id, account_id, contact_id, ticket_number, status, bot_state, started_at, last_activity_at, closed_at, handoff_at, closure_reason, assignee, merged_into_id, created_at, updated_at`

	// 1. AssignConversation
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE conversations SET assignee=$3, updated_at=CURRENT_TIMESTAMP`)).
		WithArgs(tenantID, convID, "operator-1").
		WillReturnRows(sqlmock.NewRows(convCols).AddRow(convID, tenantID, accountID, contactID, 123, "OPEN", "ACTIVE", time.Now(), time.Now(), nil, nil, "", "operator-1", nil, time.Now(), time.Now()))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO operator_audit_logs`)).
		WithArgs(tenantID, "operator-1", "ASSIGN", convID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	c, err := store.AssignConversation(tenantID, convID, "operator-1", "operator-1")
	if err != nil {
		t.Fatalf("AssignConversation: %v", err)
	}
	if c.Assignee != "operator-1" {
		t.Fatalf("expected assignee operator-1, got %q", c.Assignee)
	}

	// 2. HandoffConversation
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE conversations SET status='HANDED_OFF', handoff_at=COALESCE(handoff_at, CURRENT_TIMESTAMP), closure_reason=$3`)).
		WithArgs(tenantID, convID, "operator handoff").
		WillReturnRows(sqlmock.NewRows(convCols).AddRow(convID, tenantID, accountID, contactID, 123, "HANDED_OFF", "ACTIVE", time.Now(), time.Now(), nil, time.Now(), "operator handoff", "operator-1", nil, time.Now(), time.Now()))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO operator_audit_logs`)).
		WithArgs(tenantID, "operator-1", "HANDOFF", convID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	c2, err := store.HandoffConversation(tenantID, convID, "operator-1")
	if err != nil {
		t.Fatalf("HandoffConversation: %v", err)
	}
	if c2.Status != domain.ConversationHandedOff {
		t.Fatalf("expected status HANDED_OFF, got %s", c2.Status)
	}

	// 3. CloseConversationWithReason
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE conversations SET status='CLOSED', closed_at=COALESCE(closed_at, CURRENT_TIMESTAMP), closure_reason=$3`)).
		WithArgs(tenantID, convID, "resolved").
		WillReturnRows(sqlmock.NewRows(convCols).AddRow(convID, tenantID, accountID, contactID, 123, "CLOSED", "ACTIVE", time.Now(), time.Now(), time.Now(), time.Now(), "resolved", "operator-1", nil, time.Now(), time.Now()))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO operator_audit_logs`)).
		WithArgs(tenantID, "operator-1", "CLOSE", convID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	c3, err := store.CloseConversationWithReason(tenantID, convID, "resolved", "operator-1")
	if err != nil {
		t.Fatalf("CloseConversationWithReason: %v", err)
	}
	if c3.Status != domain.ConversationClosed || c3.ClosureReason != "resolved" {
		t.Fatalf("expected status CLOSED and reason resolved, got %s / %s", c3.Status, c3.ClosureReason)
	}

	// 4. ReopenConversation
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE conversations SET status='OPEN', closed_at=NULL, handoff_at=NULL, closure_reason=''`)).
		WithArgs(tenantID, convID).
		WillReturnRows(sqlmock.NewRows(convCols).AddRow(convID, tenantID, accountID, contactID, 123, "OPEN", "ACTIVE", time.Now(), time.Now(), nil, nil, "", "operator-1", nil, time.Now(), time.Now()))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO operator_audit_logs`)).
		WithArgs(tenantID, "operator-1", "REOPEN", convID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	c4, err := store.ReopenConversation(tenantID, convID, "operator-1")
	if err != nil {
		t.Fatalf("ReopenConversation: %v", err)
	}
	if c4.Status != domain.ConversationOpen {
		t.Fatalf("expected status OPEN, got %s", c4.Status)
	}

	// 4b. ReopenConversation - not closed (no rows)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE conversations SET status='OPEN', closed_at=NULL, handoff_at=NULL, closure_reason=''`)).
		WithArgs(tenantID, convID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err = store.ReopenConversation(tenantID, convID, "operator-1")
	if err == nil {
		t.Fatalf("expected error for non-closed conversation")
	}

	// 5. AddInternalNote
	noteID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO conversation_messages`)).
		WithArgs(tenantID, convID, domain.ActorOperator, sqlmock.AnyArg(), "", sqlmock.AnyArg(), "Internal test note", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "conversation_id", "actor", "operator_id", "operator_name", "provider", "provider_message_id", "direction", "content", "message_type", "media_url", "status", "provider_timestamp", "created_at", "updated_at", "is_internal"}).
			AddRow(noteID, tenantID, convID, domain.ActorOperator, nil, "", "internal", "note:x", "OUTGOING", "Internal test note", "TEXT", "", "SENT", time.Now(), time.Now(), time.Now(), true))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE conversations SET last_activity_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$1 AND id=$2`)).
		WithArgs(tenantID, convID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	m, err := store.AddInternalNote(tenantID, convID, domain.ActorOperator, "operator-1", "Internal test note")
	if err != nil {
		t.Fatalf("AddInternalNote: %v", err)
	}
	if !m.IsInternal || m.Content != "Internal test note" {
		t.Fatalf("invalid internal note: %+v", m)
	}

	// 6. MergeConversations
	sourceID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE conversation_messages SET conversation_id=$3 WHERE tenant_id=$1 AND conversation_id=$2`)).
		WithArgs(tenantID, sourceID, convID).
		WillReturnResult(sqlmock.NewResult(0, 5))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE conversations SET status='CLOSED', closure_reason='merged', merged_into_id=$3`)).
		WithArgs(tenantID, sourceID, convID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE conversations SET last_activity_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$1 AND id=$2`)).
		WithArgs(tenantID, convID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO operator_audit_logs`)).
		WithArgs(tenantID, "operator-1", "MERGE", sourceID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT `+convReturnCols+` FROM conversations WHERE tenant_id=$1 AND id=$2`)).
		WithArgs(tenantID, convID).
		WillReturnRows(sqlmock.NewRows(convCols).AddRow(convID, tenantID, accountID, contactID, 123, "OPEN", "ACTIVE", time.Now(), time.Now(), nil, nil, "", "operator-1", nil, time.Now(), time.Now()))
	mock.ExpectCommit()

	cMerged, err := store.MergeConversations(tenantID, convID, sourceID, "operator-1")
	if err != nil {
		t.Fatalf("MergeConversations: %v", err)
	}
	if cMerged.ID != convID {
		t.Fatalf("unexpected conversation returned: %+v", cMerged)
	}

	// 6b. MergeConversations - self merge
	_, err = store.MergeConversations(tenantID, convID, convID, "operator-1")
	if err == nil {
		t.Fatalf("expected error for merging into self")
	}

	// 7. SplitConversation
	msgID1 := uuid.New()
	msgID2 := uuid.New()
	newConvID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO conversations (tenant_id, account_id, contact_id, started_at, last_activity_at)`)).
		WithArgs(tenantID, convID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(newConvID))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE conversation_messages SET conversation_id=$3 WHERE tenant_id=$1 AND conversation_id=$2 AND id=ANY($4::uuid[])`)).
		WithArgs(tenantID, convID, newConvID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO operator_audit_logs`)).
		WithArgs(tenantID, "operator-1", "SPLIT", convID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT `+convReturnCols+` FROM conversations WHERE tenant_id=$1 AND id=$2`)).
		WithArgs(tenantID, newConvID).
		WillReturnRows(sqlmock.NewRows(convCols).AddRow(newConvID, tenantID, accountID, contactID, 124, "OPEN", "ACTIVE", time.Now(), time.Now(), nil, nil, "", "", nil, time.Now(), time.Now()))
	mock.ExpectCommit()

	cSplit, err := store.SplitConversation(tenantID, convID, []uuid.UUID{msgID1, msgID2}, "operator-1")
	if err != nil {
		t.Fatalf("SplitConversation: %v", err)
	}
	if cSplit.ID != newConvID {
		t.Fatalf("expected new conversation %s, got %s", newConvID, cSplit.ID)
	}

	// 7b. SplitConversation - empty message IDs
	_, err = store.SplitConversation(tenantID, convID, []uuid.UUID{}, "operator-1")
	if err == nil {
		t.Fatalf("expected error for empty message ids")
	}

	// 8. ListOperatorAuditLogs
	auditLogID := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, operator_id, action, conversation_id, details, created_at FROM operator_audit_logs WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`)).
		WithArgs(tenantID, 50, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "operator_id", "action", "conversation_id", "details", "created_at"}).
			AddRow(auditLogID, tenantID, "operator-1", "SPLIT", uuid.NullUUID{UUID: newConvID, Valid: true}, `{"split_count": 2}`, time.Now()))

	logs, err := store.ListOperatorAuditLogs(tenantID, 50, 0)
	if err != nil {
		t.Fatalf("ListOperatorAuditLogs: %v", err)
	}
	if len(logs) != 1 || logs[0].OperatorID != "operator-1" || logs[0].Action != "SPLIT" {
		t.Fatalf("unexpected logs: %+v", logs)
	}
}

// ensure errors package is used (for error comparisons in future tests)
var _ = errors.New
