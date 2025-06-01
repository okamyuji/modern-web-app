package repositories

import (
	"context"
	"golang-patterns/internal/domain/models"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*models.User, error)
	Save(ctx context.Context, user *models.User) error
	GetAll(ctx context.Context) ([]*models.User, error)
	Delete(ctx context.Context, id string) error
}

// Logger defines the interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}