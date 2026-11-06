package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/rinatkh/classroom14/internal/httpapi/dto"
	"github.com/rinatkh/classroom14/internal/httpapi/mapper"
	"github.com/rinatkh/classroom14/internal/logging"
	"github.com/rinatkh/classroom14/internal/user"
)

type AuthUseCase interface {
	Register(ctx context.Context, input user.RegisterInput) (user.User, error)
	Login(ctx context.Context, input user.LoginInput) (user.LoginResult, error)
}

type AuthHandler struct {
	useCase AuthUseCase
}

func NewAuthHandler(useCase AuthUseCase) *AuthHandler {
	return &AuthHandler{useCase: useCase}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	slog.DebugContext(r.Context(), "HTTP request", "operation", "auth.register")
	var request dto.RegisterRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	result, err := h.useCase.Register(r.Context(), user.RegisterInput{
		Email: request.Email, Password: request.Password,
	})
	switch {
	case errors.Is(err, user.ErrInvalidEmail), errors.Is(err, user.ErrWeakPassword):
		writeError(w, http.StatusBadRequest, err.Error())
		return
	case errors.Is(err, user.ErrAlreadyExists):
		writeError(w, http.StatusConflict, err.Error())
		return
	case err != nil:
		logging.Error(r.Context(), "auth operation failed", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.InfoContext(r.Context(), "user registered", "user_id", result.ID)
	writeJSON(w, http.StatusCreated, mapper.UserToResponse(result))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	slog.DebugContext(r.Context(), "HTTP request", "operation", "auth.login")
	var request dto.LoginRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	result, err := h.useCase.Login(r.Context(), user.LoginInput{
		Email: request.Email, Password: request.Password,
	})
	if errors.Is(err, user.ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if err != nil {
		logging.Error(r.Context(), "auth operation failed", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	slog.InfoContext(r.Context(), "auth credentials issued")
	writeJSON(w, http.StatusOK, dto.TokenResponse{Token: result.Token})
}
