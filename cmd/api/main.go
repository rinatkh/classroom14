package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rinatkh/classroom15/internal/httpapi"
	"github.com/rinatkh/classroom15/internal/task"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("переменная DATABASE_URL не задана; см. .env.example")
	}

	// На запуск подключения даём ограниченное время. Если PostgreSQL недоступен,
	// сервер не должен бесконечно висеть до первого HTTP-запроса.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("создать pool подключений: %v", err)
	}
	defer pool.Close()

	// New создаёт объект pool, но отдельный Ping подтверждает, что база реально
	// доступна и параметры подключения верны.
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("подключиться к PostgreSQL: %v", err)
	}

	store := task.NewPostgresStore(pool)
	router := httpapi.NewRouter(store)

	log.Println("Task API с PostgreSQL запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
