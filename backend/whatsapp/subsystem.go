package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/raufhm/whatsapp-testing/domain"
	"github.com/raufhm/whatsapp-testing/internal/storage"
	"github.com/raufhm/whatsapp-testing/internal/upload"
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
	S3          *storage.S3Storage
	IsConnected bool

	groupInfoMu sync.Mutex
	groupInfoAt map[string]time.Time
}

type WhatsAppManager struct {
	Instances     map[string]*WhatsAppInstance
	Container     *sqlstore.Container
	Dispatcher    domain.Dispatcher
	S3            *storage.S3Storage
	MediaStore    storage.MediaStore
	S3ObjectURL   string
	Store         domain.PlatformRepository
	ResolveTenant func(hostID string) uuid.UUID
	// Uploader drives the durable, retryable S3 archive upload for outgoing
	// media. When nil, media messages are sent without an S3 archive job.
	Uploader *upload.Manager
	Pairing  *PairingManager
	mu       sync.RWMutex
}

var ErrManagerUnavailable = fmt.Errorf("WhatsApp manager unavailable")

func NewWhatsAppManager(container *sqlstore.Container, dispatcher domain.Dispatcher, s3Store *storage.S3Storage) *WhatsAppManager {
	var mediaStore storage.MediaStore
	if s3Store != nil {
		mediaStore = s3Store
	}
	wm := &WhatsAppManager{
		Instances:  make(map[string]*WhatsAppInstance),
		Container:  container,
		Dispatcher: dispatcher,
		S3:         s3Store,
		MediaStore: mediaStore,
	}
	wm.Pairing = NewPairingManager(wm, nil)
	return wm
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
	client := whatsmeow.NewClient(device, waLog.Stdout("Client", "ERROR", true))
	instance := &WhatsAppInstance{
		Client:      client,
		Queue:       make(chan domain.MessageRequest, 100),
		Manager:     wm,
		S3:          wm.S3,
		groupInfoAt: make(map[string]time.Time),
	}
	client.AddEventHandler(instance.eventHandler)

	if err := client.Connect(); err != nil {
		wm.notifyStatus(device.ID.User, domain.StatusError, false)
		return err
	}

	if client.Store.ID != nil {
		instance.HostJID = *client.Store.ID
		instance.HostPhone = instance.resolvePhone(instance.HostJID)
		instance.IsConnected = true

		// Ensure unique instance per host: stop old if exists.
		wm.StopInstance(instance.HostPhone)

		wm.mu.Lock()
		wm.Instances[instance.HostPhone] = instance
		wm.mu.Unlock()
		go instance.worker()
		wm.notifyStatus(instance.HostPhone, domain.StatusOnline, true)
		go instance.syncJoinedGroups()
	}

	return nil
}

// removeInstance removes an instance before calling into whatsmeow. This is
// important because disconnect/logout can trigger callbacks that call back into
// the manager.
func (wm *WhatsAppManager) removeInstance(host string) *WhatsAppInstance {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	instanceKey := host
	instance, ok := wm.Instances[host]
	if !ok {
		wanted := SanitizePhoneNumber(host)
		for key, candidate := range wm.Instances {
			if SanitizePhoneNumber(key) == wanted {
				instanceKey = key
				instance = candidate
				break
			}
		}
	}
	if instance == nil {
		return nil
	}
	instance.IsConnected = false
	delete(wm.Instances, instanceKey)
	return instance
}

func (wm *WhatsAppManager) StopInstance(host string) {
	instance := wm.removeInstance(host)
	if instance != nil && instance.Client != nil {
		instance.Client.Disconnect()
	}
}

func (wm *WhatsAppManager) notifyStatus(host string, status domain.InstanceStatus, isConnected bool) {
	if wm.Dispatcher != nil {
		wm.Dispatcher.UpdateInstanceStatus(host, status, isConnected)
	}
}

func ResolveJIDPhone(jid types.JID) string {
	if jid.IsEmpty() {
		return ""
	}
	return jid.User
}

