package domain

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Operator role constants
const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleViewer   = "viewer"
)

// NormalizeRole normalizes a role string to canonical lowercase format.
func NormalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case RoleAdmin:
		return RoleAdmin
	case RoleViewer:
		return RoleViewer
	case RoleOperator:
		return RoleOperator
	default:
		return RoleOperator
	}
}

var (
	ErrPipelineNotFound                = errors.New("pipeline not found")
	ErrPipelineNameExists              = errors.New("pipeline with this name already exists")
	ErrCannotDeleteDefaultPipeline     = errors.New("cannot delete default pipeline")
	ErrPipelineContainsStages          = errors.New("pipeline contains stages; reassign or delete them first")
	ErrStageNotFound                   = errors.New("deal stage not found")
	ErrStageAssignedToContacts         = errors.New("stage is currently assigned to contacts")
	ErrDefaultPipelineCannotBeInactive = errors.New("default pipeline cannot be inactive")
)

type ConversationStatus string

const (
	ConversationOpen      ConversationStatus = "OPEN"
	ConversationBotActive ConversationStatus = "BOT_ACTIVE"
	ConversationWaiting   ConversationStatus = "WAITING"
	ConversationHandedOff ConversationStatus = "HANDED_OFF"
	ConversationClosed    ConversationStatus = "CLOSED"
)

type Actor string

const (
	ActorContact  Actor = "CONTACT"
	ActorBot      Actor = "BOT"
	ActorOperator Actor = "OPERATOR"
	ActorSystem   Actor = "SYSTEM"
)

type ActivityStatus string

const (
	ActivityPending      ActivityStatus = "PENDING"
	ActivityAcknowledged ActivityStatus = "ACKNOWLEDGED"
	ActivityDismissed    ActivityStatus = "DISMISSED"
)

