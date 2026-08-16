package upload

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whatsapp-testing/domain"
)

// Config controls the retryable upload worker. Safe defaults are supplied by
// DefaultConfig so existing deployments work without configuration changes.
type Config struct {
	Enabled        bool
	PollInterval   time.Duration
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Jitter         float64
	ClaimLimit     int
}

func DefaultConfig() Config {
	return Config{
		Enabled:        true,
		PollInterval:   5 * time.Second,
		MaxAttempts:    5,
		InitialBackoff: time.Second,
		MaxBackoff:     time.Minute,
		Jitter:         0.2,
		ClaimLimit:     10,
	}
}

// MediaStore is the object-store boundary driven by the retryable worker.
type MediaStore interface {
	Upload(data []byte, key, mimeType string) (string, error)
}

// TenantResolver maps a host id to its owning tenant. It returns uuid.Nil when
// the host is not (yet) linked to an account so the worker can still progress.
type TenantResolver func(hostID string) uuid.UUID

// MediaURLSetter is implemented by repositories that can attach the final
// archive URL to the message once the upload completes.
type MediaURLSetter interface {
	AttachMediaURL(hostID, messageID, mediaURL string)
}

// Manager owns the durable upload-job lifecycle: enqueueing on the send path
// and the bounded-backoff worker that claims, uploads, retries, and finalizes.
type Manager struct {
	repo    domain.UploadJobRepository
	media   MediaStore
	cfg     Config
	resolve TenantResolver
	logf    func(format string, args ...any)
	wg      sync.WaitGroup
}

func NewManager(repo domain.UploadJobRepository, media MediaStore, cfg Config, resolve TenantResolver) *Manager {
	base := DefaultConfig()
	if cfg.Enabled == false {
		base.Enabled = false
	} else if cfg.Enabled {
		base.Enabled = true
	}
	if cfg.PollInterval > 0 {
		base.PollInterval = cfg.PollInterval
	}
	if cfg.MaxAttempts > 0 {
		base.MaxAttempts = cfg.MaxAttempts
	}
	if cfg.InitialBackoff > 0 {
		base.InitialBackoff = cfg.InitialBackoff
	}
	if cfg.MaxBackoff > 0 {
		base.MaxBackoff = cfg.MaxBackoff
	}
	if cfg.ClaimLimit > 0 {
		base.ClaimLimit = cfg.ClaimLimit
	}
	return &Manager{repo: repo, media: media, cfg: base, resolve: resolve, logf: log.Printf}
}

// Enqueue persists a durable upload job and returns its object key. The key is
// generated once and reused for every retry so uploads converge to one object.
func (m *Manager) Enqueue(ctx context.Context, messageID, hostID, mimeType, mediaPath string) (string, error) {
	var tenant uuid.UUID
	if m.resolve != nil {
		tenant = m.resolve(hostID)
	}
	key := newObjectKey()
	job, err := m.repo.CreateUploadJob(ctx, tenant, domain.UploadJob{
		TenantID:      tenant,
		MessageID:     messageID,
		HostID:        hostID,
		ObjectKey:     key,
		MimeType:      mimeType,
		MediaPath:     mediaPath,
		Status:        domain.UploadPending,
		NextAttemptAt: time.Now().UTC(),
	})
	if err != nil {
		return "", err
	}
	m.logf("[upload] queued job=%s message=%s host=%s key=%s", job.ID, job.MessageID, job.HostID, job.ObjectKey)
	return key, nil
}

func newObjectKey() string {
	return "media/" + uuid.NewString()
}

// Start launches the worker loop. It returns immediately; shutdown is driven
// by cancelling ctx and calling Wait to drain in-flight uploads.
func (m *Manager) Start(ctx context.Context) {
	if !m.cfg.Enabled {
		m.logf("[upload] worker disabled")
		return
	}
	m.wg.Add(1)
	go m.loop(ctx)
}

// Wait blocks until the worker loop and all in-flight uploads finish.
func (m *Manager) Wait() { m.wg.Wait() }

func (m *Manager) loop(ctx context.Context) {
	defer m.wg.Done()
	ticker := time.NewTicker(m.cfg.PollInterval)
	defer ticker.Stop()
	m.poll(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.poll(ctx)
		}
	}
}

func (m *Manager) poll(ctx context.Context) {
	jobs, err := m.repo.ClaimDueUploadJobs(ctx, time.Now().UTC(), m.cfg.ClaimLimit)
	if err != nil {
		m.logf("[upload] claim error: %v", err)
		return
	}
	for _, j := range jobs {
		m.wg.Add(1)
		go m.process(ctx, j)
	}
}

func (m *Manager) process(ctx context.Context, job domain.UploadJob) {
	defer m.wg.Done()
	data, err := os.ReadFile(job.MediaPath)
	if err != nil {
		m.finalize(ctx, job, fmt.Errorf("read media: %w", err))
		return
	}
	url, err := m.media.Upload(data, job.ObjectKey, job.MimeType)
	if err != nil {
		m.finalize(ctx, job, err)
		return
	}
	if err := m.repo.MarkUploadCompleted(ctx, job.ID, url); err != nil {
		m.logf("[upload] complete persist error job=%s: %v", job.ID, err)
		return
	}
	m.logf("[upload] completed job=%s message=%s key=%s attempt=%d", job.ID, job.MessageID, job.ObjectKey, job.AttemptCount)
	if setter, ok := m.repo.(MediaURLSetter); ok {
		setter.AttachMediaURL(job.HostID, job.MessageID, url)
	}
}

// finalize applies the retry policy after a failed attempt: a transient error
// schedules the next attempt with backoff; a permanent error or hitting the
// retry limit transitions the job to FAILED.
func (m *Manager) finalize(ctx context.Context, job domain.UploadJob, cause error) {
	reason := cause.Error()
	transient := IsTransientError(cause)
	if transient && job.AttemptCount < m.cfg.MaxAttempts {
		next := time.Now().UTC().Add(Backoff(job.AttemptCount, m.cfg.InitialBackoff, m.cfg.MaxBackoff, m.cfg.Jitter))
		if err := m.repo.MarkUploadRetryable(ctx, job.ID, reason, next, job.AttemptCount); err != nil {
			m.logf("[upload] retry persist error job=%s: %v", job.ID, err)
			return
		}
		m.logf("[upload] retry queued job=%s message=%s key=%s attempt=%d next=%s error=%s",
			job.ID, job.MessageID, job.ObjectKey, job.AttemptCount, next.UTC().Format(time.RFC3339), reason)
		return
	}
	attempts := job.AttemptCount
	if err := m.repo.MarkUploadFailed(ctx, job.ID, reason, attempts); err != nil {
		m.logf("[upload] fail persist error job=%s: %v", job.ID, err)
		return
	}
	m.logf("[upload] failed job=%s message=%s key=%s attempt=%d error=%s", job.ID, job.MessageID, job.ObjectKey, attempts, reason)
}
