package models

import "errors"

// Domain errors
var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidUserName  = errors.New("invalid user name")
	ErrInvalidUserEmail = errors.New("invalid user email")
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return "validation error on field " + e.Field + ": " + e.Message
	}
	return "validation error: " + e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		Message: message,
	}
}

// NewFieldValidationError creates a new field validation error
func NewFieldValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// NotFoundError represents a not found error
type NotFoundError struct {
	Resource string
	ID       string
}

func (e NotFoundError) Error() string {
	return e.Resource + " with ID " + e.ID + " not found"
}