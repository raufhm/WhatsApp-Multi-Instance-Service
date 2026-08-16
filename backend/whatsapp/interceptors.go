package whatsapp

import (
	"context"
	"log"
	"time"

	"github.com/raufhm/whatsapp-testing/domain"
	"github.com/raufhm/whatsapp-testing/internal/bot"
)

// AsyncProjector keeps database projection work away from the whatsmeow event
// callback. A small bounded queue provides backpressure without unbounded
// goroutine creation; failed work is retried a few times for transient DB
// errors and remains visible in logs.
type AsyncProjector struct {
	projector domain.ApplicationProjector
	messages  chan domain.MessageMetadata
	receipts  chan domain.Receipt
	bot       *bot.Processor
}

func NewAsyncProjector(projector domain.ApplicationProjector, capacity int) *AsyncProjector {
	if capacity < 1 {
		capacity = 1
	}
	a := &AsyncProjector{projector: projector, messages: make(chan domain.MessageMetadata, capacity), receipts: make(chan domain.Receipt, capacity)}
	go a.run()
	return a
}

func NewAsyncProjectorWithBot(projector domain.ApplicationProjector, processor *bot.Processor, capacity int) *AsyncProjector {
	if capacity < 1 {
		capacity = 1
	}
	a := &AsyncProjector{projector: projector, messages: make(chan domain.MessageMetadata, capacity), receipts: make(chan domain.Receipt, capacity), bot: processor}
	go a.run()
	return a
}

func (a *AsyncProjector) DispatchMessage(meta domain.MessageMetadata) {
	a.messages <- meta
}
func (a *AsyncProjector) DispatchReceipt(receipt domain.Receipt) {
	a.receipts <- receipt
}

func (a *AsyncProjector) UpdateInstanceStatus(string, domain.InstanceStatus, bool) {}
func (a *AsyncProjector) UpdateGroup(domain.GroupInfo)                             {}
func (a *AsyncProjector) run() {
	for {
		select {
		case m := <-a.messages:
			a.retryMessage(m)
		case r := <-a.receipts:
			a.retryReceipt(r)
		}
	}
}
func (a *AsyncProjector) retryMessage(m domain.MessageMetadata) {
	for attempt := 0; attempt < 3; attempt++ {
		if err := a.projectMessage(m); err == nil {
			return
		}
		time.Sleep(time.Duration(attempt+1) * 25 * time.Millisecond)
	}
	log.Printf("application message projection failed: %s", m.WhatsappID)
}

func (a *AsyncProjector) projectMessage(m domain.MessageMetadata) error {
	if contextual, ok := a.projector.(domain.ContextProjector); ok {
		projected, err := contextual.ProjectMessageContext(m)
		if err != nil {
			return err
		}
		if a.bot != nil && projected.New && projected.Inbound && projected.BotEligible {
			_, err = a.bot.Handle(context.Background(), bot.Event{
				TenantID: projected.TenantID, ConversationID: projected.ConversationID,
				Host: projected.Host, Recipient: projected.Recipient, Text: projected.Text,
				At: projected.At, Eligible: projected.BotEligible, EligibilityKnown: true,
			})
			return err
		}
		return nil
	}
	return a.projector.ProjectMessage(m)
}
func (a *AsyncProjector) retryReceipt(r domain.Receipt) {
	for attempt := 0; attempt < 3; attempt++ {
		if err := a.projector.ProjectReceipt(r); err == nil {
			return
		}
		time.Sleep(time.Duration(attempt+1) * 25 * time.Millisecond)
	}
	log.Printf("application receipt projection failed: %s", r.WhatsappID)
}

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
