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

func (m *MultiDispatcher) UpdateInstanceStatus(hostID string, status domain.InstanceStatus, isConnected bool) {
	for _, d := range m.dispatchers {
		d.UpdateInstanceStatus(hostID, status, isConnected)
	}
}

func (m *MultiDispatcher) UpdateGroup(group domain.GroupInfo) {
	for _, d := range m.dispatchers {
		d.UpdateGroup(group)
	}
}

type LoggerDispatcher struct{}

func (l *LoggerDispatcher) DispatchMessage(meta domain.MessageMetadata) {
	log.Printf("[%s] [%s] MessageID: %s | Type: %s | Status: %s",
		meta.Direction, meta.HostID, meta.WhatsappID, meta.Type, meta.Status)
}

func (l *LoggerDispatcher) DispatchReceipt(receipt domain.Receipt) {}

func (l *LoggerDispatcher) UpdateInstanceStatus(hostID string, status domain.InstanceStatus, isConnected bool) {
	log.Printf("STATUS [%s]: %s", hostID, status)
}

func (l *LoggerDispatcher) UpdateGroup(group domain.GroupInfo) {
	log.Printf("GROUP UPDATE: %s (%s)", group.Name, group.GroupID)
}
