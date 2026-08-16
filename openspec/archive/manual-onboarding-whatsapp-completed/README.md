# 📋 Manual Onboarding & WhatsApp Integration Specification

**Status**: PROPOSED  
**Priority**: CRITICAL - Required for Production  
**Created**: August 16, 2025

---

## Overview

This specification defines the manual onboarding flow enhancements and WhatsApp message sending implementation required to complete the TOTP-based authentication system for production deployment.

---

## Problem Statement

### Current State

1. **WhatsApp Invitations Created but Not Sent**
   - Invitations are stored in database ✅
   - WhatsApp message sending is NOT implemented ❌
   - Users must manually request invitation code ❌

2. **Manual Onboarding Flow Gaps**
   - Token entry exists but is isolated to invitation page
   - No dedicated "Join with Code" landing page
   - No clear path for users with invitation codes

3. **Inconsistent User Experience**
   - Some users get WhatsApp (if admin manually sends)
   - Others get no communication
   - Confusion about how to join

### Impact

- ❌ Cannot launch to production
- ❌ High support burden (manual invitation sending)
- ❌ Poor user experience
- ❌ Low invitation acceptance rates

---

## Solution Summary

### 1. Automated WhatsApp Message Sending

**What**: Automatically send WhatsApp messages when invitations are created

**How**:
- Integrate with existing WhatsApp Manager
- Create message templates for invitations
- Track delivery status
- Handle errors gracefully

**Benefits**:
- ✅ Zero manual work for admins
- ✅ Instant delivery
- ✅ Higher acceptance rates
- ✅ Professional user experience

### 2. Enhanced Manual Onboarding Flow

**What**: Standalone "Join with Code" page and improved UX

**How**:
- Create dedicated `/join` page
- Add navigation from login/signup pages
- Improve token validation UX
- Add QR code option for future

**Benefits**:
- ✅ Clear entry point for invited users
- ✅ Reduced confusion
- ✅ Mobile-friendly flow
- ✅ Consistent with TanStack Router architecture

---

## Technical Architecture

### Component Diagram

```
┌─────────────────────┐
│   Admin Dashboard   │
│  (Create Invite)    │
└──────────┬──────────┘
           │ POST /api/invitations/whatsapp
           ▼
┌─────────────────────┐
│  Tenant Onboarding  │
│     Handler         │
│                     │
│  1. Create in DB    │
│  2. Send WhatsApp   │◄──────┐
│  3. Track Delivery  │        │
└──────────┬──────────┘        │
           │                   │
           ▼                   │
┌─────────────────────┐        │
│  WhatsApp Manager   │────────┘
│  (whatsmeow)        │  Send Message
│                     │────────────► WhatsApp
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│    User Phone       │
│  (Receives Message) │
│                     │
│  "Click link or     │
│   paste code at     │
│   /join"            │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   /join Page        │
│  (Enter Code)       │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ /invitation/:token  │
│  (Setup TOTP)       │
└─────────────────────┘
```

### Data Flow

```
Admin Creates Invitation
    │
    ├─▶ Database (invitations table)
    │      └─ Status: pending
    │      └─ Token: hashed
    │      └─ Expiry: 7 days
    │
    ├─▶ WhatsApp Message
    │      └─ Template: formatted
    │      └─ Link: /invitation/{token}
    │      └─ Code: plain text token
    │
    ├─▶ Delivery Tracking
    │      └─ Status: sent/failed
    │      └─ Message ID: from WhatsApp
    │      └─ Timestamp: now
    │
    └─▶ Response to Admin
           └─ Success: true/false
           └─ whatsapp_sent: boolean
```

---

## Implementation Details

### Backend Changes

#### 1. WhatsApp Templates (`internal/whatsapp/templates.go`)

```go
package whatsapp

import (
    "fmt"
    "time"
    "github.com/raufhm/whatsapp-testing/domain"
)

// BuildInvitationMessage creates formatted WhatsApp message
func BuildInvitationMessage(
    tenant domain.Tenant, 
    inv domain.Invitation, 
    token string,
    inviterName string,
) string {
    template := `👋 *You've been invited to join %s!*

Your team member %s has invited you to join their WhatsApp Operator Dashboard.

🔐 *Invitation Code*: %s
🔗 *Quick Link*: %s

⏰ This invitation expires in %d hours.

*Next Steps*:
1. Click the link above OR visit dashboard
2. Paste your invitation code
3. Complete your profile setup
4. Set up two-factor authentication
5. Download backup codes

Questions? Contact your administrator.

