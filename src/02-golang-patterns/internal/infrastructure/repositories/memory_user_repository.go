package repositories

import (
	"context"
	"encoding/base64"
	"fmt"
	"golang-patterns/internal/domain/models"
	"sort"
	"strings"
	"sync"
	"time"
)

// MemoryUserRepository implements UserRepository with advanced features
type MemoryUserRepository struct {
	users       map[string]*models.User
	emailIndex  map[string]string // email -> userID mapping
	mutex       sync.RWMutex
	idCounter   int64
}

// NewMemoryUserRepository creates a new memory repository
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users:      make(map[string]*models.User),
		emailIndex: make(map[string]string),
		mutex:      sync.RWMutex{},
		idCounter:  0,
	}
}

// === Basic CRUD Operations ===

// Create creates a new user
func (r *MemoryUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check for duplicate email
	if _, exists := r.emailIndex[user.Email]; exists {
		return nil, models.NewValidationError("email already exists")
	}

	// Generate ID and set timestamps
	r.idCounter++
	user.ID = fmt.Sprintf("user_%d", r.idCounter)
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Save user and update index
	r.users[user.ID] = user
	r.emailIndex[user.Email] = user.ID

	return user, nil
}

// GetByID gets a user by ID
func (r *MemoryUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, models.NotFoundError{Resource: "user", ID: id}
	}

	// Return a copy to prevent external modification
	userCopy := *user
	return &userCopy, nil
}

// GetAll gets all users
func (r *MemoryUserRepository) GetAll(ctx context.Context) ([]*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	users := make([]*models.User, 0, len(r.users))
	for _, user := range r.users {
		userCopy := *user
		users = append(users, &userCopy)
	}

	return users, nil
}

// Update updates an existing user
func (r *MemoryUserRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	existingUser, exists := r.users[user.ID]
	if !exists {
		return nil, models.NotFoundError{Resource: "user", ID: user.ID}
	}

	// Check for email conflict (if email is being changed)
	if user.Email != existingUser.Email {
		if _, emailExists := r.emailIndex[user.Email]; emailExists {
			return nil, models.NewValidationError("email already exists")
		}
		// Update email index
		delete(r.emailIndex, existingUser.Email)
		r.emailIndex[user.Email] = user.ID
	}

	// Update timestamps
	user.UpdatedAt = time.Now()
	user.CreatedAt = existingUser.CreatedAt // Preserve original creation time

	// Save updated user
	r.users[user.ID] = user

	userCopy := *user
	return &userCopy, nil
}

// Delete deletes a user by ID
func (r *MemoryUserRepository) Delete(ctx context.Context, id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	user, exists := r.users[id]
	if !exists {
		return models.NotFoundError{Resource: "user", ID: id}
	}

	// Remove from indexes
	delete(r.emailIndex, user.Email)
	delete(r.users, id)

	return nil
}

// Save legacy method for backward compatibility
func (r *MemoryUserRepository) Save(ctx context.Context, user *models.User) error {
	if user.ID == "" {
		_, err := r.Create(ctx, user)
		return err
	} else {
		_, err := r.Update(ctx, user)
		return err
	}
}

// === Target Specification - Advanced Query Operations ===

// GetByEmail gets a user by email
func (r *MemoryUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	userID, exists := r.emailIndex[email]
	if !exists {
		return nil, models.NotFoundError{Resource: "user", ID: "email:" + email}
	}

	user := r.users[userID]
	userCopy := *user
	return &userCopy, nil
}

// GetByDepartment gets users by department
func (r *MemoryUserRepository) GetByDepartment(ctx context.Context, department string) ([]*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var users []*models.User
	for _, user := range r.users {
		if user.Department == department {
			userCopy := *user
			users = append(users, &userCopy)
		}
	}

	return users, nil
}

// GetByPosition gets users by position
func (r *MemoryUserRepository) GetByPosition(ctx context.Context, position string) ([]*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var users []*models.User
	for _, user := range r.users {
		if user.Position == position {
			userCopy := *user
			users = append(users, &userCopy)
		}
	}

	return users, nil
}

// GetActiveUsers gets all active users
func (r *MemoryUserRepository) GetActiveUsers(ctx context.Context) ([]*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var users []*models.User
	for _, user := range r.users {
		if user.IsActive {
			userCopy := *user
			users = append(users, &userCopy)
		}
	}

	return users, nil
}

// GetInactiveUsers gets all inactive users
func (r *MemoryUserRepository) GetInactiveUsers(ctx context.Context) ([]*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var users []*models.User
	for _, user := range r.users {
		if !user.IsActive {
			userCopy := *user
			users = append(users, &userCopy)
		}
	}

	return users, nil
}

