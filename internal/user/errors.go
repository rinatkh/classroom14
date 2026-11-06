package user

import "errors"

var (
	ErrInvalidEmail       = errors.New("email is invalid")
	ErrWeakPassword       = errors.New("password must contain between 8 and 72 bytes")
	ErrAlreadyExists      = errors.New("user already exists")
	ErrNotFound           = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
