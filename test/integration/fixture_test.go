package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/rinatkh/classroom14/internal/auth"
	"github.com/rinatkh/classroom14/internal/httpapi"
	"github.com/rinatkh/classroom14/internal/httpapi/handler"
	"github.com/rinatkh/classroom14/internal/httpapi/middleware"
	"github.com/rinatkh/classroom14/internal/repository/memory"
	"github.com/rinatkh/classroom14/internal/task"
	"github.com/rinatkh/classroom14/internal/user"
	"golang.org/x/crypto/bcrypt"
)

// The previous Task API tests run unchanged for an authenticated owner.
func newAuthenticatedRouter(t *testing.T) http.Handler {
	t.Helper()
	tokens := auth.NewJWTManager("0123456789abcdef0123456789abcdef", time.Hour)
	app := httpapi.NewRouter(handler.NewAuthHandler(user.NewUseCase(memory.NewUserRepository(), auth.NewBcryptHasher(bcrypt.MinCost), tokens)), handler.NewTaskHandler(task.NewUseCase(memory.NewTaskRepository())), middleware.NewAuth(tokens))
	body := `{"email":"alice@example.com","password":"secret123"}`
	created := request(t, app, http.MethodPost, "/register", body)
	assertCode(t, created, 201)
	login := request(t, app, http.MethodPost, "/login", body)
	assertCode(t, login, 200)
	var result struct{ Token string }
	if err := json.Unmarshal(login.Body.Bytes(), &result); err != nil || result.Token == "" {
		t.Fatalf("login=%s, %v", login.Body.String(), err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Authorization", "Bearer "+result.Token)
		app.ServeHTTP(w, r)
	})
}
