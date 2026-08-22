package upload

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/smithy-go"
	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
)

// --- Backoff ---

func TestBackoffExponentialAndCapped(t *testing.T) {
	initial := time.Second
	max := time.Minute
	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{1, time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{4, 8 * time.Second},
		{7, time.Minute}, // 64s capped at 60s
		{20, time.Minute},
	}
	for _, c := range cases {
		got := Backoff(c.attempt, initial, max, 0)
		if got != c.want {
			t.Fatalf("Backoff(%d) = %s, want %s", c.attempt, got, c.want)
		}
	}
}

func TestBackoffJitterStaysWithinBand(t *testing.T) {
	initial := time.Second
	max := time.Minute
	jitter := 0.5
	for i := 1; i <= 5; i++ {
		for j := 0; j < 50; j++ {
			d := Backoff(i, initial, max, jitter)
			nominal := Backoff(i, initial, max, 0)
			bound := time.Duration(float64(nominal) * jitter)
			if d < nominal-bound || d > nominal+bound {
				t.Fatalf("attempt=%d jittered backoff %s outside band %s±%s", i, d, nominal, bound)
			}
		}
	}
}

// --- Transient classification ---

type fakeTemporary struct{ err error }

func (f fakeTemporary) Temporary() bool { return true }
func (f fakeTemporary) Error() string   { return f.err.Error() }

func TestIsTransientError(t *testing.T) {
	apiErr := &smithy.GenericAPIError{Code: "InternalError"}
	throttle := &smithy.GenericAPIError{Code: "ThrottlingException"}
	notFound := &smithy.GenericAPIError{Code: "NoSuchKey"}

	transient := []error{
		errors.New("connection refused"),
		errors.New("read: i/o timeout"),
		fakeTemporary{err: errors.New("boom")},
		apiErr,
		throttle,
	}
	for _, e := range transient {
		if !IsTransientError(e) {
			t.Fatalf("expected transient: %v", e)
		}
	}
	permanent := []error{
		nil,
		errors.New("no such file or directory"),
		errors.New("access denied: missing S3_BUCKET"),
		notFound,
	}
	for _, e := range permanent {
		if IsTransientError(e) {
			t.Fatalf("expected permanent: %v", e)
		}
	}
}

// --- Fakes ---

type fakeRepo struct {
	created     []domain.UploadJob
	completed   map[uuid.UUID]string
	retried     map[uuid.UUID]retryCall
	failed      map[uuid.UUID]failCall
	attachCalls []attachCall
}

type retryCall struct {
	err      string
	next     time.Time
	attempts int
}
type failCall struct {
	err      string
	attempts int
}
type attachCall struct{ host, msg, url string }

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		completed: map[uuid.UUID]string{},
		retried:   map[uuid.UUID]retryCall{},
		failed:    map[uuid.UUID]failCall{},
	}
}

func (f *fakeRepo) CreateUploadJob(_ context.Context, tenant uuid.UUID, job domain.UploadJob) (domain.UploadJob, error) {
	job.ID = uuid.New()
	job.TenantID = tenant
	job.CreatedAt = time.Now().UTC()
	f.created = append(f.created, job)
	return job, nil
}
func (f *fakeRepo) ClaimDueUploadJobs(context.Context, time.Time, int) ([]domain.UploadJob, error) {
	return nil, nil
}
func (f *fakeRepo) MarkUploadCompleted(_ context.Context, id uuid.UUID, url string) error {
	f.completed[id] = url
	return nil
}
func (f *fakeRepo) MarkUploadRetryable(_ context.Context, id uuid.UUID, err string, next time.Time, attempts int) error {
	f.retried[id] = retryCall{err, next, attempts}
	return nil
}
func (f *fakeRepo) MarkUploadFailed(_ context.Context, id uuid.UUID, err string, attempts int) error {
	f.failed[id] = failCall{err, attempts}
	return nil
}
func (f *fakeRepo) AttachMediaURL(host, msg, url string) {
	f.attachCalls = append(f.attachCalls, attachCall{host, msg, url})
}

// fakeRepo is both the repository and the MediaURLSetter (satisfied above).

type fakeMedia struct {
	fail     error
	uploaded map[string]string
	lastKey  string
	lastMime string
	calls    int
}

func (m *fakeMedia) Upload(data []byte, key, mime string) (string, error) {
	m.calls++
	m.lastKey = key
	m.lastMime = mime
	if m.uploaded == nil {
		m.uploaded = map[string]string{}
	}
	if m.fail != nil {
		return "", m.fail
	}
	url := "s3://bucket/" + key
	m.uploaded[key] = url
	return url, nil
}

func newTestManager(t *testing.T, repo domain.UploadJobRepository, media MediaStore, cfg Config) *Manager {
	t.Helper()
	m := NewManager(repo, media, cfg, func(host string) uuid.UUID { return uuid.Nil })
	m.logf = func(string, ...any) {}
	return m
}

