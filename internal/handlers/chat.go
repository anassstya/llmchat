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

type ChatResponse struct {
	Status   string `json:"status"`
	BotReply string `json:"bot_reply"`
}

type ChatHandler struct {
	chatService *services.ChatService
}

func NewChatHandler(chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}
func (h *ChatHandler) HandleChatMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://celadon-platypus-83c9b4.netlify.app")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")

	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	message := r.URL.Query().Get("message")
	if message == "" {
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		message = req.Message
	}
	defer r.Body.Close()

	userIDStr := r.Header.Get("X-User-ID")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	onChunk := func(text string) error {
		if _, err := fmt.Fprintf(w, "data: %s\n\n", text); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	err = h.chatService.ChatStream(r.Context(), userID, message, onChunk)
	if err != nil {
		log.Printf("chat error: user=%d, msg=%q, err=%v", userID, message, err)
		return
	}

	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func (h *ChatHandler) HandleChatHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://celadon-platypus-83c9b4.netlify.app")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
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
	json.NewEncoder(w).Encode(history)
}
