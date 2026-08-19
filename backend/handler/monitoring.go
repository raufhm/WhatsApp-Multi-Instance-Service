package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whatsapp-testing/domain"
	"github.com/raufhm/whatsapp-testing/internal/broadcast"
)

const sseHeartbeatInterval = 20 * time.Second

// monitoringHost returns the validated host_id query param, writing a 400/404
// response and returning false when the host is missing or not the tenant's.
func (d *DashboardAPIHandler) monitoringHost(w http.ResponseWriter, tenantID uuid.UUID, host string) (string, bool) {
	host = strings.TrimSpace(host)
	if host == "" {
		writeAPIError(w, 400, "INVALID_REQUEST", "host query param is required")
		return "", false
	}
	if d.Monitor == nil {
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "monitoring store unavailable")
		return "", false
	}
	ok, err := d.Monitor.AccountHasHost(tenantID, host)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to validate host")
		return "", false
	}
	if !ok {
		writeAPIError(w, 404, "NOT_FOUND", "account not found")
		return "", false
	}
	return host, true
}

// monitoringStatus returns the current state + last online/offline for an
// account, merged with live queue depth when the client is running.
func (d *DashboardAPIHandler) monitoringStatus(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	host, ok := d.monitoringHost(w, tenantID, r.URL.Query().Get("host"))
	if !ok {
		return
	}
	m, err := d.Monitor.GetInstanceMonitoring(r.Context(), tenantID, host)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to load monitoring status")
		return
	}
	if info, err := d.Manager.GetInstance(host); err == nil {
		m.IsConnected = info.IsConnected
		m.QueueSize = info.QueueSize
		if info.IsConnected {
			m.Status = domain.StatusOnline
		}
	}
	WriteJSON(w, 200, m)
}

// monitoringStatusEvents returns the status transition history newest-first.
func (d *DashboardAPIHandler) monitoringStatusEvents(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	host, ok := d.monitoringHost(w, tenantID, r.URL.Query().Get("host"))
	if !ok {
		return
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	events, err := d.Monitor.ListStatusEvents(r.Context(), tenantID, host, limit)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to load status events")
		return
	}
	if events == nil {
		events = []domain.StatusEvent{}
	}
	WriteJSON(w, 200, events)
}

// monitoringMetrics returns message volume + status breakdown for a window.
func (d *DashboardAPIHandler) monitoringMetrics(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	host, ok := d.monitoringHost(w, tenantID, r.URL.Query().Get("host"))
	if !ok {
		return
	}
	window := r.URL.Query().Get("window")
	var since time.Time
	var bucketTrunc string
	switch window {
	case "6h":
		since = time.Now().Add(-6 * time.Hour)
		bucketTrunc = "hour"
	case "24h":
		since = time.Now().Add(-24 * time.Hour)
		bucketTrunc = "hour"
	case "7d":
		since = time.Now().Add(-7 * 24 * time.Hour)
		bucketTrunc = "day"
	default:
		since = time.Now().Add(-1 * time.Hour)
		bucketTrunc = "hour"
	}
	metrics, err := d.Monitor.GetMessageMetrics(r.Context(), tenantID, host, since, bucketTrunc)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to load message metrics")
		return
	}
	if metrics.StatusBreakdown == nil {
		metrics.StatusBreakdown = map[string]int{}
	}
	WriteJSON(w, 200, metrics)
}

var monitoringErrorTypes = []string{
	domain.EventSendError,
	domain.EventProjectionFail,
	domain.EventMediaError,
	domain.EventUploadFail,
	domain.EventLoggedOut,
}

// monitoringEvents returns the error/warning event log newest-first. The
// optional type filter is a single event type or "ERROR" for all error types.
func (d *DashboardAPIHandler) monitoringEvents(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	host, ok := d.monitoringHost(w, tenantID, r.URL.Query().Get("host"))
	if !ok {
		return
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	filter := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("type")))
	var eventTypes []string
	switch filter {
	case "", "ERROR":
		eventTypes = monitoringErrorTypes
	default:
		eventTypes = []string{filter}
	}
	events, err := d.Monitor.ListInstanceEvents(r.Context(), tenantID, host, eventTypes, 0, limit)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to load events")
		return
	}
	if events == nil {
		events = []domain.InstanceLogEvent{}
	}
	WriteJSON(w, 200, events)
}

// monitoringQueueDepth returns sampled outbound queue depth over the last
// N minutes (default 60), oldest-first for a sparkline.
func (d *DashboardAPIHandler) monitoringQueueDepth(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	host, ok := d.monitoringHost(w, tenantID, r.URL.Query().Get("host"))
	if !ok {
		return
	}
	minutes, err := strconv.Atoi(r.URL.Query().Get("minutes"))
	if err != nil || minutes <= 0 {
		minutes = 60
	}
	if minutes > 1440 {
		minutes = 1440
	}
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)
	samples, err := d.Monitor.ListQueueDepth(r.Context(), tenantID, host, since, 0)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to load queue depth")
		return
	}
	if samples == nil {
		samples = []domain.QueueDepthSample{}
	}
	WriteJSON(w, 200, samples)
}

// monitoringStream is a Server-Sent Events endpoint that pushes the live event
// log for an account. It backfills events after the `since` cursor (or the
// Last-Event-ID header) then streams new events as they are recorded.
func (d *DashboardAPIHandler) monitoringStream(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	host, ok := d.monitoringHost(w, tenantID, r.URL.Query().Get("host"))
	if !ok {
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeAPIError(w, 500, "INTERNAL_ERROR", "streaming unsupported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	var since int64
	if raw := r.URL.Query().Get("since"); raw != "" {
		since, _ = strconv.ParseInt(raw, 10, 64)
	}
	if since == 0 {
		if raw := r.Header.Get("Last-Event-ID"); raw != "" {
			since, _ = strconv.ParseInt(raw, 10, 64)
		}
	}

	// Backfill events missed since the reconnect cursor.
	if since > 0 {
		backfill, err := d.Monitor.ListInstanceEvents(r.Context(), tenantID, host, nil, since, 200)
		if err == nil {
			for _, ev := range backfill {
				writeSSE(w, ev)
			}
		}
		flusher.Flush()
	}

	// Subscribe for live events. A closed channel signals a slow consumer; the
	// client reconnects and backfills via the since cursor.
	ch, unsubscribe := d.broadcaster().Subscribe(host, 0)
	defer unsubscribe()

	heartbeat := time.NewTicker(sseHeartbeatInterval)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			writeSSE(w, ev)
			flusher.Flush()
		case <-heartbeat.C:
			_, _ = w.Write([]byte(": ping\n\n"))
			flusher.Flush()
		}
	}
}

func (d *DashboardAPIHandler) broadcaster() *broadcast.Broadcaster {
	if d.Broadcaster != nil {
		return d.Broadcaster
	}
	return broadcast.NewBroadcaster()
}

func writeSSE(w http.ResponseWriter, ev domain.InstanceLogEvent) {
	data := ev.Payload
	if len(data) == 0 {
		data = []byte("{}")
	}
	// Event type is carried in the JSON payload; the stream uses default
	// "message" events so the frontend's single onmessage handler receives
	// every frame.
	_, _ = fmt.Fprintf(w, "id: %d\ndata: %s\n\n", ev.ID, data)
}
