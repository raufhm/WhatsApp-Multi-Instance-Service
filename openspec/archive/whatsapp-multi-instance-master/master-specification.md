# 📱 WhatsApp Multi-Instance Service - Master Specification

**Version**: 1.0.0  
**Date**: August 16, 2025  
**Status**: COMPREHENSIVE REVIEW COMPLETE  
**Architecture**: Event-Driven Multi-Tenant WhatsApp Platform

---

## Executive Summary

This document provides the **complete architectural specification** for the WhatsApp Multi-Instance Service - a production-ready platform for managing multiple WhatsApp Business accounts, automated bot interactions, and human operator handoff.

### Core Capabilities

✅ **Multi-Instance WhatsApp Management**
- Spawn and manage multiple WhatsApp instances per tenant
- Automatic instance lifecycle management
- Real-time status monitoring
- Message queue with backpressure handling

✅ **Intelligent Bot System**
- Rule-based conversation engine
- Pattern matching (contains, exact, prefix)
- Bot-to-human handoff
- Session state management
- Configurable rule sets per tenant

✅ **Human Operator Dashboard**
- TOTP-based secure authentication
- WhatsApp-first operator invitations
- Real-time conversation management
- Contact and conversation tracking
- Activity management and assignment

✅ **Conversation Management**
- Automatic conversation tracking
- Status lifecycle (OPEN → BOT_ACTIVE → WAITING → HANDED_OFF → CLOSED)
- Message threading
- Contact synchronization
- Merge and split capabilities

✅ **Media & Upload Management**
- Retryable S3 upload system
- Media message handling
- Bounded exponential backoff
- Durable job queue

---

## System Architecture

### High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         WhatsApp Clients                         │
│  (Multiple Tenant Instances via whatsmeow)                       │
└────────────────┬────────────────────────────────────────────────┘
                 │ Real-time Events
                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                    WhatsApp Manager                              │
│  ├─ Instance Spawning & Lifecycle                                │
│  ├─ Message Queue (per instance)                                 │
│  └─ Status Tracking                                              │
└────────────────┬────────────────────────────────────────────────┘
                 │
        ┌────────┴────────┐
        │                 │
        ▼                 ▼
┌──────────────┐  ┌──────────────────┐
│   Dispatcher │  │  Bot Processor   │
│   (Multi)    │  │  (Rule Engine)   │
└──────┬───────┘  └────────┬─────────┘
       │                   │
       │                   │ (if bot-eligible)
       ▼                   ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Async Projector                               │
│  (Database Projection with Retry Logic)                          │
└────────────────┬────────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                      PostgreSQL                                  │
│  ├─ Tenants, Operators, Sessions                                 │
│  ├─ Accounts, Contacts, Conversations                            │
│  ├─ Messages, Activities                                         │
│  ├─ Bot Rules, Bot Sessions                                      │
│  └─ TOTP, Invitations, Audit Logs                                │
└─────────────────────────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Operator Dashboard (React)                      │
│  ├─ Inbox, Conversations, Contacts                               │
│  ├─ Bot Rules Management                                         │
│  ├─ Team & Invitations                                           │
│  └─ Settings & TOTP                                              │
└─────────────────────────────────────────────────────────────────┘
```

### Component Layers

```
┌─────────────────────────────────────────────────┐
│              PRESENTATION LAYER                  │
│  Frontend (React + TanStack)                    │
│  - TanStack Router                               │
│  - TanStack Query                                │
│  - React Hook Form + Zod                         │
│  - Tailwind CSS + shadcn/ui                      │
└─────────────────────────────────────────────────┘
                    │ HTTP/REST
