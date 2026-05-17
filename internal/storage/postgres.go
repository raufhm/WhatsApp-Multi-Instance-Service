package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/raufhm/whatsapp-testing/domain"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := runMigrations(db); err != nil {
		log.Printf("DB: Migration error: %v", err)
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

func runMigrations(db *sql.DB) error {
	log.Println("DB: Starting migrations...")
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("instance: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("up: %w", err)
	}
	log.Println("DB: Migrations complete.")
	return nil
}

func (p *PostgresStore) DispatchMessage(meta domain.MessageMetadata) {
	query := `INSERT INTO whatsmeow_messages (whatsapp_id, host_id, sender, recipient, content, is_group, direction, msg_type, reaction_target, media_url, status, timestamp) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	          ON CONFLICT (whatsapp_id) DO UPDATE 
	          SET status = EXCLUDED.status`
	_, err := p.db.Exec(query, meta.WhatsappID, meta.HostID, meta.Sender, meta.Recipient, meta.Content, meta.IsGroup, string(meta.Direction), string(meta.Type), meta.ReactionTarget, meta.MediaURL, string(meta.Status), meta.Timestamp)
	if err != nil {
		log.Printf("PG Store DispatchMessage Error: %v", err)
	}
}

func (p *PostgresStore) DispatchReceipt(receipt domain.Receipt) {
	query := `INSERT INTO whatsmeow_message_receipts (whatsapp_id, recipient_id, status, timestamp) 
	          VALUES ($1, $2, $3, $4)
	          ON CONFLICT (whatsapp_id, recipient_id, status) DO NOTHING`
	_, err := p.db.Exec(query, receipt.WhatsappID, receipt.Recipient, string(receipt.Status), receipt.Timestamp)
	if err != nil {
		log.Printf("PG Store DispatchReceipt Error: %v", err)
	}

	// Update main message status if read
	if receipt.Status == domain.StatusRead || receipt.Status == domain.StatusDelivered {
		updateQuery := `UPDATE whatsmeow_messages SET status = $1 WHERE whatsapp_id = $2 AND direction = 'OUTGOING'`
		_, _ = p.db.Exec(updateQuery, string(receipt.Status), receipt.WhatsappID)
	}
}

func (p *PostgresStore) UpdateInstanceStatus(hostID string, status domain.InstanceStatus, isConnected bool) {
	query := `INSERT INTO whatsmeow_instances (host_id, status, is_connected, last_seen)
	          VALUES ($1, $2, $3, NOW())
	          ON CONFLICT (host_id) DO UPDATE
	          SET status = EXCLUDED.status,
	              is_connected = EXCLUDED.is_connected,
	              last_seen = NOW()`
	if _, err := p.db.Exec(query, hostID, string(status), isConnected); err != nil {
		log.Printf("PG Store UpdateInstanceStatus Error: %v", err)
	}
}
func (p *PostgresStore) UpdateGroup(group domain.GroupInfo) {
	participantsJSON, _ := json.Marshal(group.Participants)
	hostsJSON, _ := json.Marshal(group.Hosts)

	query := `INSERT INTO whatsmeow_groups (group_id, name, description, owner_jid, participants, hosts, participant_count, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	          ON CONFLICT (group_id) DO UPDATE 
	          SET name = EXCLUDED.name,
	              description = EXCLUDED.description,
	              participants = EXCLUDED.participants,
	              hosts = EXCLUDED.hosts,
	              participant_count = EXCLUDED.participant_count,
	              updated_at = EXCLUDED.updated_at`
	_, err := p.db.Exec(query, group.GroupID, group.Name, group.Description, group.OwnerJID, participantsJSON, hostsJSON, group.ParticipantCount, group.CreatedAt, group.UpdatedAt)
	if err != nil {
		log.Printf("PG Store UpdateGroup Error: %v", err)
	}
}
