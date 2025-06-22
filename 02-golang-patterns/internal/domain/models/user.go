package models

import (
	"regexp"
	"strings"
	"time"
)

// User represents a user in the system
type User struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Age         int        `json:"age,omitempty"`
	Department  string     `json:"department,omitempty"`
	Position    string     `json:"position,omitempty"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// UserCreateRequest represents a request to create a user
type UserCreateRequest struct {
	Name       string `json:"name" validate:"required,min=2,max=100"`
	Email      string `json:"email" validate:"required,email"`
	Age        int    `json:"age,omitempty" validate:"min=0,max=150"`
	Department string `json:"department,omitempty" validate:"max=100"`
	Position   string `json:"position,omitempty" validate:"max=100"`
}

// UserUpdateRequest represents a request to update a user
type UserUpdateRequest struct {
	Name       *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Email      *string `json:"email,omitempty" validate:"omitempty,email"`
	Age        *int    `json:"age,omitempty" validate:"omitempty,min=0,max=150"`
	Department *string `json:"department,omitempty" validate:"omitempty,max=100"`
	Position   *string `json:"position,omitempty" validate:"omitempty,max=100"`
	IsActive   *bool   `json:"is_active,omitempty"`
}

// Validate validates the user model with enhanced validation
func (u *User) Validate() error {
	// Name validation
	if strings.TrimSpace(u.Name) == "" {
		return NewValidationError("name is required")
	}
	if len(u.Name) < 2 || len(u.Name) > 100 {
		return NewValidationError("name must be between 2 and 100 characters")
	}

	// Email validation
	if strings.TrimSpace(u.Email) == "" {
		return NewValidationError("email is required")
	}
	if !isValidEmail(u.Email) {
		return NewValidationError("invalid email format")
	}

	// Age validation
	if u.Age < 0 || u.Age > 150 {
		return NewValidationError("age must be between 0 and 150")
	}

	// Department validation
	if len(u.Department) > 100 {
		return NewValidationError("department must be less than 100 characters")
	}

	// Position validation
	if len(u.Position) > 100 {
		return NewValidationError("position must be less than 100 characters")
	}

	return nil
}

// ValidateCreateRequest validates the user create request
func (req *UserCreateRequest) Validate() error {
	// Name validation
	if strings.TrimSpace(req.Name) == "" {
		return NewValidationError("name is required")
	}
	if len(req.Name) < 2 || len(req.Name) > 100 {
		return NewValidationError("name must be between 2 and 100 characters")
	}

	// Email validation
	if strings.TrimSpace(req.Email) == "" {
		return NewValidationError("email is required")
	}
	if !isValidEmail(req.Email) {
		return NewValidationError("invalid email format")
	}

	// Age validation
	if req.Age < 0 || req.Age > 150 {
		return NewValidationError("age must be between 0 and 150")
	}

	// Department validation
	if len(req.Department) > 100 {
		return NewValidationError("department must be less than 100 characters")
	}

	// Position validation
	if len(req.Position) > 100 {
		return NewValidationError("position must be less than 100 characters")
	}

	return nil
}

// ValidateUpdateRequest validates the user update request
func (req *UserUpdateRequest) Validate() error {
	// Name validation (if provided)
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return NewValidationError("name cannot be empty")
		}
		if len(*req.Name) < 2 || len(*req.Name) > 100 {
			return NewValidationError("name must be between 2 and 100 characters")
		}
	}

	// Email validation (if provided)
	if req.Email != nil {
		if strings.TrimSpace(*req.Email) == "" {
			return NewValidationError("email cannot be empty")
		}
		if !isValidEmail(*req.Email) {
			return NewValidationError("invalid email format")
		}
	}

	// Age validation (if provided)
	if req.Age != nil {
		if *req.Age < 0 || *req.Age > 150 {
			return NewValidationError("age must be between 0 and 150")
		}
	}

	// Department validation (if provided)
	if req.Department != nil && len(*req.Department) > 100 {
		return NewValidationError("department must be less than 100 characters")
	}

	// Position validation (if provided)
	if req.Position != nil && len(*req.Position) > 100 {
		return NewValidationError("position must be less than 100 characters")
	}

	return nil
}

// ToUser converts UserCreateRequest to User
func (req *UserCreateRequest) ToUser() *User {
	now := time.Now()
	return &User{
		Name:       req.Name,
		Email:      req.Email,
		Age:        req.Age,
		Department: req.Department,
		Position:   req.Position,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// ApplyUpdate applies UserUpdateRequest to existing User
func (u *User) ApplyUpdate(req *UserUpdateRequest) {
	if req.Name != nil {
		u.Name = *req.Name
	}
	if req.Email != nil {
		u.Email = *req.Email
	}
	if req.Age != nil {
		u.Age = *req.Age
	}
	if req.Department != nil {
		u.Department = *req.Department
	}
	if req.Position != nil {
		u.Position = *req.Position
	}
	if req.IsActive != nil {
		u.IsActive = *req.IsActive
	}
	u.UpdatedAt = time.Now()
}

// isValidEmail validates email format using regex
func isValidEmail(email string) bool {
	// Simple email validation regex
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(email)
}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers       int            `json:"total_users"`
	ActiveUsers      int            `json:"active_users"`
	InactiveUsers    int            `json:"inactive_users"`
	DepartmentStats  map[string]int `json:"department_stats"`
	PositionStats    map[string]int `json:"position_stats"`
	AgeDistribution  map[string]int `json:"age_distribution"`
	RecentLogins     int            `json:"recent_logins"`
	LastWeekSignups  int            `json:"last_week_signups"`
}

// UserBatch represents a batch of users for progressive loading
type UserBatch struct {
	Users      []*User `json:"users"`
	NextCursor string  `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
	Total      int     `json:"total"`
}