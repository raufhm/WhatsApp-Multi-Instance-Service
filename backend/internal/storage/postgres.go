package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/raufhm/whatsapp-testing/domain"
	"github.com/raufhm/whatsapp-testing/internal/conversation"
)

type PostgresStore struct {
	db          *sql.DB
	seen        sync.Map
	uploadLease time.Duration
}

// SetUploadClaimLease overrides the claim lease used by the upload worker.
func (p *PostgresStore) SetUploadClaimLease(lease time.Duration) {
	if lease <= 0 {
		lease = time.Minute
	}
	p.uploadLease = lease
}

// AccountHost resolves an API account within its authenticated tenant. It is
// deliberately separate from the public repository contract so lightweight
// handler fakes do not need database-specific account lookup behavior.
func (p *PostgresStore) AccountHost(tenantID uuid.UUID, account string) (string, error) {
	var host string
	err := p.db.QueryRow(`SELECT host_id FROM whatsapp_accounts WHERE tenant_id=$1 AND (id::text=$2 OR host_id=$2)`, tenantID, account).Scan(&host)
	return host, err
}

func (p *PostgresStore) AuthenticateAPIKey(key string) (uuid.UUID, error) {
	hash := sha256.Sum256([]byte(key))
	var tenant uuid.UUID
	err := p.db.QueryRow(`SELECT tenant_id FROM api_keys WHERE key_hash=$1 AND revoked_at IS NULL`, hex.EncodeToString(hash[:])).Scan(&tenant)
	return tenant, err
}

func (p *PostgresStore) RegisterAccount(tenantID uuid.UUID, hostID, displayName, provider string) (domain.WhatsAppAccount, error) {
	if provider == "" {
		provider = "whatsmeow"
	}
	query := `
INSERT INTO whatsapp_accounts (tenant_id, host_id, provider, display_name)
VALUES ($1, $2, $3, $4)
ON CONFLICT (tenant_id, provider, host_id)
DO UPDATE SET display_name = EXCLUDED.display_name, updated_at = CURRENT_TIMESTAMP
RETURNING id, tenant_id, host_id, provider, display_name, created_at, updated_at;`

	var a domain.WhatsAppAccount
	err := p.db.QueryRow(query, tenantID, hostID, provider, displayName).Scan(
		&a.ID, &a.TenantID, &a.HostID, &a.Provider, &a.DisplayName, &a.CreatedAt, &a.UpdatedAt,
	)
	return a, err
}

