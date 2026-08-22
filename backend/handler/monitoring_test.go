package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
)

// fakeMonitoringStore is an in-memory domain.MonitoringStore for dashboard
// monitoring handler tests.
type fakeMonitoringStore struct {
	mu       sync.Mutex
	hosts    map[string]bool // hostID -> belongs to tenant
	statuses []domain.StatusEvent
	events   []domain.InstanceLogEvent
	samples  []domain.QueueDepthSample
	metrics  domain.MessageMetrics
	status   domain.InstanceMonitoring
	err      error
}

func newFakeMonitoringStore(host string) *fakeMonitoringStore {
	return &fakeMonitoringStore{
		hosts: map[string]bool{host: true},
		status: domain.InstanceMonitoring{
			HostID:          host,
			Status:          domain.StatusOffline,
			IsConnected:     false,
			QueueSize:       0,
			Uptime:          "0s",
			LastConnectedAt: nil,
		},
		metrics: domain.MessageMetrics{
			StatusBreakdown: map[string]int{"DELIVERED": 3},
		},
	}
}

func (f *fakeMonitoringStore) AccountHasHost(tenantID uuid.UUID, hostID string) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.hosts[hostID], nil
}

func (f *fakeMonitoringStore) RecordStatusEvent(ctx context.Context, tenantID uuid.UUID, hostID string, status domain.InstanceStatus, isConnected bool, message string) (domain.StatusEvent, error) {
	return domain.StatusEvent{}, nil
}

func (f *fakeMonitoringStore) RecordInstanceEvent(ctx context.Context, tenantID uuid.UUID, hostID, eventType, direction string, payload any) (domain.InstanceLogEvent, error) {
	return domain.InstanceLogEvent{}, nil
}

func (f *fakeMonitoringStore) ListStatusEvents(ctx context.Context, tenantID uuid.UUID, hostID string, limit int) ([]domain.StatusEvent, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.statuses == nil {
		return []domain.StatusEvent{}, nil
	}
	if limit > 0 && len(f.statuses) > limit {
		return f.statuses[:limit], nil
	}
	return f.statuses, nil
}

func (f *fakeMonitoringStore) ListInstanceEvents(ctx context.Context, tenantID uuid.UUID, hostID string, eventTypes []string, afterID int64, limit int) ([]domain.InstanceLogEvent, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.events == nil {
		return []domain.InstanceLogEvent{}, nil
	}
	var filtered []domain.InstanceLogEvent
	for _, ev := range f.events {
		if len(eventTypes) > 0 && !containsString(eventTypes, ev.EventType) {
			continue
		}
		filtered = append(filtered, ev)
	}
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered, nil
}

