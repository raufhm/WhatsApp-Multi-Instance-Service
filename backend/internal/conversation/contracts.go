package conversation

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whatsapp-testing/domain"
)

var ErrInvalidAddress = errors.New("invalid WhatsApp address")

// NormalizeAddress returns a stable address for tenant-local identity matching.
// Device suffixes are removed while group addresses remain distinct addresses.
func NormalizeAddress(address string) (string, error) {
	address = strings.ToLower(strings.TrimSpace(address))
	if address == "" {
		return "", ErrInvalidAddress
	}
	if colon := strings.IndexByte(address, ':'); colon >= 0 {
		at := strings.IndexByte(address[colon:], '@')
		if at >= 0 {
			address = address[:colon] + address[colon+at:]
		} else {
			address = address[:colon]
		}
	}
	if !strings.Contains(address, "@") {
		address += "@s.whatsapp.net"
	}
	parts := strings.Split(address, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", ErrInvalidAddress
	}
	return address, nil
}

type Service interface {
	UpsertContact(tenantID uuid.UUID, input domain.ContactUpsert) (domain.Contact, error)
	FindOrCreateConversation(key domain.ConversationKey, now time.Time) (domain.Conversation, error)
}
