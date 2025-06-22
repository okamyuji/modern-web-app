package repositories

import (
	"context"
	"golang-patterns/internal/domain/models"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetAll(ctx context.Context) ([]*models.User, error)
	Update(ctx context.Context, user *models.User) (*models.User, error)
	Delete(ctx context.Context, id string) error
	Save(ctx context.Context, user *models.User) error // Legacy method for backward compatibility
	
	// Target specification - Advanced query operations
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByDepartment(ctx context.Context, department string) ([]*models.User, error)
	GetByPosition(ctx context.Context, position string) ([]*models.User, error)
	GetActiveUsers(ctx context.Context) ([]*models.User, error)
	GetInactiveUsers(ctx context.Context) ([]*models.User, error)
	
	// Load display - Pagination and sorting
	GetUsersWithQuery(ctx context.Context, params *models.QueryParams) (*models.PaginatedResult, error)
	GetUsersWithFilter(ctx context.Context, filter *models.UserFilter, pagination *models.PaginationParams, sort *models.SortParams) (*models.PaginatedResult, error)
	CountUsers(ctx context.Context) (int, error)
	CountUsersWithFilter(ctx context.Context, filter *models.UserFilter) (int, error)
	
	// Progressive loading
	GetUsersBatch(ctx context.Context, params *models.ProgressiveLoadParams) (*models.ProgressiveResult, error)
	GetUsersAfterCursor(ctx context.Context, cursor string, limit int) ([]*models.User, error)
	GetUsersBeforeCursor(ctx context.Context, cursor string, limit int) ([]*models.User, error)
	
	// Statistics and analytics
	GetUserStats(ctx context.Context) (*models.UserStats, error)
	GetDepartmentStats(ctx context.Context) (map[string]int, error)
	GetPositionStats(ctx context.Context) (map[string]int, error)
	GetRecentSignups(ctx context.Context, days int) ([]*models.User, error)
	
	// Bulk operations
	BulkCreate(ctx context.Context, users []*models.User) ([]*models.User, error)
	BulkUpdate(ctx context.Context, users []*models.User) ([]*models.User, error)
	BulkDelete(ctx context.Context, ids []string) error
	
	// Search operations
	SearchUsers(ctx context.Context, query string, pagination *models.PaginationParams) (*models.PaginatedResult, error)
	SearchUsersByField(ctx context.Context, field string, value string, pagination *models.PaginationParams) (*models.PaginatedResult, error)
}

// Logger defines the interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}