// === Load Display - Pagination and Sorting ===

// GetUsersWithQuery gets users with complex query parameters
func (r *MemoryUserRepository) GetUsersWithQuery(ctx context.Context, params *models.QueryParams) (*models.PaginatedResult, error) {
	return r.GetUsersWithFilter(ctx, params.Filter, params.Pagination, params.Sort)
}

// GetUsersWithFilter gets users with filtering, pagination, and sorting
func (r *MemoryUserRepository) GetUsersWithFilter(ctx context.Context, filter *models.UserFilter, pagination *models.PaginationParams, sort *models.SortParams) (*models.PaginatedResult, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Apply filtering
	var filteredUsers []*models.User
	for _, user := range r.users {
		if filter == nil || filter.Matches(user) {
			userCopy := *user
			filteredUsers = append(filteredUsers, &userCopy)
		}
	}

	// Apply sorting
	r.sortUsers(filteredUsers, sort)

	// Calculate pagination
	total := len(filteredUsers)
	startIndex := pagination.Offset
	endIndex := startIndex + pagination.PageSize

	if startIndex >= total {
		// Return empty result if offset exceeds total
		return models.NewPaginatedResult([]*models.User{}, total, pagination), nil
	}

	if endIndex > total {
		endIndex = total
	}

	paginatedUsers := filteredUsers[startIndex:endIndex]
	return models.NewPaginatedResult(paginatedUsers, total, pagination), nil
}

// CountUsers counts total users
func (r *MemoryUserRepository) CountUsers(ctx context.Context) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.users), nil
}

// CountUsersWithFilter counts users matching filter
func (r *MemoryUserRepository) CountUsersWithFilter(ctx context.Context, filter *models.UserFilter) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	count := 0
	for _, user := range r.users {
		if filter == nil || filter.Matches(user) {
			count++
		}
	}

	return count, nil
}

// === Progressive Loading ===

// GetUsersBatch gets users in batches for progressive loading
func (r *MemoryUserRepository) GetUsersBatch(ctx context.Context, params *models.ProgressiveLoadParams) (*models.ProgressiveResult, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Get all users and sort by ID for consistent cursor-based pagination
	var allUsers []*models.User
	for _, user := range r.users {
		userCopy := *user
		allUsers = append(allUsers, &userCopy)
	}

	// Sort by ID for consistent ordering
	sort.Slice(allUsers, func(i, j int) bool {
		return allUsers[i].ID < allUsers[j].ID
	})

	var startIndex int
	if params.Cursor != "" {
		// Find position of cursor
		cursorID, err := r.decodeCursor(params.Cursor)
		if err != nil {
			return nil, models.NewValidationError("invalid cursor")
		}

		for i, user := range allUsers {
			if user.ID == cursorID {
				if params.Direction == "forward" {
					startIndex = i + 1
				} else {
					startIndex = i - params.BatchSize
					if startIndex < 0 {
						startIndex = 0
					}
				}
				break
			}
		}
	}

	// Get batch
	endIndex := startIndex + params.BatchSize
	if endIndex > len(allUsers) {
		endIndex = len(allUsers)
	}

	if startIndex >= len(allUsers) {
		// No more data
		return &models.ProgressiveResult{
			Data:    []*models.User{},
			HasMore: false,
		}, nil
	}

	batchUsers := allUsers[startIndex:endIndex]

	// Determine cursors and hasMore
	var nextCursor, prevCursor string
	hasMore := endIndex < len(allUsers)

	if len(batchUsers) > 0 {
		nextCursor = r.encodeCursor(batchUsers[len(batchUsers)-1].ID)
		if startIndex > 0 {
			prevCursor = r.encodeCursor(batchUsers[0].ID)
		}
	}

	return &models.ProgressiveResult{
		Data:       batchUsers,
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasMore:    hasMore,
	}, nil
}

// GetUsersAfterCursor gets users after a specific cursor
func (r *MemoryUserRepository) GetUsersAfterCursor(ctx context.Context, cursor string, limit int) ([]*models.User, error) {
	params := &models.ProgressiveLoadParams{
		BatchSize: limit,
		Cursor:    cursor,
		Direction: "forward",
	}
	result, err := r.GetUsersBatch(ctx, params)
	if err != nil {
		return nil, err
	}
	return result.Data.([]*models.User), nil
}

