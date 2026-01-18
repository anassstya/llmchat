package services

import (
	"context"
	"errors"
	"io"
	"llm-chat-backend/internal/repository"
	"time"

	"github.com/sashabaranov/go-openai"
)

type LLMService struct {
	client *openai.Client
}

func toOpenAIRole(role string) string {
	switch role {
	case "assistant":
		return openai.ChatMessageRoleAssistant
	case "system":
		return openai.ChatMessageRoleSystem
	default:
		return openai.ChatMessageRoleUser
	}
}

func NewLLMService(apiKey string) (*LLMService, error) {
	if apiKey == "" {
		return nil, errors.New("API key is empty")
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://router.huggingface.co/v1"

	client := openai.NewClientWithConfig(config)

	return &LLMService{
		client: client,
	}, nil
}

func (s *LLMService) GetResponse(
	ctx context.Context,
	message string,
	history []repository.Message,
	onChunk func(text string) error,
) error {
	ctxChat, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var messages []openai.ChatCompletionMessage
	for _, m := range history {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    toOpenAIRole(m.Role),
			Content: m.Content,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})

	req := openai.ChatCompletionRequest{
		Model:    "meta-llama/Meta-Llama-3-8B-Instruct",
		Messages: messages,
		Stream:   true,
	}

	stream, err := s.client.CreateChatCompletionStream(ctxChat, req)
	if err != nil {
		return err
	}
	defer stream.Close()

	for {
		select {
		case <-ctxChat.Done():
			return ctxChat.Err()
		default:
		}

		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		if len(resp.Choices) == 0 {
			continue
		}

		delta := resp.Choices[0].Delta.Content
		if delta == "" {
			continue
		}

		if err := onChunk(delta); err != nil {
			return err
		}
	}
}