func (i *WhatsAppInstance) resolvePhone(jid types.JID) string {
	if jid.IsEmpty() {
		return ""
	}
	if jid.Server == types.DefaultUserServer {
		return jid.User
	}
	if jid.Server == types.HiddenUserServer && i.Client != nil && i.Client.Store != nil && i.Client.Store.LIDs != nil {
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
			if i.Manager.Dispatcher != nil {
				actor := req.Actor
				if actor == "" {
					actor = domain.ActorOperator
				}
				i.Manager.Dispatcher.DispatchMessage(domain.MessageMetadata{
					HostID: i.HostPhone, Sender: i.HostPhone, Recipient: req.Recipient,
					Content: req.Message, Direction: domain.Outgoing, Type: req.Type,
					Status: domain.StatusFailed, Actor: actor, Timestamp: time.Now(),
				})
			}
			continue
		}
		i.processRequest(req)
	}
}

func (i *WhatsAppInstance) processRequest(req domain.MessageRequest) {
	cleanRecipient := strings.TrimPrefix(strings.TrimSpace(req.Recipient), "+")
	if at := strings.IndexByte(cleanRecipient, '@'); at >= 0 {
		cleanRecipient = cleanRecipient[:at]
	}
	server := types.DefaultUserServer
	if req.IsGroup {
		server = types.GroupServer
	} else {
		cleanRecipient = SanitizePhoneNumber(cleanRecipient)
	}
	recipientJID := types.NewJID(cleanRecipient, server)

	i.Client.SendChatPresence(context.Background(), recipientJID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)

	var err error
	var resp whatsmeow.SendResponse
	var mediaURL string
	var extra []whatsmeow.SendRequestExtra
	if req.ID != "" {
		extra = append(extra, whatsmeow.SendRequestExtra{ID: req.ID})
	}

	switch req.Type {
	case domain.Image, domain.File, domain.Video, domain.Audio:
		resp, mediaURL, err = i.sendMedia(recipientJID, req, extra...)
	case domain.Reaction:
		resp, err = i.sendReaction(recipientJID, req, extra...)
	default:
		msg := &waE2E.Message{Conversation: proto.String(req.Message)}
		resp, err = i.Client.SendMessage(context.Background(), recipientJID, msg, extra...)
	}

	i.Client.SendChatPresence(context.Background(), recipientJID, types.ChatPresencePaused, types.ChatPresenceMediaText)

	actor := req.Actor
	if actor == "" {
		actor = domain.ActorOperator
	}
	sentID := req.ID
	if sentID == "" {
		sentID = resp.ID
	}
	if err == nil && i.Manager.Dispatcher != nil {
		i.Manager.Dispatcher.DispatchMessage(domain.MessageMetadata{
			WhatsappID:     sentID,
			HostID:         i.HostPhone,
			Sender:         i.HostPhone,
			Recipient:      req.Recipient,
			Content:        req.Message,
			IsGroup:        req.IsGroup,
			Direction:      domain.Outgoing,
			Type:           req.Type,
			Status:         domain.StatusSent,
			Actor:          actor,
			OperatorID:     req.OperatorID,
			OperatorName:   req.OperatorName,
			MediaURL:       mediaURL,
			ReactionTarget: req.ReactionTarget,
			Timestamp:      time.Now(),
		})
	} else if err != nil {
		log.Printf("[%s] Send error: %v", i.HostPhone, err)
		if i.Manager.Dispatcher != nil {
			i.Manager.Dispatcher.DispatchMessage(domain.MessageMetadata{
				WhatsappID:   sentID,
				HostID:       i.HostPhone,
				Sender:       i.HostPhone,
				Recipient:    req.Recipient,
				Content:      req.Message,
				Direction:    domain.Outgoing,
				Status:       domain.StatusFailed,
				Actor:        actor,
				OperatorID:   req.OperatorID,
				OperatorName: req.OperatorName,
				Timestamp:    time.Now(),
			})
		}
	}

	time.Sleep(time.Duration(1500+rand.Intn(2000)) * time.Millisecond)
}

func (i *WhatsAppInstance) sendReaction(recipient types.JID, req domain.MessageRequest, extra ...whatsmeow.SendRequestExtra) (whatsmeow.SendResponse, error) {
	msg := i.Client.BuildReaction(recipient, types.EmptyJID, req.ReactionTarget, req.Message)
	return i.Client.SendMessage(context.Background(), recipient, msg, extra...)
}

