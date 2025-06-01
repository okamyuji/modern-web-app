package usecases

import (
	"context"
	"fmt"
	"golang-patterns/internal/domain/models"
	"golang-patterns/internal/interfaces/repositories"
	"time"
)

// UserUseCase handles user business logic
type UserUseCase struct {
	userRepo repositories.UserRepository
	logger   repositories.Logger
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(userRepo repositories.UserRepository, logger repositories.Logger) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

// === Basic CRUD Operations ===

// CreateUser creates a new user with enhanced validation
func (uc *UserUseCase) CreateUser(ctx context.Context, req *models.UserCreateRequest) (*models.User, error) {
	uc.logger.Info("Creating user", "name", req.Name, "email", req.Email)

	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.Error("User creation validation failed", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check for duplicate email
	_, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, models.NewValidationError("email already exists")
	}

	// Convert request to user entity
	user := req.ToUser()

	// Create user
	createdUser, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		uc.logger.Error("Failed to create user", "error", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	uc.logger.Info("User created successfully", "id", createdUser.ID, "email", createdUser.Email)
	return createdUser, nil
}

// GetUser gets a user by ID
func (uc *UserUseCase) GetUser(ctx context.Context, id string) (*models.User, error) {
	uc.logger.Info("Getting user", "id", id)

	if id == "" {
		return nil, models.NewValidationError("user ID is required")
	}

	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get user", "id", id, "error", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UpdateUser updates an existing user
func (uc *UserUseCase) UpdateUser(ctx context.Context, id string, req *models.UserUpdateRequest) (*models.User, error) {
	uc.logger.Info("Updating user", "id", id)

	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.Error("User update validation failed", "id", id, "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing user
	existingUser, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get user for update", "id", id, "error", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check for email conflict if email is being changed
	if req.Email != nil && *req.Email != existingUser.Email {
		_, err := uc.userRepo.GetByEmail(ctx, *req.Email)
		if err == nil {
			return nil, models.NewValidationError("email already exists")
		}
	}

	// Apply updates
	existingUser.ApplyUpdate(req)

	// Update user
	updatedUser, err := uc.userRepo.Update(ctx, existingUser)
	if err != nil {
		uc.logger.Error("Failed to update user", "id", id, "error", err)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	uc.logger.Info("User updated successfully", "id", id)
	return updatedUser, nil
}

// DeleteUser deletes a user
func (uc *UserUseCase) DeleteUser(ctx context.Context, id string) error {
	uc.logger.Info("Deleting user", "id", id)

	if id == "" {
		return models.NewValidationError("user ID is required")
	}

	// Check if user exists
	_, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get user for deletion", "id", id, "error", err)
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Delete user
	err = uc.userRepo.Delete(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to delete user", "id", id, "error", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	uc.logger.Info("User deleted successfully", "id", id)
	return nil
}

// GetAllUsers gets all users (legacy method)
func (uc *UserUseCase) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	uc.logger.Info("Getting all users")

	users, err := uc.userRepo.GetAll(ctx)
	if err != nil {
		uc.logger.Error("Failed to get all users", "error", err)
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	uc.logger.Info("Retrieved all users", "count", len(users))
	return users, nil
}

// === Target Specification - Advanced Query Operations ===

// GetUserByEmail gets a user by email
func (uc *UserUseCase) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	uc.logger.Info("Getting user by email", "email", email)

	if email == "" {
		return nil, models.NewValidationError("email is required")
	}

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		uc.logger.Error("Failed to get user by email", "email", email, "error", err)
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetUsersByDepartment gets users by department
func (uc *UserUseCase) GetUsersByDepartment(ctx context.Context, department string) ([]*models.User, error) {
	uc.logger.Info("Getting users by department", "department", department)

	if department == "" {
		return nil, models.NewValidationError("department is required")
	}

	users, err := uc.userRepo.GetByDepartment(ctx, department)
	if err != nil {
		uc.logger.Error("Failed to get users by department", "department", department, "error", err)
		return nil, fmt.Errorf("failed to get users by department: %w", err)
	}

	uc.logger.Info("Retrieved users by department", "department", department, "count", len(users))
	return users, nil
}

// GetUsersByPosition gets users by position
func (uc *UserUseCase) GetUsersByPosition(ctx context.Context, position string) ([]*models.User, error) {
	uc.logger.Info("Getting users by position", "position", position)

	if position == "" {
		return nil, models.NewValidationError("position is required")
	}

	users, err := uc.userRepo.GetByPosition(ctx, position)
	if err != nil {
		uc.logger.Error("Failed to get users by position", "position", position, "error", err)
		return nil, fmt.Errorf("failed to get users by position: %w", err)
	}

	uc.logger.Info("Retrieved users by position", "position", position, "count", len(users))
	return users, nil
}

// GetActiveUsers gets all active users
func (uc *UserUseCase) GetActiveUsers(ctx context.Context) ([]*models.User, error) {
	uc.logger.Info("Getting active users")

	users, err := uc.userRepo.GetActiveUsers(ctx)
	if err != nil {
		uc.logger.Error("Failed to get active users", "error", err)
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}

	uc.logger.Info("Retrieved active users", "count", len(users))
	return users, nil
}

// GetInactiveUsers gets all inactive users
func (uc *UserUseCase) GetInactiveUsers(ctx context.Context) ([]*models.User, error) {
	uc.logger.Info("Getting inactive users")

	users, err := uc.userRepo.GetInactiveUsers(ctx)
	if err != nil {
		uc.logger.Error("Failed to get inactive users", "error", err)
		return nil, fmt.Errorf("failed to get inactive users: %w", err)
	}

	uc.logger.Info("Retrieved inactive users", "count", len(users))
	return users, nil
}

// === Load Display - Pagination and Sorting ===

// GetUsersWithQuery gets users with complex query parameters
func (uc *UserUseCase) GetUsersWithQuery(ctx context.Context, params *models.QueryParams) (*models.PaginatedResult, error) {
	uc.logger.Info("Getting users with query", "page", params.Pagination.Page, "page_size", params.Pagination.PageSize)

	// Validate query parameters
	if err := params.Pagination.Validate(); err != nil {
		return nil, fmt.Errorf("pagination validation failed: %w", err)
	}
	if err := params.Sort.Validate(); err != nil {
		return nil, fmt.Errorf("sort validation failed: %w", err)
	}
	if err := params.Filter.Validate(); err != nil {
		return nil, fmt.Errorf("filter validation failed: %w", err)
	}

	result, err := uc.userRepo.GetUsersWithQuery(ctx, params)
	if err != nil {
		uc.logger.Error("Failed to get users with query", "error", err)
		return nil, fmt.Errorf("failed to get users with query: %w", err)
	}

	uc.logger.Info("Retrieved users with query", "total", result.Total, "page", result.Page)
	return result, nil
}

// GetUsersWithPagination gets users with pagination and sorting
func (uc *UserUseCase) GetUsersWithPagination(ctx context.Context, page, pageSize int, sortField, sortOrder string) (*models.PaginatedResult, error) {
	uc.logger.Info("Getting users with pagination", "page", page, "page_size", pageSize, "sort_field", sortField, "sort_order", sortOrder)

	// Create query parameters
	pagination := models.NewPaginationParams(page, pageSize)
	sort := models.NewSortParams(sortField, sortOrder)
	filter := &models.UserFilter{} // Empty filter

	params := &models.QueryParams{
		Filter:     filter,
		Pagination: pagination,
		Sort:       sort,
	}

	return uc.GetUsersWithQuery(ctx, params)
}

// SearchUsers searches users by query string
func (uc *UserUseCase) SearchUsers(ctx context.Context, query string, page, pageSize int) (*models.PaginatedResult, error) {
	uc.logger.Info("Searching users", "query", query, "page", page, "page_size", pageSize)

	if query == "" {
		return nil, models.NewValidationError("search query is required")
	}

	pagination := models.NewPaginationParams(page, pageSize)

	result, err := uc.userRepo.SearchUsers(ctx, query, pagination)
	if err != nil {
		uc.logger.Error("Failed to search users", "query", query, "error", err)
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	uc.logger.Info("User search completed", "query", query, "total", result.Total)
	return result, nil
}

// === Progressive Loading ===

// GetUsersBatch gets users in batches for progressive loading
func (uc *UserUseCase) GetUsersBatch(ctx context.Context, batchSize int, cursor, direction string) (*models.ProgressiveResult, error) {
	uc.logger.Info("Getting users batch", "batch_size", batchSize, "cursor", cursor, "direction", direction)

	params := models.NewProgressiveLoadParams(batchSize, cursor, direction)

	// Validate parameters
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("progressive load validation failed: %w", err)
	}

	result, err := uc.userRepo.GetUsersBatch(ctx, params)
	if err != nil {
		uc.logger.Error("Failed to get users batch", "error", err)
		return nil, fmt.Errorf("failed to get users batch: %w", err)
	}

	users := result.Data.([]*models.User)
	uc.logger.Info("Retrieved users batch", "count", len(users), "has_more", result.HasMore)
	return result, nil
}

// === Statistics and Analytics ===

// GetUserStats gets comprehensive user statistics
func (uc *UserUseCase) GetUserStats(ctx context.Context) (*models.UserStats, error) {
	uc.logger.Info("Getting user statistics")

	stats, err := uc.userRepo.GetUserStats(ctx)
	if err != nil {
		uc.logger.Error("Failed to get user stats", "error", err)
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	uc.logger.Info("Retrieved user statistics", "total_users", stats.TotalUsers, "active_users", stats.ActiveUsers)
	return stats, nil
}

// GetDepartmentStats gets department-wise statistics
func (uc *UserUseCase) GetDepartmentStats(ctx context.Context) (map[string]int, error) {
	uc.logger.Info("Getting department statistics")

	stats, err := uc.userRepo.GetDepartmentStats(ctx)
	if err != nil {
		uc.logger.Error("Failed to get department stats", "error", err)
		return nil, fmt.Errorf("failed to get department stats: %w", err)
	}

	uc.logger.Info("Retrieved department statistics", "departments", len(stats))
	return stats, nil
}

// GetRecentSignups gets users who signed up recently
func (uc *UserUseCase) GetRecentSignups(ctx context.Context, days int) ([]*models.User, error) {
	uc.logger.Info("Getting recent signups", "days", days)

	if days <= 0 {
		return nil, models.NewValidationError("days must be greater than 0")
	}

	users, err := uc.userRepo.GetRecentSignups(ctx, days)
	if err != nil {
		uc.logger.Error("Failed to get recent signups", "days", days, "error", err)
		return nil, fmt.Errorf("failed to get recent signups: %w", err)
	}

	uc.logger.Info("Retrieved recent signups", "days", days, "count", len(users))
	return users, nil
}

// === Form Processing - Bulk Operations ===

// CreateUsersInBulk creates multiple users at once
func (uc *UserUseCase) CreateUsersInBulk(ctx context.Context, requests []*models.UserCreateRequest) ([]*models.User, error) {
	uc.logger.Info("Creating users in bulk", "count", len(requests))

	if len(requests) == 0 {
		return nil, models.NewValidationError("no users to create")
	}

	if len(requests) > 100 {
		return nil, models.NewValidationError("bulk create limited to 100 users")
	}

	// Validate all requests first
	for i, req := range requests {
		if err := req.Validate(); err != nil {
			return nil, fmt.Errorf("validation failed for user %d: %w", i+1, err)
		}
	}

	// Convert requests to users
	var users []*models.User
	for _, req := range requests {
		user := req.ToUser()
		users = append(users, user)
	}

	// Create users in bulk
	createdUsers, err := uc.userRepo.BulkCreate(ctx, users)
	if err != nil {
		uc.logger.Error("Failed to create users in bulk", "error", err)
		return nil, fmt.Errorf("failed to create users in bulk: %w", err)
	}

	uc.logger.Info("Users created in bulk successfully", "count", len(createdUsers))
	return createdUsers, nil
}

// UpdateUsersInBulk updates multiple users at once
func (uc *UserUseCase) UpdateUsersInBulk(ctx context.Context, updates map[string]*models.UserUpdateRequest) ([]*models.User, error) {
	uc.logger.Info("Updating users in bulk", "count", len(updates))

	if len(updates) == 0 {
		return nil, models.NewValidationError("no users to update")
	}

	if len(updates) > 100 {
		return nil, models.NewValidationError("bulk update limited to 100 users")
	}

	// Validate all requests and get existing users
	var usersToUpdate []*models.User
	for id, req := range updates {
		if err := req.Validate(); err != nil {
			return nil, fmt.Errorf("validation failed for user %s: %w", id, err)
		}

		existingUser, err := uc.userRepo.GetByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get user %s: %w", id, err)
		}

		existingUser.ApplyUpdate(req)
		usersToUpdate = append(usersToUpdate, existingUser)
	}

	// Update users in bulk
	updatedUsers, err := uc.userRepo.BulkUpdate(ctx, usersToUpdate)
	if err != nil {
		uc.logger.Error("Failed to update users in bulk", "error", err)
		return nil, fmt.Errorf("failed to update users in bulk: %w", err)
	}

	uc.logger.Info("Users updated in bulk successfully", "count", len(updatedUsers))
	return updatedUsers, nil
}

// DeleteUsersInBulk deletes multiple users at once
func (uc *UserUseCase) DeleteUsersInBulk(ctx context.Context, ids []string) error {
	uc.logger.Info("Deleting users in bulk", "count", len(ids))

	if len(ids) == 0 {
		return models.NewValidationError("no users to delete")
	}

	if len(ids) > 100 {
		return models.NewValidationError("bulk delete limited to 100 users")
	}

	// Delete users in bulk
	err := uc.userRepo.BulkDelete(ctx, ids)
	if err != nil {
		uc.logger.Error("Failed to delete users in bulk", "error", err)
		return fmt.Errorf("failed to delete users in bulk: %w", err)
	}

	uc.logger.Info("Users deleted in bulk successfully", "count", len(ids))
	return nil
}

// === Progressive Enhancement Features ===

// ActivateUser activates a user account
func (uc *UserUseCase) ActivateUser(ctx context.Context, id string) (*models.User, error) {
	uc.logger.Info("Activating user", "id", id)

	isActive := true
	req := &models.UserUpdateRequest{
		IsActive: &isActive,
	}

	user, err := uc.UpdateUser(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to activate user: %w", err)
	}

	uc.logger.Info("User activated successfully", "id", id)
	return user, nil
}

// DeactivateUser deactivates a user account
func (uc *UserUseCase) DeactivateUser(ctx context.Context, id string) (*models.User, error) {
	uc.logger.Info("Deactivating user", "id", id)

	isActive := false
	req := &models.UserUpdateRequest{
		IsActive: &isActive,
	}

	user, err := uc.UpdateUser(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate user: %w", err)
	}

	uc.logger.Info("User deactivated successfully", "id", id)
	return user, nil
}

// UpdateLastLogin updates the last login time for a user
func (uc *UserUseCase) UpdateLastLogin(ctx context.Context, id string) (*models.User, error) {
	uc.logger.Info("Updating last login", "id", id)

	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	now := time.Now()
	user.LastLoginAt = &now

	updatedUser, err := uc.userRepo.Update(ctx, user)
	if err != nil {
		uc.logger.Error("Failed to update last login", "id", id, "error", err)
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}

	uc.logger.Info("Last login updated successfully", "id", id)
	return updatedUser, nil
}

// GetUserSummary gets a summary view of user data
func (uc *UserUseCase) GetUserSummary(ctx context.Context, id string) (map[string]interface{}, error) {
	uc.logger.Info("Getting user summary", "id", id)

	user, err := uc.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"id":                id,
		"name":              user.Name,
		"email":             user.Email,
		"department":        user.Department,
		"position":          user.Position,
		"is_active":         user.IsActive,
		"member_since":      user.CreatedAt.Format("2006-01-02"),
		"last_updated":      user.UpdatedAt.Format("2006-01-02 15:04:05"),
		"has_logged_in":     user.LastLoginAt != nil,
	}

	if user.LastLoginAt != nil {
		summary["last_login"] = user.LastLoginAt.Format("2006-01-02 15:04:05")
	}

	return summary, nil
}