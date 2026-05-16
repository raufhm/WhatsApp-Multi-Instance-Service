package whatsapp

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/raufhm/whatsapp-testing/domain"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

type WhatsAppInstance struct {
	Client      *whatsmeow.Client
	Queue       chan domain.MessageRequest
	HostJID     types.JID
	HostPhone   string
	Manager     *WhatsAppManager
	IsConnected bool
}

type WhatsAppManager struct {
	Instances   map[string]*WhatsAppInstance
	Container   *sqlstore.Container
	Dispatcher  domain.Dispatcher
	LIDResolver domain.LIDResolver
	mu          sync.RWMutex
}

func NewWhatsAppManager(container *sqlstore.Container, dispatcher domain.Dispatcher, lid domain.LIDResolver) *WhatsAppManager {
	return &WhatsAppManager{
		Instances:   make(map[string]*WhatsAppInstance),
		Container:   container,
		Dispatcher:  dispatcher,
		LIDResolver: lid,
	}
}

func (wm *WhatsAppManager) Start() error {
	devices, err := wm.Container.GetAllDevices(context.Background())
	if err != nil {
		return err
	}
	for _, device := range devices {
		_ = wm.SpawnInstance(device)
	}
	return nil
}

func (wm *WhatsAppManager) SpawnInstance(device *store.Device) error {
	client := whatsmeow.NewClient(device, waLog.Stdout("Client", "WARN", true))
	instance := &WhatsAppInstance{
		Client:  client,
		Queue:   make(chan domain.MessageRequest, 100),
		Manager: wm,
	}
	client.AddEventHandler(instance.eventHandler)

	if err := client.Connect(); err != nil {
		wm.notifyStatus(device.ID.User, domain.StatusError, fmt.Sprintf("Connection failed: %v", err))
		return err
	}

	if client.Store.ID != nil {
		instance.HostJID = *client.Store.ID
		instance.HostPhone = instance.resolvePhone(instance.HostJID)
		instance.IsConnected = true

		// Ensure unique instance per host: Stop old if exists
		wm.StopInstance(instance.HostPhone)

		wm.mu.Lock()
		wm.Instances[instance.HostPhone] = instance
		wm.mu.Unlock()
		go instance.worker()
		wm.notifyStatus(instance.HostPhone, domain.StatusOnline, "Connected")
	}
	return nil
}

func (wm *WhatsAppManager) StopInstance(host string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if instance, ok := wm.Instances[host]; ok {
		instance.IsConnected = false
		instance.Client.Disconnect()
		delete(wm.Instances, host)
	}
}

func (wm *WhatsAppManager) notifyStatus(host string, status domain.InstanceStatus, msg string) {
	if wm.Dispatcher != nil {
		wm.Dispatcher.DispatchEvent(domain.InstanceEvent{
			HostID:    host,
			Status:    status,
			Message:   msg,
			Timestamp: time.Now(),
		})
	}
}

func (i *WhatsAppInstance) resolvePhone(jid types.JID) string {
	if jid.IsEmpty() {
		return ""
	}
	if jid.Server == types.DefaultUserServer {
		return jid.User
	}
	if jid.Server == types.HiddenUserServer {
		pn, err := i.Client.Store.LIDs.GetPNForLID(context.Background(), jid)
		if err == nil && pn.User != "" {
			return pn.User
		}
	}
	return jid.User
}

func (i *WhatsAppInstance) worker() {
	for req := range i.Queue {
		if !i.IsConnected {
			log.Printf("[%s] Instance offline, dropping message", i.HostPhone)
			continue
		}
		i.processRequest(req)
	}
}

