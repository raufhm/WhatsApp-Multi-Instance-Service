package test

import (
	"sync"
	"testing"
	"time"

	"github.com/raufhm/whatsapp-testing/domain"
)

type MockDispatcher struct {
	mu       sync.Mutex
	Messages []domain.MessageMetadata
	Receipts []domain.Receipt
	Statuses []domain.InstanceEvent
	Groups   []domain.GroupInfo
}

var _ domain.Dispatcher = (*MockDispatcher)(nil)

func (m *MockDispatcher) DispatchMessage(meta domain.MessageMetadata) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, meta)
}

func (m *MockDispatcher) DispatchReceipt(receipt domain.Receipt) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Receipts = append(m.Receipts, receipt)
}

func (m *MockDispatcher) UpdateInstanceStatus(hostID string, status domain.InstanceStatus, isConnected bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Statuses = append(m.Statuses, domain.InstanceEvent{
		HostID:    hostID,
		Status:    status,
		Timestamp: time.Now(),
	})
}

func (m *MockDispatcher) UpdateGroup(group domain.GroupInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Groups = append(m.Groups, group)
}

func TestDispatcherContract(t *testing.T) {
	mock := &MockDispatcher{}

	now := time.Now()
	msg := domain.MessageMetadata{
		WhatsappID: "test-id",
		HostID:     "host-1",
		Content:    "Hello",
		Timestamp:  now,
	}
	mock.DispatchMessage(msg)

	if len(mock.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(mock.Messages))
	}
	if mock.Messages[0].WhatsappID != "test-id" {
		t.Errorf("Expected test-id, got %s", mock.Messages[0].WhatsappID)
	}

	mock.UpdateInstanceStatus("host-1", domain.StatusOnline, true)

	if len(mock.Statuses) != 1 {
		t.Errorf("Expected 1 status event, got %d", len(mock.Statuses))
	}
	if mock.Statuses[0].HostID != "host-1" {
		t.Errorf("Expected host-1, got %s", mock.Statuses[0].HostID)
	}
}
