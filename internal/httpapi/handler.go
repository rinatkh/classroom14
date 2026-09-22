package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/rinatkh/classroom15/internal/task"
)

type Handler struct {
	repository task.Repository
}

func NewHandler(repository task.Repository) *Handler {
	return &Handler{repository: repository}
}

type healthResponse struct {
	Status string `json:"status"`
}

type createTaskRequest struct {
	Title string `json:"title"`
}

type patchTaskRequest struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

// ListTasks передаёт context запроса в PostgreSQL.
// Если клиент отменит HTTP-запрос, драйвер получит тот же сигнал отмены.
func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.repository.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось получить задачи")
		return
	}

	writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	result, err := h.repository.Get(r.Context(), id)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var input createTaskRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	created, err := h.repository.Create(r.Context(), input.Title)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/tasks/%d", created.ID))
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) PatchTask(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var input patchTaskRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	updated, err := h.repository.Patch(r.Context(), id, task.Changes{
		Title: input.Title,
		Done:  input.Done,
	})
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	if err := h.repository.Delete(r.Context(), id); err != nil {
		writeRepositoryError(w, err)
		return
	}

	// Для 204 body должен отсутствовать, поэтому writeJSON здесь не вызываем.
	w.WriteHeader(http.StatusNoContent)
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "id должен быть положительным целым числом")
		return 0, false
	}

	return id, true
}

func writeRepositoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, task.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, task.ErrEmptyTitle), errors.Is(err, task.ErrNoChanges):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		// Внутреннюю ошибку SQL не отправляем клиенту: она может содержать
		// подробности схемы и подключения. Позже такие ошибки попадут в лог.
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}
