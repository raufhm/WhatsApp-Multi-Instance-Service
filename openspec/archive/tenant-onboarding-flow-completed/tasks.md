# Tasks: Tenant Onboarding Flow

## Phase 1: Foundation (Week 1-2)

### Backend

- [ ] **DB Migration: Add WhatsApp verification tokens table**
  - File: `migrations/0006_whatsapp_verification.up.sql`
  - Create `whatsapp_verification_tokens` table
  - Add indexes for performance
  - Track WhatsApp number, token hash, expiry, used_at

- [ ] **DB Migration: Add password reset tokens table**
  - File: `migrations/0007_password_reset_tokens.up.sql`
  - Create `password_reset_tokens` table
  - Add indexes for performance
  - Support both WhatsApp and email reset tokens

- [ ] **DB Migration: Add WhatsApp invitations table**
  - File: `migrations/0008_whatsapp_invitations.up.sql`
  - Create `invitations` table with `channel` column (whatsapp/email/manual)
  - Add `whatsapp_number` field
  - Add `email_address` field (optional)
  - Add indexes for performance

- [ ] **DB Migration: Add WhatsApp invitation delivery tracking**
  - File: `migrations/0009_invitation_delivery.up.sql`
  - Create `whatsapp_invitation_delivery` table
  - Track delivery status: pending, sent, delivered, read, failed
  - Track retry count and failure reasons

- [ ] **DB Migration: Update operators table**
  - File: `migrations/0010_operator_fields.up.sql`
  - Add `whatsapp_number` column (unique, indexed)
  - Add `email_verified_at` column (optional for operators)
  - Add `must_change_password` column
  - Make email optional for operators

- [ ] **Implement WhatsApp message template service**
  - File: `internal/whatsapp/templates.go`
  - Define invitation message template
  - Define password reset message template
  - Define verification message template
  - Define confirmation/notification templates
  - Support localization (optional)

- [ ] **Create WhatsApp message templates**
  - Directory: `internal/whatsapp/templates/`
  - Template: `operator_invitation.txt`
  - Template: `password_reset.txt`
  - Template: `verification.txt`
  - Template: `password_changed_confirmation.txt`
  - Template: `welcome_after_signup.txt`

- [ ] **Implement WhatsApp verification token storage**
  - File: `internal/storage/whatsapp_verification.go`
  - `CreateVerificationToken(whatsappNumber, tenantID) (token string, err)`
  - `VerifyWhatsApp(tokenHash) error`
  - `GetVerificationToken(whatsappNumber) (token, expiresAt, err)`
  - `DeleteVerificationToken(whatsappNumber) error`

- [ ] **Implement password reset token storage (WhatsApp + email)**
  - File: `internal/storage/password_reset.go`
  - `CreateResetToken(operatorID, channel) (token string, err)`
  - `ValidateResetToken(tokenHash) (operatorID, channel, err)`
  - `UseResetToken(tokenHash) error`
  - `InvalidateAllResetTokens(operatorID) error`
  - Support both WhatsApp and email channels

- [ ] **Implement WhatsApp invitation storage**
  - File: `internal/storage/invitations.go`
  - `CreateWhatsAppInvitation(tenantID, whatsappNumber, role, invitedBy) (token string, err)`
  - `CreateEmailInvitation(tenantID, email, role, invitedBy) (token string, err)`
  - `CreateManualInvitation(tenantID, role, invitedBy) (token string, err)`
  - `GetInvitation(tokenHash) (invitation, err)`
  - `AcceptInvitation(tokenHash) error`
  - `RevokeInvitation(invitationID) error`
  - `ListPendingInvitations(tenantID) ([]invitation, err)`
  - `TrackDelivery(invitationID, status, metadata) error`

### Email Service (Optional Fallback Only)

- [ ] **Implement minimal email service for fallback**
  - File: `internal/email/service.go`
  - Define `EmailService` interface
  - Implement simple SMTP backend (no need for SendGrid/SES complexity)
  - Only for: tenant admin verification, password reset fallback
  - Keep it minimal to reduce vendor lock-in

- [ ] **Create minimal email templates**
  - Directory: `internal/email/templates/`
  - Template: `tenant_verification.html` (one-time for admin)
  - Template: `password_reset_fallback.html` (when WhatsApp fails)
  - Plain text fallbacks for each

---

## Phase 2: Signup Flow (Week 2-3)

### Backend

- [ ] **Implement tenant signup endpoint**
  - File: `handler/signup.go`
  - `POST /dashboard/api/signup/tenant`
  - Validate input (email, password, org name)
  - Create tenant record
  - Create admin operator record
  - Generate verification token
  - Send verification email
  - Return success response

- [ ] **Implement operator signup endpoint**
  - File: `handler/signup.go`
  - `POST /dashboard/api/signup/operator`
  - Validate input (tenant ID, email, password)
  - Verify tenant exists and is active
  - Create operator record
  - Generate verification token
  - Send verification email
  - Return success response

