package types

import "errors"

// ValidationError represents a validation error response
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents a response containing multiple validation errors
type ValidationErrorResponse struct {
	Errors []ValidationError `json:"errors"`
}

var (
	ErrInvalidToken    = errors.New("invalid token")
	ErrUserNotFound    = errors.New("user not found")
	ErrTokenExpired    = errors.New("token expired")
	ErrInvalidPassword = errors.New("invalid password")
	ErrEmailExists     = errors.New("email already exists")
	ErrInvalidEmail    = errors.New("invalid email")
)
