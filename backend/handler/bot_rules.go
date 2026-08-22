package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/raufhm/whops/domain"
)

// ValidateBotRules checks rule configuration before activation. Each rule must
// have a non-empty name, pattern, and response, and a supported match type.
func ValidateBotRules(rules []domain.BotRule) error {
	for i, r := range rules {
		if r.Name == "" {
			return fmt.Errorf("rule %d: name is required", i)
		}
		if r.Pattern == "" {
			return fmt.Errorf("rule %d (%s): pattern is required", i, r.Name)
		}
		if r.Response == "" {
			return fmt.Errorf("rule %d (%s): response is required", i, r.Name)
		}
		switch r.Match {
		case "CONTAINS", "EXACT", "PREFIX":
		default:
			return fmt.Errorf("rule %d (%s): match must be CONTAINS, EXACT, or PREFIX", i, r.Name)
		}
	}
	return nil
}

// ListBotRuleSetsHandler returns every bot ruleset version for the tenant.
func (s *Server) ListBotRuleSetsHandler(w http.ResponseWriter, r *http.Request) {
	tenant, ok := s.tenant(r)
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "valid API key required")
		return
	}
	sets, err := s.Platform.ListBotRuleSets(tenant)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
		return
	}
	WriteJSON(w, 200, sets)
}

// createBotRuleSetRequest is the JSON body for creating a new ruleset version.
type createBotRuleSetRequest struct {
	Rules []domain.BotRule `json:"rules"`
}

// CreateBotRuleSetHandler validates and persists a new bot ruleset version.
// The new version is created inactive; activate it via the activate endpoint.
func (s *Server) CreateBotRuleSetHandler(w http.ResponseWriter, r *http.Request) {
	tenant, ok := s.tenant(r)
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "valid API key required")
		return
	}
	var req createBotRuleSetRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
		return
	}
	if err := ValidateBotRules(req.Rules); err != nil {
		writeAPIError(w, 400, "VALIDATION_ERROR", err.Error())
		return
	}
	set, err := s.Platform.SaveBotRuleSet(tenant, req.Rules)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
		return
	}
	WriteJSON(w, 201, set)
}

// ActivateBotRuleSetHandler activates a specific ruleset version for the tenant.
func (s *Server) ActivateBotRuleSetHandler(w http.ResponseWriter, r *http.Request) {
	tenant, ok := s.tenant(r)
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "valid API key required")
		return
	}
	versionStr := r.URL.Query().Get("version")
	if versionStr == "" {
		writeAPIError(w, 400, "INVALID_REQUEST", "version query parameter required")
		return
	}
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		writeAPIError(w, 400, "INVALID_VERSION", "version must be an integer")
		return
	}
	set, err := s.Platform.ActivateBotRuleSetVersion(tenant, version)
	if err != nil {
		writeAPIError(w, 404, "NOT_FOUND", err.Error())
		return
	}
	WriteJSON(w, 200, set)
}

// GetActiveBotRuleSetHandler returns the currently active ruleset.
func (s *Server) GetActiveBotRuleSetHandler(w http.ResponseWriter, r *http.Request) {
	tenant, ok := s.tenant(r)
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "valid API key required")
		return
	}
	set, err := s.Platform.GetActiveBotRuleSet(tenant)
	if err != nil {
		writeAPIError(w, 404, "NOT_FOUND", "no active ruleset")
		return
	}
	WriteJSON(w, 200, set)
}
