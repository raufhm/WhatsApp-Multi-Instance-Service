package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
)

func writeAuditLogNullable(tx *sql.Tx, tenantID uuid.UUID, operatorID, action string, conversationID *uuid.UUID, details map[string]any) error {
	payload, err := json.Marshal(details)
	if err != nil {
		return err
	}
	var convID any
	if conversationID != nil {
		convID = *conversationID
	}
	_, err = tx.Exec(`INSERT INTO operator_audit_logs (tenant_id, operator_id, action, conversation_id, details) VALUES ($1,$2,$3,$4,$5)`,
		tenantID, operatorID, action, convID, payload)
	return err
}

// MergeConversations moves all messages from the source conversation onto the
// target, closes the source with closure_reason='merged', and records the
// merge in the audit log. The target conversation is returned.
func (p *PostgresStore) MergeConversations(tenantID, targetID, sourceID uuid.UUID, operatorID string) (domain.Conversation, error) {
	if sourceID == targetID {
		return domain.Conversation{}, fmt.Errorf("cannot merge a conversation into itself")
	}

	tx, err := p.db.Begin()
	if err != nil {
		return domain.Conversation{}, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE conversation_messages SET conversation_id=$3 WHERE tenant_id=$1 AND conversation_id=$2`, tenantID, sourceID, targetID); err != nil {
		return domain.Conversation{}, err
	}
	if _, err := tx.Exec(`UPDATE conversations SET status='CLOSED', closure_reason='merged', merged_into_id=$3, closed_at=COALESCE(closed_at, CURRENT_TIMESTAMP), updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$1 AND id=$2`, tenantID, sourceID, targetID); err != nil {
		return domain.Conversation{}, err
	}
	if _, err := tx.Exec(`UPDATE conversations SET last_activity_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$1 AND id=$2`, tenantID, targetID); err != nil {
		return domain.Conversation{}, err
	}
	src := sourceID
	if err := writeAuditLogNullable(tx, tenantID, operatorID, "MERGE", &src, map[string]any{"target_id": targetID, "source_id": sourceID}); err != nil {
		return domain.Conversation{}, err
	}

	var c domain.Conversation
	if err := scanConversation(tx.QueryRow(
		`SELECT `+conversationColumns+` FROM conversations WHERE tenant_id=$1 AND id=$2`, tenantID, targetID), &c); err != nil {
		return domain.Conversation{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Conversation{}, err
	}
	return c, nil
}

// SplitConversation creates a new conversation from the same account/contact as
// the source and moves the listed messages onto it.
func (p *PostgresStore) SplitConversation(tenantID, sourceID uuid.UUID, messageIDs []uuid.UUID, operatorID string) (domain.Conversation, error) {
	if len(messageIDs) == 0 {
		return domain.Conversation{}, fmt.Errorf("no message ids provided for split")
	}

	tx, err := p.db.Begin()
	if err != nil {
		return domain.Conversation{}, err
	}
	defer tx.Rollback()

	var newID uuid.UUID
	err = tx.QueryRow(
		`INSERT INTO conversations (tenant_id, account_id, contact_id, started_at, last_activity_at)
		 SELECT tenant_id, account_id, contact_id, started_at, CURRENT_TIMESTAMP FROM conversations WHERE tenant_id=$1 AND id=$2
		 RETURNING id`, tenantID, sourceID).Scan(&newID)
	if err != nil {
		return domain.Conversation{}, err
	}

	ids := make([]string, len(messageIDs))
	for i, id := range messageIDs {
		ids[i] = id.String()
	}
	array := fmt.Sprintf("{%s}", joinUUIDs(ids))
	if _, err := tx.Exec(`UPDATE conversation_messages SET conversation_id=$3 WHERE tenant_id=$1 AND conversation_id=$2 AND id=ANY($4::uuid[])`, tenantID, sourceID, newID, array); err != nil {
		return domain.Conversation{}, err
	}
	src := sourceID
	idStrs := make([]string, len(messageIDs))
	for i, id := range messageIDs {
		idStrs[i] = id.String()
	}
	if err := writeAuditLogNullable(tx, tenantID, operatorID, "SPLIT", &src, map[string]any{"new_conversation_id": newID, "message_ids": idStrs}); err != nil {
		return domain.Conversation{}, err
	}

	var c domain.Conversation
	if err := scanConversation(tx.QueryRow(
		`SELECT `+conversationColumns+` FROM conversations WHERE tenant_id=$1 AND id=$2`, tenantID, newID), &c); err != nil {
		return domain.Conversation{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Conversation{}, err
	}
	return c, nil
}

// ListOperatorAuditLogs returns operator audit entries, newest first.
func (p *PostgresStore) ListOperatorAuditLogs(tenantID uuid.UUID, limit, offset int) ([]domain.OperatorAuditLog, error) {
	rows, err := p.db.Query(`SELECT id, tenant_id, operator_id, action, conversation_id, details, created_at FROM operator_audit_logs WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []domain.OperatorAuditLog
	for rows.Next() {
		var l domain.OperatorAuditLog
		var convID uuid.NullUUID
		var details []byte
		if err := rows.Scan(&l.ID, &l.TenantID, &l.OperatorID, &l.Action, &convID, &details, &l.CreatedAt); err != nil {
			return nil, err
		}
		if convID.Valid {
			id := convID.UUID
			l.ConversationID = &id
		}
		if len(details) > 0 {
			_ = json.Unmarshal(details, &l.Details)
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func joinUUIDs(ids []string) string {
	out := ""
	for i, s := range ids {
		if i > 0 {
			out += ","
		}
		out += s
	}
	return out
}
