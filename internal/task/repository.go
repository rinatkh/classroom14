package task

import (
	"context"
	"errors"
)

var (
	ErrNotFound   = errors.New("задача не найдена")
	ErrEmptyTitle = errors.New("название задачи не должно быть пустым")
	ErrNoChanges  = errors.New("не указано ни одного поля для изменения")
)

// Repository описывает операции с задачами, которые нужны HTTP-слою.
//
// Handler не обязан знать, находится ли реализация в PostgreSQL, памяти или
// тестовом fake. На этом занятии рабочая реализация — PostgresStore.
type Repository interface {
	List(ctx context.Context) ([]Task, error)
	Get(ctx context.Context, id int64) (Task, error)
	Create(ctx context.Context, title string) (Task, error)
	Patch(ctx context.Context, id int64, changes Changes) (Task, error)
	Delete(ctx context.Context, id int64) error
}
