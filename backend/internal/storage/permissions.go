package storage

import (
	"context"

	"github.com/google/uuid"
)

// LogPermissionCheck records an operator permission check in the audit trail.
// Every check is recorded — both allowed and denied — so administrators can
// audit who attempted which action and when.
func (p *PostgresStore) LogPermissionCheck(
	ctx context.Context,
	operatorID uuid.UUID,
	action string,
	resource string,
	resourceID uuid.UUID,
	allowed bool,
	reason string,
	ip string,
	ua string,
) error {
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO operator_permission_checks
		(operator_id, action, resource, resource_id, allowed, reason, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		operatorID,
		action,
		nullString(resource),
		nullUUID(resourceID),
		allowed,
		reason,
		nullString(ip),
		nullString(ua),
	)
	return err
}

// nullUUID converts the nil UUID to SQL NULL so optional id columns stay NULL.
func nullUUID(id uuid.UUID) interface{} {
	if id == uuid.Nil {
		return nil
	}
	return id
}
