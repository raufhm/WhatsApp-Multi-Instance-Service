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
// Personal (default-server) identifiers that omit a domain are normalized to
// @s.whatsapp.net.
func NormalizeAddress(address string) (string, error) {
	return NormalizeAddressWithServer(address, "s.whatsapp.net")
}

// NormalizeAddressWithServer normalizes a WhatsApp identifier using the supplied
// server domain. Use this for group addresses (g.us) so that normalized
// identity matching does not collapse a group ID into a personal number.
func NormalizeAddressWithServer(address string, server string) (string, error) {
	address = strings.ToLower(strings.TrimSpace(address))
	address = strings.TrimPrefix(address, "+")
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
		address += "@" + server
	}
	parts := strings.Split(address, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", ErrInvalidAddress
	}
	return address, nil
}

// IsLID reports whether the address is a WhatsApp Line Identifier (LID).
// LIDs are 15-digit numeric IDs used when a real phone number is not mapped.
func IsLID(address string) bool {
	address = strings.TrimSpace(address)
	if i := strings.IndexByte(address, '@'); i >= 0 {
		address = address[:i]
	}
	address = strings.TrimPrefix(address, "+")
	if len(address) != 15 {
		return false
	}
	for i := 0; i < len(address); i++ {
		if address[i] < '0' || address[i] > '9' {
			return false
		}
	}
	return true
}

type Service interface {
	UpsertContact(tenantID uuid.UUID, input domain.ContactUpsert) (domain.Contact, error)
	FindOrCreateConversation(key domain.ConversationKey, now time.Time) (domain.Conversation, error)
}