// GetUsersBeforeCursor gets users before a specific cursor
func (r *MemoryUserRepository) GetUsersBeforeCursor(ctx context.Context, cursor string, limit int) ([]*models.User, error) {
	params := &models.ProgressiveLoadParams{
		BatchSize: limit,
		Cursor:    cursor,
		Direction: "backward",
	}
	result, err := r.GetUsersBatch(ctx, params)
	if err != nil {
		return nil, err
	}
	return result.Data.([]*models.User), nil
}

// === Statistics and Analytics ===

// GetUserStats gets comprehensive user statistics
func (r *MemoryUserRepository) GetUserStats(ctx context.Context) (*models.UserStats, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	stats := &models.UserStats{
		DepartmentStats: make(map[string]int),
		PositionStats:   make(map[string]int),
		AgeDistribution: make(map[string]int),
	}

	oneWeekAgo := time.Now().AddDate(0, 0, -7)
	oneWeekAgoRecentLogin := time.Now().AddDate(0, 0, -1)

	for _, user := range r.users {
		stats.TotalUsers++

		if user.IsActive {
			stats.ActiveUsers++
		} else {
			stats.InactiveUsers++
		}

		// Department stats
		if user.Department != "" {
			stats.DepartmentStats[user.Department]++
		}

		// Position stats
		if user.Position != "" {
			stats.PositionStats[user.Position]++
		}

		// Age distribution
		ageGroup := r.getAgeGroup(user.Age)
		stats.AgeDistribution[ageGroup]++

		// Recent signups (last week)
		if user.CreatedAt.After(oneWeekAgo) {
			stats.LastWeekSignups++
		}

		// Recent logins (last day)
		if user.LastLoginAt != nil && user.LastLoginAt.After(oneWeekAgoRecentLogin) {
			stats.RecentLogins++
		}
	}

	return stats, nil
}

// GetDepartmentStats gets department statistics
func (r *MemoryUserRepository) GetDepartmentStats(ctx context.Context) (map[string]int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	stats := make(map[string]int)
	for _, user := range r.users {
		if user.Department != "" {
			stats[user.Department]++
		}
	}

	return stats, nil
}

// GetPositionStats gets position statistics
func (r *MemoryUserRepository) GetPositionStats(ctx context.Context) (map[string]int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	stats := make(map[string]int)
	for _, user := range r.users {
		if user.Position != "" {
			stats[user.Position]++
		}
	}

	return stats, nil
}

// GetRecentSignups gets users who signed up in the last N days
func (r *MemoryUserRepository) GetRecentSignups(ctx context.Context, days int) ([]*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	cutoff := time.Now().AddDate(0, 0, -days)
	var users []*models.User

	for _, user := range r.users {
		if user.CreatedAt.After(cutoff) {
			userCopy := *user
			users = append(users, &userCopy)
		}
	}

	return users, nil
}

// === Bulk Operations ===

// BulkCreate creates multiple users
func (r *MemoryUserRepository) BulkCreate(ctx context.Context, users []*models.User) ([]*models.User, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var createdUsers []*models.User
	now := time.Now()

	for _, user := range users {
		// Check for duplicate email
		if _, exists := r.emailIndex[user.Email]; exists {
			return nil, models.NewValidationError(fmt.Sprintf("email %s already exists", user.Email))
		}

		// Generate ID and set timestamps
		r.idCounter++
		user.ID = fmt.Sprintf("user_%d", r.idCounter)
		user.CreatedAt = now
		user.UpdatedAt = now

		// Save user and update index
		r.users[user.ID] = user
		r.emailIndex[user.Email] = user.ID

		userCopy := *user
		createdUsers = append(createdUsers, &userCopy)
	}

	return createdUsers, nil
}

// BulkUpdate updates multiple users
func (r *MemoryUserRepository) BulkUpdate(ctx context.Context, users []*models.User) ([]*models.User, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var updatedUsers []*models.User
	now := time.Now()

	for _, user := range users {
		existingUser, exists := r.users[user.ID]
		if !exists {
			return nil, models.NotFoundError{Resource: "user", ID: user.ID}
		}

		// Check for email conflict
		if user.Email != existingUser.Email {
			if _, emailExists := r.emailIndex[user.Email]; emailExists {
				return nil, models.NewValidationError(fmt.Sprintf("email %s already exists", user.Email))
			}
			// Update email index
			delete(r.emailIndex, existingUser.Email)
			r.emailIndex[user.Email] = user.ID
		}

		// Update timestamps
		user.UpdatedAt = now
		user.CreatedAt = existingUser.CreatedAt

		// Save updated user
		r.users[user.ID] = user

		userCopy := *user
		updatedUsers = append(updatedUsers, &userCopy)
	}

	return updatedUsers, nil
}

