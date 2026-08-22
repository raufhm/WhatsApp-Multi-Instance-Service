// Domain model types mirroring the Go backend (domain/models.go)

export type ConversationStatus =
  | 'OPEN'
  | 'BOT_ACTIVE'
  | 'WAITING'
  | 'HANDED_OFF'
  | 'CLOSED'

export type Actor = 'CONTACT' | 'BOT' | 'OPERATOR' | 'SYSTEM'
export type Direction = 'INCOMING' | 'OUTGOING'
export type MessageType = 'TEXT' | 'IMAGE' | 'VIDEO' | 'AUDIO' | 'DOCUMENT' | 'REACTION'

export interface Conversation {
  id: string
  tenant_id: string
  account_id: string
  contact_id: string
  contact_number?: string
  ticket_number: number
  status: ConversationStatus
  bot_state: string
  started_at: string
  last_activity_at: string
  closed_at: string | null
  handoff_at: string | null
  closure_reason: string | null
  assignee: string | null
  merged_into_id: string | null
  created_at: string
  updated_at: string
}

export interface ConversationSummary extends Conversation {
  contact_name: string
  contact_number: string
  is_group: boolean
  last_message_preview: string
  last_message_actor: Actor | ''
}

export interface ConversationMessage {
  id: string
  tenant_id: string
  conversation_id: string
  actor: Actor
  operator_id?: string | null
  operator_name?: string | null
  provider: string
  provider_message_id: string
  reaction_target?: string | null
  direction: Direction
  sender_address?: string | null
  content: string
  message_type: MessageType
  media_url: string
  status: string
  provider_timestamp: string
  is_internal: boolean
  created_at: string
  updated_at: string
}

export interface Contact {
  id: string
  tenant_id: string
  name: string
  number: string
  email?: string
  tags?: string[]
  is_group?: boolean
  deal_stage?: DealStageDTO | null
  custom_values?: Record<string, unknown>
  created_at: string
  updated_at: string
}

// ── Deal Stage / CRM Pipeline ──────────────────────────────────────────

export interface Pipeline {
  id: string
  tenant_id?: string
  key?: string
  name: string
  description?: string
  is_default: boolean
  is_active?: boolean
  stages?: DealStage[]
  created_at?: string
  updated_at?: string
}

export interface DealStageDTO {
  id?: string
  key: string
  label: string
  color: string
  icon: string
  is_won: boolean
  is_lost: boolean
}

export interface DealStage {
  id: string
  tenant_id: string
  pipeline_id?: string | null
  key: string
  label: string
  color: string
  icon: string
  sort_order: number
  is_active: boolean
  is_won: boolean
  is_lost: boolean
  created_at: string
  updated_at: string
}

export interface DealStageTransition {
  id: string
  tenant_id: string
  contact_id: string
  from_stage: string | null
  to_stage: string
  note: string
  moved_by: string | null
  created_at: string
}

export type DealStageKey =
  | 'NEW_LEAD'
  | 'APPOINTMENT_SCHEDULED'
  | 'HOT_LEAD'
  | 'COLD_LEAD'
  | 'IN_PROGRESS'
  | 'CLOSED_WON'
  | 'CLOSED_LOST'

export interface Activity {
  id: string
  tenant_id: string
  conversation_id?: string | null
  contact_id?: string | null
  type: string
  summary: string
  next_action: string | null
  priority: string
  status: string
  due_at: string | null
  acknowledged_by: string | null
  acknowledged_at: string | null
  created_at: string
  updated_at: string
}

export interface BotRule {
  name: string
  pattern: string
  match: 'CONTAINS' | 'EXACT' | 'PREFIX'
  response: string
  terminal: boolean
  handoff: boolean
  enabled: boolean
}

