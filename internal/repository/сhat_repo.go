package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Message struct {
	ID        int64     `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatRepository interface {
	SaveMessage(ctx context.Context, userID int64, role, content string) error
	GetHistory(ctx context.Context, userID int64) ([]Message, error)
}

type PostgresChatRepo struct {
	db *pgxpool.Pool
}

func NewPostgresChatRepo(db *pgxpool.Pool) *PostgresChatRepo {
	return &PostgresChatRepo{db: db}
}

func (c *PostgresChatRepo) SaveMessage(ctx context.Context, userID int64, role, content string) error {
	const q = `
	INSERT INTO chat_history(user_id, role, content)
	VALUES ($1, $2, $3)
	`

	_, err := c.db.Exec(ctx, q, userID, role, content)
	if err != nil {
		return err
	}

	return nil
}

func (c *PostgresChatRepo) GetHistory(ctx context.Context, userID int64) ([]Message, error) {
	const q = `
	 SELECT id, role, content, created_at
     FROM chat_history
	 WHERE user_id=$1
	 ORDER BY created_at, id
	`

	rows, err := c.db.Query(ctx, q, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	return messages, rows.Err()
}