- [ ] **Implement email verification endpoint**
  - File: `handler/verification.go`
  - `GET /dashboard/api/verify-email/:token`
  - Validate token format
  - Find and verify token
  - Mark operator email as verified
  - Invalidate token
  - Return success/error

- [ ] **Implement resend verification email**
  - File: `handler/verification.go`
  - `POST /dashboard/api/verify-email/resend`
  - Validate email + tenant ID
  - Check rate limits
  - Generate new token
  - Send email
  - Return success

- [ ] **Add rate limiting middleware**
  - File: `handler/rate_limiter.go`
  - Implement token bucket or sliding window
  - Apply to signup, login, verification endpoints
  - Return 429 on limit exceeded

### Frontend

- [ ] **Create signup landing page**
  - File: `frontend/src/pages/SignupLanding.tsx`
  - Choose between "Create Organization" vs "Join Existing"
  - Clear value proposition
  - Links to detailed signup forms

- [ ] **Create tenant signup form**
  - File: `frontend/src/pages/SignupTenant.tsx`
  - Fields: org name, admin email, password, confirm password
  - Real-time validation
  - Password strength indicator
  - Terms & privacy checkboxes
  - Submit handler with error handling

- [ ] **Create operator signup form**
  - File: `frontend/src/pages/SignupOperator.tsx`
  - Fields: tenant ID (UUID), email, password, confirm password
  - UUID format validation
  - Real-time validation
  - Submit handler

- [ ] **Create email verification page**
  - File: `frontend/src/pages/VerifyEmail.tsx`
  - Auto-submit token on mount
  - Loading state
  - Success state (with redirect)
  - Error state (with resend option)

- [ ] **Add auth route guards**
  - File: `frontend/src/App.tsx`
  - Prevent authenticated users from accessing signup/login
  - Redirect unauthenticated users appropriately

---

## Phase 3: Password Reset Flow (Week 3-4)

### Backend

- [ ] **Implement password reset request endpoint**
  - File: `handler/password_reset.go`
  - `POST /dashboard/api/password-reset/request`
  - Validate email + tenant ID
  - Find operator
  - Generate reset token
  - Send reset email
  - Return generic success (no enumeration)

- [ ] **Implement password reset confirm endpoint**
  - File: `handler/password_reset.go`
  - `POST /dashboard/api/password-reset/confirm`
  - Validate token
  - Validate new password strength
  - Hash and update password
  - Invalidate all sessions
  - Invalidate all reset tokens
  - Send password changed notification email
  - Return success

- [ ] **Implement admin password reset**
  - File: `handler/dashboard_api.go`
  - `POST /dashboard/api/operators/:id/reset-password`
  - Admin-only endpoint
  - Generate reset token OR set temp password
  - Send email to operator
  - Audit log entry

### Frontend

- [ ] **Create forgot password page**
  - File: `frontend/src/pages/ForgotPassword.tsx`
  - Fields: email, tenant ID
  - Submit handler
  - Generic success message
  - Back to login link

- [ ] **Create reset password page**
  - File: `frontend/src/pages/ResetPassword.tsx`
  - Validate token from URL
  - Fields: new password, confirm password
  - Password strength indicator
  - Submit handler
  - Success/error states

- [ ] **Add password reset link to login**
  - File: `frontend/src/pages/Login.tsx`
  - "Forgot password?" link
  - Position prominently

---

## Phase 4: Invitation Flow (Week 4-5)

### Backend

- [ ] **Implement create invitation endpoint**
  - File: `handler/invitations.go`
  - `POST /dashboard/api/invitations`
  - Admin-only
  - Validate email, role
  - Check for existing invitation
  - Generate token
  - Send invitation email
  - Return invitation details

- [ ] **Implement list invitations endpoint**
  - File: `handler/invitations.go`
  - `GET /dashboard/api/invitations`
  - Admin-only
  - Return pending invitations for tenant

- [ ] **Implement revoke invitation endpoint**
  - File: `handler/invitations.go`
  - `DELETE /dashboard/api/invitations/:id`
  - Admin-only
  - Mark invitation as revoked
  - Return success

- [ ] **Implement accept invitation page**
  - File: `handler/invitations.go`
  - `GET /dashboard/api/invitations/accept/:token`
  - Validate token
  - Pre-fill signup form
  - Skip email verification

### Frontend

- [ ] **Create invitation management UI**
  - File: `frontend/src/pages/Invitations.tsx`
  - List pending invitations
  - "Invite operator" button
  - Revoke action
  - Resend action (optional)

- [ ] **Create invitation modal/form**
  - File: `frontend/src/components/InviteOperatorModal.tsx`
  - Fields: email, role dropdown
  - Validation
  - Submit handler
  - Success feedback

- [ ] **Create accept invitation page**
  - File: `frontend/src/pages/AcceptInvitation.tsx`
  - Validate token on mount
  - Pre-filled form (email, role)
  - Password fields only
  - Submit handler
  - Skip verification

---

## Phase 5: Tenant Setup Wizard (Week 5-6)

### Backend

