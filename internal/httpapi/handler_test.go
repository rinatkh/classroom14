package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rinatkh/classroom15/internal/task"
)

type fakeRepository struct {
	listResult   []task.Task
	getResult    task.Task
	createResult task.Task
	patchResult  task.Task
	err          error
	createdTitle string
	deletedID    int64
}

func (f *fakeRepository) List(context.Context) ([]task.Task, error) {
	return f.listResult, f.err
}

func (f *fakeRepository) Get(context.Context, int64) (task.Task, error) {
	return f.getResult, f.err
}

func (f *fakeRepository) Create(_ context.Context, title string) (task.Task, error) {
	f.createdTitle = title
	return f.createResult, f.err
}

func (f *fakeRepository) Patch(context.Context, int64, task.Changes) (task.Task, error) {
	return f.patchResult, f.err
}

func (f *fakeRepository) Delete(_ context.Context, id int64) error {
	f.deletedID = id
	return f.err
}

func TestListTasks(t *testing.T) {
	t.Parallel()

	repository := &fakeRepository{listResult: []task.Task{{ID: 1, Title: "SQL"}}}
	response := performRequest(t, NewRouter(repository), http.MethodGet, "/tasks", nil)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, ожидался %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
}

func TestCreateTask(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.January, 2, 12, 0, 0, 0, time.UTC)
	repository := &fakeRepository{createResult: task.Task{
		ID: 7, Title: "PostgreSQL", CreatedAt: now, UpdatedAt: now,
	}}
	body := bytes.NewBufferString(`{"title":"PostgreSQL"}`)
	response := performRequest(t, NewRouter(repository), http.MethodPost, "/tasks", body)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, ожидался %d; body=%s", response.Code, http.StatusCreated, response.Body.String())
	}
	if repository.createdTitle != "PostgreSQL" {
		t.Fatalf("repository получил title = %q", repository.createdTitle)
	}
	if got := response.Header().Get("Location"); got != "/tasks/7" {
		t.Fatalf("Location = %q", got)
	}
}

func TestCreateTaskRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	response := performRequest(
		t,
		NewRouter(&fakeRepository{}),
		http.MethodPost,
		"/tasks",
		bytes.NewBufferString(`{"title":`),
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, ожидался %d", response.Code, http.StatusBadRequest)
	}
}

func TestGetUnknownTask(t *testing.T) {
	t.Parallel()

	repository := &fakeRepository{err: task.ErrNotFound}
	response := performRequest(t, NewRouter(repository), http.MethodGet, "/tasks/999", nil)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, ожидался %d", response.Code, http.StatusNotFound)
	}
}

func TestGetRejectsInvalidID(t *testing.T) {
	t.Parallel()

	response := performRequest(t, NewRouter(&fakeRepository{}), http.MethodGet, "/tasks/abc", nil)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, ожидался %d", response.Code, http.StatusBadRequest)
	}
}

func TestDeleteHasNoBody(t *testing.T) {
	t.Parallel()

	repository := &fakeRepository{}
	response := performRequest(t, NewRouter(repository), http.MethodDelete, "/tasks/3", nil)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, ожидался %d", response.Code, http.StatusNoContent)
	}
	if response.Body.Len() != 0 {
		t.Fatalf("204 не должен содержать body: %q", response.Body.String())
	}
	if repository.deletedID != 3 {
		t.Fatalf("удалён id = %d", repository.deletedID)
	}
}

func TestInternalRepositoryErrorIsHidden(t *testing.T) {
	t.Parallel()

	repository := &fakeRepository{err: errors.New("password=secret")}
	response := performRequest(t, NewRouter(repository), http.MethodGet, "/tasks", nil)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, ожидался %d", response.Code, http.StatusInternalServerError)
	}
	var result map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("прочитать JSON: %v", err)
	}
	if result["error"] != "не удалось получить задачи" {
		t.Fatalf("error = %q", result["error"])
	}
}

func performRequest(
	t *testing.T,
	handler http.Handler,
	method string,
	path string,
	body *bytes.Buffer,
) *httptest.ResponseRecorder {
	t.Helper()

	if body == nil {
		body = &bytes.Buffer{}
	}
	request := httptest.NewRequest(method, path, body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
