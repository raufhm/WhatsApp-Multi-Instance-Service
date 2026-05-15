package whatsapp

import (
	"time"
)

type Direction string

const (
	Incoming Direction = "INCOMING"
	Outgoing Direction = "OUTGOING"
)

type MessageMetadata struct {
	HostID    string    `json:"host_id"`
	Sender    string    `json:"sender"`
	Recipient string    `json:"recipient"`
	Content   string    `json:"content"`
	IsGroup   bool      `json:"is_group"`
	Direction Direction `json:"direction"`
	Timestamp time.Time `json:"timestamp"`
}

type MessageInterceptor interface {
	Intercept(meta MessageMetadata)
}

type MessageRequest struct {
	Recipient string
	Message   string
	IsGroup   bool
}