func (i *WhatsAppInstance) processRequest(req domain.MessageRequest) {
	recipientJID := types.NewJID(req.Recipient, types.DefaultUserServer)
	if req.IsGroup {
		recipientJID.Server = types.GroupServer
	}

	// Human Simulation: Presence
	i.Client.SendChatPresence(context.Background(), recipientJID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)

	var err error
	var resp whatsmeow.SendResponse

	switch req.Type {
	case domain.Image, domain.File:
		resp, err = i.sendMedia(recipientJID, req)
	case domain.Reaction:
		resp, err = i.sendReaction(recipientJID, req)
	default:
		msg := &waE2E.Message{Conversation: proto.String(req.Message)}
		resp, err = i.Client.SendMessage(context.Background(), recipientJID, msg)
	}

	i.Client.SendChatPresence(context.Background(), recipientJID, types.ChatPresencePaused, types.ChatPresenceMediaText)

	if err == nil && i.Manager.Dispatcher != nil {
		i.Manager.Dispatcher.DispatchMessage(domain.MessageMetadata{
			WhatsappID: resp.ID,
			HostID:     i.HostPhone, Sender: i.HostPhone, Recipient: req.Recipient,
			Content: req.Message, IsGroup: req.IsGroup, Direction: domain.Outgoing,
			Type: req.Type, Status: domain.StatusSent, ReactionTarget: req.ReactionTarget, Timestamp: time.Now(),
		})
	} else if err != nil {
		log.Printf("[%s] Send error: %v", i.HostPhone, err)
		if i.Manager.Dispatcher != nil {
			i.Manager.Dispatcher.DispatchMessage(domain.MessageMetadata{
				HostID: i.HostPhone, Sender: i.HostPhone, Recipient: req.Recipient,
				Content: req.Message, Direction: domain.Outgoing, Status: domain.StatusFailed,
				Timestamp: time.Now(),
			})
		}
	}

	// Inter-message jitter
	time.Sleep(time.Duration(1500+rand.Intn(2000)) * time.Millisecond)
}

func (i *WhatsAppInstance) sendReaction(recipient types.JID, req domain.MessageRequest) (whatsmeow.SendResponse, error) {
	msg := i.Client.BuildReaction(recipient, types.EmptyJID, req.ReactionTarget, req.Message)
	return i.Client.SendMessage(context.Background(), recipient, msg)
}

func (i *WhatsAppInstance) sendMedia(recipient types.JID, req domain.MessageRequest) (whatsmeow.SendResponse, error) {
	data, err := os.ReadFile(req.MediaPath)
	if err != nil {
		return whatsmeow.SendResponse{}, err
	}

	var uploadResp whatsmeow.UploadResponse
	if req.Type == domain.Image {
		uploadResp, err = i.Client.Upload(context.Background(), data, whatsmeow.MediaImage)
	} else {
		uploadResp, err = i.Client.Upload(context.Background(), data, whatsmeow.MediaDocument)
	}

	if err != nil {
		return whatsmeow.SendResponse{}, err
	}

	var msg *waE2E.Message
	if req.Type == domain.Image {
		msg = &waE2E.Message{ImageMessage: &waE2E.ImageMessage{
			Caption:       proto.String(req.Message),
			URL:           proto.String(uploadResp.URL),
			DirectPath:    proto.String(uploadResp.DirectPath),
			MediaKey:      uploadResp.MediaKey,
			Mimetype:      proto.String(http.DetectContentType(data)),
			FileEncSHA256: uploadResp.FileEncSHA256,
			FileSHA256:    uploadResp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
		}}
	} else {
		msg = &waE2E.Message{DocumentMessage: &waE2E.DocumentMessage{
			Caption:       proto.String(req.Message),
			Title:         proto.String(req.Message),
			URL:           proto.String(uploadResp.URL),
			DirectPath:    proto.String(uploadResp.DirectPath),
			MediaKey:      uploadResp.MediaKey,
			Mimetype:      proto.String(http.DetectContentType(data)),
			FileEncSHA256: uploadResp.FileEncSHA256,
			FileSHA256:    uploadResp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
		}}
	}

	return i.Client.SendMessage(context.Background(), recipient, msg)
}

func (i *WhatsAppInstance) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		go i.handleIncomingMessage(v)
		if v.Info.IsGroup {
			go i.handleGroupInfo(v.Info.Chat)
		}
	case *events.Receipt:
		go i.handleReceipt(v)
	case *events.HistorySync:
		go i.handleHistorySync(v)
	case *events.LoggedOut:
		i.IsConnected = false
		log.Printf("[%s] Logged out: %s", i.HostPhone, v.Reason)
		i.Manager.notifyStatus(i.HostPhone, domain.StatusOffline, fmt.Sprintf("Logged out: %s", v.Reason))
		i.Manager.StopInstance(i.HostPhone)
	case *events.Connected:
		i.IsConnected = true
		i.Manager.notifyStatus(i.HostPhone, domain.StatusOnline, "Connected/Reconnected")
	case *events.IdentityChange:
		log.Printf("[%s] Identity change for %s", i.HostPhone, v.JID)
	case *events.JoinedGroup:
		go i.handleGroupInfo(v.GroupInfo.JID)
	case *events.GroupInfo:
		go i.handleGroupInfo(v.JID)
	}
}