- [ ] **Implement setup status endpoint**
  - File: `handler/tenant_setup.go`
  - `GET /dashboard/api/tenant/setup-status`
  - Return completed steps
  - Return pending steps

- [ ] **Implement update setup step endpoint**
  - File: `handler/tenant_setup.go`
  - `PUT /dashboard/api/tenant/setup`
  - Validate step data
  - Mark step as complete
  - Return updated status

- [ ] **Implement complete setup endpoint**
  - File: `handler/tenant_setup.go`
  - `POST /dashboard/api/tenant/complete-setup`
  - Validate all steps complete
  - Mark tenant setup_completed = true
  - Return success

### Frontend

- [ ] **Create setup wizard layout**
  - File: `frontend/src/pages/SetupWizard.tsx`
  - Multi-step form structure
  - Progress indicator
  - Back/Next navigation
  - Save & continue later

- [ ] **Create organization details step**
  - File: `frontend/src/pages/setup/OrganizationStep.tsx`
  - Fields: org name, industry, timezone, size
  - Validation
  - Save handler

- [ ] **Create WhatsApp pairing step**
  - File: `frontend/src/pages/setup/WhatsAppStep.tsx`
  - QR code display
  - Pairing status polling
  - Success confirmation

- [ ] **Create completion step**
  - File: `frontend/src/pages/setup/CompleteStep.tsx`
  - Success message
  - Summary of setup
  - "Go to dashboard" button
  - Optional: celebrate animation

- [ ] **Implement first-login redirect**
  - File: `frontend/src/App.tsx`
  - Check setup_completed on login
  - Redirect to wizard if not complete
  - Allow skip and complete later

---

## Phase 6: Polish & Testing (Week 6-7)

### Testing

- [ ] **Write backend unit tests**
  - Signup handlers
  - Verification handlers
  - Password reset handlers
  - Invitation handlers
  - Token storage functions

- [ ] **Write integration tests**
  - Full signup flow (tenant + operator)
  - Email verification flow
  - Password reset flow
  - Invitation flow
  - Setup wizard flow

- [ ] **Write frontend component tests**
  - Signup forms
  - Login form
  - Verification page
  - Reset password forms
  - Setup wizard steps

- [ ] **Write E2E tests**
  - Cypress or Playwright tests for full user journeys
  - Cover happy paths and error scenarios

### Security

- [ ] **Security review**
  - Token generation and storage
  - Rate limiting effectiveness
  - Email enumeration prevention
  - Session management
  - CSRF protection on all forms

- [ ] **Penetration testing**
  - Test for common vulnerabilities
  - SQL injection
  - XSS
  - CSRF
  - Session fixation

### UX & Accessibility

- [ ] **Accessibility audit**
  - Screen reader testing
  - Keyboard navigation
  - Color contrast
  - Focus management
  - ARIA labels

- [ ] **Responsive design testing**
  - Mobile devices
  - Tablets
  - Desktop
  - Various browsers

- [ ] **Loading states & error handling**
  - All forms have loading states
  - Error messages are clear and helpful
  - Retry mechanisms where appropriate

### Documentation

- [ ] **API documentation**
  - Update OpenAPI/Swagger spec
  - Document all new endpoints
  - Include request/response examples

- [ ] **User documentation**
  - Signup guide
  - Password reset guide
  - Invitation guide
  - Setup wizard guide

- [ ] **Admin documentation**
  - Managing operators
  - Managing invitations
  - Viewing audit logs

---

## Phase 7: Deployment & Monitoring (Week 7-8)

### Deployment

- [ ] **Update environment configuration**
  - Add email service env vars
  - Update `.env.example`
  - Update Docker configuration

- [ ] **Database migration deployment**
  - Test migrations on staging
  - Plan rollback strategy
  - Deploy to production

- [ ] **Email service configuration**
  - Configure SMTP/SendGrid/SES
  - Verify DNS records (SPF, DKIM, DMARC)
  - Test email delivery

### Monitoring

- [ ] **Add metrics collection**
  - Signup conversion rates
  - Email delivery rates
  - Verification rates
  - Password reset rates
  - Error rates by endpoint

- [ ] **Add logging**
  - Log signup attempts (without PII)
  - Log verification attempts
  - Log password reset requests
  - Log security events

- [ ] **Setup alerts**
  - Email delivery failures
  - Signup error rate spikes
  - Security event thresholds
  - Rate limit triggers

### Launch

- [ ] **Feature flag (optional)**
  - Wrap new features in feature flag
  - Enable gradually
  - Monitor impact

- [ ] **Communication**
  - Announce to existing users
  - Update documentation
  - Prepare support team

---

## Deferred (Future Phases)

- [ ] OAuth/Social login (Google, Microsoft)
- [ ] SSO/SAML integration
- [ ] Multi-factor authentication (MFA)
- [ ] Phone-based verification (SMS)
- [ ] Account self-deletion
- [ ] Data export functionality
- [ ] Advanced audit logging
- [ ] Custom branding per tenant
- [ ] White-label email templates
