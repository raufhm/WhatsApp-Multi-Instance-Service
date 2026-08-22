package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/raufhm/whatsapp-testing/domain"
	"github.com/raufhm/whatsapp-testing/internal/broadcast"
	"github.com/raufhm/whatsapp-testing/internal/conversation"
)

type PostgresStore struct {
	db          *sql.DB
	seen        sync.Map
	uploadLease time.Duration
	broadcaster *broadcast.Broadcaster
}

// SetBroadcaster wires the SSE fan-out used by the dashboard Monitoring live
// tail. When set, every recorded instance event is pushed to connected clients
// for that host.
func (p *PostgresStore) SetBroadcaster(b *broadcast.Broadcaster) {
	p.broadcaster = b
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
	if err != nil {
		return a, err
	}
	// Seed default deal stages for new tenant (idempotent)
	_ = p.SeedDefaultDealStages(tenantID)
	return a, nil
}

// SeedDefaultDealStages creates the default 7 pipeline stages for a tenant.
// It is idempotent: if stages already exist it does nothing.
func (p *PostgresStore) SeedDefaultDealStages(tenantID uuid.UUID) error {
	var pipelineID uuid.UUID
	err := p.db.QueryRow(`INSERT INTO pipelines (tenant_id, name, description, is_default, is_active)
		VALUES ($1, 'Sales Pipeline', 'Standard customer sales and deal qualification pipeline', true, true)
		ON CONFLICT (tenant_id, name) DO UPDATE SET updated_at=pipelines.updated_at
		RETURNING id`, tenantID).Scan(&pipelineID)
	if err != nil {
		_ = p.db.QueryRow(`SELECT id FROM pipelines WHERE tenant_id=$1 AND is_default=true LIMIT 1`, tenantID).Scan(&pipelineID)
	}

	defaults := []struct {
		key       string
		label     string
		color     string
		icon      string
		sortOrder int
		isWon     bool
		isLost    bool
	}{
		{"NEW_LEAD", "New Lead", "#94a3b8", "user-plus", 1, false, false},
		{"APPOINTMENT_SCHEDULED", "Appointment Scheduled", "#60a5fa", "calendar", 2, false, false},
		{"HOT_LEAD", "Hot Lead", "#f97316", "flame", 3, false, false},
		{"COLD_LEAD", "Cold Lead", "#64748b", "snowflake", 4, false, false},
		{"IN_PROGRESS", "In Progress", "#a78bfa", "spinner", 5, false, false},
		{"CLOSED_WON", "Closed Won", "#22c55e", "trophy", 6, true, false},
		{"CLOSED_LOST", "Closed Lost", "#ef4444", "x-circle", 7, false, true},
	}
	for _, d := range defaults {
		var pID *uuid.UUID
		if pipelineID != uuid.Nil {
			pID = &pipelineID
		}
		_, err := p.db.Exec(`INSERT INTO deal_stages (tenant_id, pipeline_id, key, label, color, icon, sort_order, is_won, is_lost)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT (tenant_id, key) DO UPDATE SET pipeline_id=COALESCE(deal_stages.pipeline_id, EXCLUDED.pipeline_id)`,
			tenantID, pID, d.key, d.label, d.color, d.icon, d.sortOrder, d.isWon, d.isLost)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStore) GetSubscription(tenantID uuid.UUID) (domain.Subscription, error) {
	var s domain.Subscription
	err := p.db.QueryRow(`SELECT tenant_id, plan_type, status, current_period_start, current_period_end, created_at, updated_at FROM subscriptions WHERE tenant_id=$1`, tenantID).Scan(
		&s.TenantID, &s.PlanType, &s.Status, &s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

func (p *PostgresStore) GetQuota(tenantID uuid.UUID) (domain.Quota, error) {
	var q domain.Quota
	err := p.db.QueryRow(`SELECT tenant_id, monthly_limit, current_usage, reset_at, updated_at FROM quotas WHERE tenant_id=$1`, tenantID).Scan(
		&q.TenantID, &q.MonthlyLimit, &q.CurrentUsage, &q.ResetAt, &q.UpdatedAt,
	)
	return q, err
}

func (p *PostgresStore) IncrementQuota(tenantID uuid.UUID) error {
	_, err := p.db.Exec(`UPDATE quotas SET current_usage = current_usage + 1, updated_at = CURRENT_TIMESTAMP WHERE tenant_id=$1`, tenantID)
	return err
}

func (p *PostgresStore) ListAccounts(tenantID uuid.UUID) ([]domain.WhatsAppAccount, error) {
	rows, err := p.db.Query(`SELECT a.id, a.tenant_id, a.host_id, a.provider, a.display_name, a.created_at, a.updated_at FROM whatsapp_accounts a WHERE a.tenant_id=$1 ORDER BY a.created_at`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.WhatsAppAccount, 0)
	for rows.Next() {
		var a domain.WhatsAppAccount
		if err := rows.Scan(&a.ID, &a.TenantID, &a.HostID, &a.Provider, &a.DisplayName, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

func (p *PostgresStore) GetAccount(tenantID, id uuid.UUID) (domain.WhatsAppAccount, error) {
	var a domain.WhatsAppAccount
	err := p.db.QueryRow(`SELECT id, tenant_id, host_id, provider, display_name, created_at, updated_at FROM whatsapp_accounts WHERE tenant_id=$1 AND id=$2`, tenantID, id).Scan(&a.ID, &a.TenantID, &a.HostID, &a.Provider, &a.DisplayName, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

func scanContact(row scanner, c *domain.Contact) error {
	var dealStageKey sql.NullString
	var dealStageID uuid.NullUUID
	var metadata []byte
	if err := row.Scan(&c.ID, &c.TenantID, &c.NormalizedAddress, &c.ProviderAddress, &c.DisplayName, &c.IsGroup, &dealStageKey, &dealStageID, &metadata, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return err
	}
	c.DealStageKey = dealStageKey.String
	if dealStageID.Valid {
		id := dealStageID.UUID
		c.DealStageID = &id
	}
	if len(metadata) > 0 {
		_ = json.Unmarshal(metadata, &c.Metadata)
	}
	return nil
}

func (p *PostgresStore) ListContacts(tenantID uuid.UUID, limit, offset int, search string) ([]domain.Contact, int, error) {
	countQuery := `SELECT COUNT(*) FROM (
		SELECT DISTINCT ON (normalized_address, is_group) id
		FROM contacts
		WHERE tenant_id=$1`
	countArgs := []any{tenantID}
	if search != "" {
		countQuery += ` AND (display_name ILIKE $2 OR provider_address ILIKE $2)`
		countArgs = append(countArgs, "%"+search+"%")
	}
	countQuery += `) sub`

	var total int
	if err := p.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT * FROM (
		SELECT DISTINCT ON (normalized_address, is_group)
			id, tenant_id, normalized_address, provider_address, display_name, is_group, deal_stage_key, deal_stage_id, metadata, created_at, updated_at
		FROM contacts
		WHERE tenant_id=$1`
	args := []any{tenantID}
	argIdx := 1
	if search != "" {
		argIdx++
		query += ` AND (display_name ILIKE $` + fmt.Sprint(argIdx) + ` OR provider_address ILIKE $` + fmt.Sprint(argIdx) + `)`
		args = append(args, "%"+search+"%")
	}
	argIdx++
	query += `
		ORDER BY normalized_address, is_group, updated_at DESC
	) sub
	ORDER BY updated_at DESC LIMIT $` + fmt.Sprint(argIdx) + ` OFFSET $` + fmt.Sprint(argIdx+1)
	args = append(args, limit, offset)

	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	result := make([]domain.Contact, 0)
	for rows.Next() {
		var c domain.Contact
		if err := scanContact(rows, &c); err != nil {
			return nil, 0, err
		}
		result = append(result, c)
	}
	return result, total, rows.Err()
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
	result := make([]domain.Conversation, 0)
	for rows.Next() {
		var c domain.Conversation
		if err := scanConversation(rows, &c); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

// ListConversationSummaries returns one inbox row per contact, enriched with the
// latest related conversation and its most recent timeline message. Always scoped
// by tenant. Multiple tickets for the same contact are intentionally collapsed.
func (p *PostgresStore) ListConversationSummaries(tenantID uuid.UUID, status string, limit, offset int) ([]domain.ConversationSummary, error) {
	query := `SELECT * FROM (
		SELECT DISTINCT ON (c.contact_id)
			c.id, c.tenant_id, c.account_id, c.contact_id, c.ticket_number, c.status, c.bot_state, c.started_at, c.last_activity_at, c.closed_at, c.handoff_at, COALESCE(c.closure_reason, '') AS closure_reason, COALESCE(c.assignee, '') AS assignee, c.merged_into_id, c.created_at, c.updated_at,
			COALESCE((SELECT NULLIF(wg.name, '') FROM whatsmeow_groups wg WHERE (wg.group_id = co.provider_address OR wg.group_id || '@g.us' = co.normalized_address OR wg.group_id = co.normalized_address) LIMIT 1), NULLIF(co.display_name, ''), '') AS contact_name,
			COALESCE(co.provider_address, '') AS contact_number, COALESCE(co.is_group, FALSE) AS is_group,
			COALESCE(m.content, '') AS last_message_preview, COALESCE(m.actor, '') AS last_message_actor
		FROM conversations c
		LEFT JOIN contacts co ON co.id = c.contact_id AND co.tenant_id = c.tenant_id
		LEFT JOIN LATERAL (
			SELECT content, actor FROM conversation_messages
			WHERE conversation_id = c.id
			ORDER BY provider_timestamp DESC, created_at DESC
			LIMIT 1
		) m ON TRUE
		WHERE c.tenant_id = $1`
	args := []any{tenantID}
	if status != "" {
		query += ` AND c.status=$2`
		args = append(args, status)
	}
	query += ` ORDER BY c.contact_id, `
	if status == "" {
		// The default inbox is active; prefer an active ticket over an older
		// closed ticket when both exist for the same contact.
		query += `(c.status = 'CLOSED'), `
	}
	query += `c.last_activity_at DESC
	) inbox ORDER BY last_activity_at DESC LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)
	args = append(args, limit, offset)
	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.ConversationSummary, 0)
	for rows.Next() {
		var s domain.ConversationSummary
		var closureReason, assignee, contactName, contactNumber, lastMessage, lastActor string
		var mergedInto uuid.NullUUID
		if err := rows.Scan(
			&s.ID, &s.TenantID, &s.AccountID, &s.ContactID, &s.TicketNumber, &s.Status, &s.BotState,
			&s.StartedAt, &s.LastActivityAt, &s.ClosedAt, &s.HandoffAt, &closureReason, &assignee,
			&mergedInto, &s.CreatedAt, &s.UpdatedAt,
			&contactName, &contactNumber, &s.IsGroup, &lastMessage, &lastActor,
		); err != nil {
			return nil, err
		}
		if closureReason != "" {
			s.ClosureReason = closureReason
		}
		if assignee != "" {
			s.Assignee = assignee
		}
		if mergedInto.Valid {
			s.MergedIntoID = &mergedInto.UUID
		}
		s.ContactName = contactName
		s.ContactNumber = contactNumber
		s.LastMessage = lastMessage
		if lastActor != "" {
			s.LastActor = domain.Actor(lastActor)
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func (p *PostgresStore) GetConversationTimeline(tenantID, conversationID uuid.UUID, limit, offset int) ([]domain.ConversationMessage, error) {
	rows, err := p.db.Query(`SELECT id, tenant_id, conversation_id, actor, operator_id, COALESCE(operator_name, ''), provider, provider_message_id, direction, COALESCE(sender_address, ''), COALESCE(reaction_target, ''), content, message_type, COALESCE(media_url,''), status, provider_timestamp, is_internal, created_at, updated_at FROM conversation_messages WHERE tenant_id=$1 AND conversation_id=$2 ORDER BY provider_timestamp, created_at LIMIT $3 OFFSET $4`, tenantID, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.ConversationMessage, 0)
	for rows.Next() {
		var m domain.ConversationMessage
		if err := rows.Scan(&m.ID, &m.TenantID, &m.ConversationID, &m.Actor, &m.OperatorID, &m.OperatorName, &m.Provider, &m.ProviderMessageID, &m.Direction, &m.SenderAddress, &m.ReactionTarget, &m.Content, &m.MessageType, &m.MediaURL, &m.Status, &m.ProviderTimestamp, &m.IsInternal, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

// GetContactConversationTimeline returns the complete message history across all
// conversations belonging to a contact, ordered as one chronological timeline.
func (p *PostgresStore) GetContactConversationTimeline(tenantID, contactID uuid.UUID, limit, offset int) ([]domain.ConversationMessage, error) {
	rows, err := p.db.Query(`SELECT m.id, m.tenant_id, m.conversation_id, m.actor, m.operator_id, COALESCE(m.operator_name, ''), m.provider, m.provider_message_id, m.direction, COALESCE(m.sender_address, ''), COALESCE(m.reaction_target, ''), m.content, m.message_type, COALESCE(m.media_url,''), m.status, m.provider_timestamp, m.is_internal, m.created_at, m.updated_at
		FROM conversation_messages m
		JOIN conversations c ON c.id = m.conversation_id AND c.tenant_id = m.tenant_id
		WHERE m.tenant_id=$1 AND c.contact_id=$2
		ORDER BY m.provider_timestamp, m.created_at
		LIMIT $3 OFFSET $4`, tenantID, contactID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.ConversationMessage, 0)
	for rows.Next() {
		var m domain.ConversationMessage
		if err := rows.Scan(&m.ID, &m.TenantID, &m.ConversationID, &m.Actor, &m.OperatorID, &m.OperatorName, &m.Provider, &m.ProviderMessageID, &m.Direction, &m.SenderAddress, &m.ReactionTarget, &m.Content, &m.MessageType, &m.MediaURL, &m.Status, &m.ProviderTimestamp, &m.IsInternal, &m.CreatedAt, &m.UpdatedAt); err != nil {
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
	result := make([]domain.Activity, 0)
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
	result := make([]domain.Activity, 0)
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
	var convID interface{}
	if input.ConversationID != uuid.Nil {
		convID = input.ConversationID
	}
	var a domain.Activity
	err := scanActivity(p.db.QueryRow(`INSERT INTO activities (tenant_id, contact_id, conversation_id, type, summary, next_action, priority, due_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, tenant_id, conversation_id, contact_id, type, summary, next_action, priority, status, due_at, acknowledged_by, acknowledged_at, created_at, updated_at`,
		tenantID, contactID, convID, input.Type, input.Summary, input.NextAction, input.Priority, nullTime(input.DueAt)), &a)
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

func scanPipeline(row scanner, p *domain.Pipeline) error {
	return row.Scan(&p.ID, &p.TenantID, &p.Name, &p.Description, &p.IsDefault, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
}

func (p *PostgresStore) ListPipelines(tenantID uuid.UUID, isActive *bool) ([]domain.Pipeline, error) {
	query := `SELECT id, tenant_id, name, description, is_default, is_active, created_at, updated_at FROM pipelines WHERE tenant_id=$1`
	args := []any{tenantID}
	if isActive != nil {
		query += ` AND is_active=$2`
		args = append(args, *isActive)
	}
	query += ` ORDER BY is_default DESC, name ASC, created_at ASC`
	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.Pipeline, 0)
	for rows.Next() {
		var pl domain.Pipeline
		if err := scanPipeline(rows, &pl); err != nil {
			return nil, err
		}
		result = append(result, pl)
	}
	return result, rows.Err()
}

func (p *PostgresStore) GetPipeline(tenantID, id uuid.UUID) (domain.Pipeline, error) {
	var pl domain.Pipeline
	err := scanPipeline(p.db.QueryRow(`SELECT id, tenant_id, name, description, is_default, is_active, created_at, updated_at FROM pipelines WHERE tenant_id=$1 AND id=$2`, tenantID, id), &pl)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Pipeline{}, domain.ErrPipelineNotFound
	}
	return pl, err
}

func (p *PostgresStore) GetDefaultPipeline(tenantID uuid.UUID) (domain.Pipeline, error) {
	var pl domain.Pipeline
	err := scanPipeline(p.db.QueryRow(`SELECT id, tenant_id, name, description, is_default, is_active, created_at, updated_at FROM pipelines WHERE tenant_id=$1 AND is_default=true LIMIT 1`, tenantID), &pl)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Pipeline{}, domain.ErrPipelineNotFound
	}
	return pl, err
}

func (p *PostgresStore) CreatePipeline(tenantID uuid.UUID, name, description string, isDefault bool) (domain.Pipeline, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Pipeline{}, errors.New("pipeline name is required")
	}

	tx, err := p.db.Begin()
	if err != nil {
		return domain.Pipeline{}, err
	}
	defer tx.Rollback()

	var existingID uuid.UUID
	err = tx.QueryRow(`SELECT id FROM pipelines WHERE tenant_id=$1 AND LOWER(name)=LOWER($2)`, tenantID, name).Scan(&existingID)
	if err == nil {
		return domain.Pipeline{}, domain.ErrPipelineNameExists
	} else if !errors.Is(err, sql.ErrNoRows) {
		return domain.Pipeline{}, err
	}

	if isDefault {
		if _, err := tx.Exec(`UPDATE pipelines SET is_default=false, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$1`, tenantID); err != nil {
			return domain.Pipeline{}, err
		}
	} else {
		var defaultCount int
		_ = tx.QueryRow(`SELECT COUNT(*) FROM pipelines WHERE tenant_id=$1 AND is_default=true`, tenantID).Scan(&defaultCount)
		if defaultCount == 0 {
			isDefault = true
		}
	}

	var pl domain.Pipeline
	err = tx.QueryRow(`INSERT INTO pipelines (tenant_id, name, description, is_default, is_active) VALUES ($1, $2, $3, $4, true) RETURNING id, tenant_id, name, description, is_default, is_active, created_at, updated_at`, tenantID, name, description, isDefault).Scan(&pl.ID, &pl.TenantID, &pl.Name, &pl.Description, &pl.IsDefault, &pl.IsActive, &pl.CreatedAt, &pl.UpdatedAt)
	if err != nil {
		return domain.Pipeline{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Pipeline{}, err
	}
	return pl, nil
}

func (p *PostgresStore) UpdatePipeline(tenantID, id uuid.UUID, name, description *string, isDefault, isActive *bool) (domain.Pipeline, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return domain.Pipeline{}, err
	}
	defer tx.Rollback()

	var existing domain.Pipeline
	err = scanPipeline(tx.QueryRow(`SELECT id, tenant_id, name, description, is_default, is_active, created_at, updated_at FROM pipelines WHERE tenant_id=$1 AND id=$2 FOR UPDATE`, tenantID, id), &existing)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Pipeline{}, domain.ErrPipelineNotFound
	}
	if err != nil {
		return domain.Pipeline{}, err
	}

	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return domain.Pipeline{}, errors.New("pipeline name cannot be empty")
		}
		var dupID uuid.UUID
		err = tx.QueryRow(`SELECT id FROM pipelines WHERE tenant_id=$1 AND LOWER(name)=LOWER($2) AND id <> $3`, tenantID, trimmed, id).Scan(&dupID)
		if err == nil {
			return domain.Pipeline{}, domain.ErrPipelineNameExists
		} else if !errors.Is(err, sql.ErrNoRows) {
			return domain.Pipeline{}, err
		}
		existing.Name = trimmed
	}

	if description != nil {
		existing.Description = *description
	}

	if isDefault != nil {
		if *isDefault {
			if _, err := tx.Exec(`UPDATE pipelines SET is_default=false, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$1 AND id <> $2`, tenantID, id); err != nil {
				return domain.Pipeline{}, err
			}
			existing.IsDefault = true
			existing.IsActive = true
		} else if existing.IsDefault {
			return domain.Pipeline{}, domain.ErrCannotDeleteDefaultPipeline
		}
	}

	if isActive != nil {
		if !*isActive && existing.IsDefault {
			return domain.Pipeline{}, domain.ErrDefaultPipelineCannotBeInactive
		}
		existing.IsActive = *isActive
	}

	var updated domain.Pipeline
	err = tx.QueryRow(`UPDATE pipelines SET name=$1, description=$2, is_default=$3, is_active=$4, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$5 AND id=$6 RETURNING id, tenant_id, name, description, is_default, is_active, created_at, updated_at`, existing.Name, existing.Description, existing.IsDefault, existing.IsActive, tenantID, id).Scan(&updated.ID, &updated.TenantID, &updated.Name, &updated.Description, &updated.IsDefault, &updated.IsActive, &updated.CreatedAt, &updated.UpdatedAt)
	if err != nil {
		return domain.Pipeline{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Pipeline{}, err
	}
	return updated, nil
}

func (p *PostgresStore) DeletePipeline(tenantID, id uuid.UUID) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var isDefault bool
	err = tx.QueryRow(`SELECT is_default FROM pipelines WHERE tenant_id=$1 AND id=$2`, tenantID, id).Scan(&isDefault)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrPipelineNotFound
	}
	if err != nil {
		return err
	}
	if isDefault {
		return domain.ErrCannotDeleteDefaultPipeline
	}

	var stageCount int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM deal_stages WHERE tenant_id=$1 AND pipeline_id=$2 AND is_active=true`, tenantID, id).Scan(&stageCount); err != nil {
		return err
	}
	if stageCount > 0 {
		return domain.ErrPipelineContainsStages
	}

	if _, err := tx.Exec(`DELETE FROM pipelines WHERE tenant_id=$1 AND id=$2`, tenantID, id); err != nil {
		return err
	}

	return tx.Commit()
}

func scanDealStage(row scanner, d *domain.DealStage) error {
	var pipelineID uuid.NullUUID
	if err := row.Scan(&d.ID, &d.TenantID, &pipelineID, &d.Key, &d.Label, &d.Color, &d.Icon, &d.SortOrder, &d.IsActive, &d.IsWon, &d.IsLost, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return err
	}
	if pipelineID.Valid {
		id := pipelineID.UUID
		d.PipelineID = &id
	}
	return nil
}

func (p *PostgresStore) ListDealStages(tenantID uuid.UUID, pipelineID *uuid.UUID, isActive *bool) ([]domain.DealStage, error) {
	query := `SELECT id, tenant_id, pipeline_id, key, label, color, icon, sort_order, is_active, is_won, is_lost, created_at, updated_at FROM deal_stages WHERE tenant_id=$1`
	args := []any{tenantID}
	argIdx := 1
	if pipelineID != nil {
		argIdx++
		query += ` AND pipeline_id=$` + fmt.Sprint(argIdx)
		args = append(args, *pipelineID)
	}
	if isActive != nil {
		argIdx++
		query += ` AND is_active=$` + fmt.Sprint(argIdx)
		args = append(args, *isActive)
	} else {
		query += ` AND is_active=true`
	}
	query += ` ORDER BY sort_order, created_at`
	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.DealStage, 0)
	for rows.Next() {
		var d domain.DealStage
		if err := scanDealStage(rows, &d); err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, rows.Err()
}

func (p *PostgresStore) GetDealStage(tenantID, id uuid.UUID) (domain.DealStage, error) {
	var d domain.DealStage
	err := scanDealStage(p.db.QueryRow(`SELECT id, tenant_id, pipeline_id, key, label, color, icon, sort_order, is_active, is_won, is_lost, created_at, updated_at FROM deal_stages WHERE tenant_id=$1 AND id=$2`, tenantID, id), &d)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DealStage{}, domain.ErrStageNotFound
	}
	return d, err
}

func (p *PostgresStore) CreateDealStage(tenantID uuid.UUID, pipelineID *uuid.UUID, key, label, color, icon string, sortOrder int, isWon, isLost bool) (domain.DealStage, error) {
	if pipelineID == nil {
		defaultPl, err := p.GetDefaultPipeline(tenantID)
		if err == nil {
			pipelineID = &defaultPl.ID
		}
	} else {
		if _, err := p.GetPipeline(tenantID, *pipelineID); err != nil {
			return domain.DealStage{}, domain.ErrPipelineNotFound
		}
	}

	var d domain.DealStage
	var created, updated time.Time
	var resPipelineID uuid.NullUUID
	err := p.db.QueryRow(`INSERT INTO deal_stages (tenant_id, pipeline_id, key, label, color, icon, sort_order, is_won, is_lost) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, tenant_id, pipeline_id, key, label, color, icon, sort_order, is_active, is_won, is_lost, created_at, updated_at`, tenantID, pipelineID, key, label, color, icon, sortOrder, isWon, isLost).Scan(&d.ID, &d.TenantID, &resPipelineID, &d.Key, &d.Label, &d.Color, &d.Icon, &d.SortOrder, &d.IsActive, &d.IsWon, &d.IsLost, &created, &updated)
	if err != nil {
		return domain.DealStage{}, err
	}
	if resPipelineID.Valid {
		id := resPipelineID.UUID
		d.PipelineID = &id
	}
	d.CreatedAt = created
	d.UpdatedAt = updated
	return d, nil
}

func (p *PostgresStore) UpdateDealStage(tenantID, id uuid.UUID, pipelineID *uuid.UUID, label, color, icon *string, sortOrder *int, isActive, isWon, isLost *bool) (domain.DealStage, error) {
	existing, err := p.GetDealStage(tenantID, id)
	if err != nil {
		return domain.DealStage{}, err
	}

	if pipelineID != nil {
		if _, err := p.GetPipeline(tenantID, *pipelineID); err != nil {
			return domain.DealStage{}, domain.ErrPipelineNotFound
		}
		existing.PipelineID = pipelineID
	}
	if label != nil {
		existing.Label = *label
	}
	if color != nil {
		existing.Color = *color
	}
	if icon != nil {
		existing.Icon = *icon
	}
	if sortOrder != nil {
		existing.SortOrder = *sortOrder
	}
	if isActive != nil {
		existing.IsActive = *isActive
	}
	if isWon != nil {
		existing.IsWon = *isWon
	}
	if isLost != nil {
		existing.IsLost = *isLost
	}

	var d domain.DealStage
	var created, updated time.Time
	var resPipelineID uuid.NullUUID
	err = p.db.QueryRow(`UPDATE deal_stages SET pipeline_id=$1, label=$2, color=$3, icon=$4, sort_order=$5, is_active=$6, is_won=$7, is_lost=$8, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$9 AND id=$10 RETURNING id, tenant_id, pipeline_id, key, label, color, icon, sort_order, is_active, is_won, is_lost, created_at, updated_at`, existing.PipelineID, existing.Label, existing.Color, existing.Icon, existing.SortOrder, existing.IsActive, existing.IsWon, existing.IsLost, tenantID, id).Scan(&d.ID, &d.TenantID, &resPipelineID, &d.Key, &d.Label, &d.Color, &d.Icon, &d.SortOrder, &d.IsActive, &d.IsWon, &d.IsLost, &created, &updated)
	if err != nil {
		return domain.DealStage{}, err
	}
	if resPipelineID.Valid {
		pID := resPipelineID.UUID
		d.PipelineID = &pID
	}
	d.CreatedAt = created
	d.UpdatedAt = updated
	return d, nil
}

func (p *PostgresStore) DeleteDealStage(tenantID, id uuid.UUID) error {
	stage, err := p.GetDealStage(tenantID, id)
	if err != nil {
		return err
	}

	var contactCount int
	err = p.db.QueryRow(`SELECT COUNT(*) FROM contacts WHERE tenant_id=$1 AND (deal_stage_id=$2 OR deal_stage_key=$3)`, tenantID, id, stage.Key).Scan(&contactCount)
	if err != nil {
		return err
	}
	if contactCount > 0 {
		return domain.ErrStageAssignedToContacts
	}

	_, err = p.db.Exec(`UPDATE deal_stages SET is_active=false, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$1 AND id=$2`, tenantID, id)
	return err
}

func (p *PostgresStore) MoveContactToStage(tenantID, contactID uuid.UUID, stageKey string, stageID *uuid.UUID, note string, operatorID uuid.UUID) (domain.DealStageTransition, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return domain.DealStageTransition{}, err
	}
	defer tx.Rollback()

	var resolvedKey = stageKey
	var resolvedID = stageID

	if stageID != nil {
		var key string
		err := tx.QueryRow(`SELECT key FROM deal_stages WHERE tenant_id=$1 AND id=$2`, tenantID, *stageID).Scan(&key)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.DealStageTransition{}, domain.ErrStageNotFound
		}
		if err != nil {
			return domain.DealStageTransition{}, err
		}
		resolvedKey = key
	} else if stageKey != "" {
		var id uuid.UUID
		err := tx.QueryRow(`SELECT id FROM deal_stages WHERE tenant_id=$1 AND key=$2 ORDER BY is_active DESC, sort_order ASC LIMIT 1`, tenantID, stageKey).Scan(&id)
		if err == nil {
			resolvedID = &id
		}
	}

	var oldKey sql.NullString
	if err := tx.QueryRow(`SELECT deal_stage_key FROM contacts WHERE tenant_id=$1 AND id=$2 FOR UPDATE`, tenantID, contactID).Scan(&oldKey); err != nil {
		return domain.DealStageTransition{}, err
	}

	if _, err := tx.Exec(`UPDATE contacts SET deal_stage_key=$1, deal_stage_id=$2, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$3 AND id=$4`, resolvedKey, resolvedID, tenantID, contactID); err != nil {
		return domain.DealStageTransition{}, err
	}

	var from *string
	if oldKey.Valid && oldKey.String != "" {
		from = &oldKey.String
	}

	var t domain.DealStageTransition
	var created time.Time
	var movedBy uuid.NullUUID
	var fromStage sql.NullString
	err = tx.QueryRow(`INSERT INTO deal_stage_history (tenant_id, contact_id, from_stage, to_stage, note, moved_by) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, tenant_id, contact_id, from_stage, to_stage, note, moved_by, created_at`, tenantID, contactID, from, resolvedKey, note, operatorID).Scan(&t.ID, &t.TenantID, &t.ContactID, &fromStage, &t.ToStage, &t.Note, &movedBy, &created)
	if err != nil {
		return domain.DealStageTransition{}, err
	}
	if fromStage.Valid {
		str := fromStage.String
		t.FromStage = &str
	}
	if movedBy.Valid {
		id := movedBy.UUID
		t.MovedBy = &id
	}
	t.CreatedAt = created

	if err := tx.Commit(); err != nil {
		return domain.DealStageTransition{}, err
	}
	return t, nil
}

func (p *PostgresStore) ListDealStageHistory(tenantID, contactID uuid.UUID, limit, offset int) ([]domain.DealStageTransition, error) {
	rows, err := p.db.Query(`SELECT id, tenant_id, contact_id, from_stage, to_stage, note, moved_by, created_at FROM deal_stage_history WHERE tenant_id=$1 AND contact_id=$2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`, tenantID, contactID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.DealStageTransition, 0)
	for rows.Next() {
		var t domain.DealStageTransition
		var fromStage sql.NullString
		var movedBy uuid.NullUUID
		var created time.Time
		if err := rows.Scan(&t.ID, &t.TenantID, &t.ContactID, &fromStage, &t.ToStage, &t.Note, &movedBy, &created); err != nil {
			return nil, err
		}
		if fromStage.Valid {
			str := fromStage.String
			t.FromStage = &str
		}
		if movedBy.Valid {
			id := movedBy.UUID
			t.MovedBy = &id
		}
		t.CreatedAt = created
		result = append(result, t)
	}
	return result, rows.Err()
}

func NormalizeWhatsAppAddress(address string) (string, error) {
	return conversation.NormalizeAddress(address)
}

func (p *PostgresStore) ContactHost(tenantID uuid.UUID, providerAddress string) (string, error) {
	// Not used or stub
	return "", nil
}

func (p *PostgresStore) UpsertContact(tenantID uuid.UUID, input domain.ContactUpsert) (domain.Contact, error) {
	// Use the correct server domain based on IsGroup so that group and personal
	// contacts always produce distinct normalized addresses.
	server := "s.whatsapp.net"
	if input.IsGroup {
		server = "g.us"
	}
	normalized, err := conversation.NormalizeAddressWithServer(input.ProviderAddress, server)
	if err != nil {
		return domain.Contact{}, err
	}
	metadata := input.Metadata
	if metadata == nil {
		metadata = make(map[string]any)
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return domain.Contact{}, err
	}
	var c domain.Contact
	row := p.db.QueryRow(`INSERT INTO contacts (tenant_id, normalized_address, provider_address, display_name, is_group, metadata)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (tenant_id, normalized_address, is_group) DO UPDATE SET provider_address=EXCLUDED.provider_address,
		display_name=CASE WHEN EXCLUDED.display_name <> '' THEN EXCLUDED.display_name ELSE contacts.display_name END,
		metadata=CASE WHEN EXCLUDED.metadata <> '{}'\:\:jsonb THEN EXCLUDED.metadata ELSE contacts.metadata END,
		updated_at=CURRENT_TIMESTAMP
		RETURNING id, tenant_id, normalized_address, provider_address, display_name, is_group, deal_stage_key, deal_stage_id, metadata, created_at, updated_at`,
		tenantID, normalized, input.ProviderAddress, input.DisplayName, input.IsGroup, metadataJSON)
	if err := scanContact(row, &c); err != nil {
		return domain.Contact{}, err
	}
	return c, nil
}

func (p *PostgresStore) GetContact(tenantID, id uuid.UUID) (domain.Contact, error) {
	var c domain.Contact
	row := p.db.QueryRow(`SELECT id, tenant_id, normalized_address, provider_address, display_name, is_group, deal_stage_key, deal_stage_id, metadata, created_at, updated_at FROM contacts WHERE tenant_id=$1 AND id=$2`, tenantID, id)
	err := scanContact(row, &c)
	if err == nil && c.IsGroup {
		// The provider group record is authoritative. Contact display names may
		// contain an old JID or a stale fallback from before group metadata sync.
		var groupName string
		if lookupErr := p.db.QueryRow(`SELECT name FROM whatsmeow_groups WHERE (group_id = $1 OR group_id || '@g.us' = $2 OR group_id = $2) AND name <> '' LIMIT 1`, c.ProviderAddress, c.NormalizedAddress).Scan(&groupName); lookupErr == nil && groupName != "" {
			c.DisplayName = groupName
		}
	}
	return c, err
}

func (p *PostgresStore) UpdateContact(tenantID, id uuid.UUID, input domain.ContactUpdateInput) (domain.Contact, error) {
	existing, err := p.GetContact(tenantID, id)
	if err != nil {
		return domain.Contact{}, err
	}
	metadata := existing.Metadata
	if metadata == nil {
		metadata = make(map[string]any)
	}
	if input.Email != "" {
		metadata["email"] = input.Email
	}
	if input.Tags != nil {
		metadata["tags"] = input.Tags
	}
	for k, v := range input.CustomValues {
		metadata[k] = v
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return domain.Contact{}, err
	}

	dealStageKey := existing.DealStageKey
	dealStageID := existing.DealStageID

	if input.ClearDealStage {
		dealStageKey = ""
		dealStageID = nil
	} else if input.DealStageID != nil {
		var key string
		err := p.db.QueryRow(`SELECT key FROM deal_stages WHERE tenant_id=$1 AND id=$2`, tenantID, *input.DealStageID).Scan(&key)
		if err == nil {
			dealStageKey = key
			dealStageID = input.DealStageID
		}
	} else if input.DealStageKey != "" {
		var sID uuid.UUID
		err := p.db.QueryRow(`SELECT id FROM deal_stages WHERE tenant_id=$1 AND key=$2 ORDER BY is_active DESC, sort_order ASC LIMIT 1`, tenantID, input.DealStageKey).Scan(&sID)
		dealStageKey = input.DealStageKey
		if err == nil {
			dealStageID = &sID
		} else {
			dealStageID = nil
		}
	}

	var dskArg sql.NullString
	if dealStageKey != "" {
		dskArg = sql.NullString{String: dealStageKey, Valid: true}
	}
	var dsidArg uuid.NullUUID
	if dealStageID != nil {
		dsidArg = uuid.NullUUID{UUID: *dealStageID, Valid: true}
	}

	var c domain.Contact
	row := p.db.QueryRow(`UPDATE contacts SET display_name=$1, deal_stage_key=$2, deal_stage_id=$3, metadata=$4, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$5 AND id=$6 RETURNING id, tenant_id, normalized_address, provider_address, display_name, is_group, deal_stage_key, deal_stage_id, metadata, created_at, updated_at`, input.DisplayName, dskArg, dsidArg, metadataJSON, tenantID, id)
	if err := scanContact(row, &c); err != nil {
		return domain.Contact{}, err
	}
	return c, nil
}

func (p *PostgresStore) ListContactConversations(tenantID, contactID uuid.UUID, limit, offset int) ([]domain.Conversation, error) {
	rows, err := p.db.Query(`SELECT id, tenant_id, account_id, contact_id, ticket_number, status, bot_state, started_at, last_activity_at, closed_at, handoff_at, closure_reason, assignee, merged_into_id, created_at, updated_at FROM conversations WHERE tenant_id=$1 AND contact_id=$2 ORDER BY last_activity_at DESC LIMIT $3 OFFSET $4`, tenantID, contactID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var conversations []domain.Conversation
	for rows.Next() {
		var c domain.Conversation
		err = scanConversation(rows, &c)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, c)
	}
	return conversations, rows.Err()
}

func (p *PostgresStore) ListContactFieldDefinitions(tenantID uuid.UUID) ([]domain.ContactFieldDefinition, error) {
	rows, err := p.db.Query(`SELECT id, tenant_id, key, label, field_type, options, is_required, sort_order, is_active, created_at, updated_at FROM contact_field_definitions WHERE tenant_id=$1 AND is_active=true ORDER BY sort_order, created_at`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var defs []domain.ContactFieldDefinition
	for rows.Next() {
		var d domain.ContactFieldDefinition
		var options []byte
		err = rows.Scan(&d.ID, &d.TenantID, &d.Key, &d.Label, &d.FieldType, &options, &d.IsRequired, &d.SortOrder, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if len(options) > 0 {
			_ = json.Unmarshal(options, &d.Options)
		}
		defs = append(defs, d)
	}
	return defs, rows.Err()
}

func (p *PostgresStore) GetContactFieldDefinition(tenantID, id uuid.UUID) (domain.ContactFieldDefinition, error) {
	var d domain.ContactFieldDefinition
	var options []byte
	err := p.db.QueryRow(`SELECT id, tenant_id, key, label, field_type, options, is_required, sort_order, is_active, created_at, updated_at FROM contact_field_definitions WHERE tenant_id=$1 AND id=$2`, tenantID, id).Scan(&d.ID, &d.TenantID, &d.Key, &d.Label, &d.FieldType, &options, &d.IsRequired, &d.SortOrder, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return domain.ContactFieldDefinition{}, err
	}
	if len(options) > 0 {
		_ = json.Unmarshal(options, &d.Options)
	}
	return d, nil
}

func (p *PostgresStore) CreateContactFieldDefinition(tenantID uuid.UUID, key, label, fieldType string, options []string, isRequired bool, sortOrder int) (domain.ContactFieldDefinition, error) {
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return domain.ContactFieldDefinition{}, err
	}
	var d domain.ContactFieldDefinition
	err = p.db.QueryRow(`INSERT INTO contact_field_definitions (tenant_id, key, label, field_type, options, is_required, sort_order, is_active) VALUES ($1,$2,$3,$4,$5,$6,$7,true) RETURNING id, tenant_id, key, label, field_type, options, is_required, sort_order, is_active, created_at, updated_at`, tenantID, key, label, fieldType, optionsJSON, isRequired, sortOrder).Scan(&d.ID, &d.TenantID, &d.Key, &d.Label, &d.FieldType, &optionsJSON, &d.IsRequired, &d.SortOrder, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return domain.ContactFieldDefinition{}, err
	}
	_ = json.Unmarshal(optionsJSON, &d.Options)
	return d, nil
}

func (p *PostgresStore) UpdateContactFieldDefinition(tenantID, id uuid.UUID, label, fieldType string, options []string, isRequired bool, sortOrder int, isActive bool) (domain.ContactFieldDefinition, error) {
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return domain.ContactFieldDefinition{}, err
	}
	var d domain.ContactFieldDefinition
	err = p.db.QueryRow(`UPDATE contact_field_definitions SET label=$1, field_type=$2, options=$3, is_required=$4, sort_order=$5, is_active=$6, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$7 AND id=$8 RETURNING id, tenant_id, key, label, field_type, options, is_required, sort_order, is_active, created_at, updated_at`, label, fieldType, optionsJSON, isRequired, sortOrder, isActive, tenantID, id).Scan(&d.ID, &d.TenantID, &d.Key, &d.Label, &d.FieldType, &optionsJSON, &d.IsRequired, &d.SortOrder, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return domain.ContactFieldDefinition{}, err
	}
	_ = json.Unmarshal(optionsJSON, &d.Options)
	return d, nil
}

func (p *PostgresStore) DeleteContactFieldDefinition(tenantID, id uuid.UUID) error {
	_, err := p.db.Exec(`UPDATE contact_field_definitions SET is_active=false, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$1 AND id=$2`, tenantID, id)
	return err
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
	if meta.IsGroup {
		address = meta.Recipient
	} else if meta.Direction == domain.Outgoing {
		address = meta.Recipient
	}
	if !meta.IsGroup && conversation.IsLID(address) {
		log.Printf("skipping LID contact (not a mobile number): %s", address)
		return nil
	}
	actor := domain.ActorContact
	if meta.Direction == domain.Outgoing {
		actor = domain.ActorOperator
	}
	if meta.Actor != "" {
		actor = meta.Actor
	}
	server := "s.whatsapp.net"
	if meta.IsGroup {
		server = "g.us"
	}
	normalized, err := conversation.NormalizeAddressWithServer(address, server)
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
	err = tx.QueryRow(`INSERT INTO contacts (tenant_id, normalized_address, provider_address, is_group) VALUES ($1,$2,$3,$4)
		ON CONFLICT (tenant_id, normalized_address, is_group) DO UPDATE SET provider_address=EXCLUDED.provider_address, updated_at=CURRENT_TIMESTAMP
		RETURNING id`, tenantID, normalized, address, meta.IsGroup).Scan(&contactID)
	if err != nil {
		return err
	}
	// Group names are persisted by UpdateGroup/handleGroupInfo. Do not make
	// message projection depend on the optional group metadata lookup; a failed
	// lookup must not abort this transaction and lose the message.
	var conversationID uuid.UUID
	err = tx.QueryRow(`INSERT INTO conversations (tenant_id, account_id, contact_id, is_group, started_at, last_activity_at) VALUES ($1,$2,$3,$4,$5,$5)
		ON CONFLICT DO NOTHING RETURNING id`, tenantID, accountID, contactID, meta.IsGroup, meta.Timestamp).Scan(&conversationID)
	if err == sql.ErrNoRows {
		err = tx.QueryRow(`SELECT id FROM conversations WHERE tenant_id=$1 AND account_id=$2 AND contact_id=$3 AND status <> 'CLOSED' FOR UPDATE`, tenantID, accountID, contactID).Scan(&conversationID)
	}
	if err != nil {
		return err
	}
	// History sync can replay an existing provider message. Updating sender_address
	// on conflict lets a re-sync backfill this field without creating duplicates.
	_, err = tx.Exec(`INSERT INTO conversation_messages (tenant_id, conversation_id, actor, operator_id, operator_name, provider, provider_message_id, direction, sender_address, reaction_target, content, message_type, media_url, status, provider_timestamp)
		VALUES ($1,$2,$3,$4,$5,'whatsmeow',$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (tenant_id, provider, provider_message_id) DO UPDATE SET sender_address=COALESCE(EXCLUDED.sender_address, conversation_messages.sender_address), reaction_target=COALESCE(EXCLUDED.reaction_target, conversation_messages.reaction_target), status=CASE WHEN conversation_messages.status='READ' OR (conversation_messages.status='DELIVERED' AND EXCLUDED.status='SENT') THEN conversation_messages.status ELSE EXCLUDED.status END, media_url=COALESCE(EXCLUDED.media_url, conversation_messages.media_url), updated_at=CURRENT_TIMESTAMP`,
		tenantID, conversationID, actor, meta.OperatorID, meta.OperatorName, providerID, meta.Direction, nullString(meta.Sender), nullString(meta.ReactionTarget), meta.Content, meta.Type, nullString(meta.MediaURL), meta.Status, meta.Timestamp)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE conversations SET started_at=LEAST(started_at, $1), last_activity_at=GREATEST(last_activity_at, $1), updated_at=CURRENT_TIMESTAMP WHERE id=$2`, meta.Timestamp, conversationID)
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
	if meta.IsGroup || !inbound {
		// Group conversations are keyed by the group recipient, while the
		// sender is the individual participant who authored the message.
		address = meta.Recipient
	}
	server := "s.whatsapp.net"
	if meta.IsGroup {
		server = "g.us"
	} else if conversation.IsLID(address) {
		// ProjectMessage intentionally ignores unresolved LID contacts because
		// they are not stable phone-number identities. Keep the bot-context
		// projection consistent so the same event does not surface a
		// misleading "sql: no rows" error after being skipped.
		return domain.ProjectedMessage{}, nil
	}
	normalized, err := conversation.NormalizeAddressWithServer(address, server)
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
	schemaName := os.Getenv("PG_SCHEMA")
	if schemaName == "" {
		schemaName = "public"
	}
	if _, err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{SchemaName: schemaName})
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
	p.recordMessageEvent(meta)
}

func (p *PostgresStore) recordMessageEvent(meta domain.MessageMetadata) {
	tenantID, terr := p.TenantForHost(meta.HostID)
	if terr != nil || tenantID == uuid.Nil {
		return
	}
	eventType := domain.EventMessageIn
	if meta.Direction == domain.Outgoing {
		eventType = domain.EventMessageOut
	}
	if meta.Status == domain.StatusFailed {
		eventType = domain.EventSendError
	}
	payload := map[string]any{
		"message_id": meta.WhatsappID,
		"type":       meta.Type,
		"status":     meta.Status,
	}
	if _, err := p.RecordInstanceEvent(context.Background(), tenantID, meta.HostID, eventType, string(meta.Direction), payload); err != nil {
		log.Printf("PG Store RecordMessageEvent Error: %v", err)
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

	// The host is the sender of the original message (receipt.Sender on the
	// ingest path is i.HostPhone). Mirror into the event log when it resolves.
	tenantID, terr := p.TenantForHost(receipt.Sender)
	if terr == nil && tenantID != uuid.Nil {
		payload := map[string]any{
			"message_id": receipt.WhatsappID,
			"status":     receipt.Status,
		}
		if _, err := p.RecordInstanceEvent(context.Background(), tenantID, receipt.Sender, domain.EventReceipt, "", payload); err != nil {
			log.Printf("PG Store RecordReceiptEvent Error: %v", err)
		}
	}
}

func (p *PostgresStore) UpdateInstanceStatus(hostID string, status domain.InstanceStatus, isConnected bool) {
	query := `INSERT INTO whatsmeow_instances (host_id, status, is_connected, last_seen, last_connected_at, last_disconnected_at)
	          VALUES ($1, $2, $3, NOW(), CASE WHEN $3 THEN NOW() END, CASE WHEN $3 THEN NULL ELSE NOW() END)
	          ON CONFLICT (host_id) DO UPDATE
	          SET status = EXCLUDED.status,
	              is_connected = EXCLUDED.is_connected,
	              last_seen = NOW(),
	              last_connected_at = CASE WHEN EXCLUDED.is_connected THEN NOW() ELSE whatsmeow_instances.last_connected_at END,
	              last_disconnected_at = CASE WHEN EXCLUDED.is_connected THEN whatsmeow_instances.last_disconnected_at ELSE NOW() END`
	if _, err := p.db.Exec(query, hostID, string(status), isConnected); err != nil {
		log.Printf("PG Store UpdateInstanceStatus Error: %v", err)
		return
	}

	tenantID, terr := p.TenantForHost(hostID)
	if terr != nil || tenantID == uuid.Nil {
		return
	}
	message := statusMessage(status)
	if _, err := p.RecordStatusEvent(context.Background(), tenantID, hostID, status, isConnected, message); err != nil {
		log.Printf("PG Store RecordStatusEvent Error: %v", err)
	}
}

func statusMessage(status domain.InstanceStatus) string {
	switch status {
	case domain.StatusOnline:
		return ""
	case domain.StatusOffline:
		return "Instance disconnected or logged out"
	case domain.StatusError:
		return "Instance connection error"
	default:
		return string(status)
	}
}

// RecordStatusEvent persists a status transition and broadcasts a STATUS event
// (payload and timeline row) to any live subscribers for the host.
func (p *PostgresStore) RecordStatusEvent(ctx context.Context, tenantID uuid.UUID, hostID string, status domain.InstanceStatus, isConnected bool, message string) (domain.StatusEvent, error) {
	var ev domain.StatusEvent
	err := p.db.QueryRowContext(ctx, `INSERT INTO whatsmeow_status_events (tenant_id, host_id, status, is_connected, message)
		VALUES ($1,$2,$3,$4,$5) RETURNING id, tenant_id, host_id, status, is_connected, message, occurred_at`,
		tenantID, hostID, string(status), isConnected, message,
	).Scan(&ev.ID, &ev.TenantID, &ev.HostID, &ev.Status, &ev.IsConnected, &ev.Message, &ev.OccurredAt)
	if err != nil {
		return ev, err
	}
	if _, err := p.RecordInstanceEvent(ctx, tenantID, hostID, domain.EventStatus, "", map[string]any{
		"status":       string(status),
		"is_connected": isConnected,
		"message":      message,
	}); err != nil {
		log.Printf("PG Store RecordStatusInstanceEvent Error: %v", err)
	}
	return ev, nil
}

// RecordInstanceEvent inserts a row into the tailable event log and pushes it to
// any live SSE subscriber for the host. The payload must be JSON-serializable.
func (p *PostgresStore) RecordInstanceEvent(ctx context.Context, tenantID uuid.UUID, hostID, eventType, direction string, payload any) (domain.InstanceLogEvent, error) {
	var raw json.RawMessage
	switch v := payload.(type) {
	case nil:
	case json.RawMessage:
		raw = v
	case string:
		raw = json.RawMessage(v)
	default:
		b, err := json.Marshal(payload)
		if err != nil {
			return domain.InstanceLogEvent{}, err
		}
		raw = b
	}

	var ev domain.InstanceLogEvent
	var directionCol *string
	if direction != "" {
		directionCol = &direction
	}
	err := p.db.QueryRowContext(ctx, `INSERT INTO whatsmeow_instance_events (tenant_id, host_id, event_type, direction, payload)
		VALUES ($1,$2,$3,$4,$5) RETURNING id, host_id, event_type, COALESCE(direction, ''), COALESCE(payload, '{}'::jsonb), occurred_at`,
		tenantID, hostID, eventType, directionCol, raw,
	).Scan(&ev.ID, &ev.HostID, &ev.EventType, &ev.Direction, &ev.Payload, &ev.OccurredAt)
	if err != nil {
		return ev, err
	}
	if p.broadcaster != nil {
		p.broadcaster.Publish(hostID, ev)
	}
	return ev, nil
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
	jobs := []domain.UploadJob{}
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
	jobs := []domain.UploadJob{}
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

	// Sync the group name into any matching group contact record so the inbox
	// can display the group name without an extra lookup.  Scope to is_group=true
	// so a personal contact with the same numeric prefix is never overwritten.
	if group.Name != "" {
		rawGroupID := strings.TrimSuffix(group.GroupID, "@g.us")
		groupNorm := rawGroupID + "@g.us"
		_, _ = p.db.Exec(`UPDATE contacts SET display_name=$1, updated_at=CURRENT_TIMESTAMP 
			WHERE is_group=true AND (normalized_address=$2 OR normalized_address=$3 OR provider_address=$2 OR provider_address=$3)`,
			group.Name, rawGroupID, groupNorm)
	}
}

// AccountHasHost reports whether hostID belongs to the tenant's WhatsApp
// accounts. It guards the monitoring endpoints from cross-tenant enumeration.
func (p *PostgresStore) AccountHasHost(tenantID uuid.UUID, hostID string) (bool, error) {
	var one int
	err := p.db.QueryRow(`SELECT 1 FROM whatsapp_accounts WHERE tenant_id=$1 AND host_id=$2`, tenantID, hostID).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// ListStatusEvents returns status transitions for a host newest-first.
func (p *PostgresStore) ListStatusEvents(ctx context.Context, tenantID uuid.UUID, hostID string, limit int) ([]domain.StatusEvent, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := p.db.QueryContext(ctx, `SELECT id, tenant_id, host_id, status, is_connected, message, occurred_at
		FROM whatsmeow_status_events WHERE tenant_id=$1 AND host_id=$2 ORDER BY occurred_at DESC, id DESC LIMIT $3`,
		tenantID, hostID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]domain.StatusEvent, 0, limit)
	for rows.Next() {
		var ev domain.StatusEvent
		var message sql.NullString
		if err := rows.Scan(&ev.ID, &ev.TenantID, &ev.HostID, &ev.Status, &ev.IsConnected, &message, &ev.OccurredAt); err != nil {
			return nil, err
		}
		if message.Valid {
			ev.Message = message.String
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

// ListInstanceEvents returns rows from the tailable event log, newest-first,
// optionally filtered by event types and id greater than afterID (for SSE
// backfill after a reconnect).
func (p *PostgresStore) ListInstanceEvents(ctx context.Context, tenantID uuid.UUID, hostID string, eventTypes []string, afterID int64, limit int) ([]domain.InstanceLogEvent, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := `SELECT id, host_id, event_type, COALESCE(direction, ''), COALESCE(payload, '{}'::jsonb), occurred_at
		FROM whatsmeow_instance_events WHERE tenant_id=$1 AND host_id=$2`
	args := []any{tenantID, hostID}

	if len(eventTypes) > 0 {
		placeholders := make([]string, len(eventTypes))
		for i := range eventTypes {
			placeholders[i] = fmt.Sprintf("$%d", len(args)+1+i)
		}
		query += ` AND event_type IN (` + strings.Join(placeholders, ",") + `)`
		for _, t := range eventTypes {
			args = append(args, t)
		}
	}
	if afterID > 0 {
		query += fmt.Sprintf(` AND id > $%d`, len(args)+1)
		args = append(args, afterID)
	}
	query += fmt.Sprintf(` ORDER BY occurred_at DESC, id DESC LIMIT $%d`, len(args)+1)
	args = append(args, limit)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]domain.InstanceLogEvent, 0, limit)
	for rows.Next() {
		var ev domain.InstanceLogEvent
		if err := rows.Scan(&ev.ID, &ev.HostID, &ev.EventType, &ev.Direction, &ev.Payload, &ev.OccurredAt); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

// GetInstanceMonitoring merges the persisted instance row (status, last online /
// offline timestamps) for a host, falling back to offline defaults when no row
// exists yet.
func (p *PostgresStore) GetInstanceMonitoring(ctx context.Context, tenantID uuid.UUID, hostID string) (domain.InstanceMonitoring, error) {
	m := domain.InstanceMonitoring{HostID: hostID, Status: domain.StatusOffline}
	var status sql.NullString
	var connected sql.NullBool
	var lastConn, lastDisc, lastSeen sql.NullTime

	err := p.db.QueryRowContext(ctx, `SELECT status, is_connected, last_connected_at, last_disconnected_at, last_seen
		FROM whatsmeow_instances WHERE host_id=$1`, hostID).
		Scan(&status, &connected, &lastConn, &lastDisc, &lastSeen)
	if err != nil && err != sql.ErrNoRows {
		return m, err
	}

	if status.Valid {
		switch domain.InstanceStatus(status.String) {
		case domain.StatusOnline:
			m.Status = domain.StatusOnline
		case domain.StatusError:
			m.Status = domain.StatusError
		default:
			m.Status = domain.StatusOffline
		}
	}
	if connected.Valid {
		m.IsConnected = connected.Bool
	}
	if lastConn.Valid {
		ts := lastConn.Time
		m.LastConnectedAt = &ts
	}
	if lastDisc.Valid {
		ts := lastDisc.Time
		m.LastDisconnectedAt = &ts
	}
	if m.IsConnected && lastConn.Valid {
		m.Uptime = time.Since(lastConn.Time).Truncate(time.Second).String()
	}
	return m, nil
}

// GetMessageMetrics aggregates message volume for a host over a recent window.
// bucketTrunc is the SQL date_trunc unit ("hour" or "day").
func (p *PostgresStore) GetMessageMetrics(ctx context.Context, tenantID uuid.UUID, hostID string, since time.Time, bucketTrunc string) (domain.MessageMetrics, error) {
	if bucketTrunc == "" {
		bucketTrunc = "hour"
	}
	if bucketTrunc != "day" && bucketTrunc != "hour" {
		return domain.MessageMetrics{}, fmt.Errorf("invalid bucket: %s", bucketTrunc)
	}
	m := domain.MessageMetrics{StatusBreakdown: map[string]int{}}

	// Totals + status breakdown (scoped via whatsapp_accounts to the tenant).
	rows, err := p.db.QueryContext(ctx, `SELECT m.direction, m.status, COUNT(*)
		FROM whatsmeow_messages m
		JOIN whatsapp_accounts a ON a.provider='whatsmeow' AND a.host_id = m.host_id
		WHERE a.tenant_id=$1 AND m.host_id=$2 AND m.timestamp >= $3
		GROUP BY m.direction, m.status`, tenantID, hostID, since)
	if err != nil {
		return m, err
	}
	defer rows.Close()
	for rows.Next() {
		var direction, status string
		var count int
		if err := rows.Scan(&direction, &status, &count); err != nil {
			return m, err
		}
		m.StatusBreakdown[status] += count
		switch direction {
		case string(domain.Incoming):
			m.Inbound += count
		case string(domain.Outgoing):
			m.Outbound += count
		}
		if status == string(domain.StatusFailed) {
			m.Failed += count
		}
	}
	if err := rows.Err(); err != nil {
		return m, err
	}

	// Bucketed volume for a bar chart.
	rows2, err := p.db.QueryContext(ctx, fmt.Sprintf(`SELECT date_trunc('%s', m.timestamp) AS bucket,
			SUM(CASE WHEN m.direction='INCOMING' THEN 1 ELSE 0 END) AS inbound,
			SUM(CASE WHEN m.direction='OUTGOING' THEN 1 ELSE 0 END) AS outbound
		FROM whatsmeow_messages m
		JOIN whatsapp_accounts a ON a.provider='whatsmeow' AND a.host_id = m.host_id
		WHERE a.tenant_id=$1 AND m.host_id=$2 AND m.timestamp >= $3
		GROUP BY bucket ORDER BY bucket`, bucketTrunc), tenantID, hostID, since)
	if err != nil {
		return m, err
	}
	defer rows2.Close()
	type rawBucket struct {
		Start             time.Time
		Inbound, Outbound int
	}
	var buckets []rawBucket
	for rows2.Next() {
		var b rawBucket
		if err := rows2.Scan(&b.Start, &b.Inbound, &b.Outbound); err != nil {
			return m, err
		}
		buckets = append(buckets, b)
	}
	if err := rows2.Err(); err != nil {
		return m, err
	}
	for _, b := range buckets {
		m.Buckets = append(m.Buckets, domain.MetricsBucket{Start: b.Start, Inbound: b.Inbound, Outbound: b.Outbound})
	}
	return m, nil
}

// ListQueueDepth returns sampled outbound queue depths for a host, oldest-first.
func (p *PostgresStore) ListQueueDepth(ctx context.Context, tenantID uuid.UUID, hostID string, since time.Time, limit int) ([]domain.QueueDepthSample, error) {
	if limit <= 0 || limit > 500 {
		limit = 180
	}
	rows, err := p.db.QueryContext(ctx, `SELECT id, occurred_at, COALESCE(payload->>'queue_size','0')::int
		FROM whatsmeow_instance_events
		WHERE tenant_id=$1 AND host_id=$2 AND event_type=$3 AND occurred_at >= $4
		ORDER BY occurred_at ASC LIMIT $5`, tenantID, hostID, domain.EventQueueDepth, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	samples := make([]domain.QueueDepthSample, 0, limit)
	for rows.Next() {
		var s domain.QueueDepthSample
		if err := rows.Scan(&s.ID, &s.Timestamp, &s.QueueSize); err != nil {
			return nil, err
		}
		samples = append(samples, s)
	}
	return samples, rows.Err()
}