func (i *WhatsAppInstance) handleReceipt(v *events.Receipt) {
	if i.Manager.Dispatcher == nil {
		return
	}

	status := domain.StatusSent
	switch v.Type {
	case events.ReceiptTypeDelivered:
		status = domain.StatusDelivered
	case events.ReceiptTypeRead, events.ReceiptTypeReadSelf:
		status = domain.StatusRead
	case events.ReceiptTypePlayed:
		status = domain.StatusRead
	default:
		// Ignore other types like "inactive" for now
		return
	}

	// For receipts, Sender is the one who sent the receipt (the recipient of the message)
	senderPhone := i.resolvePhone(v.Sender)
	chatID := senderPhone
	if v.IsGroup {
		chatID = v.Chat.User
	}

	for _, id := range v.MessageIDs {
		i.Manager.Dispatcher.DispatchReceipt(domain.Receipt{
			WhatsappID: id,
			Sender:     i.HostPhone,
			Recipient:  chatID,
			Status:     status,
			Timestamp:  v.Timestamp,
		})
	}
}

func (i *WhatsAppInstance) handleGroupInfo(groupJID types.JID) {
	if i.Manager.Dispatcher == nil {
		return
	}

	info, err := i.Client.GetGroupInfo(context.Background(), groupJID)
	if err != nil {
		log.Printf("[%s] GetGroupInfo error: %v", i.HostPhone, err)
		return
	}

	var participants []string
	var hostsInGroup []string

	i.Manager.mu.RLock()
	allHosts := make(map[string]bool)
	for phone := range i.Manager.Instances {
		allHosts[phone] = true
	}
	i.Manager.mu.RUnlock()

	for _, p := range info.Participants {
		phone := i.resolvePhone(p.JID)
		if phone != "" {
			participants = append(participants, phone)
			if allHosts[phone] {
				hostsInGroup = append(hostsInGroup, phone)
			}
		}
	}

	i.Manager.Dispatcher.UpdateGroup(domain.GroupInfo{
		GroupID:          info.JID.User,
		Name:             info.Name,
		Description:      info.Topic, // Use Topic as Description
		OwnerJID:         info.OwnerJID.String(),
		Participants:     participants,
		Hosts:            hostsInGroup,
		ParticipantCount: len(participants),
		CreatedAt:        info.GroupCreated,
		UpdatedAt:        time.Now(),
	})
}

func (i *WhatsAppInstance) handleHistorySync(v *events.HistorySync) {
	if i.Manager.Dispatcher == nil {
		return
	}

	for _, conv := range v.Data.GetConversations() {
		chatJID, _ := types.ParseJID(conv.GetID())
		if chatJID.IsEmpty() {
			continue
		}

		for _, hMsg := range conv.GetMessages() {
			wi := hMsg.GetMessage()
			if wi == nil {
				continue
			}
			m := wi.GetMessage()
			if m == nil {
				continue
			}

			content := ""
			msgType := domain.Text
			reactionTarget := ""

			if m.GetConversation() != "" {
				content = m.GetConversation()
			} else if m.GetExtendedTextMessage().GetText() != "" {
				content = m.GetExtendedTextMessage().GetText()
			} else if m.GetImageMessage() != nil {
				content = "[Image]"
				msgType = domain.Image
			} else if m.GetDocumentMessage() != nil {
				content = "[Document]"
				msgType = domain.File
			} else if m.GetReactionMessage() != nil {
				content = m.GetReactionMessage().GetText()
				msgType = domain.Reaction
				reactionTarget = m.GetReactionMessage().GetKey().GetID()
			}

			if content == "" {
				continue
			}

			// Identify sender/recipient
			senderJID := chatJID
			direction := domain.Incoming
			if wi.GetKey().GetFromMe() {
				direction = domain.Outgoing
			} else if chatJID.Server == types.GroupServer && wi.GetKey().GetParticipant() != "" {
				p, _ := types.ParseJID(wi.GetKey().GetParticipant())
				if !p.IsEmpty() {
					senderJID = p
				}
			}

			sender := i.resolvePhone(senderJID)
			recipient := i.HostPhone
			if direction == domain.Outgoing {
				recipient = i.resolvePhone(chatJID)
			} else if chatJID.Server == types.GroupServer {
				recipient = chatJID.User
			}

			i.Manager.Dispatcher.DispatchMessage(domain.MessageMetadata{
				WhatsappID: wi.GetKey().GetID(),
				HostID:     i.HostPhone, Sender: sender, Recipient: recipient,
				Content: content, IsGroup: chatJID.Server == types.GroupServer, Direction: direction,
				Type: msgType, Status: domain.StatusSent, ReactionTarget: reactionTarget,
				Timestamp: time.Unix(int64(wi.GetMessageTimestamp()), 0),
			})
		}
	}
}

