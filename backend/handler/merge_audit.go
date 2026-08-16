package handler

import (
	"net/http"

	"github.com/google/uuid"
)

// mergeRequest is the JSON body for merging conversations.
type mergeRequest struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
}

// MergeConversationsHandler moves messages from source onto target and closes source.
func (s *Server) MergeConversationsHandler(w http.ResponseWriter, r *http.Request) {
	tenant, ok := s.tenant(r)
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "valid API key required")
		return
	}
	var req mergeRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
		return
	}
	sourceID, err := uuid.Parse(req.SourceID)
	if err != nil {
		writeAPIError(w, 400, "INVALID_ID", "valid source_id required")
		return
	}
	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		writeAPIError(w, 400, "INVALID_ID", "valid target_id required")
		return
	}
	c, err := s.Platform.MergeConversations(tenant, targetID, sourceID, s.operatorID(r))
	if err != nil {
		writeAPIError(w, 400, "MERGE_FAILED", err.Error())
		return
	}
	WriteJSON(w, 200, c)
}

// splitRequest is the JSON body for splitting a conversation.
type splitRequest struct {
	MessageIDs []string `json:"message_ids"`
}

// SplitConversationHandler creates a new conversation and moves the listed
// messages onto it.
func (s *Server) SplitConversationHandler(w http.ResponseWriter, r *http.Request) {
	tenant, ok := s.tenant(r)
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "valid API key required")
		return
	}
	sourceID, err := parseConversationID(r)
	if err != nil {
		writeAPIError(w, 400, "INVALID_ID", "valid conversation id required")
		return
	}
	var req splitRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
		return
	}
	if len(req.MessageIDs) == 0 {
		writeAPIError(w, 400, "INVALID_REQUEST", "at least one message_id is required")
		return
	}
	messageIDs := make([]uuid.UUID, 0, len(req.MessageIDs))
	for _, idStr := range req.MessageIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			writeAPIError(w, 400, "INVALID_ID", "message_ids contains an invalid uuid")
			return
		}
		messageIDs = append(messageIDs, id)
	}
	c, err := s.Platform.SplitConversation(tenant, sourceID, messageIDs, s.operatorID(r))
	if err != nil {
		writeAPIError(w, 400, "SPLIT_FAILED", err.Error())
		return
	}
	WriteJSON(w, 201, c)
}

// ListOperatorAuditLogsHandler returns the operator audit log, newest first.
func (s *Server) ListOperatorAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	tenant, ok := s.tenant(r)
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "valid API key required")
		return
	}
	limit, offset, valid := page(r)
	if !valid {
		writeAPIError(w, 400, "INVALID_PAGINATION", "limit must be between 1 and 100 and offset non-negative")
		return
	}
	logs, err := s.Platform.ListOperatorAuditLogs(tenant, limit, offset)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
		return
	}
	WriteJSON(w, 200, logs)
}
