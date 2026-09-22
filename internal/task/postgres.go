package task

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX — небольшой набор методов pgx, который нужен PostgresStore.
// Такую реализацию можно создать и из *pgxpool.Pool, и из pgx.Tx.
// В первом уроке транзакции разбираются на уровне SQL, но интерфейс уже не
// привязывает хранилище к одному конкретному типу подключения.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// PostgresStore выполняет SQL-запросы к таблице tasks.
//
// В отличие от прошлого Store здесь нет map, nextID и mutex. PostgreSQL сам
// управляет параллельным доступом, выдаёт ID и сохраняет данные между
// перезапусками Go-сервера.
type PostgresStore struct {
	db DBTX
}

func NewPostgresStore(db DBTX) *PostgresStore {
	return &PostgresStore{db: db}
}

// List возвращает задачи в стабильном порядке: сначала новые.
//
// Query используют, когда ожидают несколько строк. Rows обязательно закрываем,
// а после цикла проверяем rows.Err(): ошибка может возникнуть не только при
// запуске запроса, но и во время чтения результата.
func (s *PostgresStore) List(ctx context.Context) ([]Task, error) {
	const query = `
		SELECT id, title, done, created_at, updated_at
		FROM tasks
		ORDER BY id DESC`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("получить список задач: %w", err)
	}
	defer rows.Close()

	tasks := make([]Task, 0)
	for rows.Next() {
		var current Task
		if err := rows.Scan(
			&current.ID,
			&current.Title,
			&current.Done,
			&current.CreatedAt,
			&current.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("прочитать задачу: %w", err)
		}
		tasks = append(tasks, current)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("прочитать список задач: %w", err)
	}

	return tasks, nil
}

// Get получает одну строку по primary key.
// QueryRow откладывает ошибку до Scan, поэтому проверяем именно результат Scan.
func (s *PostgresStore) Get(ctx context.Context, id int64) (Task, error) {
	const query = `
		SELECT id, title, done, created_at, updated_at
		FROM tasks
		WHERE id = $1`

	var result Task
	err := scanTask(s.db.QueryRow(ctx, query, id), &result)
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, fmt.Errorf("получить задачу %d: %w", id, err)
	}

	return result, nil
}

// Create сохраняет Task и сразу получает значения, которые сформировала база.
//
// $1 — параметр запроса. Не склеиваем SQL со строкой title: параметры отделяют
// код запроса от данных и защищают от SQL injection.
func (s *PostgresStore) Create(ctx context.Context, title string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, ErrEmptyTitle
	}

	const query = `
		INSERT INTO tasks (title)
		VALUES ($1)
		RETURNING id, title, done, created_at, updated_at`

	var created Task
	if err := scanTask(s.db.QueryRow(ctx, query, title), &created); err != nil {
		return Task{}, fmt.Errorf("создать задачу: %w", err)
	}

	return created, nil
}

// Patch изменяет только те поля, которые клиент действительно прислал.
//
// nil передаётся в PostgreSQL как NULL. COALESCE оставляет прежнее значение,
// если параметр NULL. Trigger из миграции отдельно обновляет updated_at.
func (s *PostgresStore) Patch(
	ctx context.Context,
	id int64,
	changes Changes,
) (Task, error) {
	if changes.Title == nil && changes.Done == nil {
		return Task{}, ErrNoChanges
	}

	if changes.Title != nil {
		trimmed := strings.TrimSpace(*changes.Title)
		if trimmed == "" {
			return Task{}, ErrEmptyTitle
		}
		changes.Title = &trimmed
	}

	const query = `
		UPDATE tasks
		SET
			title = COALESCE($2, title),
			done = COALESCE($3, done)
		WHERE id = $1
		RETURNING id, title, done, created_at, updated_at`

	var updated Task
	err := scanTask(
		s.db.QueryRow(ctx, query, id, changes.Title, changes.Done),
		&updated,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, fmt.Errorf("изменить задачу %d: %w", id, err)
	}

	return updated, nil
}

// Delete использует RETURNING, чтобы отличить удаление от отсутствующей строки.
// DELETE без RETURNING тоже возможен, но тогда пришлось бы проверять RowsAffected.
func (s *PostgresStore) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1 RETURNING id`

	var deletedID int64
	err := s.db.QueryRow(ctx, query, id).Scan(&deletedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("удалить задачу %d: %w", id, err)
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTask(row rowScanner, task *Task) error {
	return row.Scan(
		&task.ID,
		&task.Title,
		&task.Done,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
}