func (p *PostgresStore) ListAccounts(tenantID uuid.UUID) ([]domain.WhatsAppAccount, error) {
	rows, err := p.db.Query(`SELECT a.id, a.tenant_id, a.host_id, a.provider, a.display_name, a.created_at, a.updated_at FROM whatsapp_accounts a WHERE a.tenant_id=$1 ORDER BY a.created_at`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.WhatsAppAccount
	for rows.Next() {
		var a domain.WhatsAppAccount
		if err := rows.Scan(&a.ID, &a.TenantID, &a.HostID, &a.Provider, &a.DisplayName, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

func (p *PostgresStore) ListContacts(tenantID uuid.UUID, limit, offset int) ([]domain.Contact, error) {
	rows, err := p.db.Query(`SELECT id, tenant_id, normalized_address, provider_address, display_name, metadata, created_at, updated_at FROM contacts WHERE tenant_id=$1 ORDER BY updated_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.Contact
	for rows.Next() {
		var c domain.Contact
		var metadata []byte
		if err := rows.Scan(&c.ID, &c.TenantID, &c.NormalizedAddress, &c.ProviderAddress, &c.DisplayName, &metadata, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		if len(metadata) > 0 {
			_ = json.Unmarshal(metadata, &c.Metadata)
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (p *PostgresStore) ListConversations(tenantID uuid.UUID, status string, limit, offset int) ([]domain.Conversation, error) {
	query := `SELECT id, tenant_id, account_id, contact_id, ticket_number, status, bot_state, started_at, last_activity_at, closed_at, handoff_at, closure_reason, assignee, merged_into_id, created_at, updated_at FROM conversations WHERE tenant_id=$1`
	args := []any{tenantID}
	if status != "" {
		query += ` AND status=$2`
		args = append(args, status)
	}
	query += ` ORDER BY last_activity_at DESC LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)
	args = append(args, limit, offset)
	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.Conversation
	for rows.Next() {
		var c domain.Conversation
		if err := scanConversation(rows, &c); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (p *PostgresStore) GetConversationTimeline(tenantID, conversationID uuid.UUID, limit, offset int) ([]domain.ConversationMessage, error) {
	rows, err := p.db.Query(`SELECT id, tenant_id, conversation_id, actor, provider, provider_message_id, direction, content, message_type, COALESCE(media_url,''), status, provider_timestamp, is_internal, created_at, updated_at FROM conversation_messages WHERE tenant_id=$1 AND conversation_id=$2 ORDER BY provider_timestamp, created_at LIMIT $3 OFFSET $4`, tenantID, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.ConversationMessage
	for rows.Next() {
		var m domain.ConversationMessage
		if err := rows.Scan(&m.ID, &m.TenantID, &m.ConversationID, &m.Actor, &m.Provider, &m.ProviderMessageID, &m.Direction, &m.Content, &m.MessageType, &m.MediaURL, &m.Status, &m.ProviderTimestamp, &m.IsInternal, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

func (p *PostgresStore) ListActivities(tenantID uuid.UUID, status string, limit, offset int) ([]domain.Activity, error) {
	query := `SELECT id, tenant_id, conversation_id, contact_id, type, summary, next_action, priority, status, due_at, acknowledged_by, acknowledged_at, created_at, updated_at FROM activities WHERE tenant_id=$1`
	args := []any{tenantID}
	if status != "" {
		query += ` AND status=$2`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)
	args = append(args, limit, offset)
	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.Activity
	for rows.Next() {
		var a domain.Activity
		if err := scanActivity(rows, &a); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

// ListContactActivities returns a contact's activities newest-first, including
// activities created against any of the contact's conversations.
func (p *PostgresStore) ListContactActivities(tenantID, contactID uuid.UUID, limit, offset int) ([]domain.Activity, error) {
	rows, err := p.db.Query(`SELECT id, tenant_id, conversation_id, contact_id, type, summary, next_action, priority, status, due_at, acknowledged_by, acknowledged_at, created_at, updated_at FROM activities WHERE tenant_id=$1 AND contact_id=$2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`, tenantID, contactID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.Activity
	for rows.Next() {
		var a domain.Activity
		if err := scanActivity(rows, &a); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

// CreateContactActivity stores a follow-up activity scoped to a contact rather
// than a single conversation (conversation_id remains NULL).
func (p *PostgresStore) CreateContactActivity(tenantID, contactID uuid.UUID, input domain.ContactActivityInput) (domain.Activity, error) {
	var a domain.Activity
	err := scanActivity(p.db.QueryRow(`INSERT INTO activities (tenant_id, contact_id, type, summary, next_action, priority, due_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, tenant_id, conversation_id, contact_id, type, summary, next_action, priority, status, due_at, acknowledged_by, acknowledged_at, created_at, updated_at`,
		tenantID, contactID, input.Type, input.Summary, input.NextAction, input.Priority, nullTime(input.DueAt)), &a)
	return a, err
}

func nullTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return *t
}

// scanActivity scans a row in the canonical activity column order
// (id, tenant_id, conversation_id, contact_id, type, ..., updated_at).
func scanActivity(row scanner, a *domain.Activity) error {
	var conv, contact uuid.NullUUID
	var by sql.NullString
	var due, ack sql.NullTime
	if err := row.Scan(&a.ID, &a.TenantID, &conv, &contact, &a.Type, &a.Summary, &a.NextAction, &a.Priority, &a.Status, &due, &by, &ack, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return err
	}
	if conv.Valid {
		a.ConversationID = conv.UUID
	}
	if contact.Valid {
		a.ContactID = contact.UUID
	}
	if due.Valid {
		a.DueAt = &due.Time
	}
	if by.Valid {
		a.AcknowledgedBy = by.String
	}
	if ack.Valid {
		a.AcknowledgedAt = &ack.Time
	}
	return nil
}

func (p *PostgresStore) AcknowledgeActivity(tenantID, activityID uuid.UUID, actor string, at time.Time) (domain.Activity, error) {
	var a domain.Activity
	err := scanActivity(p.db.QueryRow(`UPDATE activities SET status='ACKNOWLEDGED', acknowledged_by=COALESCE(acknowledged_by,$1), acknowledged_at=COALESCE(acknowledged_at,$2), updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$3 AND id=$4 RETURNING id,tenant_id,conversation_id,contact_id,type,summary,next_action,priority,status,due_at,acknowledged_by,acknowledged_at,created_at,updated_at`, actor, at, tenantID, activityID), &a)
	return a, err
}

func NormalizeWhatsAppAddress(address string) (string, error) {
	return conversation.NormalizeAddress(address)
}

func (p *PostgresStore) ContactHost(tenantID uuid.UUID, providerAddress string) (string, error) {
	// Not used or stub
	return "", nil
}

func (p *PostgresStore) UpsertContact(tenantID uuid.UUID, input domain.ContactUpsert) (domain.Contact, error) {
	normalized, err := conversation.NormalizeAddress(input.ProviderAddress)
	if err != nil {
		return domain.Contact{}, err
	}
	metadata, err := json.Marshal(input.Metadata)
	if err != nil {
		return domain.Contact{}, err
	}
	var c domain.Contact
	var metadataJSON []byte
	err = p.db.QueryRow(`INSERT INTO contacts (tenant_id, normalized_address, provider_address, display_name, metadata)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (tenant_id, normalized_address) DO UPDATE SET provider_address=EXCLUDED.provider_address,
		display_name=CASE WHEN EXCLUDED.display_name <> '' THEN EXCLUDED.display_name ELSE contacts.display_name END,
		metadata=CASE WHEN EXCLUDED.metadata <> '{}'::jsonb THEN EXCLUDED.metadata ELSE contacts.metadata END,
		updated_at=CURRENT_TIMESTAMP
		RETURNING id, tenant_id, normalized_address, provider_address, display_name, metadata, created_at, updated_at`,
		tenantID, normalized, input.ProviderAddress, input.DisplayName, metadata).Scan(&c.ID, &c.TenantID, &c.NormalizedAddress, &c.ProviderAddress, &c.DisplayName, &metadataJSON, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return domain.Contact{}, err
	}
	if len(metadataJSON) > 0 {
		_ = json.Unmarshal(metadataJSON, &c.Metadata)
	}
	return c, nil
}

func (p *PostgresStore) GetContact(tenantID, id uuid.UUID) (domain.Contact, error) {
	var c domain.Contact
	var metadata []byte
	err := p.db.QueryRow(`SELECT id, tenant_id, normalized_address, provider_address, display_name, metadata, created_at, updated_at FROM contacts WHERE tenant_id=$1 AND id=$2`, tenantID, id).Scan(&c.ID, &c.TenantID, &c.NormalizedAddress, &c.ProviderAddress, &c.DisplayName, &metadata, &c.CreatedAt, &c.UpdatedAt)
	if err == nil && len(metadata) > 0 {
		err = json.Unmarshal(metadata, &c.Metadata)
	}
	return c, err
}

func (p *PostgresStore) FindOrCreateConversation(key domain.ConversationKey, now time.Time) (domain.Conversation, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return domain.Conversation{}, err
	}
	var c domain.Conversation
	err = scanConversation(tx.QueryRow(`SELECT id, tenant_id, account_id, contact_id, ticket_number, status, bot_state, started_at, last_activity_at, closed_at, handoff_at, closure_reason, assignee, merged_into_id, created_at, updated_at FROM conversations WHERE tenant_id=$1 AND account_id=$2 AND contact_id=$3 AND status <> 'CLOSED' FOR UPDATE`, key.TenantID, key.AccountID, key.ContactID), &c)
	if err == sql.ErrNoRows {
		// DO NOTHING makes the unique active-conversation index the arbiter for
		// concurrent callbacks; the follow-up select reads the winning row.
		err = scanConversation(tx.QueryRow(`INSERT INTO conversations (tenant_id, account_id, contact_id, started_at, last_activity_at) VALUES ($1,$2,$3,$4,$4) ON CONFLICT DO NOTHING RETURNING id, tenant_id, account_id, contact_id, ticket_number, status, bot_state, started_at, last_activity_at, closed_at, handoff_at, closure_reason, assignee, merged_into_id, created_at, updated_at`, key.TenantID, key.AccountID, key.ContactID, now), &c)
		if err == sql.ErrNoRows {
			err = scanConversation(tx.QueryRow(`SELECT id, tenant_id, account_id, contact_id, ticket_number, status, bot_state, started_at, last_activity_at, closed_at, handoff_at, closure_reason, assignee, merged_into_id, created_at, updated_at FROM conversations WHERE tenant_id=$1 AND account_id=$2 AND contact_id=$3 AND status <> 'CLOSED' FOR UPDATE`, key.TenantID, key.AccountID, key.ContactID), &c)
		}
	}
	if err != nil {
		_ = tx.Rollback()
		return domain.Conversation{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.Conversation{}, err
	}
	return c, nil
}

type scanner interface{ Scan(...any) error }

func scanConversation(row scanner, c *domain.Conversation) error {
	var closureReason sql.NullString
	var assignee sql.NullString
	var mergedInto uuid.NullUUID
	if err := row.Scan(&c.ID, &c.TenantID, &c.AccountID, &c.ContactID, &c.TicketNumber, &c.Status, &c.BotState, &c.StartedAt, &c.LastActivityAt, &c.ClosedAt, &c.HandoffAt, &closureReason, &assignee, &mergedInto, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return err
	}
	if closureReason.Valid {
		c.ClosureReason = closureReason.String
	} else {
		c.ClosureReason = ""
	}
	if assignee.Valid {
		c.Assignee = assignee.String
	} else {
		c.Assignee = ""
	}
	if mergedInto.Valid {
		c.MergedIntoID = &mergedInto.UUID
	} else {
		c.MergedIntoID = nil
	}
	return nil
}

func (p *PostgresStore) GetConversation(tenantID, id uuid.UUID) (domain.Conversation, error) {
	var c domain.Conversation
	err := scanConversation(p.db.QueryRow(`SELECT id, tenant_id, account_id, contact_id, ticket_number, status, bot_state, started_at, last_activity_at, closed_at, handoff_at, closure_reason, assignee, merged_into_id, created_at, updated_at FROM conversations WHERE tenant_id=$1 AND id=$2`, tenantID, id), &c)
	return c, err
}

// ProjectMessage atomically upserts the contact, active conversation, and
// timeline message. The account lookup is also the tenant boundary: events
// from hosts that have not been linked to an account are ignored.
func (p *PostgresStore) ProjectMessage(meta domain.MessageMetadata) error {
	providerID := meta.WhatsappID
	if providerID == "" {
		sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%s|%s", meta.HostID, meta.Sender, meta.Recipient, meta.Content, meta.Timestamp.UTC().Format(time.RFC3339Nano))))
		providerID = "local:" + hex.EncodeToString(sum[:])
	}
	address := meta.Sender
	actor := domain.ActorContact
	if meta.Direction == domain.Outgoing {
		address = meta.Recipient
		actor = domain.ActorOperator
	}
	if meta.Actor != "" {
		actor = meta.Actor
	}
	normalized, err := conversation.NormalizeAddress(address)
	if err != nil {
		return nil // malformed transport metadata must not block event handling
	}
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var tenantID, accountID uuid.UUID
	err = tx.QueryRow(`SELECT tenant_id, id FROM whatsapp_accounts WHERE provider='whatsmeow' AND host_id=$1`, meta.HostID).Scan(&tenantID, &accountID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	var contactID uuid.UUID
	err = tx.QueryRow(`INSERT INTO contacts (tenant_id, normalized_address, provider_address) VALUES ($1,$2,$3)
		ON CONFLICT (tenant_id, normalized_address) DO UPDATE SET provider_address=EXCLUDED.provider_address, updated_at=CURRENT_TIMESTAMP
		RETURNING id`, tenantID, normalized, address).Scan(&contactID)
	if err != nil {
		return err
	}
	var conversationID uuid.UUID
	err = tx.QueryRow(`INSERT INTO conversations (tenant_id, account_id, contact_id, started_at, last_activity_at) VALUES ($1,$2,$3,$4,$4)
		ON CONFLICT DO NOTHING RETURNING id`, tenantID, accountID, contactID, meta.Timestamp).Scan(&conversationID)
	if err == sql.ErrNoRows {
		err = tx.QueryRow(`SELECT id FROM conversations WHERE tenant_id=$1 AND account_id=$2 AND contact_id=$3 AND status <> 'CLOSED' FOR UPDATE`, tenantID, accountID, contactID).Scan(&conversationID)
	}
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO conversation_messages (tenant_id, conversation_id, actor, provider, provider_message_id, direction, content, message_type, media_url, status, provider_timestamp)
		VALUES ($1,$2,$3,'whatsmeow',$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (tenant_id, provider, provider_message_id) DO UPDATE SET status=CASE WHEN conversation_messages.status='READ' OR (conversation_messages.status='DELIVERED' AND EXCLUDED.status='SENT') THEN conversation_messages.status ELSE EXCLUDED.status END, media_url=COALESCE(EXCLUDED.media_url, conversation_messages.media_url), updated_at=CURRENT_TIMESTAMP`,
		tenantID, conversationID, actor, providerID, meta.Direction, meta.Content, meta.Type, nullString(meta.MediaURL), meta.Status, meta.Timestamp)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE conversations SET last_activity_at=$1, updated_at=CURRENT_TIMESTAMP WHERE id=$2`, meta.Timestamp, conversationID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// ProjectMessageContext projects an event and returns the resolved application
// context for the asynchronous bot worker. The process-local seen set avoids
// re-running automation for duplicate callbacks; the timeline uniqueness
// constraint remains the durable duplicate guard.
func (p *PostgresStore) ProjectMessageContext(meta domain.MessageMetadata) (domain.ProjectedMessage, error) {
	if err := p.ProjectMessage(meta); err != nil {
		return domain.ProjectedMessage{}, err
	}
	address := meta.Sender
	inbound := meta.Direction == domain.Incoming
	if !inbound {
		address = meta.Recipient
	}
	normalized, err := conversation.NormalizeAddress(address)
	if err != nil {
		return domain.ProjectedMessage{}, nil
	}
	var tenantID, accountID, contactID, conversationID uuid.UUID
	var status domain.ConversationStatus
	var botState string
	err = p.db.QueryRow(`SELECT tenant_id, id FROM whatsapp_accounts WHERE provider='whatsmeow' AND host_id=$1`, meta.HostID).Scan(&tenantID, &accountID)
	if err == sql.ErrNoRows {
		return domain.ProjectedMessage{}, nil
	}
	if err != nil {
		return domain.ProjectedMessage{}, err
	}
	err = p.db.QueryRow(`SELECT id FROM contacts WHERE tenant_id=$1 AND normalized_address=$2`, tenantID, normalized).Scan(&contactID)
	if err != nil {
		return domain.ProjectedMessage{}, err
	}
	err = p.db.QueryRow(`SELECT id, status, bot_state FROM conversations WHERE tenant_id=$1 AND account_id=$2 AND contact_id=$3 ORDER BY created_at DESC LIMIT 1`, tenantID, accountID, contactID).Scan(&conversationID, &status, &botState)
	if err != nil {
		return domain.ProjectedMessage{}, err
	}
	providerID := meta.WhatsappID
	if providerID == "" {
		providerID = fmt.Sprintf("%s|%s|%s|%s", meta.HostID, address, meta.Content, meta.Timestamp.UTC().Format(time.RFC3339Nano))
	}
	_, newEvent := p.seen.LoadOrStore(tenantID.String()+":"+providerID, struct{}{})
	return domain.ProjectedMessage{
		TenantID: tenantID, ConversationID: conversationID, Host: meta.HostID,
		Recipient: address, Text: meta.Content, At: meta.Timestamp, Inbound: inbound,
		New: !newEvent, BotEligible: inbound && status != domain.ConversationHandedOff && status != domain.ConversationClosed && botState == "ACTIVE",
	}, nil
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func (p *PostgresStore) SaveBotSession(tenantID, conversationID uuid.UUID, version string, state map[string]any) error {
	payload, err := json.Marshal(state)
	if err != nil {
		return err
	}
	_, err = p.db.Exec(`INSERT INTO bot_sessions (tenant_id, conversation_id, rule_version, state)
		VALUES ($1,$2,$3,$4) ON CONFLICT (conversation_id) DO UPDATE SET rule_version=EXCLUDED.rule_version, state=EXCLUDED.state, updated_at=CURRENT_TIMESTAMP`, tenantID, conversationID, version, payload)
	return err
}

// CloseConversation changes the lifecycle and creates the follow-up item in
// one transaction. The partial unique index on pending activities makes
// repeated terminal/handoff events harmless.
func (p *PostgresStore) CloseConversation(tenantID, conversationID uuid.UUID, status domain.ConversationStatus, reason string, at time.Time) (domain.Activity, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return domain.Activity{}, err
	}
	defer tx.Rollback()
	if status != domain.ConversationClosed && status != domain.ConversationHandedOff {
		return domain.Activity{}, fmt.Errorf("invalid closing status %q", status)
	}
	_, err = tx.Exec(`UPDATE conversations SET status=$1, closed_at=CASE WHEN $1='CLOSED' THEN $2 ELSE closed_at END, handoff_at=CASE WHEN $1='HANDED_OFF' THEN $2 ELSE handoff_at END, closure_reason=$3, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$4 AND id=$5 AND status <> 'CLOSED'`, status, at, reason, tenantID, conversationID)
	if err != nil {
		return domain.Activity{}, err
	}
	var activity domain.Activity
	err = scanActivity(tx.QueryRow(`INSERT INTO activities (tenant_id, conversation_id, contact_id, type, summary, next_action, priority)
		SELECT $1, $2, (SELECT contact_id FROM conversations WHERE id=$2), CASE WHEN $3='HANDED_OFF' THEN 'HANDOFF' ELSE 'SESSION_CLOSED' END,
		COALESCE((SELECT string_agg(actor || ': ' || content, ' | ' ORDER BY provider_timestamp, created_at) FROM conversation_messages WHERE tenant_id=$1 AND conversation_id=$2), 'No messages recorded'),
		CASE WHEN $3='HANDED_OFF' THEN 'Assign an operator and follow up' ELSE 'Review the completed session' END, 'NORMAL'
		ON CONFLICT (conversation_id, type) WHERE status='PENDING' DO UPDATE SET updated_at=CURRENT_TIMESTAMP
		RETURNING id, tenant_id, conversation_id, contact_id, type, summary, next_action, priority, status, due_at, acknowledged_by, acknowledged_at, created_at, updated_at`, tenantID, conversationID, status), &activity)
	if err != nil {
		return domain.Activity{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.Activity{}, err
	}
	return activity, nil
}

func (p *PostgresStore) CloseTimedOut(tenantID uuid.UUID, timeout time.Duration, at time.Time) error {
	rows, err := p.db.Query(`SELECT id FROM conversations WHERE tenant_id=$1 AND status IN ('OPEN','BOT_ACTIVE','WAITING') AND last_activity_at < $2`, tenantID, at.Add(-timeout))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return err
		}
		if _, err := p.CloseConversation(tenantID, id, domain.ConversationClosed, "session timeout", at); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (p *PostgresStore) CloseAllTimedOut(timeout time.Duration, at time.Time) error {
	rows, err := p.db.Query(`SELECT DISTINCT tenant_id FROM conversations WHERE status IN ('OPEN','BOT_ACTIVE','WAITING') AND last_activity_at < $1`, at.Add(-timeout))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var tenantID uuid.UUID
		if err := rows.Scan(&tenantID); err != nil {
			return err
		}
		if err := p.CloseTimedOut(tenantID, timeout, at); err != nil {
			return err
		}
	}
	return rows.Err()
}

// ProjectReceipt reconciles delivery state in the application timeline. The
// raw receipt is persisted by DispatchReceipt independently.
func (p *PostgresStore) ProjectReceipt(receipt domain.Receipt) error {
	if receipt.WhatsappID == "" || receipt.Sender == "" {
		return nil
	}
	_, err := p.db.Exec(`UPDATE conversation_messages m SET status=$1, updated_at=CURRENT_TIMESTAMP
		FROM whatsapp_accounts a WHERE a.tenant_id=m.tenant_id AND a.host_id=$2 AND m.provider='whatsmeow' AND m.provider_message_id=$3
		AND (m.status NOT IN ('READ') OR $1='READ')`, receipt.Status, receipt.Sender, receipt.WhatsappID)
	return err
}

func NewPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := runMigrations(db); err != nil {
		log.Printf("DB: Migration error: %v", err)
		return nil, err
	}

	return &PostgresStore{db: db, uploadLease: time.Minute}, nil
}

func runMigrations(db *sql.DB) error {
	log.Println("DB: Starting migrations...")
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("instance: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("up: %w", err)
	}
	log.Println("DB: Migrations complete.")
	return nil
}

func (p *PostgresStore) DispatchMessage(meta domain.MessageMetadata) {
	query := `INSERT INTO whatsmeow_messages (whatsapp_id, host_id, sender, recipient, content, is_group, direction, msg_type, reaction_target, media_url, status, timestamp) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	          ON CONFLICT (whatsapp_id) DO UPDATE 
	          SET status = EXCLUDED.status`
	_, err := p.db.Exec(query, meta.WhatsappID, meta.HostID, meta.Sender, meta.Recipient, meta.Content, meta.IsGroup, string(meta.Direction), string(meta.Type), meta.ReactionTarget, meta.MediaURL, string(meta.Status), meta.Timestamp)
	if err != nil {
		log.Printf("PG Store DispatchMessage Error: %v", err)
	}
}

func (p *PostgresStore) DispatchReceipt(receipt domain.Receipt) {
	query := `INSERT INTO whatsmeow_message_receipts (whatsapp_id, recipient_id, status, timestamp) 
	          VALUES ($1, $2, $3, $4)
	          ON CONFLICT (whatsapp_id, recipient_id, status) DO NOTHING`
	_, err := p.db.Exec(query, receipt.WhatsappID, receipt.Recipient, string(receipt.Status), receipt.Timestamp)
	if err != nil {
		log.Printf("PG Store DispatchReceipt Error: %v", err)
	}

	// Update main message status if read
	if receipt.Status == domain.StatusRead || receipt.Status == domain.StatusDelivered {
		updateQuery := `UPDATE whatsmeow_messages SET status = $1 WHERE whatsapp_id = $2 AND direction = 'OUTGOING'`
		_, _ = p.db.Exec(updateQuery, string(receipt.Status), receipt.WhatsappID)
	}
}

func (p *PostgresStore) UpdateInstanceStatus(hostID string, status domain.InstanceStatus, isConnected bool) {
	query := `INSERT INTO whatsmeow_instances (host_id, status, is_connected, last_seen)
	          VALUES ($1, $2, $3, NOW())
	          ON CONFLICT (host_id) DO UPDATE
	          SET status = EXCLUDED.status,
	              is_connected = EXCLUDED.is_connected,
	              last_seen = NOW()`
	if _, err := p.db.Exec(query, hostID, string(status), isConnected); err != nil {
		log.Printf("PG Store UpdateInstanceStatus Error: %v", err)
	}
}

// TenantForHost resolves the tenant that owns a linked WhatsApp host. It is
// used by the upload worker to stamp durable upload jobs with their tenant.
func (p *PostgresStore) TenantForHost(hostID string) (uuid.UUID, error) {
	var tenant uuid.UUID
	err := p.db.QueryRow(`SELECT tenant_id FROM whatsapp_accounts WHERE provider='whatsmeow' AND host_id=$1`, hostID).Scan(&tenant)
	return tenant, err
}

const uploadJobColumns = `id, tenant_id, message_id, host_id, object_key, mime_type, media_path, status, attempt_count, next_attempt_at, last_error, media_url, lease_until, created_at, updated_at`

func scanUploadJob(row scanner, j *domain.UploadJob) error {
	var lastErr, mediaURL sql.NullString
	var lease sql.NullTime
	if err := row.Scan(&j.ID, &j.TenantID, &j.MessageID, &j.HostID, &j.ObjectKey, &j.MimeType, &j.MediaPath, &j.Status, &j.AttemptCount, &j.NextAttemptAt, &lastErr, &mediaURL, &lease, &j.CreatedAt, &j.UpdatedAt); err != nil {
		return err
	}
	if lastErr.Valid {
		j.LastError = lastErr.String
	}
	if mediaURL.Valid {
		j.MediaURL = mediaURL.String
	}
	if lease.Valid {
		j.LeaseUntil = &lease.Time
	}
	return nil
}

// CreateUploadJob persists a durable upload job. The object key is supplied by
// the caller and reused for every retry so uploads converge to one object.
func (p *PostgresStore) CreateUploadJob(ctx context.Context, tenantID uuid.UUID, job domain.UploadJob) (domain.UploadJob, error) {
	var j domain.UploadJob
	err := scanUploadJob(p.db.QueryRowContext(ctx, `INSERT INTO upload_jobs (tenant_id, message_id, host_id, object_key, mime_type, media_path, status, next_attempt_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING `+uploadJobColumns,
		tenantID, job.MessageID, job.HostID, job.ObjectKey, job.MimeType, job.MediaPath, job.Status, job.NextAttemptAt), &j)
	return j, err
}

// ClaimDueUploadJobs atomically leases a batch of due jobs to this worker. The
// FOR UPDATE SKIP LOCKED subquery prevents multiple workers from claiming the
// same job, and PROCESSING rows with an expired lease are reclaimed so a crash
// between S3 success and persistence converges on the next attempt.
func (p *PostgresStore) ClaimDueUploadJobs(ctx context.Context, now time.Time, limit int) ([]domain.UploadJob, error) {
	if limit < 1 {
		limit = 1
	}
	lease := p.uploadLease
	if lease <= 0 {
		lease = time.Minute
	}
	leaseUntil := now.Add(lease)
	rows, err := p.db.QueryContext(ctx, `WITH due AS (
		SELECT id FROM upload_jobs
		WHERE (status='PENDING' AND next_attempt_at <= $1)
		   OR (status='PROCESSING' AND lease_until < $1)
		ORDER BY next_attempt_at
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	)
	UPDATE upload_jobs SET status='PROCESSING', lease_until=$3, attempt_count=attempt_count+1, updated_at=CURRENT_TIMESTAMP
	WHERE id IN (SELECT id FROM due)
	RETURNING `+uploadJobColumns, now, limit, leaseUntil)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []domain.UploadJob
	for rows.Next() {
		var j domain.UploadJob
		if err := scanUploadJob(rows, &j); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (p *PostgresStore) MarkUploadCompleted(ctx context.Context, id uuid.UUID, mediaURL string) error {
	_, err := p.db.ExecContext(ctx, `UPDATE upload_jobs SET status='COMPLETED', media_url=$2, last_error=NULL, lease_until=NULL, updated_at=CURRENT_TIMESTAMP WHERE id=$1`, id, mediaURL)
	return err
}

func (p *PostgresStore) MarkUploadRetryable(ctx context.Context, id uuid.UUID, lastError string, nextAttemptAt time.Time, attemptCount int) error {
	_, err := p.db.ExecContext(ctx, `UPDATE upload_jobs SET status='PENDING', last_error=$2, next_attempt_at=$3, attempt_count=$4, lease_until=NULL, updated_at=CURRENT_TIMESTAMP WHERE id=$1`, id, lastError, nextAttemptAt, attemptCount)
	return err
}

func (p *PostgresStore) MarkUploadFailed(ctx context.Context, id uuid.UUID, lastError string, attemptCount int) error {
	_, err := p.db.ExecContext(ctx, `UPDATE upload_jobs SET status='FAILED', last_error=$2, attempt_count=$3, lease_until=NULL, updated_at=CURRENT_TIMESTAMP WHERE id=$1`, id, lastError, attemptCount)
	return err
}

// ListUploadJobs returns tenant upload jobs optionally filtered by status.
func (p *PostgresStore) ListUploadJobs(tenantID uuid.UUID, status string, limit, offset int) ([]domain.UploadJob, error) {
	const columns = `id, tenant_id, message_id, host_id, object_key, mime_type, media_path, status, attempt_count, next_attempt_at, last_error, media_url, lease_until, created_at, updated_at`
	var rows *sql.Rows
	var err error
	if status != "" {
		rows, err = p.db.Query(`SELECT `+columns+` FROM upload_jobs WHERE tenant_id=$1 AND status=$2 ORDER BY next_attempt_at DESC LIMIT $3 OFFSET $4`, tenantID, status, limit, offset)
	} else {
		rows, err = p.db.Query(`SELECT `+columns+` FROM upload_jobs WHERE tenant_id=$1 ORDER BY next_attempt_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []domain.UploadJob
	for rows.Next() {
		var j domain.UploadJob
		if err := scanUploadJob(rows, &j); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

// AttachMediaURL persists the archive URL on the message once the S3 upload
// completes, mirroring the existing receipt-reconciliation pattern. No-op when
// the message id is empty (e.g. the outbound send itself failed).
func (p *PostgresStore) AttachMediaURL(hostID, messageID, mediaURL string) {
	if messageID == "" {
		return
	}
	_, _ = p.db.Exec(`UPDATE whatsmeow_messages SET media_url=$1 WHERE whatsapp_id=$2`, mediaURL, messageID)
	_, _ = p.db.Exec(`UPDATE conversation_messages m SET media_url=$1 FROM whatsapp_accounts a WHERE a.tenant_id=m.tenant_id AND a.host_id=$2 AND m.provider='whatsmeow' AND m.provider_message_id=$3`, mediaURL, hostID, messageID)
}

func (p *PostgresStore) UpdateGroup(group domain.GroupInfo) {
	participantsJSON, _ := json.Marshal(group.Participants)
	hostsJSON, _ := json.Marshal(group.Hosts)

	query := `INSERT INTO whatsmeow_groups (group_id, name, description, owner_jid, participants, hosts, participant_count, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	          ON CONFLICT (group_id) DO UPDATE 
	          SET name = EXCLUDED.name,
	              description = EXCLUDED.description,
	              participants = EXCLUDED.participants,
	              hosts = EXCLUDED.hosts,
	              participant_count = EXCLUDED.participant_count,
	              updated_at = EXCLUDED.updated_at`
	_, err := p.db.Exec(query, group.GroupID, group.Name, group.Description, group.OwnerJID, participantsJSON, hostsJSON, group.ParticipantCount, group.CreatedAt, group.UpdatedAt)
	if err != nil {
		log.Printf("PG Store UpdateGroup Error: %v", err)
	}
}
