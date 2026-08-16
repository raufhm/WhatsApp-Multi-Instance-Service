// Domain model types mirroring the Go backend (domain/models.go)

export type ConversationStatus =
  | 'OPEN'
  | 'BOT_ACTIVE'
  | 'WAITING'
  | 'HANDED_OFF'
  | 'CLOSED'

export type Actor = 'CONTACT' | 'BOT' | 'OPERATOR' | 'SYSTEM'
export type Direction = 'INCOMING' | 'OUTGOING'
export type MessageType = 'TEXT' | 'IMAGE' | 'VIDEO' | 'AUDIO' | 'DOCUMENT'

export interface Conversation {
  id: string
  tenant_id: string
  account_id: string
  contact_id: string
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

export interface ConversationMessage {
  id: string
  tenant_id: string
  conversation_id: string
  actor: Actor
  provider: string
  provider_message_id: string
  direction: Direction
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
  created_at: string
  updated_at: string
}

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

export interface Tenant {
  id: string
  name: string
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
  identifier: string
  whatsapp_number?: string | null
  email?: string | null
  role: string
  totp_setup: TOTPSetupData
}

// Tenant Setup Wizard
export interface TenantSetupStatus {
  tenant_id: string
  tenant_name: string
  completed: boolean
  is_setup_complete?: boolean
  current_step: number // 1 to 4
  setup_step?: number
  organization_details?: {
    name: string
    business_type?: string
    timezone?: string
    support_hours?: string
    website?: string
  }
  whatsapp_connected?: boolean
  has_whatsapp_account?: boolean
  team_invited_count?: number
}

export interface TenantSetupUpdate {
  step: number
  business_type?: string
  timezone?: string
  support_hours?: string
  website?: string
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
