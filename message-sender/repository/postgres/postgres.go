package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"message-sender/config"
	"message-sender/model"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(cfg *config.DatabaseConfig) (*Repository, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) GetUnsentMessages(ctx context.Context, limit int) ([]model.Message, error) {
	query := `
		SELECT id, content, recipient, is_sent, sent_at, message_id
		FROM messages
		WHERE is_sent = false
		ORDER BY id ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query unsent messages: %w", err)
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var msg model.Message
		var sentAt sql.NullTime
		var messageID sql.NullString

		if err := rows.Scan(&msg.ID, &msg.Content, &msg.Recipient, &msg.IsSent, &sentAt, &messageID); err != nil {
			return nil, fmt.Errorf("failed to scan message row: %w", err)
		}

		if sentAt.Valid {
			msg.SentAt = sentAt.Time
		}

		if messageID.Valid {
			msg.MessageID = messageID.String
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating message rows: %w", err)
	}

	return messages, nil
}

func (r *Repository) MarkMessageAsSent(ctx context.Context, id uint, messageID string, sentAt time.Time) error {
	query := `
		UPDATE messages
		SET is_sent = true, message_id = $1, sent_at = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, messageID, sentAt, id)
	if err != nil {
		return fmt.Errorf("failed to mark message as sent: %w", err)
	}

	return nil
}

func (r *Repository) GetSentMessages(ctx context.Context, page, limit int) ([]model.Message, int, error) {
	offset := (page - 1) * limit

	var total int
	countQuery := `SELECT COUNT(*) FROM messages WHERE is_sent = true`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to get sent messages count: %w", err)
	}

	query := `
		SELECT id, content, recipient, is_sent, sent_at, message_id
		FROM messages
		WHERE is_sent = true
		ORDER BY sent_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query sent messages: %w", err)
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var msg model.Message
		var sentAt sql.NullTime
		var messageID sql.NullString

		if err := rows.Scan(&msg.ID, &msg.Content, &msg.Recipient, &msg.IsSent, &sentAt, &messageID); err != nil {
			return nil, 0, fmt.Errorf("failed to scan message row: %w", err)
		}

		if sentAt.Valid {
			msg.SentAt = sentAt.Time
		}

		if messageID.Valid {
			msg.MessageID = messageID.String
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating message rows: %w", err)
	}

	return messages, total, nil
}

func (r *Repository) GetMessageByID(ctx context.Context, id uint) (*model.Message, error) {
	query := `
		SELECT id, content, recipient, is_sent, sent_at, message_id
		FROM messages
		WHERE id = $1
	`

	var msg model.Message
	var sentAt sql.NullTime
	var messageID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID, &msg.Content, &msg.Recipient, &msg.IsSent, &sentAt, &messageID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message by ID: %w", err)
	}

	if sentAt.Valid {
		msg.SentAt = sentAt.Time
	}

	if messageID.Valid {
		msg.MessageID = messageID.String
	}

	return &msg, nil
}

func (r *Repository) SaveMessage(ctx context.Context, message *model.Message) error {
	query := `
		INSERT INTO messages (content, recipient, is_sent)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	return r.db.QueryRowContext(ctx, query, message.Content, message.Recipient, message.IsSent).Scan(&message.ID)
}

func (r *Repository) InitSchema(ctx context.Context) error {
	schema := `
		CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			content VARCHAR(160) NOT NULL,
			recipient VARCHAR(15) NOT NULL,
			is_sent BOOLEAN DEFAULT FALSE,
			sent_at TIMESTAMP,
			message_id VARCHAR(36)
		)
	`

	_, err := r.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}
