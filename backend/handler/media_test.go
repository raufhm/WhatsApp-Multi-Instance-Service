package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
	"github.com/raufhm/whops/internal/storage"
)

func contextWithTenant(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKey{}, tenantID)
}

type mediaTestRepo struct {
	apiRepoStub
	objects map[string]domain.MediaObject
}

func (m *mediaTestRepo) RecordMediaObject(ctx context.Context, tenantID uuid.UUID, objectKey, mimeType string, size int64) error {
	m.objects[objectKey] = domain.MediaObject{
		ID:        uuid.New(),
		TenantID:  tenantID,
		ObjectKey: objectKey,
		MimeType:  mimeType,
		Size:      size,
	}
	return nil
}

func (m *mediaTestRepo) GetMediaObject(ctx context.Context, tenantID uuid.UUID, objectKey string) (domain.MediaObject, error) {
	obj, ok := m.objects[objectKey]
	if !ok {
		return domain.MediaObject{}, errors.New("not found")
	}
	if obj.TenantID != uuid.Nil && tenantID != uuid.Nil && obj.TenantID != tenantID {
		return domain.MediaObject{}, errors.New("not found")
	}
	return obj, nil
}

func TestAPIMediaServing(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "media_api_test_*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)

	diskStore, err := storage.NewDiskStore(tempDir)
	if err != nil {
		t.Fatalf("disk store error: %v", err)
	}

	tenant1 := uuid.New()
	tenant2 := uuid.New()

	repo := &mediaTestRepo{
		apiRepoStub: apiRepoStub{tenant: tenant1},
		objects:     make(map[string]domain.MediaObject),
	}

	key1 := "media/tenant1-image.png"
	data1 := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A}
	_ = diskStore.Put(context.Background(), key1, "image/png", data1)
	_ = repo.RecordMediaObject(context.Background(), tenant1, key1, "image/png", int64(len(data1)))

	key2 := "media/tenant2-image.png"
	data2 := []byte{0xFF, 0xD8, 0xFF}
	_ = diskStore.Put(context.Background(), key2, "image/jpeg", data2)
	_ = repo.RecordMediaObject(context.Background(), tenant2, key2, "image/jpeg", int64(len(data2)))

	srv := &Server{
		Platform:   repo,
		MediaStore: diskStore,
	}

	// 1. Unauthenticated -> 401
	{
		r := httptest.NewRequest("GET", "/api/v1/media/"+key1, nil)
		w := httptest.NewRecorder()
		srv.APIHandler(w, r)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	}

	// 2. Authenticated as tenant1 -> 200 for key1
	{
		r := httptest.NewRequest("GET", "/api/v1/media/"+key1, nil)
		r.Header.Set("Authorization", "Bearer good")
		w := httptest.NewRecorder()
		srv.APIHandler(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if w.Header().Get("Content-Type") != "image/png" {
			t.Errorf("expected Content-Type image/png, got %q", w.Header().Get("Content-Type"))
		}
		if w.Header().Get("Content-Length") != "6" {
			t.Errorf("expected Content-Length 6, got %q", w.Header().Get("Content-Length"))
		}
		if !bytes.Equal(w.Body.Bytes(), data1) {
			t.Errorf("body mismatch: expected %v, got %v", data1, w.Body.Bytes())
		}
	}

	// 3. Query param token authentication -> 200
	{
		r := httptest.NewRequest("GET", "/api/v1/media/"+key1+"?token=good", nil)
		w := httptest.NewRecorder()
		srv.APIHandler(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 with query token, got %d", w.Code)
		}
	}

	// 4. Cross-tenant isolation: Tenant 1 accessing Tenant 2's media -> 404 (NOT 403)
	{
		r := httptest.NewRequest("GET", "/api/v1/media/"+key2, nil)
		r.Header.Set("Authorization", "Bearer good")
		w := httptest.NewRecorder()
		srv.APIHandler(w, r)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for cross-tenant object, got %d", w.Code)
		}
	}

	// 5. Non-existent media -> 404
	{
		r := httptest.NewRequest("GET", "/api/v1/media/media/nonexistent.jpg", nil)
		r.Header.Set("Authorization", "Bearer good")
		w := httptest.NewRecorder()
		srv.APIHandler(w, r)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	}
}

func TestDashboardMediaEndpoints(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dash_media_test_*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)

	diskStore, err := storage.NewDiskStore(tempDir)
	if err != nil {
		t.Fatalf("disk store error: %v", err)
	}

	tenant1 := uuid.New()
	tenant2 := uuid.New()

	repo := &mediaTestRepo{
		objects: make(map[string]domain.MediaObject),
	}

	dash := &DashboardAPIHandler{
		Platform:   repo,
		MediaStore: diskStore,
	}

	// 1. Upload media via POST /dashboard/api/media
	var uploadedKey string
	var uploadedURL string
	{
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "sample-photo.jpg")
		if err != nil {
			t.Fatalf("create form file error: %v", err)
		}
		fileBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
		_, _ = part.Write(fileBytes)
		_ = writer.Close()

		r := httptest.NewRequest("POST", "/dashboard/api/media", body)
		r.Header.Set("Content-Type", writer.FormDataContentType())
		r = r.WithContext(contextWithTenant(r.Context(), tenant1))

		w := httptest.NewRecorder()
		dash.ServeHTTP(w, r)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d, body=%s", w.Code, w.Body.String())
		}

		var resp map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("json unmarshal error: %v", err)
		}

		uploadedKey, _ = resp["media_key"].(string)
		uploadedURL, _ = resp["media_url"].(string)
		if uploadedKey == "" || uploadedURL == "" {
			t.Fatalf("missing media_key or media_url in response: %+v", resp)
		}
		if resp["size"] != float64(len(fileBytes)) {
			t.Errorf("expected size %d, got %v", len(fileBytes), resp["size"])
		}
	}

	// 2. Fetch media via GET /dashboard/api/media/{key} with owner tenant -> 200
	{
		r := httptest.NewRequest("GET", "/dashboard/api/media/"+uploadedKey, nil)
		r = r.WithContext(contextWithTenant(r.Context(), tenant1))

		w := httptest.NewRecorder()
		dash.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if w.Header().Get("Content-Length") != "6" {
			t.Errorf("expected Content-Length 6, got %q", w.Header().Get("Content-Length"))
		}
	}

	// 3. Cross-tenant isolation on dashboard endpoint -> 404
	{
		r := httptest.NewRequest("GET", "/dashboard/api/media/"+uploadedKey, nil)
		r = r.WithContext(contextWithTenant(r.Context(), tenant2))

		w := httptest.NewRecorder()
		dash.ServeHTTP(w, r)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for different tenant, got %d", w.Code)
		}
	}

	// 4. Request without session context -> 401
	{
		r := httptest.NewRequest("GET", "/dashboard/api/media/"+uploadedKey, nil)
		w := httptest.NewRecorder()
		dash.ServeHTTP(w, r)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 without session context, got %d", w.Code)
		}
	}
}