— WhatsApp Operator Dashboard`

    hoursLeft := time.Until(inv.ExpiresAt).Hours()
    
    // Use short link in production
    inviteURL := fmt.Sprintf("https://dashboard.example.com/invitation/%s", token)
    
    return fmt.Sprintf(template,
        tenant.Name,
        inviterName,
        token,
        inviteURL,
        int(hoursLeft),
    )
}
```

#### 2. Handler Update (`handler/tenant_onboarding.go`)

```go
func (d *DashboardHandler) handleCreateWhatsAppInvitation(
    w http.ResponseWriter, 
    r *http.Request, 
    tenantID, callerID uuid.UUID,
) {
    var req createWhatsAppInvitationRequest
    if err := DecodeJSONBody(r, &req); err != nil {
        writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
        return
    }

    // Validate phone number
    number := normalizePhoneNumber(req.WhatsappNumber)
    if number == "" {
        writeAPIError(w, 400, "VALIDATION_ERROR", "invalid whatsapp_number")
        return
    }

    // Create invitation in database
    inv, token, err := d.Auth.CreateInvitation(
        tenantID, 
        &callerID, 
        number, 
        "whatsapp", 
        req.Role, 
        number, 
        "",
    )
    if err != nil {
        writeAPIError(w, 500, "INTERNAL_ERROR", err.Error())
        return
    }

    // ✅ NEW: Send WhatsApp message
    whatsappSent := false
    var messageID string
    var sendErr error

    tenant, tenantErr := d.Auth.GetTenant(tenantID)
    caller, callerErr := d.Auth.GetOperatorByID(tenantID, callerID)
    
    if tenantErr == nil && callerErr == nil {
        inviterName := caller.Name
        if inviterName == "" {
            inviterName = "your team"
        }
        
        message := whatsapp.BuildInvitationMessage(tenant, inv, token, inviterName)
        
        sendErr = d.WhatsApp.SendInvitation(number, message)
        if sendErr != nil {
            log.Printf("WhatsApp invitation failed: %v", sendErr)
            d.Auth.TrackInvitationDelivery(inv.ID, "failed", "", sendErr.Error())
        } else {
            whatsappSent = true
            messageID = "pending" // Will be updated with webhook
            d.Auth.TrackInvitationDelivery(inv.ID, "sent", messageID, "")
        }
    }

    // Return response with WhatsApp status
    response := map[string]any{
        "invitation": inv,
        "invite_token": token,
        "invite_url": "/dashboard/invitation/" + token,
        "whatsapp_sent": whatsappSent,
    }
    
    if !whatsappSent {
        response["whatsapp_error"] = sendErr.Error()
        response["manual_instructions"] = "Share this code manually: " + token
    }

    WriteJSON(w, http.StatusCreated, response)
}
```

#### 3. Storage Layer (`internal/storage/tenant_onboarding.go`)

```go
// TrackInvitationDelivery records WhatsApp message delivery status
func (p *PostgresStore) TrackInvitationDelivery(
    invitationID uuid.UUID,
    status string,
    messageID string,
    errorMessage string,
) error {
    _, err := p.db.Exec(`
        INSERT INTO whatsapp_invitation_delivery 
        (invitation_id, status, message_id, error_message, sent_at)
        VALUES ($1, $2, $3, $4, NOW())
    `, invitationID, status, messageID, errorMessage)
    
    return err
}
```

### Frontend Changes

#### 1. Join Page (`frontend/src/pages/JoinWithCode.tsx`)

```tsx
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useMutation } from '@tanstack/react-query'
import { onboardingApi } from '@/lib/apiClient'

const joinSchema = z.object({
  invitationToken: z.string()
    .min(1, 'Invitation code is required')
    .uuid('Must be a valid invitation code'),
})

type JoinInput = z.infer<typeof joinSchema>

export const Route = createFileRoute('/join')({
  component: JoinWithCodePage,
})

function JoinWithCodePage() {
  const navigate = useNavigate()
  const [error, setError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<JoinInput>({
    resolver: zodResolver(joinSchema),
  })

  const validateToken = useMutation({
    mutationFn: async (token: string) => {
      return await onboardingApi.getInvitationInfo(token)
    },
    onSuccess: (data) => {
      navigate({ 
        to: '/invitation/$token',
        params: { token: data.data.invitation_token }
      })
    },
    onError: (err: any) => {
      setError(err?.response?.data?.message || 'Invalid invitation code')
    },
  })

  const onSubmit = async (data: JoinInput) => {
    await validateToken.mutateAsync(data.invitationToken)
  }

  return (
    // UI implementation as shown in proposal.md
  )
}
```

