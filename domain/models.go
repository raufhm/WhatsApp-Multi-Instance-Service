package domain

import (
	"time"
)

type Direction string

const (
	Incoming Direction = "INCOMING"
	Outgoing Direction = "OUTGOING"
)

type MessageType string

const (
	Text     MessageType = "TEXT"
	Image    MessageType = "IMAGE"
	File     MessageType = "FILE"
	Reaction MessageType = "REACTION"
)

type MessageStatus string

const (
	StatusPending   MessageStatus = "PENDING"
	StatusSent      MessageStatus = "SENT"
	StatusDelivered MessageStatus = "DELIVERED"
	StatusRead      MessageStatus = "READ"
	StatusFailed    MessageStatus = "FAILED"
)

type MessageMetadata struct {
	WhatsappID     string        `json:"whatsapp_id"`
	HostID         string        `json:"host_id"`
	Sender         string        `json:"sender"`
	Recipient      string        `json:"recipient"` // Group JID or User JID
	Content        string        `json:"content"`
	IsGroup        bool          `json:"is_group"`
	Direction      Direction     `json:"direction"`
	Type           MessageType   `json:"type"`
	Status         MessageStatus `json:"status"`
	MediaURL       string        `json:"media_url,omitempty"`
	ReactionTarget string        `json:"reaction_target,omitempty"` // Message ID of the message being reacted to
	Timestamp      time.Time     `json:"timestamp"`
}

type Receipt struct {
	WhatsappID string        `json:"whatsapp_id"`
	Sender     string        `json:"sender"`
	Recipient  string        `json:"recipient"`
	Status     MessageStatus `json:"status"`
	Timestamp  time.Time     `json:"timestamp"`
}

type MessageRequest struct {
	Recipient      string      `json:"recipient"`
	Message        string      `json:"message"`
	IsGroup        bool        `json:"is_group"`
	Type           MessageType `json:"type"`
	MediaPath      string      `json:"media_path,omitempty"`
	ReactionTarget string      `json:"reaction_target,omitempty"` // Message ID for sending reactions
}

type InstanceStatus string

const (
	StatusOnline  InstanceStatus = "ONLINE"
	StatusOffline InstanceStatus = "OFFLINE"
	StatusError   InstanceStatus = "ERROR"
)

type InstanceEvent struct {
	HostID    string         `json:"host_id"`
	Status    InstanceStatus `json:"status"`
	Message   string         `json:"message"`
	Timestamp time.Time      `json:"timestamp"`
}

type InstanceInfo struct {
	HostPhone   string         `json:"host_phone"`
	Status      InstanceStatus `json:"status"`
	IsConnected bool           `json:"is_connected"`
	QueueSize   int            `json:"queue_size"`
}

type GroupInfo struct {
	GroupID          string    `json:"group_id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	OwnerJID         string    `json:"owner_jid"`
	Participants     []string  `json:"participants"` // Phone numbers
	Hosts            []string  `json:"hosts"`        // Bot phone numbers
	ParticipantCount int       `json:"participant_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Interfaces
type Dispatcher interface {
	DispatchMessage(meta MessageMetadata)
	DispatchReceipt(receipt Receipt)
	DispatchEvent(event InstanceEvent)
	UpdateGroup(group GroupInfo)
}

type MessageInterceptor interface {
	Intercept(meta MessageMetadata)
}

type AlertService interface {
	Alert(event InstanceEvent)
}

type LIDResolver interface {
}