func (i *WhatsAppInstance) handleIncomingMessage(v *events.Message) {
	log.Printf("[%s] Received message event: %s from %s", i.HostPhone, v.Info.ID, v.Info.Sender)

	content := ""
	msgType := domain.Text
	reactionTarget := ""

	if v.Message.GetConversation() != "" {
		content = v.Message.GetConversation()
	} else if v.Message.GetExtendedTextMessage().GetText() != "" {
		content = v.Message.GetExtendedTextMessage().GetText()
	} else if v.Message.GetImageMessage() != nil {
		content = "[Image]"
		msgType = domain.Image
	} else if v.Message.GetDocumentMessage() != nil {
		content = "[Document]"
		msgType = domain.File
	} else if v.Message.GetReactionMessage() != nil {
		content = v.Message.GetReactionMessage().GetText()
		msgType = domain.Reaction
		reactionTarget = v.Message.GetReactionMessage().GetKey().GetID()
	}

	if content == "" {
		log.Printf("[%s] Empty or unsupported message type for ID %s", i.HostPhone, v.Info.ID)
		return
	}

	sender := i.resolvePhone(v.Info.Sender)
	recipient := i.HostPhone
	direction := domain.Incoming

	if v.Info.IsFromMe {
		direction = domain.Outgoing
		recipient = i.resolvePhone(v.Info.Chat)
	} else if v.Info.IsGroup {
		recipient = v.Info.Chat.User
	}

	if i.Manager.Dispatcher != nil {
		i.Manager.Dispatcher.DispatchMessage(domain.MessageMetadata{
			WhatsappID: v.Info.ID,
			HostID:     i.HostPhone, Sender: sender, Recipient: recipient,
			Content: content, IsGroup: v.Info.IsGroup, Direction: direction,
			Type: msgType, Status: domain.StatusSent, ReactionTarget: reactionTarget, Timestamp: v.Info.Timestamp,
		})
	}
}

func (wm *WhatsAppManager) SendMessageRequest(host string, req domain.MessageRequest) error {
	wm.mu.RLock()
	instance, ok := wm.Instances[host]
	wm.mu.RUnlock()
	if !ok {
		return fmt.Errorf("host %s not found", host)
	}
	instance.Queue <- req
	return nil
}

func (wm *WhatsAppManager) ListInstances() []domain.InstanceInfo {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	var list []domain.InstanceInfo
	for _, i := range wm.Instances {
		status := domain.StatusOnline
		if !i.IsConnected {
			status = domain.StatusOffline
		}
		list = append(list, domain.InstanceInfo{
			HostPhone:   i.HostPhone,
			Status:      status,
			IsConnected: i.IsConnected,
			QueueSize:   len(i.Queue),
		})
	}
	return list
}

func (wm *WhatsAppManager) GetInstance(host string) (domain.InstanceInfo, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	i, ok := wm.Instances[host]
	if !ok {
		return domain.InstanceInfo{}, fmt.Errorf("not found")
	}
	status := domain.StatusOnline
	if !i.IsConnected {
		status = domain.StatusOffline
	}
	return domain.InstanceInfo{
		HostPhone:   i.HostPhone,
		Status:      status,
		IsConnected: i.IsConnected,
		QueueSize:   len(i.Queue),
	}, nil
}