// BulkDelete deletes multiple users
func (r *MemoryUserRepository) BulkDelete(ctx context.Context, ids []string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// First check all users exist
	for _, id := range ids {
		if _, exists := r.users[id]; !exists {
			return models.NotFoundError{Resource: "user", ID: id}
		}
	}

	// Delete all users
	for _, id := range ids {
		user := r.users[id]
		delete(r.emailIndex, user.Email)
		delete(r.users, id)
	}

	return nil
}

// === Search Operations ===

// SearchUsers searches users by name or email
func (r *MemoryUserRepository) SearchUsers(ctx context.Context, query string, pagination *models.PaginationParams) (*models.PaginatedResult, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var matchedUsers []*models.User
	lowerQuery := strings.ToLower(query)

	for _, user := range r.users {
		if strings.Contains(strings.ToLower(user.Name), lowerQuery) ||
			strings.Contains(strings.ToLower(user.Email), lowerQuery) ||
			strings.Contains(strings.ToLower(user.Department), lowerQuery) ||
			strings.Contains(strings.ToLower(user.Position), lowerQuery) {
			userCopy := *user
			matchedUsers = append(matchedUsers, &userCopy)
		}
	}

	// Apply pagination
	total := len(matchedUsers)
	startIndex := pagination.Offset
	endIndex := startIndex + pagination.PageSize

	if startIndex >= total {
		return models.NewPaginatedResult([]*models.User{}, total, pagination), nil
	}

	if endIndex > total {
		endIndex = total
	}

	paginatedUsers := matchedUsers[startIndex:endIndex]
	return models.NewPaginatedResult(paginatedUsers, total, pagination), nil
}

// SearchUsersByField searches users by specific field
func (r *MemoryUserRepository) SearchUsersByField(ctx context.Context, field string, value string, pagination *models.PaginationParams) (*models.PaginatedResult, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var matchedUsers []*models.User
	lowerValue := strings.ToLower(value)

	for _, user := range r.users {
		var fieldValue string
		switch field {
		case "name":
			fieldValue = strings.ToLower(user.Name)
		case "email":
			fieldValue = strings.ToLower(user.Email)
		case "department":
			fieldValue = strings.ToLower(user.Department)
		case "position":
			fieldValue = strings.ToLower(user.Position)
		default:
			return nil, models.NewValidationError("invalid search field: " + field)
		}

		if strings.Contains(fieldValue, lowerValue) {
			userCopy := *user
			matchedUsers = append(matchedUsers, &userCopy)
		}
	}

	// Apply pagination
	total := len(matchedUsers)
	startIndex := pagination.Offset
	endIndex := startIndex + pagination.PageSize

	if startIndex >= total {
		return models.NewPaginatedResult([]*models.User{}, total, pagination), nil
	}

	if endIndex > total {
		endIndex = total
	}

	paginatedUsers := matchedUsers[startIndex:endIndex]
	return models.NewPaginatedResult(paginatedUsers, total, pagination), nil
}

// === Helper Methods ===

// sortUsers sorts users based on sort parameters
func (r *MemoryUserRepository) sortUsers(users []*models.User, sortParams *models.SortParams) {
	if sortParams == nil {
		return
	}

	sort.Slice(users, func(i, j int) bool {
		var result bool
		switch sortParams.Field {
		case "id":
			result = users[i].ID < users[j].ID
		case "name":
			result = strings.ToLower(users[i].Name) < strings.ToLower(users[j].Name)
		case "email":
			result = strings.ToLower(users[i].Email) < strings.ToLower(users[j].Email)
		case "created_at":
			result = users[i].CreatedAt.Before(users[j].CreatedAt)
		case "updated_at":
			result = users[i].UpdatedAt.Before(users[j].UpdatedAt)
		default:
			result = users[i].CreatedAt.Before(users[j].CreatedAt)
		}

		if sortParams.Order == "desc" {
			return !result
		}
		return result
	})
}

// getAgeGroup categorizes age into groups
func (r *MemoryUserRepository) getAgeGroup(age int) string {
	switch {
	case age < 18:
		return "under_18"
	case age < 25:
		return "18_24"
	case age < 35:
		return "25_34"
	case age < 45:
		return "35_44"
	case age < 55:
		return "45_54"
	case age < 65:
		return "55_64"
	default:
		return "65_plus"
	}
}

// encodeCursor encodes a user ID as a cursor
func (r *MemoryUserRepository) encodeCursor(userID string) string {
	return base64.StdEncoding.EncodeToString([]byte(userID))
}

// decodeCursor decodes a cursor to get user ID
func (r *MemoryUserRepository) decodeCursor(cursor string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}