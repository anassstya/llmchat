package handlers

import (
	"encoding/json"
	"fmt"
	"llm-chat-backend/internal/services"
	"log"
	"net/http"
	"strconv"
)

type ChatRequest struct {
	Message string `json:"message"`
}

type ChatHandler struct {
	chatService *services.ChatService
}

func NewChatHandler(chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

func (h *ChatHandler) HandleChatMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	userIDStr := r.Header.Get("X-User-ID")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	message := r.URL.Query().Get("message")
	if message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	onChunk := func(text string) error {
		b, _ := json.Marshal(map[string]string{"delta": text})
		_, err := fmt.Fprintf(w, "data: %s\n\n", b)
		if err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	if err := h.chatService.ChatStream(r.Context(), userID, message, onChunk); err != nil {
		log.Printf("chat error: user=%d, msg=%q, err=%v", userID, message, err)
		return
	}

	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func (h *ChatHandler) HandleChatHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	history, err := h.chatService.Repo.GetHistory(r.Context(), userID)
	if err != nil {
		log.Printf("history error: user=%d, err=%v", userID, err)
		http.Error(w, "Failed to load history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(history)
}
