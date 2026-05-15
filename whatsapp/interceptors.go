package whatsapp

import (
	"log"

	"github.com/raufhm/whatsapp-testing/domain"
)

type MultiDispatcher struct {
	dispatchers []domain.Dispatcher
}

func NewMultiDispatcher(dispatchers ...domain.Dispatcher) *MultiDispatcher {
	return &MultiDispatcher{dispatchers: dispatchers}
}

func (m *MultiDispatcher) DispatchMessage(meta domain.MessageMetadata) {
	for _, d := range m.dispatchers {
		d.DispatchMessage(meta)
	}
}

func (m *MultiDispatcher) DispatchReceipt(receipt domain.Receipt) {
	for _, d := range m.dispatchers {
		d.DispatchReceipt(receipt)
	}
}

func (m *MultiDispatcher) DispatchEvent(event domain.InstanceEvent) {
	for _, d := range m.dispatchers {
		d.DispatchEvent(event)
	}
}

func (m *MultiDispatcher) UpdateGroup(group domain.GroupInfo) {
	for _, d := range m.dispatchers {
		d.UpdateGroup(group)
	}
}

type LoggerDispatcher struct{}

func (l *LoggerDispatcher) DispatchMessage(meta domain.MessageMetadata) {
	log.Printf("[%s] [%s] %s -> %s | Type: %s | Content: %s | Status: %s",
		meta.Direction, meta.HostID, meta.Sender, meta.Recipient, meta.Type, meta.Content, meta.Status)
}

func (l *LoggerDispatcher) DispatchReceipt(receipt domain.Receipt) {
	log.Printf("RECEIPT [%s]: %s -> %s | Status: %s",
		receipt.WhatsappID, receipt.Sender, receipt.Recipient, receipt.Status)
}

func (l *LoggerDispatcher) DispatchEvent(event domain.InstanceEvent) {
	log.Printf("EVENT [%s]: %s - %s", event.Status, event.HostID, event.Message)
}

func (l *LoggerDispatcher) UpdateGroup(group domain.GroupInfo) {
	log.Printf("GROUP UPDATE: %s (%s) | Participants: %d", group.Name, group.GroupID, group.ParticipantCount)
}

// Deprecated: Use MultiDispatcher
type MultiInterceptor struct {
	MultiDispatcher
}

// Deprecated: Use LoggerDispatcher
type LoggerInterceptor struct {
	LoggerDispatcher
}
