package whatsapp

import "github.com/raufhm/whatsapp-testing/domain"

// BotSender adapts the application bot to the provider's bounded outbound
// queue. The manager is assigned after construction to avoid a dispatcher
// construction cycle.
type BotSender struct {
	manager *WhatsAppManager
}

func NewBotSender() *BotSender { return &BotSender{} }

func (s *BotSender) SetManager(manager *WhatsAppManager) { s.manager = manager }

func (s *BotSender) SendMessage(host string, request domain.MessageRequest) error {
	if s.manager == nil {
		return ErrManagerUnavailable
	}
	return s.manager.SendMessageRequest(host, request)
}
