package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rinatkh/classroom14/internal/auth"
	"github.com/rinatkh/classroom14/internal/httpapi"
	"github.com/rinatkh/classroom14/internal/httpapi/handler"
	"github.com/rinatkh/classroom14/internal/httpapi/middleware"
	"github.com/rinatkh/classroom14/internal/logging"
	"github.com/rinatkh/classroom14/internal/repository/postgres"
	"github.com/rinatkh/classroom14/internal/task"
	"github.com/rinatkh/classroom14/internal/user"
)

func main() {
	if err := logging.Configure("task-api"); err != nil {
		logging.Fatal("invalid logging configuration", "error", err)
	}
	if err := run(); err != nil {
		logging.Fatal("task-api stopped", "error", err)
	}
}

func run() error {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return errors.New("JWT_SECRET должен содержать не менее 32 байт; см. README")
	}
	ttl, err := time.ParseDuration(envOr("TOKEN_TTL", "2h"))
	if err != nil || ttl <= 0 {
		return errors.New("TOKEN_TTL должен быть положительной duration")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL обязателен; см. docs/setup.md")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return errors.New("не удалось создать pool PostgreSQL; проверьте DATABASE_URL")
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return errors.New("PostgreSQL недоступен; проверьте подключение и docs/setup.md")
	}
	slog.Info("PostgreSQL connected")

	// Главная замена занятия: две concrete реализации Repository в main.
	// UseCase, handlers, JWT и middleware ниже остаются прежними.
	userRepository := postgres.NewUserRepository(pool)
	taskRepository := postgres.NewTaskRepository(pool)
	passwordHasher := auth.NewBcryptHasher(0)
	tokenManager := auth.NewJWTManager(secret, ttl)
	authUseCase := user.NewUseCase(userRepository, passwordHasher, tokenManager)
	taskUseCase := task.NewUseCase(taskRepository)
	authHandler := handler.NewAuthHandler(authUseCase)
	taskHandler := handler.NewTaskHandler(taskUseCase)
	authMiddleware := middleware.NewAuth(tokenManager)
	router := httpapi.NewRouter(authHandler, taskHandler, authMiddleware)
	server := &http.Server{Addr: envOr("HTTP_ADDR", ":8080"), Handler: router, ReadHeaderTimeout: 5 * time.Second}
	slog.Info("HTTP server starting", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