func (i *WhatsAppInstance) sendMedia(recipient types.JID, req domain.MessageRequest, extra ...whatsmeow.SendRequestExtra) (whatsmeow.SendResponse, string, error) {
	var data []byte
	var err error

	if req.MediaKey != "" {
		if i.Manager.MediaStore == nil {
			return whatsmeow.SendResponse{}, "", errors.New("no media store configured")
		}
		rc, oerr := i.Manager.MediaStore.Open(context.Background(), req.MediaKey)
		if oerr != nil {
			return whatsmeow.SendResponse{}, "", fmt.Errorf("open media from store: %w", oerr)
		}
		defer rc.Close()
		data, err = io.ReadAll(rc)
	} else if req.MediaPath != "" {
		data, err = os.ReadFile(req.MediaPath)
	} else {
		return whatsmeow.SendResponse{}, "", errors.New("neither media_key nor media_path provided")
	}

	if err != nil {
		return whatsmeow.SendResponse{}, "", err
	}
	mimeType := http.DetectContentType(data)

	var uploadResp whatsmeow.UploadResponse
	switch req.Type {
	case domain.Image:
		uploadResp, err = i.Client.Upload(context.Background(), data, whatsmeow.MediaImage)
	case domain.Video:
		uploadResp, err = i.Client.Upload(context.Background(), data, whatsmeow.MediaVideo)
	case domain.Audio:
		uploadResp, err = i.Client.Upload(context.Background(), data, whatsmeow.MediaAudio)
	default:
		uploadResp, err = i.Client.Upload(context.Background(), data, whatsmeow.MediaDocument)
	}

	if err != nil {
		return whatsmeow.SendResponse{}, "", err
	}

	var msg *waE2E.Message
	switch req.Type {
	case domain.Image:
		msg = &waE2E.Message{ImageMessage: &waE2E.ImageMessage{
			Caption:       proto.String(req.Message),
			URL:           proto.String(uploadResp.URL),
			DirectPath:    proto.String(uploadResp.DirectPath),
			MediaKey:      uploadResp.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploadResp.FileEncSHA256,
			FileSHA256:    uploadResp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
		}}
	case domain.Video:
		msg = &waE2E.Message{VideoMessage: &waE2E.VideoMessage{
			Caption:       proto.String(req.Message),
			URL:           proto.String(uploadResp.URL),
			DirectPath:    proto.String(uploadResp.DirectPath),
			MediaKey:      uploadResp.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploadResp.FileEncSHA256,
			FileSHA256:    uploadResp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
		}}
	case domain.Audio:
		msg = &waE2E.Message{AudioMessage: &waE2E.AudioMessage{
			URL:           proto.String(uploadResp.URL),
			DirectPath:    proto.String(uploadResp.DirectPath),
			MediaKey:      uploadResp.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploadResp.FileEncSHA256,
			FileSHA256:    uploadResp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
		}}
	default:
		msg = &waE2E.Message{DocumentMessage: &waE2E.DocumentMessage{
			Caption:       proto.String(req.Message),
			Title:         proto.String(req.Message),
			URL:           proto.String(uploadResp.URL),
			DirectPath:    proto.String(uploadResp.DirectPath),
			MediaKey:      uploadResp.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploadResp.FileEncSHA256,
			FileSHA256:    uploadResp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
		}}
	}

	resp, err := i.Client.SendMessage(context.Background(), recipient, msg, extra...)
	if err != nil {
		return resp, "", err
	}

	var mediaURL string
	if req.MediaKey != "" {
		mediaURL = storage.ResolveMediaURL(req.MediaKey, i.Manager.S3ObjectURL, "")
		if setter, ok := i.Manager.Store.(upload.MediaURLSetter); ok {
			setter.AttachMediaURL(i.HostPhone, resp.ID, mediaURL)
		}
	} else if req.MediaPath != "" && i.Manager.Uploader != nil {
		if key, kerr := i.Manager.Uploader.Enqueue(context.Background(), resp.ID, i.HostPhone, mimeType, req.MediaPath); kerr != nil {
			log.Printf("[%s] Upload job enqueue error: %v", i.HostPhone, kerr)
		} else {
			log.Printf("[%s] Archived media enqueued key=%s message=%s", i.HostPhone, key, resp.ID)
		}
	}

	return resp, mediaURL, nil
}

