// Package user содержит внутреннюю модель пользователя и сценарии регистрации
// и входа. Пакет не знает про HTTP, JSON, JWT-библиотеку или способ хранения.
package user

// User — внутренняя модель приложения. PasswordHash никогда не попадает в DTO
// ответа и не записывается в token.
type User struct {
	ID           int64
	Email        string
	PasswordHash string
}

type RegisterInput struct {
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginResult struct {
	Token string
}