func containsString(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

func (f *fakeMonitoringStore) GetInstanceMonitoring(ctx context.Context, tenantID uuid.UUID, hostID string) (domain.InstanceMonitoring, error) {
	if f.err != nil {
		return domain.InstanceMonitoring{}, f.err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.status, nil
}

func (f *fakeMonitoringStore) GetMessageMetrics(ctx context.Context, tenantID uuid.UUID, hostID string, since time.Time, bucketTrunc string) (domain.MessageMetrics, error) {
	if f.err != nil {
		return domain.MessageMetrics{}, f.err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	m := f.metrics
	if m.StatusBreakdown == nil {
		m.StatusBreakdown = map[string]int{}
	}
	return m, nil
}

func (f *fakeMonitoringStore) ListQueueDepth(ctx context.Context, tenantID uuid.UUID, hostID string, since time.Time, limit int) ([]domain.QueueDepthSample, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.samples == nil {
		return []domain.QueueDepthSample{}, nil
	}
	return f.samples, nil
}

// fakeManager implements the accountManager subset with a single known host.
type fakeManager struct {
	host string
	info domain.InstanceInfo
	mu   sync.Mutex
}

func newFakeManager(host string) *fakeManager {
	return &fakeManager{
		host: host,
		info: domain.InstanceInfo{
			HostPhone:   "15550001111",
			Status:      domain.StatusOnline,
			IsConnected: true,
			QueueSize:   4,
		},
	}
}

func (m *fakeManager) ListInstances() []domain.InstanceInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	return []domain.InstanceInfo{m.info}
}

func (m *fakeManager) GetInstance(host string) (domain.InstanceInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if host != m.host {
		return domain.InstanceInfo{}, errors.New("instance not found")
	}
	return m.info, nil
}

func (m *fakeManager) Disconnect(host string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if host != m.host {
		return errors.New("instance not found")
	}
	m.info.IsConnected = false
	m.info.Status = domain.StatusOffline
	return nil
}

func (m *fakeManager) Reconnect(host string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if host != m.host {
		return errors.New("instance not found")
	}
	m.info.IsConnected = true
	m.info.Status = domain.StatusOnline
	return nil
}

// newMonitoringSession signs a session in and returns the cookie-authed
// DashboardAPIHandler wired with the fakes.
func newMonitoringSession(t *testing.T, tenantID uuid.UUID, role string) (*fakeAuth, *domain.Session, *DashboardAPIHandler) {
	t.Helper()
	auth := newFakeAuthWithRole(tenantID, "op@monitoring.test", "secret", role)
	op := auth.operators["op@monitoring.test"].op
	session, err := auth.CreateSession(op.ID, tenantID, time.Hour)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	api := &DashboardAPIHandler{Auth: auth}
	return auth, &session, api
}

func monitoringRequest(cookie *http.Cookie, path string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.AddCookie(cookie)
	return req
}

// monitoringHandler wraps the raw DashboardAPIHandler with the session
// middleware so tenant/operator context is populated from the cookie.
func monitoringHandler(api *DashboardAPIHandler) http.Handler {
	return DashboardSessionMiddleware(api.Auth, api)
}

func TestMonitoringAPI(t *testing.T) {
	tenantID := uuid.New()
	host := "host-abc-123"
	_, adminSession, adminAPI := newMonitoringSession(t, tenantID, "admin")
	adminHandler := monitoringHandler(adminAPI)
	cookie := &http.Cookie{Name: sessionCookieName, Value: adminSession.ID.String()}

	store := newFakeMonitoringStore(host)
	store.statuses = []domain.StatusEvent{
		{ID: 1, TenantID: tenantID, HostID: host, Status: domain.StatusOnline, IsConnected: true, OccurredAt: time.Now()},
		{ID: 2, TenantID: tenantID, HostID: host, Status: domain.StatusOffline, IsConnected: false, OccurredAt: time.Now()},
	}
	store.events = []domain.InstanceLogEvent{
		{ID: 1, HostID: host, EventType: domain.EventSendError, Direction: "OUT", Payload: json.RawMessage(`{"message_id":"m1"}`), OccurredAt: time.Now()},
		{ID: 2, HostID: host, EventType: domain.EventStatus, Direction: "IN", Payload: json.RawMessage(`{}`), OccurredAt: time.Now()},
	}
	store.samples = []domain.QueueDepthSample{
		{ID: 1, Timestamp: time.Now().Add(-time.Minute), QueueSize: 2},
		{ID: 2, Timestamp: time.Now(), QueueSize: 4},
	}
	adminAPI.Monitor = store
	adminAPI.Manager = newFakeManager(host)

	t.Run("status for owned host", func(t *testing.T) {
		rr := httptest.NewRecorder()
		adminHandler.ServeHTTP(rr, monitoringRequest(cookie, "/dashboard/api/monitoring/status?host="+host))
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("bad json: %v", err)
		}
		if body["host_id"] != host {
			t.Fatalf("expected host_id %q, got %v", host, body["host_id"])
		}
		if body["status"] != "ONLINE" {
			t.Fatalf("expected merged ONLINE status, got %v", body["status"])
		}
		if body["queue_size"] != float64(4) {
			t.Fatalf("expected merged queue_size 4, got %v", body["queue_size"])
		}
	})

	t.Run("status missing host is 400", func(t *testing.T) {
		rr := httptest.NewRecorder()
		adminHandler.ServeHTTP(rr, monitoringRequest(cookie, "/dashboard/api/monitoring/status"))
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("status for host outside tenant is 404", func(t *testing.T) {
		rr := httptest.NewRecorder()
		adminHandler.ServeHTTP(rr, monitoringRequest(cookie, "/dashboard/api/monitoring/status?host=other-host"))
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("monitor unavailable is 503", func(t *testing.T) {
		api := &DashboardAPIHandler{Auth: adminAPI.Auth}
		rr := httptest.NewRecorder()
		monitoringHandler(api).ServeHTTP(rr, monitoringRequest(cookie, "/dashboard/api/monitoring/status?host="+host))
		if rr.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("status events respects limit", func(t *testing.T) {
		rr := httptest.NewRecorder()
		adminHandler.ServeHTTP(rr, monitoringRequest(cookie, "/dashboard/api/monitoring/status-events?host="+host+"&limit=1"))
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}
		var events []domain.StatusEvent
		if err := json.Unmarshal(rr.Body.Bytes(), &events); err != nil {
			t.Fatalf("bad json: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
	})

	t.Run("events filters by ERROR type", func(t *testing.T) {
		rr := httptest.NewRecorder()
		adminHandler.ServeHTTP(rr, monitoringRequest(cookie, "/dashboard/api/monitoring/events?host="+host+"&type=ERROR"))
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}
		var events []domain.InstanceLogEvent
		if err := json.Unmarshal(rr.Body.Bytes(), &events); err != nil {
			t.Fatalf("bad json: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 error event, got %d", len(events))
		}
		if events[0].EventType != domain.EventSendError {
			t.Fatalf("expected SEND_ERROR, got %s", events[0].EventType)
		}
	})

	t.Run("events single type filter", func(t *testing.T) {
		rr := httptest.NewRecorder()
		adminHandler.ServeHTTP(rr, monitoringRequest(cookie, "/dashboard/api/monitoring/events?host="+host+"&type=SEND_ERROR"))
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}
		var events []domain.InstanceLogEvent
		if err := json.Unmarshal(rr.Body.Bytes(), &events); err != nil {
			t.Fatalf("bad json: %v", err)
		}
		if len(events) != 1 || events[0].EventType != domain.EventSendError {
			t.Fatalf("expected 1 SEND_ERROR event, got %d", len(events))
		}
	})

	t.Run("metrics returns breakdown", func(t *testing.T) {
		rr := httptest.NewRecorder()
		adminHandler.ServeHTTP(rr, monitoringRequest(cookie, "/dashboard/api/monitoring/metrics?host="+host+"&window=24h"))
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}
		var m domain.MessageMetrics
		if err := json.Unmarshal(rr.Body.Bytes(), &m); err != nil {
			t.Fatalf("bad json: %v", err)
		}
		if m.StatusBreakdown["DELIVERED"] != 3 {
			t.Fatalf("expected DELIVERED=3, got %v", m.StatusBreakdown)
		}
	})

	t.Run("queue depth returns samples", func(t *testing.T) {
		rr := httptest.NewRecorder()
		adminHandler.ServeHTTP(rr, monitoringRequest(cookie, "/dashboard/api/monitoring/queue-depth?host="+host+"&minutes=30"))
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}
		var samples []domain.QueueDepthSample
		if err := json.Unmarshal(rr.Body.Bytes(), &samples); err != nil {
			t.Fatalf("bad json: %v", err)
		}
		if len(samples) != 2 {
			t.Fatalf("expected 2 samples, got %d", len(samples))
		}
	})

	t.Run("viewer role has monitoring access", func(t *testing.T) {
		_, viewerSession, viewerAPI := newMonitoringSession(t, tenantID, "viewer")
		viewerAPI.Monitor = store
		viewerAPI.Manager = newFakeManager(host)
		vCookie := &http.Cookie{Name: sessionCookieName, Value: viewerSession.ID.String()}
		rr := httptest.NewRecorder()
		monitoringHandler(viewerAPI).ServeHTTP(rr, monitoringRequest(vCookie, "/dashboard/api/monitoring/status?host="+host))
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 for viewer, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("no session is unauthorized", func(t *testing.T) {
		api := &DashboardAPIHandler{Auth: adminAPI.Auth, Monitor: store, Manager: newFakeManager(host)}
		apiHandler := DashboardSessionMiddleware(adminAPI.Auth, api)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/dashboard/api/monitoring/status?host="+host, nil)
		apiHandler.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
		}
	})
}

func TestMonitoringStreamSSE(t *testing.T) {
	tenantID := uuid.New()
	host := "host-stream-1"
	_, adminSession, adminAPI := newMonitoringSession(t, tenantID, "admin")
	adminAPI.Monitor = newFakeMonitoringStore(host)
	adminAPI.Manager = newFakeManager(host)
	apiHandler := DashboardSessionMiddleware(adminAPI.Auth, adminAPI)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/dashboard/api/monitoring/stream?host="+host, nil)
	req = req.WithContext(ctx)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: adminSession.ID.String()})

	rr := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		defer close(done)
		apiHandler.ServeHTTP(rr, req)
	}()

	deadline := time.After(2 * time.Second)
	for {
		if ct := rr.Header().Get("Content-Type"); ct == "text/event-stream" {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("expected text/event-stream content type, timed out")
		case <-time.After(10 * time.Millisecond):
		}
	}
	cancel()
	<-done

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}