func (i *WhatsAppInstance) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		go i.handleIncomingMessage(v)
		// Do not query GetGroupInfo for every group message. Group metadata is
		// handled by explicit GroupInfo/JoinedGroup events; message ingestion
		// must remain independent of WhatsApp metadata rate limits.
	case *events.Receipt:
		go i.handleReceipt(v)
	case *events.HistorySync:
		go i.handleHistorySync(v)
	case *events.LoggedOut:
		i.IsConnected = false
		log.Printf("[%s] Logged out: %s", i.HostPhone, v.Reason)
		i.Manager.notifyStatus(i.HostPhone, domain.StatusOffline, false)
		i.Manager.StopInstance(i.HostPhone)
	case *events.Connected:
		i.IsConnected = true
		i.Manager.notifyStatus(i.HostPhone, domain.StatusOnline, true)
		go i.syncJoinedGroups()
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
		// Ignore other types like "inactive".
		return
	}

	// For receipts, Sender is the one who sent the receipt (the recipient of the original message).
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

func (i *WhatsAppInstance) syncJoinedGroups() {
	if i.Manager.Dispatcher == nil || i.Client == nil {
		return
	}

	// Refresh all groups with one request after reconnect. This avoids one
	// GetGroupInfo call per conversation and keeps WhatsApp rate-limit usage
	// bounded during a resync.
	groups, err := i.Client.GetJoinedGroups(context.Background())
	if err != nil {
		log.Printf("[%s] GetJoinedGroups error: %v", i.HostPhone, err)
		return
	}
	for _, info := range groups {
		if info != nil {
			i.dispatchGroupInfo(info)
		}
	}
}

func (i *WhatsAppInstance) handleGroupInfo(groupJID types.JID) {
	if i.Manager.Dispatcher == nil || groupJID.IsEmpty() {
		return
	}

	// Group metadata is stable and WhatsApp rate-limits GetGroupInfo. Incoming
	// messages can otherwise trigger one request per message, especially during
	// history sync. Deduplicate and refresh at most once per ten minutes.
	groupKey := groupJID.String()
	now := time.Now()
	i.groupInfoMu.Lock()
	last, seen := i.groupInfoAt[groupKey]
	if seen && now.Sub(last) < 10*time.Minute {
		i.groupInfoMu.Unlock()
		return
	}
	i.groupInfoAt[groupKey] = now
	i.groupInfoMu.Unlock()

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

	i.dispatchGroupInfo(info)
}

func (i *WhatsAppInstance) dispatchGroupInfo(info *types.GroupInfo) {
	if info == nil || i.Manager.Dispatcher == nil {
		return
	}

	var participants []string
	var hostsInGroup []string
	i.Manager.mu.RLock()
	allHosts := make(map[string]bool, len(i.Manager.Instances))
	for phone := range i.Manager.Instances {
		allHosts[phone] = true
	}
	i.Manager.mu.RUnlock()
	for _, participant := range info.Participants {
		phone := i.resolvePhone(participant.JID)
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
		Description:      info.Topic,
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
		if chatJID.IsEmpty() || chatJID.Server == types.BroadcastServer {
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

			content, msgType, reactionTarget := parseMessageContent(m)
			content = i.normalizeMentionNumbers(content, m)

			if content == "" {
				continue
			}

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

			var mediaURL string
			if i.Manager != nil && i.Manager.MediaStore != nil && i.Client != nil {
				mediaURL = i.downloadMessageMedia(m, wi.GetKey().GetID())
			}

			i.Manager.Dispatcher.DispatchMessage(domain.MessageMetadata{
				WhatsappID:     wi.GetKey().GetID(),
				HostID:         i.HostPhone,
				Sender:         sender,
				Recipient:      recipient,
				Content:        content,
				IsGroup:        chatJID.Server == types.GroupServer,
				Direction:      direction,
				Type:           msgType,
				Status:         domain.StatusSent,
				MediaURL:       mediaURL,
				ReactionTarget: reactionTarget,
				Timestamp:      time.Unix(int64(wi.GetMessageTimestamp()), 0),
			})
		}
	}
}

