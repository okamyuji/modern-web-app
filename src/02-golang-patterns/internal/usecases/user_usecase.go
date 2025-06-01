package usecases

import (
	"context"
	"fmt"
	"golang-patterns/internal/domain/models"
	"golang-patterns/internal/interfaces/repositories"
)

// UserUseCase handles user business logic
type UserUseCase struct {
	userRepo repositories.UserRepository
	logger   repositories.Logger
}

// NewUserUseCase creates a new UserUseCase
func NewUserUseCase(userRepo repositories.UserRepository, logger repositories.Logger) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

// GetUser gets a user by ID
func (uc *UserUseCase) GetUser(ctx context.Context, id string) (*models.User, error) {
	uc.logger.Info("Getting user", "id", id)
	
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return user, nil
}

// CreateUser creates a new user
func (uc *UserUseCase) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	uc.logger.Info("Creating user", "name", user.Name)
	
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	err := uc.userRepo.Save(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}
	
	return user, nil
}

// GetAllUsers gets all users
func (uc *UserUseCase) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	uc.logger.Info("Getting all users")
	
	users, err := uc.userRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	
	return users, nil
}