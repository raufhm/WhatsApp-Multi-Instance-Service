package handler

import (
	"net/http"

	"github.com/raufhm/whops/domain"
)

// internalNoteRequest is the JSON body for creating an internal note.
type internalNoteRequest struct {
	Content string       `json:"content"`
	Actor   domain.Actor `json:"actor"`
}

// AddInternalNoteHandler records an operator note on the conversation timeline
// without sending it to the WhatsApp contact.
func (s *Server) AddInternalNoteHandler(w http.ResponseWriter, r *http.Request) {
	tenant, ok := s.tenant(r)
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "valid API key required")
		return
	}
	convID, err := parseConversationID(r)
	if err != nil {
		writeAPIError(w, 400, "INVALID_ID", "valid conversation id required")
		return
	}
	var req internalNoteRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
		return
	}
	if req.Content == "" {
		writeAPIError(w, 400, "INVALID_REQUEST", "content is required")
		return
	}
	actor := req.Actor
	if actor == "" {
		actor = domain.ActorOperator
	}
	msg, err := s.Platform.AddInternalNote(tenant, convID, actor, s.operatorID(r), req.Content)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
		return
	}
	WriteJSON(w, 201, msg)
}