func (i *WhatsAppInstance) downloadMessageMedia(m *waE2E.Message, msgID string) string {
	if i.Manager == nil || i.Manager.MediaStore == nil || i.Client == nil || m == nil {
		return ""
	}

	var downloadable whatsmeow.DownloadableMessage
	var defaultMime string
	if img := m.GetImageMessage(); img != nil {
		downloadable = img
		defaultMime = img.GetMimetype()
		if defaultMime == "" {
			defaultMime = "image/jpeg"
		}
	} else if vid := m.GetVideoMessage(); vid != nil {
		downloadable = vid
		defaultMime = vid.GetMimetype()
		if defaultMime == "" {
			defaultMime = "video/mp4"
		}
	} else if aud := m.GetAudioMessage(); aud != nil {
		downloadable = aud
		defaultMime = aud.GetMimetype()
		if defaultMime == "" {
			defaultMime = "audio/ogg"
		}
	} else if doc := m.GetDocumentMessage(); doc != nil {
		downloadable = doc
		defaultMime = doc.GetMimetype()
		if defaultMime == "" {
			defaultMime = "application/octet-stream"
		}
	}

	if downloadable == nil {
		return ""
	}

	data, err := i.Client.Download(context.Background(), downloadable)
	if err != nil {
		log.Printf("[%s] Error downloading media for message %s: %v", i.HostPhone, msgID, err)
		return ""
	}

	mimeType := http.DetectContentType(data)
	if mimeType == "application/octet-stream" && defaultMime != "" {
		mimeType = defaultMime
	}

	key := fmt.Sprintf("media/%s", uuid.NewString())
	if err := i.Manager.MediaStore.Put(context.Background(), key, mimeType, data); err != nil {
		log.Printf("[%s] Error storing media: %v", i.HostPhone, err)
		return ""
	}

	var tenantID uuid.UUID
	if i.Manager.ResolveTenant != nil {
		tenantID = i.Manager.ResolveTenant(i.HostPhone)
	}
	if i.Manager.Store != nil && tenantID != uuid.Nil {
		_ = i.Manager.Store.RecordMediaObject(context.Background(), tenantID, key, mimeType, int64(len(data)))
	}

	return storage.ResolveMediaURL(key, i.Manager.S3ObjectURL, "")
}

func (i *WhatsAppInstance) handleIncomingMessage(v *events.Message) {
	// WhatsApp Status updates are broadcast timeline items, not customer chats.
	// Do not project them into contacts, conversations, or the inbox.
	if v == nil || v.Info.Chat.Server == types.BroadcastServer {
		return
	}
	content, msgType, reactionTarget := parseMessageContent(v.Message)
	content = i.normalizeMentionNumbers(content, v.Message)

	if content == "" {
		return
	}

	log.Printf("[%s] Received message: %s from %s", i.HostPhone, v.Info.ID, v.Info.Sender)

	// For group events, Sender should be the participant, but some WhatsApp
	// addressing modes expose the group JID there. Prefer SenderAlt (the
	// participant's phone-number JID) when Sender is the group/chat JID.
	senderJID := v.Info.Sender
	if v.Info.IsGroup && (senderJID.Server == types.GroupServer || senderJID == v.Info.Chat) {
		if !v.Info.SenderAlt.IsEmpty() {
			senderJID = v.Info.SenderAlt
		} else if v.SourceWebMsg != nil {
			// History/unavailable-message events may retain the participant
			// in the original message key even when MessageInfo.Sender is the
			// group JID.
			if participant, err := types.ParseJID(v.SourceWebMsg.GetKey().GetParticipant()); err == nil && !participant.IsEmpty() {
				senderJID = participant
			}
		}
	}
	sender := i.resolvePhone(senderJID)
	recipient := i.HostPhone
	direction := domain.Incoming

	if v.Info.IsFromMe {
		direction = domain.Outgoing
		recipient = i.resolvePhone(v.Info.Chat)
	} else if v.Info.IsGroup {
		recipient = v.Info.Chat.User
	}

	var mediaURL string
	if i.Manager != nil && i.Manager.MediaStore != nil && i.Client != nil {
		mediaURL = i.downloadMessageMedia(v.Message, v.Info.ID)
	}

	if i.Manager.Dispatcher != nil {
		i.Manager.Dispatcher.DispatchMessage(domain.MessageMetadata{
			WhatsappID:     v.Info.ID,
			HostID:         i.HostPhone,
			Sender:         sender,
			Recipient:      recipient,
			Content:        content,
			IsGroup:        v.Info.IsGroup,
			Direction:      direction,
			Type:           msgType,
			Status:         domain.StatusSent,
			MediaURL:       mediaURL,
			ReactionTarget: reactionTarget,
			Timestamp:      v.Info.Timestamp,
		})
	}
}