┌─────────────────────────────────────────────────┐
│               API LAYER                          │
│  Handlers (handler/*.go)                        │
│  - WhatsApp Legacy API (/api/*)                  │
│  - Versioned API (/api/v1/*)                     │
│  - Dashboard API (/dashboard/api/*)              │
│  - Session Middleware                            │
│  - TOTP Authentication                           │
└─────────────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────────────┐
│             BUSINESS LOGIC LAYER                 │
│  Core Services:                                  │
│  - WhatsApp Manager (whatsapp/*.go)              │
│  - Bot Engine (internal/bot/*.go)                │
│  - Conversation Service (internal/conversation/*)│
│  - Upload Manager (internal/upload/*.go)         │
│  - TOTP Service (internal/totp/*.go)             │
└─────────────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────────────┐
│              DATA ACCESS LAYER                   │
│  Storage (internal/storage/*.go)                │
│  - PostgreSQL Repository                         │
│  - S3 Storage                                    │
│  - Migrations                                    │
└─────────────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────────────┐
│               DOMAIN LAYER                       │
│  Models (domain/models.go)                      │
│  - Entities & Value Objects                      │
│  - Domain Events                                 │
│  - Repository Interfaces                         │
└─────────────────────────────────────────────────┘
```

---

## Core Features - Complete Inventory

### 1. Multi-Instance WhatsApp Management

**Status**: ✅ **PRODUCTION READY**

**Capabilities**:
- Spawn multiple WhatsApp instances per tenant
- Automatic instance discovery on startup
- Instance lifecycle management (start, stop, restart)
- Real-time connection status tracking
- Message queue per instance (bounded, 100 messages)
- Event handling (messages, receipts, group updates)
- Automatic reconnection on failure

**Key Components**:
```
whatsapp/
├── subsystem.go        # WhatsAppManager, WhatsAppInstance
├── bot_sender.go       # BotSender adapter
├── interceptors.go     # AsyncProjector, MultiDispatcher
└── interceptors_test.go
```

**Key Functions**:
- `WhatsAppManager.Start()` - Spawn all instances from DB
- `WhatsAppManager.SpawnInstance()` - Create new instance
- `WhatsAppManager.StopInstance()` - Graceful shutdown
- `WhatsAppInstance.worker()` - Process message queue
- `WhatsAppInstance.eventHandler()` - Handle WhatsApp events

**Database Tables**:
- `accounts` - WhatsApp account instances
- `whatsapp_instance_status` - Real-time status tracking

**API Endpoints**:
- `GET /api/v1/accounts` - List all accounts
- `POST /api/onboard` - Onboard new WhatsApp number
- `POST /api/send` - Send message (legacy)

---

### 2. Intelligent Bot System

**Status**: ✅ **PRODUCTION READY**

**Capabilities**:
- Rule-based conversation engine
- Pattern matching strategies:
  - `CONTAINS` - Pattern anywhere in text
  - `EXACT` - Exact match
  - `PREFIX` - Starts with pattern
- Bot session state management
- Rule versioning per tenant
- Automatic handoff to human operators
- Session timeout management
- Per-conversation event serialization

**Key Components**:
```
internal/bot/
├── engine.go           # Bot engine, rule evaluation
└── (processor in memory)
```

**Key Types**:
```go
type Rule struct {
    Name, Pattern, Response string
    Match MatchType  // CONTAINS, EXACT, PREFIX
    Terminal, Handoff bool
}

type Decision struct {
    Response string
    Rule     Rule
    Terminal bool
    Handoff  bool
}

type Event struct {
    TenantID, ConversationID uuid.UUID
    Host, Recipient, Text    string
    At                       time.Time
    Eligible                 bool
}
```

**Key Functions**:
- `Engine.Evaluate(text)` - Pattern matching
- `Processor.Handle(event)` - Process inbound message
- `bot.Processor` - Serializes events per conversation

**Database Tables**:
- `bot_rule_sets` - Versioned rule sets
- `bot_rules` - Individual rules
- `bot_sessions` - Conversation state

**API Endpoints**:
- `GET /api/bots` - List bot configurations
- `GET /api/bots/detail` - Get bot details
- `GET /dashboard/api/bot-rules` - Get active rules
- `POST /dashboard/api/bot-rules/activate` - Activate rule set

---

### 3. Conversation Management

**Status**: ✅ **PRODUCTION READY**

**Capabilities**:
- Automatic conversation creation
- Status lifecycle management:
  - `OPEN` - New conversation
  - `BOT_ACTIVE` - Bot handling
  - `WAITING` - Awaiting human
  - `HANDED_OFF` - Transferred to operator
  - `CLOSED` - Resolved
- Contact synchronization
- Message threading
- Conversation assignment
- Merge conversations
- Split conversations
- Activity tracking

**Key Components**:
```
internal/conversation/
├── contracts.go        # Service interface, normalization
└── contracts_test.go

internal/storage/
├── postgres.go         # Conversation CRUD
├── merge_audit.go      # Merge tracking
└── notes.go            # Internal notes
```

**Key Types**:
```go
type Conversation struct {
    ID, TenantID, AccountID, ContactID uuid.UUID
    TicketNumber int64
    Status       ConversationStatus
    BotState     string
    Assignee     string
    MergedIntoID *uuid.UUID
    // ... timestamps
}

type ConversationMessage struct {
    ID, TenantID, ConversationID uuid.UUID
    Actor         Actor  // CONTACT, BOT, OPERATOR, SYSTEM
    Direction     string // incoming, outgoing
    Content       string
    MessageType   string // text, image, video, etc.
    MediaURL      string
    Status        string
    IsInternal    bool   // Operator notes
}
```

**Database Tables**:
- `conversations` - Conversation tracking
- `conversation_messages` - All messages
- `contacts` - Contact database
- `activities` - Action items
- `conversation_assignments` - Operator assignments
- `merge_audit_logs` - Merge history

**API Endpoints**:
- `GET /api/v1/conversations` - List conversations
- `GET /api/v1/tickets/:id` - Get conversation detail
- `POST /dashboard/api/conversations/:id/assign` - Assign operator
- `POST /dashboard/api/conversations/:id/close` - Close conversation
- `POST /dashboard/api/conversations/merge` - Merge conversations
- `POST /dashboard/api/conversations/:id/note` - Add internal note

---

### 4. Operator Dashboard & Authentication

**Status**: ✅ **PRODUCTION READY** (TOTP Complete)

**Capabilities**:
- TOTP-based authentication (no passwords)
- WhatsApp-first operator invitations
- Backup codes for recovery
- Admin TOTP reset
- Role-based access (admin, operator)
- Real-time inbox
- Conversation management UI
- Team management
- Bot rules configuration

**Key Components**:
```
Frontend:
frontend/src/pages/
├── Login.tsx              # TOTP login
├── Recovery.tsx           # Backup code recovery
├── SignupTenant.tsx       # Tenant signup
├── OperatorInvitation.tsx # Accept invitation
├── Inbox.tsx              # Conversation inbox
├── ConversationDetail.tsx # Conversation view
├── Team.tsx               # Team management
├── BotRules.tsx           # Bot configuration
├── SetupWizard.tsx        # Tenant onboarding
└── TotpSettings.tsx       # TOTP management

Backend:
handler/
├── dashboard.go           # Dashboard routes
├── dashboard_api.go       # API handlers
├── tenant_onboarding.go   # TOTP, invitations
└── operator_actions.go    # Operator management

internal/storage/
├── auth.go                # TOTP, sessions
├── tenant_onboarding.go   # Invitations, recovery
└── operator_actions.go    # Operator CRUD
```

**Database Tables**:
- `operators` - Dashboard users
- `sessions` - Auth sessions
- `totp_secrets` - Encrypted TOTP secrets
- `totp_backup_codes` - Recovery codes
- `invitations` - WhatsApp/email invitations
- `recovery_audit_log` - Recovery tracking

**API Endpoints**:
- `POST /dashboard/api/login` - TOTP login
- `POST /dashboard/api/login/backup-code` - Backup code login
- `POST /dashboard/api/logout` - End session
- `GET /dashboard/api/me` - Current user
- `POST /dashboard/api/totp/setup` - Setup TOTP
- `POST /dashboard/api/totp/verify-setup` - Verify TOTP
- `GET /dashboard/api/invitations` - List invitations
- `POST /dashboard/api/invitations/whatsapp` - Send WhatsApp invite
- `POST /dashboard/api/operators/:id/totp-reset` - Admin reset

---

### 5. Media Upload & S3 Integration

**Status**: ✅ **PRODUCTION READY**

**Capabilities**:
- Retryable S3 upload system
- Durable job queue
- Bounded exponential backoff
- Configurable retry limits
- Media URL tracking
- WhatsApp media message handling

**Key Components**:
```
internal/upload/
├── upload.go             # Upload manager, worker
├── backoff.go            # Exponential backoff logic
└── (integration in subsystem.go)
```

**Key Types**:
```go
type UploadJob struct {
    ID, TenantID uuid.UUID
    ContentType  string
    Data         []byte
    Attempts     int
    Status       string
    // ... timestamps
}
```

**Key Functions**:
- `upload.Manager.Start()` - Start worker
- `upload.Manager.Enqueue()` - Add upload job
- `upload.backoff()` - Calculate delay

**Database Tables**:
- `upload_jobs` - Retryable upload queue

**Configuration**:
```yaml
upload:
  worker_enabled: true
  poll_interval: 5s
  max_attempts: 5
  initial_backoff: 1s
  max_backoff: 5m
  jitter: 0.5
```

---

### 6. TOTP Authentication System

**Status**: ✅ **PRODUCTION READY** (Archived)

**Capabilities**:
- AES-256-GCM encrypted secrets
- RFC 6238 TOTP generation/verification
- QR code setup
- Manual entry fallback
- 10 single-use backup codes
- Recovery via admin reset
- Complete audit logging

**Documentation**: See `openspec/archive/tenant-onboarding-flow-completed/`

---

## Data Models - Complete Reference

### Core Entities

```go
// Tenant - Organization
type Tenant struct {
    ID              uuid.UUID
    Name            string
    SetupStep       int
    IsSetupComplete bool
    OrgDetails      map[string]any
}

// WhatsAppAccount - Connected WhatsApp instance
type WhatsAppAccount struct {
    ID, TenantID  uuid.UUID
    HostID, Provider, DisplayName string
}

// Contact - WhatsApp contact
type Contact struct {
    ID, TenantID       uuid.UUID
    NormalizedAddress  string
    ProviderAddress    string
    DisplayName        string
    Metadata           map[string]any
}

// Conversation - Chat thread
type Conversation struct {
    ID, TenantID, AccountID, ContactID uuid.UUID
    TicketNumber    int64
    Status          ConversationStatus  // OPEN, BOT_ACTIVE, etc.
    BotState        string
    Assignee        string
    MergedIntoID    *uuid.UUID
}

// ConversationMessage - Message in thread
type ConversationMessage struct {
    ID, TenantID, ConversationID uuid.UUID
    Actor          Actor  // CONTACT, BOT, OPERATOR, SYSTEM
    Direction      string // incoming, outgoing
    Content        string
    MessageType    string // text, image, video, document, etc.
    MediaURL       string
    Status         string
    IsInternal     bool  // Operator notes
}

// Operator - Dashboard user
type Operator struct {
    ID, TenantID       uuid.UUID
    Email              string
    WhatsappNumber     string
    Name               string
    Role               string  // admin, operator
    IsActive           bool
    TotpVerifiedAt     *time.Time
    TotpSetupRequired  bool
    EmailVerifiedAt    *time.Time
}

// BotRule - Conversation rule
type BotRule struct {
    Name     string
    Pattern  string
    Match    string  // CONTAINS, EXACT, PREFIX
    Response string
    Terminal bool    // End conversation
    Handoff  bool    // Transfer to human
    Enabled  bool
}

// Activity - Action item
type Activity struct {
    ID, TenantID, ConversationID uuid.UUID
    Type       string  // FOLLOW_UP, CALLBACK, etc.
    Summary    string
    NextAction string
    Priority   string  // HIGH, MEDIUM, LOW
    Status     ActivityStatus
    DueAt      *time.Time
}
```

### Value Objects

```go
// ConversationKey - Unique conversation identifier
type ConversationKey struct {
    TenantID, AccountID, ContactID uuid.UUID
}

// MessageMetadata - Message context
type MessageMetadata struct {
    WhatsappID  string
    TenantID    uuid.UUID
    Host        string
    Sender      string
    Recipient   string
    Content     string
    Direction   string
    Type        string
    Status      InstanceStatus
    Timestamp   time.Time
    Actor       Actor
}

// MessageRequest - Outbound message
type MessageRequest struct {
    Recipient  string
    Message    string
    Type       string
    MediaURL   string
    Metadata   map[string]any
    Actor      Actor
}
```

---

## API Reference

### Legacy WhatsApp API (v1)

**Base**: `/api/v1/`  
**Auth**: `X-API-Key` header or `Authorization: Bearer <key>`

#### Accounts
```
GET    /api/v1/accounts              # List all accounts
POST   /api/v1/accounts/:host/messages  # Send message
```

#### Contacts
```
GET    /api/v1/contacts              # List contacts
GET    /api/v1/contacts/:id          # Get contact detail
```

#### Conversations
```
GET    /api/v1/conversations         # List conversations
GET    /api/v1/tickets/:id           # Get conversation
```

### Dashboard API

**Base**: `/dashboard/api/`  
**Auth**: Session cookie (HttpOnly, Secure)

#### Authentication
```
POST   /dashboard/api/login          # TOTP login
POST   /dashboard/api/login/backup-code  # Backup code
POST   /dashboard/api/logout         # End session
GET    /dashboard/api/me             # Current user
```

#### TOTP Management
```
GET    /dashboard/api/totp/setup/:token  # Get TOTP setup info
POST   /dashboard/api/totp/verify-setup  # Verify TOTP code
GET    /dashboard/api/account/totp       # Get TOTP status
POST   /dashboard/api/account/totp/regenerate-backup-codes
```

#### Invitations
```
GET    /dashboard/api/invitations          # List invitations
POST   /dashboard/api/invitations/whatsapp # Send WhatsApp invite
POST   /dashboard/api/invitations/email    # Send email invite
DELETE /dashboard/api/invitations/:id      # Revoke invitation
```

#### Conversations
```
GET    /dashboard/api/conversations        # List conversations
GET    /dashboard/api/conversations/:id    # Get detail
POST   /dashboard/api/conversations/:id/assign
POST   /dashboard/api/conversations/:id/close
POST   /dashboard/api/conversations/:id/reopen
POST   /dashboard/api/conversations/merge
POST   /dashboard/api/conversations/:id/note
```

#### Bot Rules
```
GET    /dashboard/api/bot-rules            # Get active rules
POST   /dashboard/api/bot-rules/activate   # Activate rule set
```

#### Team
```
GET    /dashboard/api/operators            # List operators
POST   /dashboard/api/operators/:id/totp-reset  # Admin reset
```

---

## Configuration

### Environment Variables

```bash
# Server
PORT=8080
LOG_LEVEL=info

# Database
PGDSN=postgres://user:pass@localhost:5432/dbname?sslmode=disable

# AWS S3
AWS_REGION=us-east-1
AWS_S3_BUCKET=your-bucket

# WhatsApp
BOT_FALLBACK_REPLY=Sorry, I didn't understand that.
BOT_SESSION_TIMEOUT=24h
BOT_RULES_VERSION=default

# Upload Worker
UPLOAD_LEASE=1h
UPLOAD_WORKER_ENABLED=true
UPLOAD_POLL_INTERVAL=5s
UPLOAD_MAX_ATTEMPTS=5
UPLOAD_INITIAL_BACKOFF=1s
UPLOAD_MAX_BACKOFF=5m
UPLOAD_JITTER=0.5

# TOTP
TOTP_ENCRYPTION_KEY=<32-byte-hex-key>
```

### Config File (Optional)

```yaml
server:
  port: "8080"
  log_level: "info"

database:
  pgdsn: "postgres://..."

aws:
  region: "us-east-1"
  s3_bucket: "your-bucket"

bot:
  fallback_reply: "Sorry, I didn't understand that."
  session_timeout: "24h"
  rules_version: "default"

upload:
  worker_enabled: true
  poll_interval: "5s"
  max_attempts: 5
  initial_backoff: "1s"
  max_backoff: "5m"
  jitter: 0.5

totp:
  encryption_key: "${TOTP_ENCRYPTION_KEY}"
```

---

## Deployment Architecture

### Production Deployment

```
┌─────────────────────────────────────────────────────────┐
│                    Load Balancer                         │
│              (nginx, ALB, or Cloudflare)                 │
└────────────────┬────────────────────────────────────────┘
                 │
        ┌────────┴────────┐
        │                 │
        ▼                 ▼
┌──────────────┐  ┌──────────────┐
│   App Node 1 │  │   App Node 2 │
│   (Go)       │  │   (Go)       │
│              │  │              │
│  - WhatsApp  │  │  - WhatsApp  │
│    Instances │  │    Instances │
│  - Bot Engine│  │  - Bot Engine│
│  - Upload    │  │  - Upload    │
│    Worker    │  │    Worker    │
└──────┬───────┘  └──────┬───────┘
       │                 │
       └────────┬────────┘
                │
        ┌───────┴───────┐
        │               │
        ▼               ▼
┌──────────────┐  ┌──────────────┐
│  PostgreSQL  │  │     S3       │
│  (RDS/EC2)   │  │   Storage    │
└──────────────┘  └──────────────┘
                │
        ┌───────┴───────┐
        │               │
        ▼               ▼
┌──────────────┐  ┌──────────────┐
│   Frontend   │  │   Monitoring │
│   (CDN)      │  │   (Sentry,   │
│              │  │    Datadog)  │
└──────────────┘  └──────────────┘
```

### Scaling Considerations

**WhatsApp Instances**:
- Each instance: ~10-50MB RAM
- Recommended: 10-20 instances per node
- Scale horizontally by adding nodes

**Database**:
- Connection pooling required
- Recommended: RDS with read replicas
- Index optimization critical

**Upload Worker**:
- Can be disabled if not using media
- Separate worker process recommended
- Queue-based scaling

---

## Monitoring & Observability

### Key Metrics

```sql
-- Instance uptime
SELECT host, status, COUNT(*) 
FROM whatsapp_instance_status 
GROUP BY host, status;

-- Message volume
SELECT DATE(timestamp), COUNT(*) 
FROM conversation_messages 
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY DATE(timestamp);

-- Bot effectiveness
SELECT 
  COUNT(CASE WHEN actor = 'BOT' THEN 1 END) * 100.0 / COUNT(*) as bot_percentage,
  COUNT(CASE WHEN status = 'HANDED_OFF' THEN 1 END) as handoffs
FROM conversations
WHERE started_at > NOW() - INTERVAL '7 days';

-- Upload job status
SELECT status, COUNT(*) 
FROM upload_jobs 
GROUP BY status;
```

### Logging

```go
// Instance status changes
log.Printf("STATUS [%s]: %s", hostID, status)

// Message processing
log.Printf("[%s] [%s] MessageID: %s | Type: %s | Status: %s",
    meta.Direction, meta.HostID, meta.WhatsappID, meta.Type, meta.Status)

// Errors
log.Printf("application message projection failed: %s", m.WhatsappID)
```

### Alerts

- WhatsApp instance down > 5 minutes
- Upload job failures > 10%
- Database connection pool exhaustion
- High error rate (> 5% of requests)

---

## Security

### Authentication & Authorization

- TOTP-based operator authentication
- Session-based dashboard access
- API key authentication for external APIs
- Role-based access control (admin, operator)

### Data Protection

- AES-256-GCM encryption for TOTP secrets
- Bcrypt hashing for backup codes
- HTTPS enforcement in production
- Secure cookies (HttpOnly, Secure, SameSite)

### Rate Limiting

- Login attempts: 5 per 15 minutes
- Backup code attempts: 3 per 15 minutes
- API requests: Configurable per tenant

### Audit Logging

- All authentication events
- Operator actions (assign, close, merge)
- TOTP resets
- Recovery events

---

## Testing Strategy

### Test Coverage

```
Backend:
- Unit tests: totp_test.go, engine_test.go
- Integration tests: tenant_onboarding_test.go
- Handler tests: dashboard_test.go, http_test.go

Frontend:
- Component tests: *.test.tsx
- E2E tests: (recommended with Playwright)
```

### Test Commands

```bash
# Backend tests
go test ./...

# Frontend tests
cd frontend && npm test

# E2E tests (recommended)
npx playwright test
```

---

## Gaps & Recommended Improvements

### Critical Gaps

1. **WhatsApp Invitation Messages Not Sent** - See separate spec
2. **Manual Onboarding UX Gaps** - See separate spec

### Recommended Enhancements

1. **WebSocket for Real-Time Updates**
   - Replace polling with WebSocket
   - Real-time inbox updates
   - Live conversation notifications

2. **Advanced Bot Features**
   - NLP integration (Dialogflow, Rasa)
   - Multi-turn conversations
   - Context-aware responses

3. **Analytics Dashboard**
   - Conversation metrics
   - Bot performance
   - Operator efficiency

4. **Webhook Integrations**
   - CRM integration
   - Ticketing systems
   - Third-party notifications

5. **Mobile App**
   - React Native operator app
   - Push notifications
   - Offline support

---

## Related Specifications

- [`openspec/archive/tenant-onboarding-flow-completed/`](openspec/archive/tenant-onboarding-flow-completed/) - TOTP authentication
- [`openspec/changes/manual-onboarding-whatsapp/`](openspec/changes/manual-onboarding-whatsapp/) - WhatsApp invitations
- [`openspec/changes/operator-dashboard/`](openspec/changes/operator-dashboard/) - Operator dashboard

---

## Appendix: Technology Stack

### Backend
- **Language**: Go 1.21+
- **WhatsApp Library**: whatsmeow (go.mau.fi/whatsmeow)
- **Database**: PostgreSQL 14+
- **Storage**: AWS S3
- **Queue**: In-memory channels (upgrade to Redis recommended)

### Frontend
- **Framework**: React 18+
- **Router**: TanStack Router v1
- **State**: TanStack Query v5
- **Forms**: React Hook Form v7 + Zod v3
- **UI**: Tailwind CSS + shadcn/ui
- **Build**: Vite

### Infrastructure
- **Container**: Docker (recommended)
- **Orchestration**: Kubernetes (optional)
- **Monitoring**: Sentry, Datadog (recommended)
- **CI/CD**: GitHub Actions (recommended)

---

**This specification is complete and covers all implemented features. Gaps are documented in separate specifications for AI agent implementation.**

**Next**: Review gap specifications and prioritize implementation.
