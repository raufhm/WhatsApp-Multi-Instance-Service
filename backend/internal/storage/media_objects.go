package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
)

// RecordMediaObject records a newly stored media file for the tenant.
// If the object key already exists, it updates the mime_type and size.
func (p *PostgresStore) RecordMediaObject(ctx context.Context, tenantID uuid.UUID, objectKey, mimeType string, size int64) error {
	var tenantVal any
	if tenantID == uuid.Nil {
		tenantVal = nil
	} else {
		tenantVal = tenantID
	}

	query := `
		INSERT INTO media_objects (tenant_id, object_key, mime_type, size)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (object_key) DO UPDATE
		SET mime_type = EXCLUDED.mime_type,
		    size = EXCLUDED.size
	`
	_, err := p.db.ExecContext(ctx, query, tenantVal, objectKey, mimeType, size)
	if err != nil {
		return fmt.Errorf("record media object: %w", err)
	}
	return nil
}

// GetMediaObject looks up a media object by tenant ID and object key.
// If the object key does not belong to the given tenant (or is not found), it returns an error.
func (p *PostgresStore) GetMediaObject(ctx context.Context, tenantID uuid.UUID, objectKey string) (domain.MediaObject, error) {
	var obj domain.MediaObject
	var dbTenantID sql.NullString

	query := `
		SELECT id, tenant_id, object_key, mime_type, size, created_at
		FROM media_objects
		WHERE object_key = $1
	`
	err := p.db.QueryRowContext(ctx, query, objectKey).Scan(
		&obj.ID,
		&dbTenantID,
		&obj.ObjectKey,
		&obj.MimeType,
		&obj.Size,
		&obj.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.MediaObject{}, fmt.Errorf("media object not found: %s", objectKey)
		}
		return domain.MediaObject{}, fmt.Errorf("get media object: %w", err)
	}

	if dbTenantID.Valid && dbTenantID.String != "" {
		parsedTenant, err := uuid.Parse(dbTenantID.String)
		if err == nil {
			obj.TenantID = parsedTenant
		}
	}

	// Enforce tenant isolation: if object has a tenant and doesn't match caller tenant, return not found
	if obj.TenantID != uuid.Nil && tenantID != uuid.Nil && obj.TenantID != tenantID {
		return domain.MediaObject{}, fmt.Errorf("media object not found: %s", objectKey)
	}

	return obj, nil
}
