package task

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestIntegrationPostgresStoreLifecycle(t *testing.T) {
	pool, ctx := testPool(t)
	if _, err := pool.Exec(ctx, "TRUNCATE task_comments, tasks RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("очистить таблицу: %v", err)
	}

	store := NewPostgresStore(pool)
	created, err := store.Create(ctx, "Проверить PostgreSQL")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID != 1 || created.Done {
		t.Fatalf("неожиданная созданная задача: %+v", created)
	}

	// Небольшая пауза делает проверку trigger детерминированной даже на машине,
	// где два SQL-запроса могут получить почти одинаковые timestamps.
	if _, err := pool.Exec(ctx, "SELECT pg_sleep(0.01)"); err != nil {
		t.Fatalf("подготовить проверку updated_at: %v", err)
	}

	done := true
	updated, err := store.Patch(ctx, created.ID, Changes{Done: &done})
	if err != nil {
		t.Fatalf("Patch: %v", err)
	}
	if !updated.Done {
		t.Fatal("Patch не изменил done")
	}
	if !updated.UpdatedAt.After(created.UpdatedAt) {
		t.Fatalf("trigger не обновил updated_at: before=%v after=%v", created.UpdatedAt, updated.UpdatedAt)
	}

	items, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("List вернул %d строк, ожидалась 1", len(items))
	}

	if err := store.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get после Delete: %v, ожидался ErrNotFound", err)
	}
}

func TestIntegrationSchemaRules(t *testing.T) {
	pool, ctx := testPool(t)
	if _, err := pool.Exec(ctx, "TRUNCATE task_comments, tasks RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("очистить таблицы: %v", err)
	}

	// CHECK защищает таблицу даже при INSERT напрямую, в обход handler.
	if _, err := pool.Exec(ctx, "INSERT INTO tasks (title) VALUES ('   ')"); err == nil {
		t.Fatal("title из пробелов должен нарушать CHECK")
	}

	// FOREIGN KEY не разрешает комментарий к отсутствующей задаче.
	if _, err := pool.Exec(ctx,
		"INSERT INTO task_comments (task_id, body) VALUES (999, 'Нет такой задачи')",
	); err == nil {
		t.Fatal("комментарий с неизвестным task_id должен нарушать FOREIGN KEY")
	}

	var taskID int64
	if err := pool.QueryRow(ctx,
		"INSERT INTO tasks (title) VALUES ('Проверить CASCADE') RETURNING id",
	).Scan(&taskID); err != nil {
		t.Fatalf("создать задачу: %v", err)
	}
	if _, err := pool.Exec(ctx,
		"INSERT INTO task_comments (task_id, body) VALUES ($1, 'Комментарий')",
		taskID,
	); err != nil {
		t.Fatalf("создать комментарий: %v", err)
	}
	if _, err := pool.Exec(ctx, "DELETE FROM tasks WHERE id = $1", taskID); err != nil {
		t.Fatalf("удалить задачу: %v", err)
	}

	var comments int
	if err := pool.QueryRow(ctx,
		"SELECT count(*) FROM task_comments WHERE task_id = $1",
		taskID,
	).Scan(&comments); err != nil {
		t.Fatalf("посчитать комментарии: %v", err)
	}
	if comments != 0 {
		t.Fatalf("ON DELETE CASCADE оставил %d комментариев", comments)
	}
}

func testPool(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL не задан: пропускаем тест с настоящей PostgreSQL")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("создать pool: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("подключиться к тестовой PostgreSQL: %v", err)
	}

	return pool, ctx
}
