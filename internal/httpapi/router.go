package httpapi

import (
	"net/http"

	"github.com/rinatkh/classroom15/internal/task"
)

// NewRouter связывает внешний HTTP-договор с методами Handler.
//
// Маршруты не изменились после перехода с map на PostgreSQL. Клиент по-прежнему
// работает с тем же API. Изменился только компонент, который handler вызывает
// для чтения и сохранения задач.
func NewRouter(repository task.Repository) http.Handler {
	handler := NewHandler(repository)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("GET /tasks", handler.ListTasks)
	mux.HandleFunc("POST /tasks", handler.CreateTask)
	mux.HandleFunc("GET /tasks/{id}", handler.GetTask)
	mux.HandleFunc("PATCH /tasks/{id}", handler.PatchTask)
	mux.HandleFunc("DELETE /tasks/{id}", handler.DeleteTask)

	return mux
}
