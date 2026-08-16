package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
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
	SetupStep       int            `json:"setup_step"`
	IsSetupComplete bool           `json:"is_setup_complete"`
	OrgDetails      map[string]any `json:"org_details,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}
type WhatsAppAccount struct {
	ID, TenantID                  uuid.UUID
	HostID, Provider, DisplayName string
	CreatedAt, UpdatedAt          time.Time
}
type Contact struct {
	ID, TenantID                                    uuid.UUID
	NormalizedAddress, ProviderAddress, DisplayName string
	Metadata                                        map[string]any
	CreatedAt, UpdatedAt                            time.Time
}
type Conversation struct {
	ID, TenantID, AccountID, ContactID uuid.UUID
	TicketNumber                       int64
	Status                             ConversationStatus
	BotState                           string
	StartedAt, LastActivityAt          time.Time
	ClosedAt, HandoffAt                *time.Time
	ClosureReason                      string
	Assignee                           string
	MergedIntoID                       *uuid.UUID
	CreatedAt, UpdatedAt               time.Time
}
type ConversationMessage struct {
	ID, TenantID, ConversationID                                                   uuid.UUID
	Actor                                                                          Actor
	Provider, ProviderMessageID, Direction, Content, MessageType, MediaURL, Status string
	ProviderTimestamp                                                              *time.Time
	IsInternal                                                                     bool
	CreatedAt, UpdatedAt                                                           time.Time
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
	ID        uuid.UUID   `json:"id"`
	TenantID  uuid.UUID   `json:"tenant_id"`
	Version   int         `json:"version"`
	Rules     []BotRule   `json:"rules"`
	IsActive  bool        `json:"is_active"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
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
	SetupStep       int            `json:"setup_step"`
	IsSetupComplete bool           `json:"is_setup_complete"`
	OrgDetails      map[string]any `json:"org_details,omitempty"`
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

// ContactActivityInput is the tenant-scoped payload for creating a follow-up
// activity directly against a contact (i.e. not tied to a single conversation).
type ContactActivityInput struct {
	Type       string
	Summary    string
	NextAction string
	Priority   string
	DueAt      *time.Time
}

type ContactUpsert struct {
	TenantID                     uuid.UUID
	ProviderAddress, DisplayName string
	Metadata                     map[string]any
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
	ListContacts(tenantID uuid.UUID, limit, offset int) ([]Contact, error)
	GetContact(tenantID, id uuid.UUID) (Contact, error)
	ListConversations(tenantID uuid.UUID, status string, limit, offset int) ([]Conversation, error)
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
	Recipient      string      `json:"recipient"`
	Message        string      `json:"message"`
	IsGroup        bool        `json:"is_group"`
	Type           MessageType `json:"type"`
	MediaPath      string      `json:"media_path,omitempty"`
	MediaKey       string      `json:"media_key,omitempty"`
	ReactionTarget string      `json:"reaction_target,omitempty"`
	Actor          Actor       `json:"-"`
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
