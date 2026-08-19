package whatsapp

import (
	"testing"

	"github.com/raufhm/whatsapp-testing/domain"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func TestParseMessageContent(t *testing.T) {
	tests := []struct {
		name            string
		msg             *waE2E.Message
		expectedContent string
		expectedType    domain.MessageType
		expectedTarget  string
	}{
		{
			name:            "plain text conversation",
			msg:             &waE2E.Message{Conversation: proto.String("Hello world")},
			expectedContent: "Hello world",
			expectedType:    domain.Text,
		},
		{
			name: "extended text message",
			msg: &waE2E.Message{
				ExtendedTextMessage: &waE2E.ExtendedTextMessage{
					Text: proto.String("Extended text"),
				},
			},
			expectedContent: "Extended text",
			expectedType:    domain.Text,
		},
		{
			name: "image with caption",
			msg: &waE2E.Message{
				ImageMessage: &waE2E.ImageMessage{
					Caption: proto.String("Look at this photo"),
				},
			},
			expectedContent: "Look at this photo",
			expectedType:    domain.Image,
		},
		{
			name: "image without caption",
			msg: &waE2E.Message{
				ImageMessage: &waE2E.ImageMessage{},
			},
			expectedContent: "[Image]",
			expectedType:    domain.Image,
		},
		{
			name: "video with caption",
			msg: &waE2E.Message{
				VideoMessage: &waE2E.VideoMessage{
					Caption: proto.String("Watch this video"),
				},
			},
			expectedContent: "Watch this video",
			expectedType:    domain.Video,
		},
		{
			name: "audio message",
			msg: &waE2E.Message{
				AudioMessage: &waE2E.AudioMessage{},
			},
			expectedContent: "[Audio]",
			expectedType:    domain.Audio,
		},
		{
			name: "document with file name",
			msg: &waE2E.Message{
				DocumentMessage: &waE2E.DocumentMessage{
					FileName: proto.String("invoice.pdf"),
				},
			},
			expectedContent: "invoice.pdf",
			expectedType:    domain.File,
		},
		{
			name: "reaction message",
			msg: &waE2E.Message{
				ReactionMessage: &waE2E.ReactionMessage{
					Text: proto.String("👍"),
					Key: &waCommon.MessageKey{
						ID: proto.String("target-msg-123"),
					},
				},
			},
			expectedContent: "👍",
			expectedType:    domain.Reaction,
			expectedTarget:  "target-msg-123",
		},
		{
			name:            "nil message",
			msg:             nil,
			expectedContent: "",
			expectedType:    domain.Text,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			content, msgType, reactionTarget := parseMessageContent(tc.msg)
			if content != tc.expectedContent {
				t.Errorf("expected content %q, got %q", tc.expectedContent, content)
			}
			if msgType != tc.expectedType {
				t.Errorf("expected type %q, got %q", tc.expectedType, msgType)
			}
			if reactionTarget != tc.expectedTarget {
				t.Errorf("expected reaction target %q, got %q", tc.expectedTarget, reactionTarget)
			}
		})
	}
}
