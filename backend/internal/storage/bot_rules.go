package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
)

// scanBotRuleSet scans a bot_rule_sets row into a BotRuleSet, unmarshaling the
// rules JSONB column into the struct slice.
func scanBotRuleSet(row scanner, set *domain.BotRuleSet) error {
	var rulesJSON []byte
	if err := row.Scan(&set.ID, &set.TenantID, &set.Version, &rulesJSON, &set.IsActive, &set.CreatedAt, &set.UpdatedAt); err != nil {
		return err
	}
	if len(rulesJSON) > 0 {
		if err := json.Unmarshal(rulesJSON, &set.Rules); err != nil {
			return fmt.Errorf("unmarshal bot rules: %w", err)
		}
	}
	return nil
}

const botRuleSetColumns = `id, tenant_id, version, rules, is_active, created_at, updated_at`

// GetActiveBotRuleSet returns the currently active ruleset for a tenant. When
// no ruleset is active, sql.ErrNoRows is returned so callers can fall back to
// compiled-in defaults.
func (p *PostgresStore) GetActiveBotRuleSet(tenantID uuid.UUID) (domain.BotRuleSet, error) {
	var set domain.BotRuleSet
	err := scanBotRuleSet(p.db.QueryRow(
		`SELECT `+botRuleSetColumns+` FROM bot_rule_sets WHERE tenant_id=$1 AND is_active=TRUE`, tenantID), &set)
	return set, err
}

// SaveBotRuleSet persists a new ruleset version (is_active=FALSE) so prior
// versions remain available for rollback. The version is computed under a lock
// to keep the per-tenant version sequence gapless.
func (p *PostgresStore) SaveBotRuleSet(tenantID uuid.UUID, rules []domain.BotRule) (domain.BotRuleSet, error) {
	rulesJSON, err := json.Marshal(rules)
	if err != nil {
		return domain.BotRuleSet{}, err
	}
	tx, err := p.db.Begin()
	if err != nil {
		return domain.BotRuleSet{}, err
	}
	defer tx.Rollback()

	var maxVersion sql.NullInt64
	if err := tx.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM bot_rule_sets WHERE tenant_id=$1 FOR UPDATE`, tenantID).Scan(&maxVersion); err != nil {
		return domain.BotRuleSet{}, err
	}
	nextVersion := int(maxVersion.Int64) + 1

	var set domain.BotRuleSet
	if err := scanBotRuleSet(tx.QueryRow(
		`INSERT INTO bot_rule_sets (tenant_id, version, rules, is_active) VALUES ($1,$2,$3,FALSE)
		 RETURNING `+botRuleSetColumns, tenantID, nextVersion, rulesJSON), &set); err != nil {
		return domain.BotRuleSet{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.BotRuleSet{}, err
	}
	return set, nil
}

// ListBotRuleSets returns every ruleset version for a tenant, newest first.
func (p *PostgresStore) ListBotRuleSets(tenantID uuid.UUID) ([]domain.BotRuleSet, error) {
	rows, err := p.db.Query(`SELECT `+botRuleSetColumns+` FROM bot_rule_sets WHERE tenant_id=$1 ORDER BY version DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sets := []domain.BotRuleSet{}
	for rows.Next() {
		var set domain.BotRuleSet
		if err := scanBotRuleSet(rows, &set); err != nil {
			return nil, err
		}
		sets = append(sets, set)
	}
	return sets, rows.Err()
}

// ActivateBotRuleSetVersion atomically deactivates the current active ruleset
// and activates the named version, retaining all prior versions for rollback.
func (p *PostgresStore) ActivateBotRuleSetVersion(tenantID uuid.UUID, version int) (domain.BotRuleSet, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return domain.BotRuleSet{}, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE bot_rule_sets SET is_active=FALSE, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$1 AND is_active=TRUE`, tenantID); err != nil {
		return domain.BotRuleSet{}, err
	}

	var set domain.BotRuleSet
	err = scanBotRuleSet(tx.QueryRow(
		`UPDATE bot_rule_sets SET is_active=TRUE, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$1 AND version=$2
		 RETURNING `+botRuleSetColumns, tenantID, version), &set)
	if err == sql.ErrNoRows {
		return domain.BotRuleSet{}, fmt.Errorf("ruleset version %d not found", version)
	}
	if err != nil {
		return domain.BotRuleSet{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.BotRuleSet{}, err
	}
	return set, nil
}
