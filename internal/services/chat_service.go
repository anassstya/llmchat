package services

import (
	"context"
	"llm-chat-backend/internal/repository"
)

type ChatService struct {
	Repo repository.ChatRepository
	Llm  *LLMService
}

func NewChatService(repo repository.ChatRepository, llm *LLMService) *ChatService {
	return &ChatService{
		Repo: repo,
		Llm:  llm,
	}
}

func (s *ChatService) ChatStream(ctx context.Context, userID int64, message string, onChunk func(string) error) error {
	if err := s.Repo.SaveMessage(ctx, userID, "user", message); err != nil {
		return err
	}

	history, err := s.Repo.GetHistory(ctx, userID)
	if err != nil {
		return err
	}

	return s.Llm.GetResponse(ctx, message, history, onChunk)
}