export interface BotRuleSet {
  id: string
  tenant_id: string
  version: number
  rules: BotRule[]
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface OperatorAuditLog {
  id: string
  tenant_id: string
  operator_id: string
  action: string
  conversation_id: string | null
  details: Record<string, unknown>
  created_at: string
}

export interface Paginated<T> {
  items: T[]
  total: number
  limit: number
  offset: number
}

export interface UploadJob {
  id: string
  tenant_id: string
  message_id: string
  host_id: string
  object_key: string
  mime_type: string
  media_path: string
  status: 'PENDING' | 'PROCESSING' | 'COMPLETED' | 'FAILED'
  attempt_count: number
  next_attempt_at: string
  last_error: string
  media_url: string
  lease_until: string | null
  created_at: string
  updated_at: string
}

export interface MediaUploadResponse {
  media_key: string
  media_url: string
  mime_type: string
  size: number
}

// Tenant and Operator types
export type OperatorRole = 'ADMIN' | 'OPERATOR' | 'VIEWER' | 'admin' | 'operator' | 'viewer'

export interface Operator {
  id: string
  tenant_id: string
  email?: string | null
  name: string
  role: string
  whatsapp_number?: string | null
  is_active: boolean
  totp_verified_at?: string | null
  totp_setup_required?: boolean
  email_verified_at?: string | null
  last_login_at?: string | null
  created_at?: string
  updated_at?: string
}

export interface ContactFieldDefinition {
  id: string
  tenant_id: string
  key: string
  label: string
  field_type: 'text' | 'number' | 'date' | 'select' | 'checkbox'
  options?: string[]
  is_required: boolean
  sort_order: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface Tenant {
  id: string
  name: string
  slug?: string
  setup_completed?: boolean
  setup_step?: number
  created_at?: string
  updated_at?: string
}

// TOTP and Authentication types
export interface TOTPSetupData {
  secret: string
  otpauth_url: string
  qr_code_svg?: string
  qr_code_data_url?: string
  temp_token?: string
  manual_entry_key?: string
}

export interface TOTPVerifyResult {
  verified: boolean
  backup_codes?: string[]
  token?: string
  user?: Operator
  session_id?: string
  tenant_id?: string
  tenant_slug?: string
  tenant_name?: string
}

export interface TOTPStatus {
  enabled: boolean
  verified_at: string | null
  backup_codes_remaining: number
  setup_required?: boolean
}

export interface BackupCodesResponse {
  backup_codes: string[]
}

// Invitations
export type InvitationStatus = 'PENDING' | 'ACCEPTED' | 'REVOKED' | 'EXPIRED'
export type InvitationChannel = 'WHATSAPP' | 'EMAIL' | 'MANUAL'

export interface Invitation {
  id: string
  tenant_id: string
  tenant_name?: string
  invitation_channel: InvitationChannel
  identifier: string // phone or email
  whatsapp_number?: string | null
  email?: string | null
  role: OperatorRole
  status: InvitationStatus
  token?: string
  setup_url?: string
  delivery_status?: 'SENT' | 'DELIVERED' | 'READ' | 'FAILED' | 'PENDING'
  expires_at: string
  created_at: string
  accepted_at?: string | null
}

export interface InvitationDetails {
  id: string
  token: string
  tenant_id: string
  tenant_name: string
  tenant_slug?: string
  identifier: string
  whatsapp_number?: string | null
  email?: string | null
  role: string
  totp_setup: TOTPSetupData
}

// Tenant Setup Wizard
export interface TenantSetupStatus {
  tenant_id: string
  tenant_name?: string
  tenant_slug?: string
  completed: boolean
  is_setup_complete?: boolean
  setup_step?: number
  org_details?: {
    name: string
    business_type?: string
    timezone?: string
    support_hours?: string
    website?: string
  }
  has_whatsapp_account?: boolean
  team_count?: number
  has_bot_rules?: boolean
}

export interface TenantSetupUpdate {
  setup_step: number
  tenant_name?: string
  org_details?: {
    business_type?: string
    timezone?: string
    support_hours?: string
    website?: string
  }
}

// QR Pairing types
export type PairingStatus = 'awaiting_scan' | 'connected' | 'failed' | 'cancelled' | 'expired'

export interface PairingSnapshot {
  id?: string
  status: PairingStatus | string
  qr_data_url?: string
  host_phone?: string
  error?: string
}

// Monitoring & observability types
export type InstanceStatus = 'ONLINE' | 'OFFLINE' | 'ERROR'

export interface StatusEvent {
  id: number
  tenant_id: string
  host_id: string
  status: InstanceStatus
  is_connected: boolean
  message: string
  occurred_at: string
}

export type InstanceEventType =
  | 'MESSAGE_IN'
  | 'MESSAGE_OUT'
  | 'RECEIPT'
  | 'STATUS'
  | 'QUEUE_DEPTH'
  | 'SEND_ERROR'
  | 'PROJECTION_FAILED'
  | 'MEDIA_ERROR'
  | 'UPLOAD_FAILED'
  | 'LOGGED_OUT'

export interface InstanceLogEvent {
  id: number
  host_id: string
  event_type: InstanceEventType
  direction: string | null
  payload: Record<string, unknown> | null
  occurred_at: string
}

export interface MetricsBucket {
  start: string
  inbound: number
  outbound: number
}

export interface MessageMetrics {
  inbound: number
  outbound: number
  failed: number
  status_breakdown: Record<string, number>
  buckets: MetricsBucket[]
}

export interface QueueDepthSample {
  id: number
  timestamp: string
  queue_size: number
}

export interface InstanceMonitoring {
  host_id: string
  status: InstanceStatus
  is_connected: boolean
  queue_size: number
  uptime: string
  last_connected_at: string | null
  last_disconnected_at: string | null
}
