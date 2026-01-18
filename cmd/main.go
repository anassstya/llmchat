package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"llm-chat-backend/internal/config"
	"llm-chat-backend/internal/handlers"
	"llm-chat-backend/internal/repository"
	"llm-chat-backend/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func runMigrations(databaseURL string) error {
	m, err := migrate.New(
		"file://"+filepath.ToSlash("./migrations"),
		databaseURL,
	)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	if err := runMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	startCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbpool, err := pgxpool.New(startCtx, cfg.DatabaseURL)

	if err != nil {
		log.Fatalf("Failed to create db pool: %v", err)
	}

	if err := dbpool.Ping(startCtx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	userRepo := repository.NewPostgresUserRepo(dbpool)
	authService := services.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	llmService, err := services.NewLLMService(cfg.OpenAIKey)
	if err != nil {
		log.Fatal("Failed to create LLM service:", err)
	}

	chatRepo := repository.NewPostgresChatRepo(dbpool)
	chatService := services.NewChatService(chatRepo, llmService)
	chatHandler := handlers.NewChatHandler(chatService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/api/auth/register", authHandler.HandleRegister)
	r.Post("/api/auth/login", authHandler.HandleLogin)
	r.Get("/api/chat/message", chatHandler.HandleChatMessage)
	r.Post("/api/chat/message", chatHandler.HandleChatMessage)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8050"
	}

	srv := &http.Server{
		Addr: ":" + port,
		// ...
	}

	log.Fatal(srv.ListenAndServe())
}
