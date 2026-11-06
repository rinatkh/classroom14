package task

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

const maxActiveTasks = 20

// UseCase выполняет бизнес-сценарии Task API.
// HTTP, JSON и конкретное хранилище остаются за пределами этого пакета.
type UseCase struct {
	repository Repository
	// Защищает последовательность CountActive → Create/Update в одном процессе.
	// Mutex самого repository защищает map, но не составное business decision.
	changeMu sync.Mutex
}

// NewUseCase получает зависимость снаружи. Это базовая dependency injection.
func NewUseCase(repository Repository) *UseCase {
	return &UseCase{repository: repository}
}

// Create применяет бизнес-правила, затем просит Repository сохранить Task.
func (uc *UseCase) Create(ctx context.Context, userID int64, input CreateInput) (Task, error) {
	if userID <= 0 {
		return Task{}, ErrForbidden
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return Task{}, ErrEmptyTitle
	}

	uc.changeMu.Lock()
	defer uc.changeMu.Unlock()
	if err := uc.ensureActiveSlot(ctx, userID); err != nil {
		return Task{}, err
	}

	created, err := uc.repository.Create(ctx, userID, CreateInput{Title: title})
	if err != nil {
		return Task{}, fmt.Errorf("создать задачу: %w", err)
	}
	return created, nil
}

func (uc *UseCase) Get(ctx context.Context, userID, id int64) (Task, error) {
	if userID <= 0 {
		return Task{}, ErrForbidden
	}
	result, err := uc.repository.Get(ctx, id)
	if err != nil {
		return Task{}, fmt.Errorf("получить задачу: %w", err)
	}
	if result.OwnerID != userID {
		return Task{}, ErrForbidden
	}
	return result, nil
}

func (uc *UseCase) List(ctx context.Context, userID int64) ([]Task, error) {
	if userID <= 0 {
		return nil, ErrForbidden
	}
	result, err := uc.repository.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("получить список задач: %w", err)
	}
	return result, nil
}

// Update показывает настоящую роль UseCase: одно HTTP-действие требует
// сначала прочитать текущее состояние, проверить правила и только затем писать.
func (uc *UseCase) Update(
	ctx context.Context,
	userID, id int64,
	input UpdateInput,
) (Task, error) {
	if input.Title == nil && input.Done == nil {
		return Task{}, ErrNoChanges
	}
	uc.changeMu.Lock()
	defer uc.changeMu.Unlock()

	current, err := uc.Get(ctx, userID, id)
	if err != nil {
		return Task{}, fmt.Errorf("получить задачу перед изменением: %w", err)
	}

	if input.Title != nil {
		trimmed := strings.TrimSpace(*input.Title)
		if trimmed == "" {
			return Task{}, ErrEmptyTitle
		}
		input.Title = &trimmed
	}

	if input.Done != nil && *input.Done && current.Done {
		return Task{}, ErrAlreadyComplete
	}
	if input.Done != nil && !*input.Done && current.Done {
		if err := uc.ensureActiveSlot(ctx, userID); err != nil {
			return Task{}, err
		}
	}

	updated, err := uc.repository.Update(ctx, id, input)
	if err != nil {
		return Task{}, fmt.Errorf("изменить задачу: %w", err)
	}
	return updated, nil
}

func (uc *UseCase) Delete(ctx context.Context, userID, id int64) error {
	if _, err := uc.Get(ctx, userID, id); err != nil {
		return err
	}
	if err := uc.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("удалить задачу: %w", err)
	}
	return nil
}

func (uc *UseCase) ensureActiveSlot(ctx context.Context, userID int64) error {
	count, err := uc.repository.CountActive(ctx, userID)
	if err != nil {
		return fmt.Errorf("посчитать активные задачи: %w", err)
	}
	if count >= maxActiveTasks {
		return ErrActiveLimit
	}
	return nil
}
