package whatsapp

import (
	"fmt"
	"time"

	"github.com/raufhm/whatsapp-testing/domain"
)

// BuildInvitationMessage formats a WhatsApp invitation message with instructions and TOTP onboarding details.
func BuildInvitationMessage(tenant domain.Tenant, inv domain.Invitation, token string, inviterName string) string {
	inviter := inviterName
	if inviter == "" {
		inviter = "Your Administrator"
	}

	hoursRemaining := int(time.Until(inv.ExpiresAt).Hours())
	if hoursRemaining < 0 {
		hoursRemaining = 0
	}

	return fmt.Sprintf(
		"👋 Hello!\n\n"+
			"You have been invited by *%s* to join *%s* on *whops*.\n\n"+
			"📋 *Your Role:* %s\n"+
			"🔑 *Your Invitation Code:* `%s`\n\n"+
			"To get started and set up your account:\n"+
			"1. Visit the join page in your browser: /dashboard/invitation/%s\n"+
			"2. Scan your personal TOTP QR code using Google Authenticator, Authy, or 1Password\n"+
			"3. Enter the 6-digit verification code to complete setup\n\n"+
			"⚠️ *Important:* This invitation will expire in %d hours. Do not share this code with anyone.",
		inviter,
		tenant.Name,
		inv.Role,
		token,
		token,
		hoursRemaining,
	)
}
