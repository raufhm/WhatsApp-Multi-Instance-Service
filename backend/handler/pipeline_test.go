package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whatsapp-testing/domain"
)

type pipelineTestRepoStub struct {
	apiRepoStub
	pipelines       map[uuid.UUID]domain.Pipeline
	stages          map[uuid.UUID]domain.DealStage
	contactStages   map[uuid.UUID]string
	contactStageIDs map[uuid.UUID]*uuid.UUID
}

func newPipelineTestRepo(tenantID uuid.UUID) *pipelineTestRepoStub {
	defaultPlID := uuid.New()
	return &pipelineTestRepoStub{
		apiRepoStub: apiRepoStub{tenant: tenantID},
		pipelines: map[uuid.UUID]domain.Pipeline{
			defaultPlID: {
				ID:          defaultPlID,
				TenantID:    tenantID,
				Name:        "Sales Pipeline",
				Description: "Standard pipeline",
				IsDefault:   true,
				IsActive:    true,
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
			},
		},
		stages:          make(map[uuid.UUID]domain.DealStage),
		contactStages:   make(map[uuid.UUID]string),
		contactStageIDs: make(map[uuid.UUID]*uuid.UUID),
	}
}

func (r *pipelineTestRepoStub) ListPipelines(tenantID uuid.UUID, isActive *bool) ([]domain.Pipeline, error) {
	var result []domain.Pipeline
	for _, p := range r.pipelines {
		if p.TenantID == tenantID {
			if isActive == nil || p.IsActive == *isActive {
				result = append(result, p)
			}
		}
	}
	return result, nil
}

func (r *pipelineTestRepoStub) GetPipeline(tenantID, id uuid.UUID) (domain.Pipeline, error) {
	p, ok := r.pipelines[id]
	if !ok || p.TenantID != tenantID {
		return domain.Pipeline{}, domain.ErrPipelineNotFound
	}
	return p, nil
}

func (r *pipelineTestRepoStub) GetDefaultPipeline(tenantID uuid.UUID) (domain.Pipeline, error) {
	for _, p := range r.pipelines {
		if p.TenantID == tenantID && p.IsDefault {
			return p, nil
		}
	}
	return domain.Pipeline{}, domain.ErrPipelineNotFound
}