type Tenant struct {
	ID              uuid.UUID      `json:"id"`
	Name            string         `json:"name"`
	Slug            string         `json:"slug"`
	SetupStep       int            `json:"setup_step"`
	IsSetupComplete bool           `json:"is_setup_complete"`
	OrgDetails      map[string]any `json:"org_details,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}
type WhatsAppAccount struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	HostID      string    `json:"host_id"`
	Provider    string    `json:"provider"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type Contact struct {
	ID, TenantID                                    uuid.UUID
	NormalizedAddress, ProviderAddress, DisplayName string
	IsGroup                                         bool
	DealStageKey                                    string
	DealStageID                                     *uuid.UUID
	Metadata                                        map[string]any
	CreatedAt, UpdatedAt                            time.Time
}
type Conversation struct {
	ID             uuid.UUID          `json:"id"`
	TenantID       uuid.UUID          `json:"tenant_id"`
	AccountID      uuid.UUID          `json:"account_id"`
	ContactID      uuid.UUID          `json:"contact_id"`
	TicketNumber   int64              `json:"ticket_number"`
	Status         ConversationStatus `json:"status"`
	BotState       string             `json:"bot_state"`
	StartedAt      time.Time          `json:"started_at"`
	LastActivityAt time.Time          `json:"last_activity_at"`
	ClosedAt       *time.Time         `json:"closed_at"`
	HandoffAt      *time.Time         `json:"handoff_at"`
	ClosureReason  string             `json:"closure_reason"`
	Assignee       string             `json:"assignee"`
	MergedIntoID   *uuid.UUID         `json:"merged_into_id"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

// ConversationSummary is the enriched inbox list entry: a Conversation plus
// the contact's display identity and the most recent timeline message, so a
// conversation-centric UI can render a user list without N+1 lookups.
type ConversationSummary struct {
	Conversation
	ContactName   string `json:"contact_name"`
	ContactNumber string `json:"contact_number"`
	IsGroup       bool   `json:"is_group"`
	LastMessage   string `json:"last_message_preview"`
	LastActor     Actor  `json:"last_message_actor"`
}

type ConversationMessage struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	ConversationID    uuid.UUID  `json:"conversation_id"`
	Actor             Actor      `json:"actor"`
	OperatorID        *uuid.UUID `json:"operator_id,omitempty"`
	OperatorName      string     `json:"operator_name,omitempty"`
	Provider          string     `json:"provider"`
	ProviderMessageID string     `json:"provider_message_id"`
	Direction         string     `json:"direction"`
	// SenderAddress is the participant who authored a group message.
	// It is nullable in storage for historical messages.
	SenderAddress     string     `json:"sender_address,omitempty"`
	ReactionTarget    string     `json:"reaction_target,omitempty"`
	Content           string     `json:"content"`
	MessageType       string     `json:"message_type"`
	MediaURL          string     `json:"media_url"`
	Status            string     `json:"status"`
	ProviderTimestamp *time.Time `json:"provider_timestamp"`
	IsInternal        bool       `json:"is_internal"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}
type BotRule struct {
	Name     string `json:"name"`
	Pattern  string `json:"pattern"`
	Match    string `json:"match"` // "CONTAINS", "EXACT", "PREFIX"
	Response string `json:"response"`
	Terminal bool   `json:"terminal"`
	Handoff  bool   `json:"handoff"`
	Enabled  bool   `json:"enabled"`
}
type BotRuleSet struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Version   int       `json:"version"`
	Rules     []BotRule `json:"rules"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type OperatorAuditLog struct {
	ID             uuid.UUID      `json:"id"`
	TenantID       uuid.UUID      `json:"tenant_id"`
	OperatorID     string         `json:"operator_id"`
	Action         string         `json:"action"` // 'ASSIGN', 'CLOSE', 'REOPEN', 'HANDOFF', 'MERGE', 'SPLIT', 'UPDATE_RULES'
	ConversationID *uuid.UUID     `json:"conversation_id,omitempty"`
	Details        map[string]any `json:"details"`
	CreatedAt      time.Time      `json:"created_at"`
}

// Operator is a dashboard user, scoped to a tenant, that manages conversations.
type Operator struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	Email             string     `json:"email,omitempty"`
	WhatsappNumber    string     `json:"whatsapp_number,omitempty"`
	Name              string     `json:"name"`
	Role              string     `json:"role"`
	IsActive          bool       `json:"is_active"`
	TotpVerifiedAt    *time.Time `json:"totp_verified_at,omitempty"`
	TotpSetupRequired bool       `json:"totp_setup_required"`
	EmailVerifiedAt   *time.Time `json:"email_verified_at,omitempty"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type Invitation struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	Role           string     `json:"role"`
	Channel        string     `json:"channel"`
	Recipient      string     `json:"recipient"`
	WhatsappNumber string     `json:"whatsapp_number,omitempty"`
	Email          string     `json:"email,omitempty"`
	TokenHash      string     `json:"-"`
	Status         string     `json:"status"`
	ExpiresAt      time.Time  `json:"expires_at"`
	AcceptedAt     *time.Time `json:"accepted_at,omitempty"`
	CreatedBy      *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type BackupCode struct {
	ID         uuid.UUID  `json:"id"`
	OperatorID uuid.UUID  `json:"operator_id"`
	CodeHash   string     `json:"-"`
	UsedAt     *time.Time `json:"used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type RecoveryToken struct {
	ID         uuid.UUID  `json:"id"`
	OperatorID uuid.UUID  `json:"operator_id"`
	TokenHash  string     `json:"-"`
	ExpiresAt  time.Time  `json:"expires_at"`
	UsedAt     *time.Time `json:"used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type EmailVerificationToken struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	OperatorID uuid.UUID  `json:"operator_id"`
	TokenHash  string     `json:"-"`
	ExpiresAt  time.Time  `json:"expires_at"`
	UsedAt     *time.Time `json:"used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type RecoveryAuditLog struct {
	ID         uuid.UUID      `json:"id"`
	TenantID   uuid.UUID      `json:"tenant_id"`
	OperatorID *uuid.UUID     `json:"operator_id,omitempty"`
	Action     string         `json:"action"`
	IPAddress  string         `json:"ip_address,omitempty"`
	UserAgent  string         `json:"user_agent,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

type TenantSetupStatus struct {
	TenantName         string         `json:"tenant_name,omitempty"`
	TenantSlug         string         `json:"tenant_slug,omitempty"`
	SetupStep          int            `json:"setup_step"`
	IsSetupComplete    bool           `json:"is_setup_complete"`
	OrgDetails         map[string]any `json:"org_details,omitempty"`
	HasWhatsappAccount bool           `json:"has_whatsapp_account"`
	TeamCount          int            `json:"team_count"`
	HasBotRules        bool           `json:"has_bot_rules"`
}

// Session represents an authenticated operator dashboard session.
type Session struct {
	ID         uuid.UUID `json:"id"`
	OperatorID uuid.UUID `json:"operator_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}
type BotSession struct {
	ID, TenantID, ConversationID uuid.UUID
	RuleVersion                  string
	State                        map[string]any
	CreatedAt, UpdatedAt         time.Time
}
type Activity struct {
	ID             uuid.UUID      `json:"id"`
	TenantID       uuid.UUID      `json:"tenant_id"`
	ConversationID uuid.UUID      `json:"conversation_id"`
	ContactID      uuid.UUID      `json:"contact_id"`
	Type           string         `json:"type"`
	Summary        string         `json:"summary"`
	NextAction     string         `json:"next_action"`
	Priority       string         `json:"priority"`
	Status         ActivityStatus `json:"status"`
	DueAt          *time.Time     `json:"due_at"`
	AcknowledgedBy string         `json:"acknowledged_by"`
	AcknowledgedAt *time.Time     `json:"acknowledged_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// Pipeline represents a tenant-scoped CRM pipeline containing stages.
type Pipeline struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsDefault   bool      `json:"is_default"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DealStage represents a tenant-configured pipeline stage.
type DealStage struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	PipelineID *uuid.UUID `json:"pipeline_id,omitempty"`
	Key        string     `json:"key"`
	Label      string     `json:"label"`
	Color      string     `json:"color"`
	Icon       string     `json:"icon"`
	SortOrder  int        `json:"sort_order"`
	IsActive   bool       `json:"is_active"`
	IsWon      bool       `json:"is_won"`
	IsLost     bool       `json:"is_lost"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// DealStageTransition records a contact moving from one stage to another.
type DealStageTransition struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	ContactID uuid.UUID  `json:"contact_id"`
	FromStage *string    `json:"from_stage,omitempty"`
	ToStage   string     `json:"to_stage"`
	Note      string     `json:"note,omitempty"`
	MovedBy   *uuid.UUID `json:"moved_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// DealStageInput is the tenant-scoped payload for creating/updating a deal stage.
type DealStageInput struct {
	PipelineID *uuid.UUID `json:"pipeline_id,omitempty"`
	Key        string     `json:"key"`
	Label      string     `json:"label"`
	Color      string     `json:"color"`
	Icon       string     `json:"icon"`
	SortOrder  int        `json:"sort_order"`
	IsWon      bool       `json:"is_won"`
	IsLost     bool       `json:"is_lost"`
}

// MoveContactStageInput is the payload for moving a contact to a deal stage.
type MoveContactStageInput struct {
	StageKey    string     `json:"stage_key"`
	StageID     *uuid.UUID `json:"stage_id,omitempty"`
	DealStageID *uuid.UUID `json:"deal_stage_id,omitempty"`
	Note        string     `json:"note"`
}

// ContactActivityInput is the tenant-scoped payload for creating a follow-up
// activity directly against a contact (i.e. not tied to a single conversation).
type ContactActivityInput struct {
	Type           string
	Summary        string
	NextAction     string
	Priority       string
	DueAt          *time.Time
	ConversationID uuid.UUID
}

type ContactUpsert struct {
	TenantID                     uuid.UUID
	ProviderAddress, DisplayName string
	IsGroup                      bool
	Metadata                     map[string]any
}

// ContactUpdateInput is the tenant-scoped payload for editing a contact's
// display-facing details and custom field values.
type ContactUpdateInput struct {
	DisplayName    string
	Email          string
	Tags           []string
	CustomValues   map[string]any
	DealStageKey   string
	DealStageID    *uuid.UUID
	ClearDealStage bool
}

// ContactFieldDefinition describes a tenant-defined custom field for contacts.
type ContactFieldDefinition struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	Key        string    `json:"key"`
	Label      string    `json:"label"`
	FieldType  string    `json:"field_type"`
	Options    []string  `json:"options,omitempty"`
	IsRequired bool      `json:"is_required"`
	SortOrder  int       `json:"sort_order"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ConversationKey struct{ TenantID, AccountID, ContactID uuid.UUID }

// ContactRepository and ConversationRepository are application persistence boundaries.
// Implementations must scope every operation by TenantID.
type ContactRepository interface {
	UpsertContact(tenantID uuid.UUID, input ContactUpsert) (Contact, error)
	GetContact(tenantID, id uuid.UUID) (Contact, error)
}
type ConversationRepository interface {
	FindOrCreateConversation(key ConversationKey, now time.Time) (Conversation, error)
	GetConversation(tenantID, id uuid.UUID) (Conversation, error)
}
type ApplicationRepository interface {
	ContactRepository
	ConversationRepository
}

// AccountRegistrar encapsulates the registration of a WhatsApp host account for a tenant.
type AccountRegistrar interface {
	RegisterAccount(tenantID uuid.UUID, hostID, displayName, provider string) (WhatsAppAccount, error)
}

// PlatformRepository is the tenant-scoped read/write boundary used by the HTTP API.
// Implementations must never infer a tenant from a resource identifier alone.
type PlatformRepository interface {
	AccountRegistrar
	AuthenticateAPIKey(key string) (uuid.UUID, error)
	ListAccounts(tenantID uuid.UUID) ([]WhatsAppAccount, error)
	GetAccount(tenantID, id uuid.UUID) (WhatsAppAccount, error)
	ListContacts(tenantID uuid.UUID, limit, offset int, search string) ([]Contact, int, error)
	GetContact(tenantID, id uuid.UUID) (Contact, error)
	UpdateContact(tenantID, id uuid.UUID, input ContactUpdateInput) (Contact, error)
	ListContactConversations(tenantID, contactID uuid.UUID, limit, offset int) ([]Conversation, error)
	ListContactFieldDefinitions(tenantID uuid.UUID) ([]ContactFieldDefinition, error)
	GetContactFieldDefinition(tenantID, id uuid.UUID) (ContactFieldDefinition, error)
	CreateContactFieldDefinition(tenantID uuid.UUID, key, label, fieldType string, options []string, isRequired bool, sortOrder int) (ContactFieldDefinition, error)
	UpdateContactFieldDefinition(tenantID, id uuid.UUID, label, fieldType string, options []string, isRequired bool, sortOrder int, isActive bool) (ContactFieldDefinition, error)
	DeleteContactFieldDefinition(tenantID, id uuid.UUID) error

	ListConversations(tenantID uuid.UUID, status string, limit, offset int) ([]Conversation, error)
	GetConversation(tenantID, id uuid.UUID) (Conversation, error)
	ListConversationSummaries(tenantID uuid.UUID, status string, limit, offset int) ([]ConversationSummary, error)
	GetConversationTimeline(tenantID, conversationID uuid.UUID, limit, offset int) ([]ConversationMessage, error)
	ListActivities(tenantID uuid.UUID, status string, limit, offset int) ([]Activity, error)
	AcknowledgeActivity(tenantID, activityID uuid.UUID, actor string, at time.Time) (Activity, error)
	ListContactActivities(tenantID, contactID uuid.UUID, limit, offset int) ([]Activity, error)
	CreateContactActivity(tenantID, contactID uuid.UUID, input ContactActivityInput) (Activity, error)

	// Bot rules versioning and management
	GetActiveBotRuleSet(tenantID uuid.UUID) (BotRuleSet, error)
	SaveBotRuleSet(tenantID uuid.UUID, rules []BotRule) (BotRuleSet, error)
	ListBotRuleSets(tenantID uuid.UUID) ([]BotRuleSet, error)
	ActivateBotRuleSetVersion(tenantID uuid.UUID, version int) (BotRuleSet, error)

	// Operator ticket actions
	AssignConversation(tenantID, conversationID uuid.UUID, assignee string, operatorID string) (Conversation, error)
	HandoffConversation(tenantID, conversationID uuid.UUID, operatorID string) (Conversation, error)
	CloseConversationWithReason(tenantID, conversationID uuid.UUID, reason string, operatorID string) (Conversation, error)
	ReopenConversation(tenantID, conversationID uuid.UUID, operatorID string) (Conversation, error)
	AddInternalNote(tenantID, conversationID uuid.UUID, actor Actor, operatorID, content string) (ConversationMessage, error)
	MergeConversations(tenantID, targetID, sourceID uuid.UUID, operatorID string) (Conversation, error)
	SplitConversation(tenantID, sourceID uuid.UUID, messageIDs []uuid.UUID, operatorID string) (Conversation, error)
	ListOperatorAuditLogs(tenantID uuid.UUID, limit, offset int) ([]OperatorAuditLog, error)

	// Media objects
	RecordMediaObject(ctx context.Context, tenantID uuid.UUID, objectKey, mimeType string, size int64) error
	GetMediaObject(ctx context.Context, tenantID uuid.UUID, objectKey string) (MediaObject, error)

	// Pipelines
	ListPipelines(tenantID uuid.UUID, isActive *bool) ([]Pipeline, error)
	GetPipeline(tenantID, id uuid.UUID) (Pipeline, error)
	GetDefaultPipeline(tenantID uuid.UUID) (Pipeline, error)
	CreatePipeline(tenantID uuid.UUID, name, description string, isDefault bool) (Pipeline, error)
	UpdatePipeline(tenantID, id uuid.UUID, name, description *string, isDefault, isActive *bool) (Pipeline, error)
	DeletePipeline(tenantID, id uuid.UUID) error

	// Deal stages
	ListDealStages(tenantID uuid.UUID, pipelineID *uuid.UUID, isActive *bool) ([]DealStage, error)
	GetDealStage(tenantID, id uuid.UUID) (DealStage, error)
	CreateDealStage(tenantID uuid.UUID, pipelineID *uuid.UUID, key, label, color, icon string, sortOrder int, isWon, isLost bool) (DealStage, error)
	UpdateDealStage(tenantID, id uuid.UUID, pipelineID *uuid.UUID, label, color, icon *string, sortOrder *int, isActive, isWon, isLost *bool) (DealStage, error)
	DeleteDealStage(tenantID, id uuid.UUID) error
	MoveContactToStage(tenantID, contactID uuid.UUID, stageKey string, stageID *uuid.UUID, note string, operatorID uuid.UUID) (DealStageTransition, error)
	ListDealStageHistory(tenantID, contactID uuid.UUID, limit, offset int) ([]DealStageTransition, error)

	// Subscriptions and Quotas
	GetSubscription(tenantID uuid.UUID) (Subscription, error)
	GetQuota(tenantID uuid.UUID) (Quota, error)
	IncrementQuota(tenantID uuid.UUID) error
}

type Subscription struct {
	TenantID           uuid.UUID
	PlanType           string
	Status             string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type Quota struct {
	TenantID     uuid.UUID
	MonthlyLimit int
	CurrentUsage int
	ResetAt      time.Time
	UpdatedAt    time.Time
}

type Direction string

const (
	Incoming Direction = "INCOMING"
	Outgoing Direction = "OUTGOING"
)

type MessageType string

const (
	Text     MessageType = "TEXT"
	Image    MessageType = "IMAGE"
	Video    MessageType = "VIDEO"
	Audio    MessageType = "AUDIO"
	File     MessageType = "FILE"
	Reaction MessageType = "REACTION"
)

type MessageStatus string

const (
	StatusPending   MessageStatus = "PENDING"
	StatusSent      MessageStatus = "SENT"
	StatusDelivered MessageStatus = "DELIVERED"
	StatusRead      MessageStatus = "READ"
	StatusFailed    MessageStatus = "FAILED"
)

type MessageMetadata struct {
	WhatsappID     string        `json:"whatsapp_id"`
	HostID         string        `json:"host_id"`
	Sender         string        `json:"sender"`
	Recipient      string        `json:"recipient"`
	Content        string        `json:"content"`
	IsGroup        bool          `json:"is_group"`
	Direction      Direction     `json:"direction"`
	Type           MessageType   `json:"type"`
	Status         MessageStatus `json:"status"`
	Actor          Actor         `json:"actor,omitempty"`
	OperatorID     *uuid.UUID    `json:"operator_id,omitempty"`
	OperatorName   string        `json:"operator_name,omitempty"`
	MediaURL       string        `json:"media_url,omitempty"`
	ReactionTarget string        `json:"reaction_target,omitempty"`
	Timestamp      time.Time     `json:"timestamp"`
}

type Receipt struct {
	WhatsappID string        `json:"whatsapp_id"`
	Sender     string        `json:"sender"`
	Recipient  string        `json:"recipient"`
	Status     MessageStatus `json:"status"`
	Timestamp  time.Time     `json:"timestamp"`
}

type MessageRequest struct {
	ID             string      `json:"id,omitempty"`
	Recipient      string      `json:"recipient"`
	Message        string      `json:"message"`
	IsGroup        bool        `json:"is_group"`
	Type           MessageType `json:"type"`
	MediaPath      string      `json:"media_path,omitempty"`
	MediaKey       string      `json:"media_key,omitempty"`
	ReactionTarget string      `json:"reaction_target,omitempty"`
	Actor          Actor       `json:"-"`
	OperatorID     *uuid.UUID  `json:"operator_id,omitempty"`
	OperatorName   string      `json:"operator_name,omitempty"`
}

type MediaObject struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	ObjectKey string    `json:"object_key"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

type InstanceStatus string

const (
	StatusOnline  InstanceStatus = "ONLINE"
	StatusOffline InstanceStatus = "OFFLINE"
	StatusError   InstanceStatus = "ERROR"
)

type InstanceEvent struct {
	HostID    string         `json:"host_id"`
	Status    InstanceStatus `json:"status"`
	Message   string         `json:"message"`
	Timestamp time.Time      `json:"timestamp"`
}

type InstanceInfo struct {
	HostPhone   string         `json:"host_phone"`
	Status      InstanceStatus `json:"status"`
	IsConnected bool           `json:"is_connected"`
	QueueSize   int            `json:"queue_size"`
}

type GroupInfo struct {
	GroupID          string    `json:"group_id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	OwnerJID         string    `json:"owner_jid"`
	Participants     []string  `json:"participants"`
	Hosts            []string  `json:"hosts"`
	ParticipantCount int       `json:"participant_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Dispatcher interface {
	DispatchMessage(meta MessageMetadata)
	DispatchReceipt(receipt Receipt)
	UpdateInstanceStatus(hostID string, status InstanceStatus, isConnected bool)
	UpdateGroup(group GroupInfo)
}

// Instance event types recorded in the monitoring tailable event log.
const (
	EventMessageIn      = "MESSAGE_IN"
	EventMessageOut     = "MESSAGE_OUT"
	EventReceipt        = "RECEIPT"
	EventStatus         = "STATUS"
	EventQueueDepth     = "QUEUE_DEPTH"
	EventSendError      = "SEND_ERROR"
	EventProjectionFail = "PROJECTION_FAILED"
	EventMediaError     = "MEDIA_ERROR"
	EventUploadFail     = "UPLOAD_FAILED"
	EventLoggedOut      = "LOGGED_OUT"
)

// StatusEvent is a single status transition (ONLINE → OFFLINE → ERROR) with a
// human-readable message when one is available.
type StatusEvent struct {
	ID          int64          `json:"id"`
	TenantID    uuid.UUID      `json:"tenant_id"`
	HostID      string         `json:"host_id"`
	Status      InstanceStatus `json:"status"`
	IsConnected bool           `json:"is_connected"`
	Message     string         `json:"message"`
	OccurredAt  time.Time      `json:"occurred_at"`
}

// InstanceLogEvent is a single row in the tailable per-account event log. The
// payload holds sanitized metadata (message id, type, status) and never message
// content.
type InstanceLogEvent struct {
	ID         int64           `json:"id"`
	HostID     string          `json:"host_id"`
	EventType  string          `json:"event_type"`
	Direction  string          `json:"direction"`
	Payload    json.RawMessage `json:"payload"`
	OccurredAt time.Time       `json:"occurred_at"`
}

// MetricsBucket aggregates message volume over one bucket interval.
type MetricsBucket struct {
	Start    time.Time `json:"start"`
	Inbound  int       `json:"inbound"`
	Outbound int       `json:"outbound"`
}

// MessageMetrics is the per-account message volume surface.
type MessageMetrics struct {
	Inbound         int             `json:"inbound"`
	Outbound        int             `json:"outbound"`
	Failed          int             `json:"failed"`
	StatusBreakdown map[string]int  `json:"status_breakdown"`
	Buckets         []MetricsBucket `json:"buckets"`
}

// QueueDepthSample tracks the outbound queue depth over time.
type QueueDepthSample struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	QueueSize int       `json:"queue_size"`
}

// InstanceMonitoring is the current state + last online/offline timestamps for
// an account, merged from the instances row and (when live) the running client.
type InstanceMonitoring struct {
	HostID             string         `json:"host_id"`
	Status             InstanceStatus `json:"status"`
	IsConnected        bool           `json:"is_connected"`
	QueueSize          int            `json:"queue_size"`
	Uptime             string         `json:"uptime"`
	LastConnectedAt    *time.Time     `json:"last_connected_at"`
	LastDisconnectedAt *time.Time     `json:"last_disconnected_at"`
}

// MonitoringStore exposes the tailable event log, status history, and metrics
// aggregates used by the dashboard monitoring surface. All queries are
// tenant-scoped.
type MonitoringStore interface {
	RecordStatusEvent(ctx context.Context, tenantID uuid.UUID, hostID string, status InstanceStatus, isConnected bool, message string) (StatusEvent, error)
	RecordInstanceEvent(ctx context.Context, tenantID uuid.UUID, hostID, eventType, direction string, payload any) (InstanceLogEvent, error)
	ListStatusEvents(ctx context.Context, tenantID uuid.UUID, hostID string, limit int) ([]StatusEvent, error)
	ListInstanceEvents(ctx context.Context, tenantID uuid.UUID, hostID string, eventTypes []string, afterID int64, limit int) ([]InstanceLogEvent, error)
	GetInstanceMonitoring(ctx context.Context, tenantID uuid.UUID, hostID string) (InstanceMonitoring, error)
	GetMessageMetrics(ctx context.Context, tenantID uuid.UUID, hostID string, since time.Time, bucketTrunc string) (MessageMetrics, error)
	ListQueueDepth(ctx context.Context, tenantID uuid.UUID, hostID string, since time.Time, limit int) ([]QueueDepthSample, error)
	AccountHasHost(tenantID uuid.UUID, hostID string) (bool, error)
}

// UploadJobStatus is the lifecycle of a durable outgoing-media S3 upload.
// A job moves PENDING → PROCESSING → COMPLETED, or PROCESSING → PENDING
// (retryable error) or PROCESSING → FAILED (retry limit / permanent error).
type UploadJobStatus string

const (
	UploadPending    UploadJobStatus = "PENDING"
	UploadProcessing UploadJobStatus = "PROCESSING"
	UploadCompleted  UploadJobStatus = "COMPLETED"
	UploadFailed     UploadJobStatus = "FAILED"
)

// UploadJob is a durable, retryable S3 media upload. ObjectKey is generated once
// and reused for every attempt so retries converge to a single archive object.
// MediaPath is a durable payload reference (the local source file).
type UploadJob struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	MessageID     string
	HostID        string
	ObjectKey     string
	MimeType      string
	MediaPath     string
	Status        UploadJobStatus
	AttemptCount  int
	NextAttemptAt time.Time
	LastError     string
	MediaURL      string
	LeaseUntil    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// UploadJobRepository persists and transitions outgoing-media upload jobs. Jobs
// are tenant-scoped on creation; claiming is performed by the application worker
// across tenants and is atomic to prevent concurrent processing.
type UploadJobRepository interface {
	CreateUploadJob(ctx context.Context, tenantID uuid.UUID, job UploadJob) (UploadJob, error)
	ClaimDueUploadJobs(ctx context.Context, now time.Time, limit int) ([]UploadJob, error)
	MarkUploadCompleted(ctx context.Context, id uuid.UUID, mediaURL string) error
	MarkUploadRetryable(ctx context.Context, id uuid.UUID, lastError string, nextAttemptAt time.Time, attemptCount int) error
	MarkUploadFailed(ctx context.Context, id uuid.UUID, lastError string, attemptCount int) error
}

// ApplicationProjector receives transport events for projection into the
// tenant-scoped conversation model. Implementations must be idempotent.
type ApplicationProjector interface {
	ProjectMessage(meta MessageMetadata) error
	ProjectReceipt(receipt Receipt) error
}

// ProjectedMessage is the application context produced for a transport event.
// New is true only when the event should be offered to automation.
type ProjectedMessage struct {
	TenantID, ConversationID  uuid.UUID
	Host, Recipient, Text     string
	At                        time.Time
	Inbound, New, BotEligible bool
}

type ContextProjector interface {
	ProjectMessageContext(meta MessageMetadata) (ProjectedMessage, error)
}

type ActivityRepository interface {
	CloseConversation(tenantID, conversationID uuid.UUID, status ConversationStatus, reason string, at time.Time) (Activity, error)
}
