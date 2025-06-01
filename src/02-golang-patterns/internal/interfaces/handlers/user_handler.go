package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"golang-patterns/internal/domain/models"
	"golang-patterns/internal/usecases"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	userUseCase *usecases.UserUseCase
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userUseCase *usecases.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

// GetUser handles GET /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// Get user ID from path parameters
	vars := mux.Vars(r)
	userID := vars["id"]
	
	if userID == "" {
		WriteJSONError(w, http.StatusBadRequest, "INVALID_ID", "User ID is required")
		return
	}
	
	// Get user from use case
	user, err := h.userUseCase.GetUser(ctx, userID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			WriteJSONError(w, http.StatusRequestTimeout, "TIMEOUT", "Request timed out")
			return
		}
		
		// Check if it's a not found error
		var notFoundErr models.NotFoundError
		if errors.As(err, &notFoundErr) {
			WriteJSONError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}
		
		WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}
	
	WriteJSONResponse(w, http.StatusOK, user)
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		WriteJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
		return
	}
	
	// Create user through use case
	createdUser, err := h.userUseCase.CreateUser(ctx, &user)
	if err != nil {
		if _, ok := err.(*models.ValidationError); ok {
			WriteJSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		
		WriteJSONError(w, http.StatusInternalServerError, "CREATION_FAILED", "Failed to create user")
		return
	}
	
	WriteJSONResponse(w, http.StatusCreated, createdUser)
}

// GetAllUsers handles GET /users
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	users, err := h.userUseCase.GetAllUsers(ctx)
	if err != nil {
		WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get users")
		return
	}
	
	WriteJSONResponse(w, http.StatusOK, users)
}