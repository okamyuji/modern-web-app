package models

import "time"

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates the user model
func (u *User) Validate() error {
	if u.Name == "" {
		return ErrInvalidUserName
	}
	if u.Email == "" {
		return ErrInvalidUserEmail
	}
	return nil
}