func (r *pipelineTestRepoStub) CreatePipeline(tenantID uuid.UUID, name, description string, isDefault bool) (domain.Pipeline, error) {
	for _, p := range r.pipelines {
		if p.TenantID == tenantID && p.Name == name {
			return domain.Pipeline{}, domain.ErrPipelineNameExists
		}
	}
	if isDefault {
		for id, p := range r.pipelines {
			if p.TenantID == tenantID {
				p.IsDefault = false
				r.pipelines[id] = p
			}
		}
	}
	pl := domain.Pipeline{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		IsDefault:   isDefault,
		IsActive:    true,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	r.pipelines[pl.ID] = pl
	return pl, nil
}

func (r *pipelineTestRepoStub) UpdatePipeline(tenantID, id uuid.UUID, name, description *string, isDefault, isActive *bool) (domain.Pipeline, error) {
	p, ok := r.pipelines[id]
	if !ok || p.TenantID != tenantID {
		return domain.Pipeline{}, domain.ErrPipelineNotFound
	}
	if name != nil {
		for pid, existing := range r.pipelines {
			if pid != id && existing.TenantID == tenantID && existing.Name == *name {
				return domain.Pipeline{}, domain.ErrPipelineNameExists
			}
		}
		p.Name = *name
	}
	if description != nil {
		p.Description = *description
	}
	if isDefault != nil {
		if *isDefault {
			for pid, existing := range r.pipelines {
				if pid != id && existing.TenantID == tenantID {
					existing.IsDefault = false
					r.pipelines[pid] = existing
				}
			}
			p.IsDefault = true
		} else if p.IsDefault {
			return domain.Pipeline{}, domain.ErrCannotDeleteDefaultPipeline
		}
	}
	if isActive != nil {
		if !*isActive && p.IsDefault {
			return domain.Pipeline{}, domain.ErrDefaultPipelineCannotBeInactive
		}
		p.IsActive = *isActive
	}
	p.UpdatedAt = time.Now().UTC()
	r.pipelines[id] = p
	return p, nil
}

func (r *pipelineTestRepoStub) DeletePipeline(tenantID, id uuid.UUID) error {
	p, ok := r.pipelines[id]
	if !ok || p.TenantID != tenantID {
		return domain.ErrPipelineNotFound
	}
	if p.IsDefault {
		return domain.ErrCannotDeleteDefaultPipeline
	}
	for _, s := range r.stages {
		if s.PipelineID != nil && *s.PipelineID == id && s.IsActive {
			return domain.ErrPipelineContainsStages
		}
	}
	delete(r.pipelines, id)
	return nil
}

func (r *pipelineTestRepoStub) ListDealStages(tenantID uuid.UUID, pipelineID *uuid.UUID, isActive *bool) ([]domain.DealStage, error) {
	var result []domain.DealStage
	for _, s := range r.stages {
		if s.TenantID == tenantID {
			if pipelineID != nil && (s.PipelineID == nil || *s.PipelineID != *pipelineID) {
				continue
			}
			if isActive != nil && s.IsActive != *isActive {
				continue
			}
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *pipelineTestRepoStub) GetDealStage(tenantID, id uuid.UUID) (domain.DealStage, error) {
	s, ok := r.stages[id]
	if !ok || s.TenantID != tenantID {
		return domain.DealStage{}, domain.ErrStageNotFound
	}
	return s, nil
}

func (r *pipelineTestRepoStub) CreateDealStage(tenantID uuid.UUID, pipelineID *uuid.UUID, key, label, color, icon string, sortOrder int, isWon, isLost bool) (domain.DealStage, error) {
	if pipelineID != nil {
		if _, ok := r.pipelines[*pipelineID]; !ok {
			return domain.DealStage{}, domain.ErrPipelineNotFound
		}
	}
	s := domain.DealStage{
		ID:         uuid.New(),
		TenantID:   tenantID,
		PipelineID: pipelineID,
		Key:        key,
		Label:      label,
		Color:      color,
		Icon:       icon,
		SortOrder:  sortOrder,
		IsActive:   true,
		IsWon:      isWon,
		IsLost:     isLost,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	r.stages[s.ID] = s
	return s, nil
}

func (r *pipelineTestRepoStub) UpdateDealStage(tenantID, id uuid.UUID, pipelineID *uuid.UUID, label, color, icon *string, sortOrder *int, isActive, isWon, isLost *bool) (domain.DealStage, error) {
	s, ok := r.stages[id]
	if !ok || s.TenantID != tenantID {
		return domain.DealStage{}, domain.ErrStageNotFound
	}
	if pipelineID != nil {
		if _, ok := r.pipelines[*pipelineID]; !ok {
			return domain.DealStage{}, domain.ErrPipelineNotFound
		}
		s.PipelineID = pipelineID
	}
	if label != nil {
		s.Label = *label
	}
	if color != nil {
		s.Color = *color
	}
	if icon != nil {
		s.Icon = *icon
	}
	if sortOrder != nil {
		s.SortOrder = *sortOrder
	}
	if isActive != nil {
		s.IsActive = *isActive
	}
	if isWon != nil {
		s.IsWon = *isWon
	}
	if isLost != nil {
		s.IsLost = *isLost
	}
	s.UpdatedAt = time.Now().UTC()
	r.stages[id] = s
	return s, nil
}

func (r *pipelineTestRepoStub) DeleteDealStage(tenantID, id uuid.UUID) error {
	s, ok := r.stages[id]
	if !ok || s.TenantID != tenantID {
		return domain.ErrStageNotFound
	}
	for _, stageID := range r.contactStageIDs {
		if stageID != nil && *stageID == id {
			return domain.ErrStageAssignedToContacts
		}
	}
	s.IsActive = false
	r.stages[id] = s
	return nil
}

func (r *pipelineTestRepoStub) MoveContactToStage(tenantID, contactID uuid.UUID, stageKey string, stageID *uuid.UUID, note string, operatorID uuid.UUID) (domain.DealStageTransition, error) {
	if stageID != nil {
		s, ok := r.stages[*stageID]
		if !ok || s.TenantID != tenantID {
			return domain.DealStageTransition{}, domain.ErrStageNotFound
		}
		stageKey = s.Key
	}
	r.contactStages[contactID] = stageKey
	r.contactStageIDs[contactID] = stageID
	from := "PREV"
	return domain.DealStageTransition{
		ID:        uuid.New(),
		TenantID:  tenantID,
		ContactID: contactID,
		FromStage: &from,
		ToStage:   stageKey,
		Note:      note,
		MovedBy:   &operatorID,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func TestPipelineCRUDOperations(t *testing.T) {
	tenantID := uuid.New()
	repo := newPipelineTestRepo(tenantID)
	srv := &Server{Platform: repo}

	// 1. List Pipelines
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pipelines", nil)
	req.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	srv.APIHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list pipelines failed: %d %s", w.Code, w.Body.String())
	}
	var list []domain.Pipeline
	_ = json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 || list[0].Name != "Sales Pipeline" {
		t.Fatalf("expected default pipeline in list, got %+v", list)
	}

	// 2. Create Pipeline
	createBody := []byte(`{"name":"Support Pipeline","description":"Customer support tickets"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/pipelines", bytes.NewReader(createBody))
	req.Header.Set("X-API-Key", "good")
	w = httptest.NewRecorder()
	srv.APIHandler(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create pipeline failed: %d %s", w.Code, w.Body.String())
	}
	var created domain.Pipeline
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	if created.Name != "Support Pipeline" || created.IsDefault {
		t.Fatalf("unexpected created pipeline: %+v", created)
	}

	// 2b. Duplicate Name should fail 409
	req = httptest.NewRequest(http.MethodPost, "/api/v1/pipelines", bytes.NewReader(createBody))
	req.Header.Set("X-API-Key", "good")
	w = httptest.NewRecorder()
	srv.APIHandler(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate pipeline name, got %d", w.Code)
	}

	// 3. Get Pipeline
	req = httptest.NewRequest(http.MethodGet, "/api/v1/pipelines/"+created.ID.String(), nil)
	req.Header.Set("X-API-Key", "good")
	w = httptest.NewRecorder()
	srv.APIHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get pipeline failed: %d %s", w.Code, w.Body.String())
	}

	// 4. Update Pipeline
	updateBody := []byte(`{"name":"Tier 2 Support Pipeline"}`)
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/pipelines/"+created.ID.String(), bytes.NewReader(updateBody))
	req.Header.Set("X-API-Key", "good")
	w = httptest.NewRecorder()
	srv.APIHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update pipeline failed: %d %s", w.Code, w.Body.String())
	}
	var updated domain.Pipeline
	_ = json.Unmarshal(w.Body.Bytes(), &updated)
	if updated.Name != "Tier 2 Support Pipeline" {
		t.Fatalf("expected updated name, got %q", updated.Name)
	}

	// 5. Delete Pipeline
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/pipelines/"+created.ID.String(), nil)
	req.Header.Set("X-API-Key", "good")
	w = httptest.NewRecorder()
	srv.APIHandler(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete pipeline failed: %d %s", w.Code, w.Body.String())
	}
}

func TestDealStageWithPipeline(t *testing.T) {
	tenantID := uuid.New()
	repo := newPipelineTestRepo(tenantID)
	srv := &Server{Platform: repo}

	var defaultPlID uuid.UUID
	for id, p := range repo.pipelines {
		if p.IsDefault {
			defaultPlID = id
		}
	}

	// 1. Create Stage in Pipeline
	stageBody, _ := json.Marshal(map[string]any{
		"pipeline_id": defaultPlID.String(),
		"key":         "QUALIFIED",
		"label":       "Qualified Lead",
		"color":       "#10b981",
		"icon":        "check",
		"sort_order":  1,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/deal-stages", bytes.NewReader(stageBody))
	req.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	srv.APIHandler(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create deal stage failed: %d %s", w.Code, w.Body.String())
	}
	var stage domain.DealStage
	_ = json.Unmarshal(w.Body.Bytes(), &stage)
	if stage.Key != "QUALIFIED" || stage.PipelineID == nil || *stage.PipelineID != defaultPlID {
		t.Fatalf("unexpected stage: %+v", stage)
	}

	// 2. List Deal Stages by pipeline_id
	req = httptest.NewRequest(http.MethodGet, "/api/v1/deal-stages?pipeline_id="+defaultPlID.String(), nil)
	req.Header.Set("X-API-Key", "good")
	w = httptest.NewRecorder()
	srv.APIHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list deal stages failed: %d %s", w.Code, w.Body.String())
	}
	var stages []domain.DealStage
	_ = json.Unmarshal(w.Body.Bytes(), &stages)
	if len(stages) != 1 || stages[0].Key != "QUALIFIED" {
		t.Fatalf("expected 1 stage, got %+v", stages)
	}

	// 3. Move contact to stage
	contactID := uuid.New()
	moveBody, _ := json.Marshal(map[string]any{
		"stage_id": stage.ID.String(),
		"note":     "Lead qualified after call",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/contacts/"+contactID.String()+"/move-stage", bytes.NewReader(moveBody))
	req.Header.Set("X-API-Key", "good")
	w = httptest.NewRecorder()
	srv.APIHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("move contact stage failed: %d %s", w.Code, w.Body.String())
	}
	var transition domain.DealStageTransition
	_ = json.Unmarshal(w.Body.Bytes(), &transition)
	if transition.ToStage != "QUALIFIED" {
		t.Fatalf("expected to_stage QUALIFIED, got %q", transition.ToStage)
	}
}
