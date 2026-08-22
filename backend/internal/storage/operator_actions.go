package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
)

// conversationColumns lists the conversation projection columns in scan order.
const conversationColumns = `id, tenant_id, account_id, contact_id, ticket_number, status, bot_state, started_at, last_activity_at, closed_at, handoff_at, closure_reason, assignee, merged_into_id, created_at, updated_at`

func writeAuditLog(tx *sql.Tx, tenantID uuid.UUID, operatorID, action string, conversationID uuid.UUID, details map[string]any) error {
	payload, err := json.Marshal(details)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO operator_audit_logs (tenant_id, operator_id, action, conversation_id, details) VALUES ($1,$2,$3,$4,$5)`,
		tenantID, operatorID, action, conversationID, payload)
	return err
}

// AssignConversation assigns an operator to an open conversation and audits
// the change. Closed conversations cannot be reassigned.
func (p *PostgresStore) AssignConversation(tenantID, conversationID uuid.UUID, assignee string, operatorID string) (domain.Conversation, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return domain.Conversation{}, err
	}
	defer tx.Rollback()

	var c domain.Conversation
	err = scanConversation(tx.QueryRow(
		`UPDATE conversations SET assignee=$3, updated_at=CURRENT_TIMESTAMP
		 WHERE tenant_id=$1 AND id=$2 AND status<>'CLOSED'
		 RETURNING `+conversationColumns, tenantID, conversationID, assignee), &c)
	if err == sql.ErrNoRows {
		return domain.Conversation{}, fmt.Errorf("conversation not found or is closed")
	}
	if err != nil {
		return domain.Conversation{}, err
	}
	if err := writeAuditLog(tx, tenantID, operatorID, "ASSIGN", conversationID, map[string]any{"assignee": assignee}); err != nil {
		return domain.Conversation{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Conversation{}, err
	}
	return c, nil
}

// HandoffConversation moves an open conversation to HANDED_OFF, recording the
// reason for the transfer.
func (p *PostgresStore) HandoffConversation(tenantID, conversationID uuid.UUID, operatorID string) (domain.Conversation, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return domain.Conversation{}, err
	}
	defer tx.Rollback()

	var c domain.Conversation
	err = scanConversation(tx.QueryRow(
		`UPDATE conversations SET status='HANDED_OFF', handoff_at=COALESCE(handoff_at, CURRENT_TIMESTAMP), closure_reason=$3, updated_at=CURRENT_TIMESTAMP
		 WHERE tenant_id=$1 AND id=$2 AND status<>'CLOSED'
		 RETURNING `+conversationColumns, tenantID, conversationID, "operator handoff"), &c)
	if err == sql.ErrNoRows {
		return domain.Conversation{}, fmt.Errorf("conversation not found or is closed")
	}
	if err != nil {
		return domain.Conversation{}, err
	}
	if err := writeAuditLog(tx, tenantID, operatorID, "HANDOFF", conversationID, map[string]any{"reason": "operator handoff"}); err != nil {
		return domain.Conversation{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Conversation{}, err
	}
	return c, nil
}

// DeleteConversation permanently removes a tenant-owned conversation and its
// messages, bot session, and activities through database cascades.
func (p *PostgresStore) DeleteConversation(tenantID, conversationID uuid.UUID) error {
	result, err := p.db.Exec(`DELETE FROM conversations WHERE tenant_id=$1 AND id=$2`, tenantID, conversationID)
	if err != nil {
		return err
	}
	if n, err := result.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// RequestConversationDeletion records an operator's request for an admin to
// permanently remove the conversation.
func (p *PostgresStore) RequestConversationDeletion(tenantID, conversationID uuid.UUID, operatorID string) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var exists bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM conversations WHERE tenant_id=$1 AND id=$2)`, tenantID, conversationID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}
	if err := writeAuditLog(tx, tenantID, operatorID, "DELETE_REQUESTED", conversationID, map[string]any{"reason": "operator requested deletion"}); err != nil {
		return err
	}
	return tx.Commit()
}

// CloseConversationWithReason closes an open conversation with an explicit
// reason. Unlike CloseConversation it does not create a follow-up activity.
func (p *PostgresStore) CloseConversationWithReason(tenantID, conversationID uuid.UUID, reason string, operatorID string) (domain.Conversation, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return domain.Conversation{}, err
	}
	defer tx.Rollback()

	var c domain.Conversation
	err = scanConversation(tx.QueryRow(
		`UPDATE conversations SET status='CLOSED', closed_at=COALESCE(closed_at, CURRENT_TIMESTAMP), closure_reason=$3, updated_at=CURRENT_TIMESTAMP
		 WHERE tenant_id=$1 AND id=$2 AND status<>'CLOSED'
		 RETURNING `+conversationColumns, tenantID, conversationID, reason), &c)
	if err == sql.ErrNoRows {
		return domain.Conversation{}, fmt.Errorf("conversation not found or is already closed")
	}
	if err != nil {
		return domain.Conversation{}, err
	}
	if err := writeAuditLog(tx, tenantID, operatorID, "CLOSE", conversationID, map[string]any{"reason": reason}); err != nil {
		return domain.Conversation{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Conversation{}, err
	}
	return c, nil
}

// ReopenConversation reopens a closed conversation, clearing closure metadata.
func (p *PostgresStore) ReopenConversation(tenantID, conversationID uuid.UUID, operatorID string) (domain.Conversation, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return domain.Conversation{}, err
	}
	defer tx.Rollback()

	var c domain.Conversation
	err = scanConversation(tx.QueryRow(
		`UPDATE conversations SET status='OPEN', closed_at=NULL, handoff_at=NULL, closure_reason='', updated_at=CURRENT_TIMESTAMP
		 WHERE tenant_id=$1 AND id=$2 AND status='CLOSED'
		 RETURNING `+conversationColumns, tenantID, conversationID), &c)
	if err == sql.ErrNoRows {
		return domain.Conversation{}, fmt.Errorf("conversation is not closed or not found")
	}
	if err != nil {
		return domain.Conversation{}, err
	}
	if err := writeAuditLog(tx, tenantID, operatorID, "REOPEN", conversationID, map[string]any{}); err != nil {
		return domain.Conversation{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Conversation{}, err
	}
	return c, nil
}
