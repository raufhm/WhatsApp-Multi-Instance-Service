package storage

import (
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whatsapp-testing/domain"
)

// conversationMessageColumns lists the conversation_messages projection columns
// in scan order, with COALESCE so nullable media_url scans into a string.
const conversationMessageColumns = `id, tenant_id, conversation_id, actor, provider, provider_message_id, direction, content, message_type, COALESCE(media_url,''), status, provider_timestamp, created_at, updated_at, is_internal`

func scanConversationMessage(row scanner, m *domain.ConversationMessage) error {
	return row.Scan(&m.ID, &m.TenantID, &m.ConversationID, &m.Actor, &m.Provider, &m.ProviderMessageID, &m.Direction, &m.Content, &m.MessageType, &m.MediaURL, &m.Status, &m.ProviderTimestamp, &m.CreatedAt, &m.UpdatedAt, &m.IsInternal)
}

// AddInternalNote records an operator note on the conversation timeline without
// sending it to the WhatsApp contact. The note uses a synthetic provider so the
// timeline uniqueness constraint does not collide with real message ids.
func (p *PostgresStore) AddInternalNote(tenantID, conversationID uuid.UUID, actor domain.Actor, operatorID, content string) (domain.ConversationMessage, error) {
	if actor == "" {
		actor = domain.ActorOperator
	}
	providerID := "note:" + uuid.NewString()
	now := time.Now().UTC()

	tx, err := p.db.Begin()
	if err != nil {
		return domain.ConversationMessage{}, err
	}
	defer tx.Rollback()

	var m domain.ConversationMessage
	err = scanConversationMessage(tx.QueryRow(
		`INSERT INTO conversation_messages (tenant_id, conversation_id, actor, provider, provider_message_id, direction, content, message_type, media_url, status, provider_timestamp, is_internal)
		 VALUES ($1,$2,$3,'internal',$4,'OUTGOING',$5,'TEXT',NULL,'SENT',$6,TRUE)
		 RETURNING `+conversationMessageColumns,
		tenantID, conversationID, actor, providerID, content, now), &m)
	if err != nil {
		return domain.ConversationMessage{}, err
	}
	if _, err := tx.Exec(`UPDATE conversations SET last_activity_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$1 AND id=$2`, tenantID, conversationID); err != nil {
		return domain.ConversationMessage{}, err
	}
	_ = operatorID // operatorID is retained for audit callers; the note itself is the audit record
	if err := tx.Commit(); err != nil {
		return domain.ConversationMessage{}, err
	}
	return m, nil
}
