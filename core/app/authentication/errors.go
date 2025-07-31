package authentication

import "errors"

// Auth-specific errors
var (
	ErrInvalidToken    = errors.New("invalid token")
	ErrUserNotFound    = errors.New("user not found")
	ErrTokenExpired    = errors.New("token expired")
	ErrInvalidPassword = errors.New("invalid password")
	ErrEmailExists     = errors.New("email already exists")
	ErrInvalidEmail    = errors.New("invalid email")
)
