package whatsapp

import (
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
	"github.com/raufhm/whops/internal/bot"
)

type projectorProbe struct {
	mu       sync.Mutex
	messages []domain.MessageMetadata
	receipts []domain.Receipt
}

type botProjectionProbe struct {
	mu       sync.Mutex
	requests []domain.MessageRequest
}

func (p *botProjectionProbe) ProjectMessage(domain.MessageMetadata) error { return nil }
func (p *botProjectionProbe) ProjectReceipt(domain.Receipt) error         { return nil }
func (p *botProjectionProbe) ProjectMessageContext(domain.MessageMetadata) (domain.ProjectedMessage, error) {
	return domain.ProjectedMessage{TenantID: uuid.New(), ConversationID: uuid.New(), Host: "host", Recipient: "1555", Text: "hello", At: time.Now(), Inbound: true, New: true, BotEligible: true}, nil
}
func (p *botProjectionProbe) SendMessage(_ string, request domain.MessageRequest) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.requests = append(p.requests, request)
	return nil
}
func (p *botProjectionProbe) SaveBotSession(uuid.UUID, uuid.UUID, string, map[string]any) error {
	return nil
}
func (p *botProjectionProbe) CloseConversation(uuid.UUID, uuid.UUID, domain.ConversationStatus, string, time.Time) (domain.Activity, error) {
	return domain.Activity{}, nil
}
func (p *botProjectionProbe) GetActiveBotRuleSet(uuid.UUID) (domain.BotRuleSet, error) {
	return domain.BotRuleSet{}, sql.ErrNoRows
}

func (p *projectorProbe) ProjectMessage(m domain.MessageMetadata) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages = append(p.messages, m)
	return nil
}
func (p *projectorProbe) ProjectReceipt(r domain.Receipt) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.receipts = append(p.receipts, r)
	return nil
}

func TestAsyncProjectorProjectsMessagesAndReceipts(t *testing.T) {
	probe := &projectorProbe{}
	projector := NewAsyncProjector(probe, 2)
	projector.DispatchMessage(domain.MessageMetadata{WhatsappID: "m-1"})
	projector.DispatchReceipt(domain.Receipt{WhatsappID: "m-1"})

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		probe.mu.Lock()
		ready := len(probe.messages) == 1 && len(probe.receipts) == 1
		probe.mu.Unlock()
		if ready {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("projection worker did not process both events")
}

func TestAsyncProjectorTriggersBotAndPreservesBotActor(t *testing.T) {
	probe := &botProjectionProbe{}
	processor := bot.NewProcessor(bot.NewEngine(bot.Config{Fallback: "reply"}), probe, probe)
	projector := NewAsyncProjectorWithBot(probe, processor, 2)
	projector.DispatchMessage(domain.MessageMetadata{WhatsappID: "in-1", Direction: domain.Incoming, Content: "hello"})

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		probe.mu.Lock()
		if len(probe.requests) == 1 {
			request := probe.requests[0]
			probe.mu.Unlock()
			if request.Actor != domain.ActorBot || request.Message != "reply" {
				t.Fatalf("bot request = %+v", request)
			}
			return
		}
		probe.mu.Unlock()
		time.Sleep(time.Millisecond)
	}
	t.Fatal("bot was not triggered by projected inbound message")
}

var _ bot.Sender = (*botProjectionProbe)(nil)
var _ bot.SessionStore = (*botProjectionProbe)(nil)
var _ domain.ContextProjector = (*botProjectionProbe)(nil)
var _ domain.ApplicationProjector = (*botProjectionProbe)(nil)
