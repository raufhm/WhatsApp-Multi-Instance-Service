package test

import (
	"sync"
	"testing"
	"time"

	"github.com/raufhm/whatsapp-testing/domain"
)

type MockDispatcher struct {
	Messages []domain.MessageMetadata
	Events   []domain.InstanceEvent
	mu       sync.Mutex
}

func (m *MockDispatcher) DispatchMessage(meta domain.MessageMetadata) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, meta)
}

func (m *MockDispatcher) DispatchEvent(event domain.InstanceEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Events = append(m.Events, event)
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

	evt := domain.InstanceEvent{
		HostID:    "host-1",
		Status:    domain.StatusOnline,
		Timestamp: now,
	}

	mock.DispatchEvent(evt)

	if len(mock.Events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(mock.Events))
	}
}