#### 2. Route Registration (`frontend/src/App.tsx`)

```tsx
// Add import
import { JoinWithCode } from './pages/JoinWithCode'

// Add route
const joinRoute = createFileRoute('/join')({
  component: JoinWithCode,
})

// Add to route tree
export const routeTree = rootRoute.addChildren([
  // Public routes
  loginRoute,
  recoveryRoute,
  signupTenantRoute,
  signupOperatorRoute,
  invitationRoute,
  joinRoute, // ✅ NEW
  // Protected routes...
])
```

#### 3. Navigation Links

Add to `Login.tsx` and `SignupChoice.tsx`:

```tsx
<div className="mt-6 text-center">
  <p className="text-sm text-gray-600">Have an invitation code?</p>
  <Link 
    to="/join" 
    className="text-primary-600 hover:text-primary-700 font-medium"
  >
    Join your team →
  </Link>
</div>
```

---

## Testing Requirements

### Unit Tests

- [ ] WhatsApp message template formatting
- [ ] Phone number normalization
- [ ] Token validation
- [ ] Delivery tracking

### Integration Tests

- [ ] End-to-end invitation flow
- [ ] WhatsApp message sending (mocked)
- [ ] Manual code entry flow
- [ ] Error scenarios

### E2E Tests

```typescript
// e2e/whatsapp-invitation.spec.ts
test('admin creates invitation and WhatsApp sent', async ({ page }) => {
  // Login as admin
  await page.goto('/dashboard/team')
  
  // Click invite button
  await page.click('text=Invite Operator')
  
  // Fill form
  await page.fill('[name="whatsappNumber"]', '+1234567890')
  await page.click('button:has-text("Send Invitation")')
  
  // Verify success message
  await expect(page.locator('text=Invitation sent via WhatsApp')).toBeVisible()
})

// e2e/manual-join.spec.ts
test('user joins with invitation code', async ({ page }) => {
  await page.goto('/join')
  
  // Enter token
  await page.fill('[name="invitationToken"]', 'valid-token-123')
  await page.click('button:has-text("Continue")')
  
  // Should navigate to invitation setup
  await expect(page).toHaveURL(/\/invitation\/.+\/setup/)
})
```

---

## Monitoring & Metrics

### Key Metrics Dashboard

```sql
-- WhatsApp Delivery Rate
SELECT 
  DATE(sent_at) as date,
  COUNT(*) as total_sent,
  COUNT(CASE WHEN status = 'sent' THEN 1 END) as successful,
  ROUND(COUNT(CASE WHEN status = 'sent' THEN 1 END) * 100.0 / COUNT(*), 2) as success_rate
FROM whatsapp_invitation_delivery
WHERE sent_at > NOW() - INTERVAL '30 days'
GROUP BY DATE(sent_at)
ORDER BY date DESC;

-- Invitation Acceptance Rate
SELECT 
  DATE(created_at) as date,
  COUNT(*) as total_invitations,
  COUNT(CASE WHEN status = 'accepted' THEN 1 END) as accepted,
  ROUND(COUNT(CASE WHEN status = 'accepted' THEN 1 END) * 100.0 / COUNT(*), 2) as acceptance_rate
FROM invitations
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

### Alerts

- WhatsApp delivery rate < 90% → Page on-call
- Invitation acceptance rate < 60% → Notify product team
- Error rate spike > 20% → Investigate immediately

---

## Rollback Plan

If issues occur:

1. **Disable WhatsApp Sending**:
   ```bash
   # Set env var to disable
   export WHATSAPP_INVITATIONS_ENABLED=false
   ```

2. **Fallback to Manual**:
   - Admins manually share codes
   - Users use /join page

3. **Revert Deployment**:
   ```bash
   git revert <commit-hash>
   ```

---

## Success Criteria

- [ ] WhatsApp messages sent for 100% of invitations
- [ ] Delivery rate > 95%
- [ ] Invitation acceptance rate > 70%
- [ ] Manual join flow working
- [ ] No critical bugs
- [ ] Support tickets reduced by 40%

---

## Related Documents

- [`proposal.md`](proposal.md) - Full implementation proposal
- [`../tenant-onboarding-flow/README.md`](../tenant-onboarding-flow/README.md) - TOTP overview
- [`../tenant-onboarding-flow/UI-MIGRATION-TANSTACK.md`](../tenant-onboarding-flow/UI-MIGRATION-TANSTACK.md) - Frontend architecture

---

**Next Step**: Review proposal.md and begin Phase 1 implementation.
