package whatsapp

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whatsapp-testing/domain"
)

func TestBuildInvitationMessage(t *testing.T) {
	tenant := domain.Tenant{
		ID:   uuid.New(),
		Name: "Acme Corp",
	}
	inv := domain.Invitation{
		ID:        uuid.New(),
		Role:      "OPERATOR",
		ExpiresAt: time.Now().Add(48 * time.Hour),
	}
	token := "abc-123-token"

	msg := BuildInvitationMessage(tenant, inv, token, "Alice Admin")

	if !strings.Contains(msg, "Alice Admin") {
		t.Errorf("expected message to contain inviter name, got: %s", msg)
	}
	if !strings.Contains(msg, "Acme Corp") {
		t.Errorf("expected message to contain tenant name, got: %s", msg)
	}
	if !strings.Contains(msg, "OPERATOR") {
		t.Errorf("expected message to contain role, got: %s", msg)
	}
	if !strings.Contains(msg, "abc-123-token") {
		t.Errorf("expected message to contain token, got: %s", msg)
	}
	if !strings.Contains(msg, "/dashboard/invitation/abc-123-token") {
		t.Errorf("expected message to contain invitation URL, got: %s", msg)
	}
}
