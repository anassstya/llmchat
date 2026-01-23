package services

import (
	"context"
	"llm-chat-backend/internal/repository"
	"strings"
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

	var fullResponse strings.Builder

	wrapChunk := func(text string) error {
		fullResponse.WriteString(text)
		return onChunk(text)
	}

	if err := s.Llm.GetResponse(ctx, message, history, wrapChunk); err != nil {
		return err
	}

	return s.Repo.SaveMessage(ctx, userID, "assistant", fullResponse.String())
}

func (s *ChatService) GetHistory(ctx context.Context, userID int64) ([]repository.Message, error) {
	return s.Repo.GetHistory(ctx, userID)
}
