package user

import (
	"context"
	"errors"
	"strings"
)

// PasswordHasher и TokenIssuer — небольшие интерфейсы потребителя. UseCase
// зависит от нужного поведения, а не от bcrypt и конкретной JWT-библиотеки.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type TokenIssuer interface {
	Issue(userID int64) (string, error)
}

type UseCase struct {
	repository Repository
	hasher     PasswordHasher
	tokens     TokenIssuer
}

func NewUseCase(repository Repository, hasher PasswordHasher, tokens TokenIssuer) *UseCase {
	return &UseCase{repository: repository, hasher: hasher, tokens: tokens}
}

func (uc *UseCase) Register(ctx context.Context, input RegisterInput) (User, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	password := input.Password
	if !strings.Contains(email, "@") || strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@") {
		return User{}, ErrInvalidEmail
	}
	if len(password) < 8 || len(password) > 72 {
		return User{}, ErrWeakPassword
	}

	if _, err := uc.repository.GetByEmail(ctx, email); err == nil {
		return User{}, ErrAlreadyExists
	} else if !errors.Is(err, ErrNotFound) {
		return User{}, err
	}

	hash, err := uc.hasher.Hash(password)
	if err != nil {
		return User{}, err
	}

	return uc.repository.Create(ctx, User{Email: email, PasswordHash: hash})
}

func (uc *UseCase) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	if len(input.Password) < 8 || len(input.Password) > 72 {
		return LoginResult{}, ErrInvalidCredentials
	}
	email := strings.ToLower(strings.TrimSpace(input.Email))
	value, err := uc.repository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	if err := uc.hasher.Compare(value.PasswordHash, input.Password); err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	token, err := uc.tokens.Issue(value.ID)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{Token: token}, nil
}