func TestEnqueueCreatesDurableJobAndIdempotentKey(t *testing.T) {
	repo := newFakeRepo()
	dir := t.TempDir()
	path := filepath.Join(dir, "a.jpg")
	if err := os.WriteFile(path, []byte("jpeg"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := newTestManager(t, repo, &fakeMedia{}, DefaultConfig())

	key1, err := m.Enqueue(context.Background(), "msg-1", "host-1", "image/jpeg", path)
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected 1 created job, got %d", len(repo.created))
	}
	job := repo.created[0]
	if job.ObjectKey != key1 {
		t.Fatalf("returned key %q != job key %q", key1, job.ObjectKey)
	}
	if job.Status != domain.UploadPending || job.MessageID != "msg-1" || job.HostID != "host-1" || job.MimeType != "image/jpeg" {
		t.Fatalf("unexpected job: %+v", job)
	}
}

func TestProcessRetriesTransientThenTerminalFailure(t *testing.T) {
	repo := newFakeRepo()
	dir := t.TempDir()
	path := filepath.Join(dir, "a.jpg")
	if err := os.WriteFile(path, []byte("jpeg"), 0o644); err != nil {
		t.Fatal(err)
	}
	media := &fakeMedia{fail: errors.New("connection reset by peer")}
	cfg := DefaultConfig()
	cfg.MaxAttempts = 3
	m := newTestManager(t, repo, media, cfg)

	job := domain.UploadJob{ID: uuid.New(), MessageID: "m", HostID: "h", ObjectKey: "media/k", MimeType: "image/jpeg", MediaPath: path, AttemptCount: 1}

	// attempt 1 -> transient, below limit -> retry
	m.wg.Add(1)
	m.process(context.Background(), job)
	if _, ok := repo.retried[job.ID]; !ok {
		t.Fatalf("expected retry scheduled, got %+v", repo.failed)
	}
	if repo.retried[job.ID].attempts != 1 {
		t.Fatalf("retry attempt count = %d, want 1", repo.retried[job.ID].attempts)
	}

	// attempt 2 == max -> terminal failure
	job.AttemptCount = 3
	m.wg.Add(1)
	m.process(context.Background(), job)
	if _, ok := repo.failed[job.ID]; !ok {
		t.Fatalf("expected terminal failure, got %+v", repo.retried)
	}
	if repo.failed[job.ID].attempts != 3 {
		t.Fatalf("failed attempt count = %d, want 3", repo.failed[job.ID].attempts)
	}
}

func TestProcessPermanentErrorFailsImmediately(t *testing.T) {
	repo := newFakeRepo()
	dir := t.TempDir()
	path := filepath.Join(dir, "a.jpg")
	if err := os.WriteFile(path, []byte("jpeg"), 0o644); err != nil {
		t.Fatal(err)
	}
	media := &fakeMedia{fail: &smithy.GenericAPIError{Code: "NoSuchKey"}}
	m := newTestManager(t, repo, media, DefaultConfig())
	job := domain.UploadJob{ID: uuid.New(), ObjectKey: "media/k", MediaPath: path, AttemptCount: 1}
	m.wg.Add(1)
	m.process(context.Background(), job)
	if _, ok := repo.failed[job.ID]; !ok {
		t.Fatalf("expected permanent failure, retried=%+v", repo.retried)
	}
	if _, ok := repo.retried[job.ID]; ok {
		t.Fatal("permanent error must not be retried")
	}
}

func TestProcessReusesIdempotentKeyAndAttachesURL(t *testing.T) {
	repo := newFakeRepo()
	dir := t.TempDir()
	path := filepath.Join(dir, "a.jpg")
	if err := os.WriteFile(path, []byte("jpeg"), 0o644); err != nil {
		t.Fatal(err)
	}
	media := &fakeMedia{}
	m := newTestManager(t, repo, media, DefaultConfig())
	job := domain.UploadJob{ID: uuid.New(), MessageID: "m-1", HostID: "host-1", ObjectKey: "media/stable-key", MimeType: "image/jpeg", MediaPath: path, AttemptCount: 1}
	m.wg.Add(1)
	m.process(context.Background(), job)

	if media.lastKey != "media/stable-key" {
		t.Fatalf("object key not reused: got %q", media.lastKey)
	}
	if _, ok := repo.completed[job.ID]; !ok {
		t.Fatal("job not marked completed")
	}
	if len(repo.attachCalls) != 1 || repo.attachCalls[0].url != "s3://bucket/media/stable-key" {
		t.Fatalf("message URL not attached: %+v", repo.attachCalls)
	}
	if repo.attachCalls[0].msg != "m-1" || repo.attachCalls[0].host != "host-1" {
		t.Fatalf("attach args wrong: %+v", repo.attachCalls[0])
	}
}

func TestProcessMissingPayloadFails(t *testing.T) {
	repo := newFakeRepo()
	media := &fakeMedia{}
	m := newTestManager(t, repo, media, DefaultConfig())
	job := domain.UploadJob{ID: uuid.New(), ObjectKey: "media/k", MediaPath: filepath.Join(t.TempDir(), "gone.bin"), AttemptCount: 1}
	m.wg.Add(1)
	m.process(context.Background(), job)
	if _, ok := repo.failed[job.ID]; !ok {
		t.Fatalf("missing payload should fail, retried=%+v", repo.retried)
	}
}

// Compile-time check that fakeRepo satisfies both interfaces.
var _ domain.UploadJobRepository = (*fakeRepo)(nil)
var _ MediaURLSetter = (*fakeRepo)(nil)
