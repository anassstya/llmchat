package services

import (
	"context"
	"errors"
	"llm-chat-backend/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash string) (int64, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
}

type AuthService struct {
	users UserRepository
}

func NewAuthService(users UserRepository) *AuthService {
	return &AuthService{users: users}
}

func (s *AuthService) Register(ctx context.Context, email string, password string) (int64, error) {
	if email == "" || password == "" {
		return 0, errors.New("email and password required")
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	passwordHash := string(hashBytes)

	return s.users.Create(ctx, email, passwordHash)
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	if email == "" || password == "" {
		return domain.User{}, errors.New("email and password required")
	}

	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return domain.User{}, errors.New("invalid credentials")
	}

	return u, nil
}
