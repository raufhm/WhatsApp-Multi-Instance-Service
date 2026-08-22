package bot

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
)

type testSender struct {
	requests []domain.MessageRequest
	err      error
}

func (s *testSender) SendMessage(_ string, r domain.MessageRequest) error {
	if s.err != nil {
		return s.err
	}
	s.requests = append(s.requests, r)
	return nil
}

type testStore struct {
	sessions int
	closes   int
	last     domain.ConversationStatus
	rules    []domain.BotRule
}

func (s *testStore) SaveBotSession(uuid.UUID, uuid.UUID, string, map[string]any) error {
	s.sessions++
	return nil
}
func (s *testStore) CloseConversation(_ uuid.UUID, _ uuid.UUID, status domain.ConversationStatus, _ string, _ time.Time) (domain.Activity, error) {
	s.closes++
	s.last = status
	return domain.Activity{Status: domain.ActivityPending}, nil
}
func (s *testStore) GetActiveBotRuleSet(uuid.UUID) (domain.BotRuleSet, error) {
	if len(s.rules) > 0 {
		return domain.BotRuleSet{Version: 1, Rules: s.rules}, nil
	}
	return domain.BotRuleSet{}, errors.New("no rules")
}

func TestEngineMatchingAndFallback(t *testing.T) {
	e := NewEngine(Config{Fallback: "fallback", Rules: []Rule{{Name: "help", Pattern: "HELP", Match: MatchExact, Response: "How can I help?"}}})
	if got := e.Evaluate(" help ").Response; got != "How can I help?" {
		t.Fatalf("response = %q", got)
	}
	if got := e.Evaluate("unknown").Response; got != "fallback" {
		t.Fatalf("fallback = %q", got)
	}
}

func TestProcessorQueuesBotReplyAndClosesOnce(t *testing.T) {
	sender := &testSender{}
	store := &testStore{}
	p := NewProcessor(NewEngine(Config{Rules: []Rule{{Name: "bye", Pattern: "bye", Response: "Goodbye", Match: MatchExact, Terminal: true}}}), sender, store)
	e := Event{TenantID: uuid.New(), ConversationID: uuid.New(), Host: "host", Recipient: "123", Text: "bye", At: time.Now()}
	decision, err := p.Handle(context.Background(), e)
	if err != nil || !decision.Terminal {
		t.Fatalf("decision=%+v err=%v", decision, err)
	}
	if len(sender.requests) != 1 || sender.requests[0].Actor != domain.ActorBot {
		t.Fatalf("requests=%+v", sender.requests)
	}
	if store.sessions != 1 || store.closes != 1 || store.last != domain.ConversationClosed {
		t.Fatalf("store=%+v", store)
	}
}

func TestProcessorDoesNotPersistAfterSendFailure(t *testing.T) {
	sender := &testSender{err: errors.New("offline")}
	store := &testStore{}
	p := NewProcessor(NewEngine(Config{Fallback: "hello"}), sender, store)
	_, err := p.Handle(context.Background(), Event{ConversationID: uuid.New(), Host: "host", Recipient: "123", Text: "x", At: time.Now()})
	if err == nil || store.sessions != 0 {
		t.Fatalf("err=%v store=%+v", err, store)
	}
}

func TestProcessorUsesDatabaseRules(t *testing.T) {
	sender := &testSender{}
	store := &testStore{
		rules: []domain.BotRule{
			{Name: "db-hi", Pattern: "hello", Match: "EXACT", Response: "Hello from DB!", Enabled: true},
		},
	}
	p := NewProcessor(NewEngine(Config{Fallback: "fallback"}), sender, store)
	e := Event{TenantID: uuid.New(), ConversationID: uuid.New(), Host: "host", Recipient: "123", Text: "hello", At: time.Now()}
	_, err := p.Handle(context.Background(), e)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(sender.requests) != 1 || sender.requests[0].Message != "Hello from DB!" {
		t.Fatalf("expected DB response, got requests: %+v", sender.requests)
	}
}
