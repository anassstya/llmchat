package repository

import (
	"context"
	"errors"
	"llm-chat-backend/internal/domain"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepo struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepo(db *pgxpool.Pool) *PostgresUserRepo {
	return &PostgresUserRepo{db: db}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (p *PostgresUserRepo) Create(ctx context.Context, email, passwordHash string) (int64, error) {
	email = normalizeEmail(email)

	const q = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id;
	`

	var id int64

	err := p.db.QueryRow(ctx, q, email, passwordHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, errors.New("email already taken")
		}
		return 0, err
	}

	return id, nil
}

func (p *PostgresUserRepo) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	email = normalizeEmail(email)

	const q = `
		SELECT id, email, password_hash
		FROM users
		WHERE lower(email) = $1
		LIMIT 1 
	
	`

	var u domain.User
	err := p.db.QueryRow(ctx, q, email).Scan(&u.ID, &u.Email, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, errors.New("user not found")
		}
		return domain.User{}, err
	}

	return u, nil
}
