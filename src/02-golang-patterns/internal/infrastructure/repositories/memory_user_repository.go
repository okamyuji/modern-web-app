package repositories

import (
	"context"
	"golang-patterns/internal/domain/models"
	"sync"
	"time"
)

// MemoryUserRepository implements UserRepository using in-memory storage
type MemoryUserRepository struct {
	users map[string]*models.User
	mutex sync.RWMutex
}

// NewMemoryUserRepository creates a new MemoryUserRepository
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users: make(map[string]*models.User),
		mutex: sync.RWMutex{},
	}
}

// GetByID gets a user by ID
func (r *MemoryUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	user, exists := r.users[id]
	if !exists {
		return nil, models.NotFoundError{Resource: "user", ID: id}
	}
	
	return user, nil
}

// Save saves a user
func (r *MemoryUserRepository) Save(ctx context.Context, user *models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if user.ID == "" {
		user.ID = generateID()
		user.CreatedAt = time.Now()
	}
	user.UpdatedAt = time.Now()
	
	r.users[user.ID] = user
	return nil
}

// GetAll gets all users
func (r *MemoryUserRepository) GetAll(ctx context.Context) ([]*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	users := make([]*models.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	
	return users, nil
}

// Delete deletes a user by ID
func (r *MemoryUserRepository) Delete(ctx context.Context, id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.users[id]; !exists {
		return models.NotFoundError{Resource: "user", ID: id}
	}
	
	delete(r.users, id)
	return nil
}

// generateID generates a simple ID (in real app, use UUID)
func generateID() string {
	return "user_" + time.Now().Format("20060102150405")
}