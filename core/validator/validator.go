package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the go-playground validator with Base-specific functionality
type Validator struct {
	validate *validator.Validate
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// ValidationErrors is a slice of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (v ValidationErrors) Error() string {
	var messages []string
	for _, err := range v {
		messages = append(messages, err.Message)
	}
	return strings.Join(messages, "; ")
}

// New creates a new validator instance
func New() *Validator {
	v := validator.New()

	// Register custom tag name function to use json tags for field names
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		if name == "" {
			return fld.Name
		}
		return name
	})

	return &Validator{validate: v}
}

// Validate validates a struct and returns user-friendly errors
func (v *Validator) Validate(data interface{}) ValidationErrors {
	var validationErrors ValidationErrors

	err := v.validate.Struct(data)
	if err == nil {
		return nil
	}

	// Handle validation errors
	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, err := range validatorErrors {
			validationErrors = append(validationErrors, ValidationError{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Value:   fmt.Sprintf("%v", err.Value()),
				Message: v.getErrorMessage(err),
			})
		}
	}

	return validationErrors
}

// ValidateVar validates a single variable
func (v *Validator) ValidateVar(field interface{}, tag string) ValidationErrors {
	err := v.validate.Var(field, tag)
	if err == nil {
		return nil
	}

	var validationErrors ValidationErrors
	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, err := range validatorErrors {
			validationErrors = append(validationErrors, ValidationError{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Value:   fmt.Sprintf("%v", err.Value()),
				Message: v.getErrorMessage(err),
			})
		}
	}

	return validationErrors
}

// getErrorMessage returns a user-friendly error message for a validation error
func (v *Validator) getErrorMessage(fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()
	param := fe.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, param)
	case "numeric":
		return fmt.Sprintf("%s must be a number", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// RegisterValidation adds a custom validation rule
func (v *Validator) RegisterValidation(tag string, fn validator.Func) error {
	return v.validate.RegisterValidation(tag, fn)
}

// Global validator instance
var defaultValidator = New()

// Validate validates using the default validator instance
func Validate(data interface{}) ValidationErrors {
	return defaultValidator.Validate(data)
}

// ValidateVar validates a single variable using the default validator instance
func ValidateVar(field interface{}, tag string) ValidationErrors {
	return defaultValidator.ValidateVar(field, tag)
}
