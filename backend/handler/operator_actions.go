package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
)

// assignRequest is the JSON body for assigning an operator to a conversation.
type assignRequest struct {
	Assignee string `json:"assignee"`
}

// AssignConversationHandler assigns an operator to an open conversation.
func (s *Server) AssignConversationHandler(w http.ResponseWriter, r *http.Request) {
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
	var req assignRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
		return
	}
	if req.Assignee == "" {
		writeAPIError(w, 400, "INVALID_REQUEST", "assignee is required")
		return
	}
	c, err := s.Platform.AssignConversation(tenant, convID, req.Assignee, s.operatorID(r))
	if err != nil {
		writeAPIError(w, 404, "NOT_FOUND", err.Error())
		return
	}
	WriteJSON(w, 200, c)
}

// handoffRequest is the JSON body for handing off a conversation.
type handoffRequest struct {
	Reason string `json:"reason"`
}

// HandoffConversationHandler moves a conversation to HANDED_OFF.
func (s *Server) HandoffConversationHandler(w http.ResponseWriter, r *http.Request) {
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
	var req handoffRequest
	_ = DecodeJSONBody(r, &req)
	c, err := s.Platform.HandoffConversation(tenant, convID, s.operatorID(r))
	if err != nil {
		writeAPIError(w, 404, "NOT_FOUND", err.Error())
		return
	}
	WriteJSON(w, 200, c)
}

// CloseConversationHandler closes an open conversation with an explicit reason.
func (s *Server) CloseConversationHandler(w http.ResponseWriter, r *http.Request) {
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
	var req handoffRequest
	_ = DecodeJSONBody(r, &req)
	reason := req.Reason
	if reason == "" {
		reason = "operator closed"
	}
	c, err := s.Platform.CloseConversationWithReason(tenant, convID, reason, s.operatorID(r))
	if err != nil {
		writeAPIError(w, 404, "NOT_FOUND", err.Error())
		return
	}
	WriteJSON(w, 200, c)
}

// DeleteConversationHandler permanently removes a conversation and its timeline.
func (s *Server) DeleteConversationHandler(w http.ResponseWriter, r *http.Request) {
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
	callerID, err := uuid.Parse(s.operatorID(r))
	if err != nil {
		writeAPIError(w, 403, "FORBIDDEN", "an authenticated operator is required")
		return
	}
	caller, err := s.Auth.GetOperatorByID(tenant, callerID)
	if err != nil {
		writeAPIError(w, 403, "FORBIDDEN", "unable to verify operator role")
		return
	}
	if caller.Role == domain.RoleAdmin {
		if err := s.Platform.DeleteConversation(tenant, convID); err != nil {
			writeAPIError(w, 404, "NOT_FOUND", "conversation not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if caller.Role != domain.RoleOperator {
		writeAPIError(w, 403, "FORBIDDEN", "only admins can delete conversations")
		return
	}
	if err := s.Platform.RequestConversationDeletion(tenant, convID, callerID.String()); err != nil {
		writeAPIError(w, 404, "NOT_FOUND", "conversation not found")
		return
	}
	WriteJSON(w, http.StatusAccepted, map[string]any{"status": "deletion_requested"})
}

// ReopenConversationHandler reopens a closed conversation.
func (s *Server) ReopenConversationHandler(w http.ResponseWriter, r *http.Request) {
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
	c, err := s.Platform.ReopenConversation(tenant, convID, s.operatorID(r))
	if err != nil {
		writeAPIError(w, 404, "NOT_FOUND", err.Error())
		return
	}
	WriteJSON(w, 200, c)
}

// parseConversationID extracts the conversation id from the ?id= query param.
func parseConversationID(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(r.URL.Query().Get("id"))
}

// operatorID returns the operator identity from the session when available,
// falling back to the X-Actor header, defaulting to "api" for direct API calls.
func (s *Server) operatorID(r *http.Request) string {
	if s.Auth != nil {
		if c, err := r.Cookie(sessionCookieName); err == nil && c.Value != "" {
			if sid, err := uuid.Parse(c.Value); err == nil && sid != uuid.Nil {
				if sess, err := s.Auth.GetSessionByID(sid); err == nil && sess.OperatorID != uuid.Nil {
					return sess.OperatorID.String()
				}
			}
		}
	}
	actor := r.Header.Get("X-Actor")
	if actor == "" {
		actor = "api"
	}
	return actor
}
