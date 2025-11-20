package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorCode represents a typed error code
type ErrorCode int

const (
	// General errors
	CodeInternal ErrorCode = iota + 1000
	CodeNotFound
	CodeUnauthorized
	CodeForbidden
	CodeBadRequest
	CodeConflict
	CodeValidation
	CodeTimeout
	CodeRateLimit

	// Database errors
	CodeDatabaseConnection ErrorCode = iota + 2000
	CodeDatabaseQuery
	CodeDatabaseConstraint
	CodeDatabaseMigration

	// Storage errors
	CodeStorageUpload ErrorCode = iota + 3000
	CodeStorageDownload
	CodeStorageDelete
	CodeStorageNotFound
	CodeStorageQuotaExceeded

	// Email errors
	CodeEmailSend ErrorCode = iota + 4000
	CodeEmailTemplate
	CodeEmailConfiguration

	// Authentication errors
	CodeAuthInvalidToken ErrorCode = iota + 5000
	CodeAuthExpiredToken
	CodeAuthInvalidCredentials
	CodeAuthTokenGeneration

	// Module errors
	CodeModuleNotFound ErrorCode = iota + 6000
	CodeModuleAlreadyRegistered
	CodeModuleInitialization
	CodeModuleDependency
)

// Error represents a structured error with code and metadata
type Error struct {
	Code     ErrorCode      `json:"code"`
	Message  string         `json:"message"`
	Details  string         `json:"details,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Cause    error          `json:"-"` // Don't serialize the underlying error
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Cause
}

// WithDetails adds details to the error
func (e *Error) WithDetails(details string) *Error {
	e.Details = details
	return e
}

// WithMetadata adds metadata to the error
func (e *Error) WithMetadata(key string, value any) *Error {
	if e.Metadata == nil {
		e.Metadata = make(map[string]any)
	}
	e.Metadata[key] = value
	return e
}

// WithCause adds the underlying cause
func (e *Error) WithCause(cause error) *Error {
	e.Cause = cause
	return e
}

// HTTPStatus returns the appropriate HTTP status code for the error
func (e *Error) HTTPStatus() int {
	switch e.Code {
	case CodeNotFound, CodeStorageNotFound, CodeModuleNotFound:
		return http.StatusNotFound
	case CodeUnauthorized, CodeAuthInvalidToken, CodeAuthExpiredToken, CodeAuthInvalidCredentials:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeBadRequest, CodeValidation:
		return http.StatusBadRequest
	case CodeConflict, CodeModuleAlreadyRegistered:
		return http.StatusConflict
	case CodeTimeout:
		return http.StatusRequestTimeout
	case CodeRateLimit:
		return http.StatusTooManyRequests
	case CodeStorageQuotaExceeded:
		return http.StatusInsufficientStorage
	default:
		return http.StatusInternalServerError
	}
}

// MarshalJSON implements json.Marshaler
func (e *Error) MarshalJSON() ([]byte, error) {
	type alias Error
	return json.Marshal(&struct {
		*alias
		HTTPStatus int `json:"http_status"`
	}{
		alias:      (*alias)(e),
		HTTPStatus: e.HTTPStatus(),
	})
}

// New creates a new error with the given code and message
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with a Base error
func Wrap(err error, code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// Is checks if the error is of a specific code
func Is(err error, code ErrorCode) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	return false
}

// GetCode extracts the error code from an error
func GetCode(err error) ErrorCode {
	if e, ok := err.(*Error); ok {
		return e.Code
	}
	return CodeInternal
}

// Pre-defined common errors
var (
	ErrInternal     = New(CodeInternal, "Internal server error")
	ErrNotFound     = New(CodeNotFound, "Resource not found")
	ErrUnauthorized = New(CodeUnauthorized, "Unauthorized")
	ErrForbidden    = New(CodeForbidden, "Forbidden")
	ErrBadRequest   = New(CodeBadRequest, "Bad request")
	ErrConflict     = New(CodeConflict, "Resource already exists")
	ErrValidation   = New(CodeValidation, "Validation failed")
	ErrTimeout      = New(CodeTimeout, "Request timeout")
	ErrRateLimit    = New(CodeRateLimit, "Rate limit exceeded")

	ErrDatabaseConnection = New(CodeDatabaseConnection, "Database connection failed")
	ErrDatabaseQuery      = New(CodeDatabaseQuery, "Database query failed")
	ErrDatabaseConstraint = New(CodeDatabaseConstraint, "Database constraint violation")
	ErrDatabaseMigration  = New(CodeDatabaseMigration, "Database migration failed")

	ErrStorageUpload   = New(CodeStorageUpload, "Storage upload failed")
	ErrStorageDownload = New(CodeStorageDownload, "Storage download failed")
	ErrStorageDelete   = New(CodeStorageDelete, "Storage delete failed")
	ErrStorageNotFound = New(CodeStorageNotFound, "File not found in storage")
	ErrStorageQuota    = New(CodeStorageQuotaExceeded, "Storage quota exceeded")

	ErrEmailSend          = New(CodeEmailSend, "Email send failed")
	ErrEmailTemplate      = New(CodeEmailTemplate, "Email template error")
	ErrEmailConfiguration = New(CodeEmailConfiguration, "Email configuration error")

	ErrAuthInvalidToken       = New(CodeAuthInvalidToken, "Invalid authentication token")
	ErrAuthExpiredToken       = New(CodeAuthExpiredToken, "Authentication token expired")
	ErrAuthInvalidCredentials = New(CodeAuthInvalidCredentials, "Invalid credentials")
	ErrAuthTokenGeneration    = New(CodeAuthTokenGeneration, "Token generation failed")

	ErrModuleNotFound          = New(CodeModuleNotFound, "Module not found")
	ErrModuleAlreadyRegistered = New(CodeModuleAlreadyRegistered, "Module already registered")
	ErrModuleInitialization    = New(CodeModuleInitialization, "Module initialization failed")
	ErrModuleDependency        = New(CodeModuleDependency, "Module dependency error")
)

// ValidationError represents a validation error with field-specific details
type ValidationError struct {
	*Error
	Fields map[string][]string `json:"fields"`
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		Error:  New(CodeValidation, message),
		Fields: make(map[string][]string),
	}
}

// AddField adds a field error to the validation error
func (v *ValidationError) AddField(field, message string) *ValidationError {
	if v.Fields[field] == nil {
		v.Fields[field] = []string{}
	}
	v.Fields[field] = append(v.Fields[field], message)
	return v
}

// HasFields returns true if there are field errors
func (v *ValidationError) HasFields() bool {
	return len(v.Fields) > 0
}

// ErrorHandler provides common error handling utilities
type ErrorHandler struct {
	logger Logger
}

// Logger interface for error logging
type Logger interface {
	Error(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Info(msg string, fields ...any)
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger Logger) *ErrorHandler {
	return &ErrorHandler{logger: logger}
}

// Handle handles an error by logging it appropriately
func (h *ErrorHandler) Handle(err error) {
	if err == nil {
		return
	}

	if baseErr, ok := err.(*Error); ok {
		switch baseErr.Code {
		case CodeInternal, CodeDatabaseConnection, CodeDatabaseMigration:
			h.logger.Error("Internal error occurred",
				"code", baseErr.Code,
				"message", baseErr.Message,
				"details", baseErr.Details,
			)
		case CodeNotFound, CodeUnauthorized, CodeForbidden:
			h.logger.Warn("Client error occurred",
				"code", baseErr.Code,
				"message", baseErr.Message,
			)
		default:
			h.logger.Info("Error occurred",
				"code", baseErr.Code,
				"message", baseErr.Message,
			)
		}
	} else {
		h.logger.Error("Unhandled error occurred", "error", err.Error())
	}
}

// Recovery creates a panic recovery handler
func Recovery(logger Logger) func() {
	return func() {
		if r := recover(); r != nil {
			logger.Error("Panic recovered", "panic", r)
		}
	}
}