// normalizeMentionNumbers replaces WhatsApp LIDs in rendered text with the
// corresponding phone-number JIDs when the local LID mapping is available.
func (i *WhatsAppInstance) normalizeMentionNumbers(content string, m *waE2E.Message) string {
	if content == "" || m == nil {
		return content
	}
	var mentioned []string
	if text := m.GetExtendedTextMessage(); text != nil && text.GetContextInfo() != nil {
		mentioned = text.GetContextInfo().GetMentionedJID()
	} else if image := m.GetImageMessage(); image != nil && image.GetContextInfo() != nil {
		mentioned = image.GetContextInfo().GetMentionedJID()
	} else if video := m.GetVideoMessage(); video != nil && video.GetContextInfo() != nil {
		mentioned = video.GetContextInfo().GetMentionedJID()
	} else if document := m.GetDocumentMessage(); document != nil && document.GetContextInfo() != nil {
		mentioned = document.GetContextInfo().GetMentionedJID()
	}
	for _, raw := range mentioned {
		jid, err := types.ParseJID(raw)
		if err != nil || jid.IsEmpty() {
			continue
		}
		phone := i.resolvePhone(jid)
		if phone != "" && phone != jid.User {
			content = strings.ReplaceAll(content, "@"+jid.User, "@"+phone)
		}
	}
	return content
}

func parseMessageContent(m *waE2E.Message) (content string, msgType domain.MessageType, reactionTarget string) {
	if m == nil {
		return "", domain.Text, ""
	}
	msgType = domain.Text

	switch {
	case m.GetConversation() != "":
		content = m.GetConversation()
	case m.GetExtendedTextMessage().GetText() != "":
		content = m.GetExtendedTextMessage().GetText()
	case m.GetImageMessage() != nil:
		content = m.GetImageMessage().GetCaption()
		if content == "" {
			content = "[Image]"
		}
		msgType = domain.Image
	case m.GetVideoMessage() != nil:
		content = m.GetVideoMessage().GetCaption()
		if content == "" {
			content = "[Video]"
		}
		msgType = domain.Video
	case m.GetAudioMessage() != nil:
		content = "[Audio]"
		msgType = domain.Audio
	case m.GetDocumentMessage() != nil:
		content = m.GetDocumentMessage().GetCaption()
		if content == "" {
			content = m.GetDocumentMessage().GetTitle()
		}
		if content == "" {
			content = m.GetDocumentMessage().GetFileName()
		}
		if content == "" {
			content = "[Document]"
		}
		msgType = domain.File
	case m.GetReactionMessage() != nil:
		content = m.GetReactionMessage().GetText()
		msgType = domain.Reaction
		reactionTarget = m.GetReactionMessage().GetKey().GetID()
	case m.GetProtocolMessage() != nil:
		return "", domain.Text, ""
	}

	return content, msgType, reactionTarget
}

