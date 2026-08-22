package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
)

type MatchType string

const (
	MatchContains MatchType = "CONTAINS"
	MatchExact    MatchType = "EXACT"
	MatchPrefix   MatchType = "PREFIX"
)

type Rule struct {
	Name, Pattern, Response string
	Match                   MatchType
	Terminal, Handoff       bool
}

type Config struct {
	Rules          []Rule
	Fallback       string
	RuleVersion    string
	SessionTimeout time.Duration
}

type Decision struct {
	Response string
	Rule     Rule
	Terminal bool
	Handoff  bool
}

type Engine struct{ config Config }

func NewEngine(config Config) *Engine {
	if config.RuleVersion == "" {
		config.RuleVersion = "default"
	}
	return &Engine{config: config}
}

func (e *Engine) Evaluate(text string) Decision {
	value := strings.ToLower(strings.TrimSpace(text))
	for _, rule := range e.config.Rules {
		pattern := strings.ToLower(strings.TrimSpace(rule.Pattern))
		matched := false
		switch rule.Match {
		case MatchExact:
			matched = value == pattern
		case MatchPrefix:
			matched = strings.HasPrefix(value, pattern)
		default:
			matched = pattern != "" && strings.Contains(value, pattern)
		}
		if matched {
			return Decision{Response: rule.Response, Rule: rule, Terminal: rule.Terminal, Handoff: rule.Handoff}
		}
	}
	return Decision{Response: e.config.Fallback}
}

type Sender interface {
	SendMessage(host string, request domain.MessageRequest) error
}
type SessionStore interface {
	SaveBotSession(tenantID, conversationID uuid.UUID, version string, state map[string]any) error
	CloseConversation(tenantID, conversationID uuid.UUID, status domain.ConversationStatus, reason string, at time.Time) (domain.Activity, error)
	GetActiveBotRuleSet(tenantID uuid.UUID) (domain.BotRuleSet, error)
}

type Event struct {
	TenantID, ConversationID   uuid.UUID
	Host, Recipient, Text      string
	At                         time.Time
	Eligible, EligibilityKnown bool
}

var ErrNoRecipient = errors.New("bot recipient is empty")

// Processor serializes events per conversation. This prevents a slow send from
// allowing a later inbound message to observe stale bot state.
type Processor struct {
	engine *Engine
	sender Sender
	store  SessionStore
	mu     sync.Mutex
	locks  map[uuid.UUID]*sync.Mutex
}

func NewProcessor(engine *Engine, sender Sender, store SessionStore) *Processor {
	return &Processor{engine: engine, sender: sender, store: store, locks: make(map[uuid.UUID]*sync.Mutex)}
}

func (p *Processor) lock(id uuid.UUID) *sync.Mutex {
	p.mu.Lock()
	defer p.mu.Unlock()
	if l := p.locks[id]; l != nil {
		return l
	}
	l := &sync.Mutex{}
	p.locks[id] = l
	return l
}

func mapRules(domainRules []domain.BotRule) []Rule {
	rules := make([]Rule, 0, len(domainRules))
	for _, r := range domainRules {
		if !r.Enabled {
			continue
		}
		rules = append(rules, Rule{
			Name:     r.Name,
			Pattern:  r.Pattern,
			Response: r.Response,
			Match:    MatchType(r.Match),
			Terminal: r.Terminal,
			Handoff:  r.Handoff,
		})
	}
	return rules
}

func (p *Processor) Handle(ctx context.Context, event Event) (Decision, error) {
	if err := ctx.Err(); err != nil {
		return Decision{}, err
	}
	if event.Recipient == "" {
		return Decision{}, ErrNoRecipient
	}
	if event.EligibilityKnown && !event.Eligible {
		return Decision{}, nil
	}
	l := p.lock(event.ConversationID)
	l.Lock()
	defer l.Unlock()

	var d Decision
	ruleVersion := p.engine.config.RuleVersion
	if ruleVersion == "" {
		ruleVersion = "default"
	}
	if p.store != nil {
		if ruleSet, err := p.store.GetActiveBotRuleSet(event.TenantID); err == nil && len(ruleSet.Rules) > 0 {
			ruleVersion = fmt.Sprintf("db:v%d", ruleSet.Version)
			tempEngine := NewEngine(Config{Rules: mapRules(ruleSet.Rules), Fallback: p.engine.config.Fallback})
			d = tempEngine.Evaluate(event.Text)
		} else {
			d = p.engine.Evaluate(event.Text)
		}
	} else {
		d = p.engine.Evaluate(event.Text)
	}

	if d.Response != "" {
		err := p.sender.SendMessage(event.Host, domain.MessageRequest{Recipient: event.Recipient, Message: d.Response, Type: domain.Text, Actor: domain.ActorBot})
		if err != nil {
			return d, err
		}
	}
	state := map[string]any{"last_rule": d.Rule.Name, "last_at": event.At.UTC().Format(time.RFC3339Nano)}
	if p.store != nil {
		if err := p.store.SaveBotSession(event.TenantID, event.ConversationID, ruleVersion, state); err != nil {
			return d, err
		}
		if d.Terminal || d.Handoff {
			status := domain.ConversationClosed
			reason := "terminal rule"
			if d.Handoff {
				status = domain.ConversationHandedOff
				reason = "handoff rule"
			}
			_, err := p.store.CloseConversation(event.TenantID, event.ConversationID, status, reason, event.At)
			return d, err
		}
	}
	return d, nil
}