func (wm *WhatsAppManager) SendMessageRequest(host string, req domain.MessageRequest) error {
	wm.mu.RLock()
	instance, ok := wm.Instances[host]
	wm.mu.RUnlock()
	if !ok {
		return fmt.Errorf("host %s not found", host)
	}

	if wm.ResolveTenant != nil {
		tenantID := wm.ResolveTenant(host)
		if tenantID != uuid.Nil {
			quota, err := wm.Store.GetQuota(tenantID)
			if err == nil && quota.CurrentUsage >= quota.MonthlyLimit {
				return fmt.Errorf("quota exceeded")
			}
			_ = wm.Store.IncrementQuota(tenantID)
		}
	}

	if req.ID == "" {
		if instance.Client != nil {
			req.ID = instance.Client.GenerateMessageID()
		} else {
			req.ID = whatsmeow.GenerateMessageID()
		}
	}

	if strings.HasSuffix(req.Recipient, "@g.us") || strings.Contains(req.Recipient, "-") {
		req.IsGroup = true
	}

	actor := req.Actor
	if actor == "" {
		actor = domain.ActorOperator
	}

	meta := domain.MessageMetadata{
		WhatsappID:     req.ID,
		HostID:         instance.HostPhone,
		Sender:         instance.HostPhone,
		Recipient:      req.Recipient,
		Content:        req.Message,
		IsGroup:        req.IsGroup,
		Direction:      domain.Outgoing,
		Type:           req.Type,
		Status:         domain.StatusPending,
		Actor:          actor,
		MediaURL:       "",
		ReactionTarget: req.ReactionTarget,
		Timestamp:      time.Now(),
	}

	// Persist to store synchronously if available so any immediate query sees it right away.
	if proj, ok := wm.Store.(domain.ApplicationProjector); ok {
		_ = proj.ProjectMessage(meta)
	}

	// Immediately persist outgoing message with PENDING status so it echoes instantly in the inbox.
	if wm.Dispatcher != nil {
		wm.Dispatcher.DispatchMessage(meta)
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

func (wm *WhatsAppManager) Disconnect(host string) error {
	// Dashboard disconnect means unlink this companion device from the actual
	// WhatsApp phone, not merely close the websocket connection.
	instance := wm.removeInstance(host)
	if instance != nil && instance.Client != nil {
		if err := instance.Client.Logout(context.Background()); err != nil {
			// Ensure the local connection is still stopped if the remote unlink
			// request fails, but return the error to the dashboard.
			instance.Client.Disconnect()
			wm.notifyStatus(host, domain.StatusOffline, false)
			return err
		}
	}
	wm.notifyStatus(host, domain.StatusOffline, false)
	return nil
}

func (wm *WhatsAppManager) Reconnect(host string) error {
	wm.mu.Lock()
	if _, ok := wm.Instances[host]; ok {
		wm.mu.Unlock()
		return fmt.Errorf("instance already connected")
	}
	wm.mu.Unlock()

	devices, err := wm.Container.GetAllDevices(context.Background())
	if err != nil {
		return err
	}
	for _, device := range devices {
		if device.ID.User == host {
			return wm.SpawnInstance(device)
		}
	}
	return fmt.Errorf("no saved device for host %s", host)
}

// SanitizePhoneNumber cleans a phone number or JID string into digits only.
func SanitizePhoneNumber(phone string) string {
	if at := strings.IndexByte(phone, '@'); at >= 0 {
		phone = phone[:at]
	}
	var b strings.Builder
	for _, r := range phone {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// SendInvitation sends an invitation text message using any available connected instance.
func (wm *WhatsAppManager) SendInvitation(to, message string) error {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	var activeInstance *WhatsAppInstance
	for _, instance := range wm.Instances {
		if instance.IsConnected && instance.Client != nil {
			activeInstance = instance
			break
		}
	}

	if activeInstance == nil {
		return fmt.Errorf("no connected WhatsApp instance available")
	}

	cleanTo := SanitizePhoneNumber(to)
	if cleanTo == "" {
		return fmt.Errorf("invalid recipient phone number: %s", to)
	}
	recipientJID := types.NewJID(cleanTo, types.DefaultUserServer)
	msg := &waE2E.Message{Conversation: proto.String(message)}
	_, err := activeInstance.Client.SendMessage(context.Background(), recipientJID, msg)
	return err
}
