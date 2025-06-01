package handlers

import (
	"context"
	"encoding/json"
	"golang-patterns/internal/domain/models"
	"golang-patterns/internal/usecases"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	userUseCase *usecases.UserUseCase
}

// NewUserHandler creates a new user handler
func NewUserHandler(userUseCase *usecases.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

// === Basic CRUD Operations ===

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var req models.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
		return
	}

	user, err := h.userUseCase.CreateUser(ctx, &req)
	if err != nil {
		if _, ok := err.(*models.ValidationError); ok {
			WriteJSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		WriteJSONError(w, http.StatusInternalServerError, "CREATION_FAILED", "Failed to create user")
		return
	}

	WriteJSONResponse(w, http.StatusCreated, user)
}

// GetUser handles GET /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		WriteJSONError(w, http.StatusBadRequest, "INVALID_ID", "User ID is required")
		return
	}

	user, err := h.userUseCase.GetUser(ctx, userID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			WriteJSONError(w, http.StatusRequestTimeout, "TIMEOUT", "Request timed out")
			return
		}

		if _, ok := err.(models.NotFoundError); ok {
			WriteJSONError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}

		WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	WriteJSONResponse(w, http.StatusOK, user)
}

// UpdateUser handles PUT /users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		WriteJSONError(w, http.StatusBadRequest, "INVALID_ID", "User ID is required")
		return
	}

	var req models.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
		return
	}

	user, err := h.userUseCase.UpdateUser(ctx, userID, &req)
	if err != nil {
		if _, ok := err.(*models.ValidationError); ok {
			WriteJSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}

		if _, ok := err.(models.NotFoundError); ok {
			WriteJSONError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}

		WriteJSONError(w, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update user")
		return
	}

	WriteJSONResponse(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		WriteJSONError(w, http.StatusBadRequest, "INVALID_ID", "User ID is required")
		return
	}

	err := h.userUseCase.DeleteUser(ctx, userID)
	if err != nil {
		if _, ok := err.(models.NotFoundError); ok {
			WriteJSONError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}

		WriteJSONError(w, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete user")
		return
	}

	WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// GetAllUsers handles GET /users (legacy method)
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

// === Target Specification - Advanced Query Operations ===

// GetUserByEmail handles GET /users/email/{email}
func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	email := vars["email"]

	user, err := h.userUseCase.GetUserByEmail(ctx, email)
	if err != nil {
		if _, ok := err.(models.NotFoundError); ok {
			WriteJSONError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}

		WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	WriteJSONResponse(w, http.StatusOK, user)
}

// GetUsersByDepartment handles GET /users/department/{department}
func (h *UserHandler) GetUsersByDepartment(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	department := vars["department"]

	users, err := h.userUseCase.GetUsersByDepartment(ctx, department)
	if err != nil {
		WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get users by department")
		return
	}

	WriteJSONResponse(w, http.StatusOK, users)
}

// GetUsersByPosition handles GET /users/position/{position}
func (h *UserHandler) GetUsersByPosition(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	position := vars["position"]

	users, err := h.userUseCase.GetUsersByPosition(ctx, position)
	if err != nil {
		WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get users by position")
		return
	}

	WriteJSONResponse(w, http.StatusOK, users)
}

// GetActiveUsers handles GET /users/active
func (h *UserHandler) GetActiveUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	users, err := h.userUseCase.GetActiveUsers(ctx)
	if err != nil {
		WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get active users")
		return
	}

	WriteJSONResponse(w, http.StatusOK, users)
}

// GetInactiveUsers handles GET /users/inactive
func (h *UserHandler) GetInactiveUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	users, err := h.userUseCase.GetInactiveUsers(ctx)
	if err != nil {
		WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get inactive users")
		return
	}

	WriteJSONResponse(w, http.StatusOK, users)
}

// === Load Display - Pagination and Sorting ===

// GetUsersWithPagination handles GET /users/paginated
func (h *UserHandler) GetUsersWithPagination(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Parse query parameters
	queryParams := r.URL.Query()
	params := make(map[string]string)
	for key, values := range queryParams {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	// Create query parameters from request
	queryParamsObj, err := models.NewQueryParamsFromRequest(params)
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, "INVALID_PARAMS", err.Error())
		return
	}

	result, err := h.userUseCase.GetUsersWithQuery(ctx, queryParamsObj)
	if err != nil {
		WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get users")
		return
	}

	WriteJSONResponse(w, http.StatusOK, result)
}

// SearchUsers handles GET /users/search
func (h *UserHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	query := r.URL.Query().Get("q")
	if query == "" {
		WriteJSONError(w, http.StatusBadRequest, "MISSING_QUERY", "Search query is required")
		return
	}

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 10
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	result, err := h.userUseCase.SearchUsers(ctx, query, page, pageSize)
	if err != nil {
		WriteJSONError(w, http.StatusInternalServerError, "SEARCH_FAILED", "Failed to search users")
		return
	}

	WriteJSONResponse(w, http.StatusOK, result)
}

// === Progressive Loading ===

// GetUsersBatch handles GET /users/batch
func (h *UserHandler) GetUsersBatch(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Parse query parameters
	batchSize := 20
	if batchSizeStr := r.URL.Query().Get("batch_size"); batchSizeStr != "" {
		if bs, err := strconv.Atoi(batchSizeStr); err == nil && bs > 0 && bs <= 50 {
			batchSize = bs
		}
	}

	cursor := r.URL.Query().Get("cursor")
	direction := r.URL.Query().Get("direction")
	if direction == "" {
		direction = "forward"
	}

	result, err := h.userUseCase.GetUsersBatch(ctx, batchSize, cursor, direction)
	if err != nil {
		WriteJSONError(w, http.StatusInternalServerError, "BATCH_FAILED", "Failed to get users batch")
		return
	}

	WriteJSONResponse(w, http.StatusOK, result)
}

// === Statistics and Analytics ===

// GetUserStats handles GET /users/stats
func (h *UserHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	stats, err := h.userUseCase.GetUserStats(ctx)
	if err != nil {
		WriteJSONError(w, http.StatusInternalServerError, "STATS_FAILED", "Failed to get user statistics")
		return
	}

	WriteJSONResponse(w, http.StatusOK, stats)
}

// GetDepartmentStats handles GET /users/stats/departments
func (h *UserHandler) GetDepartmentStats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	stats, err := h.userUseCase.GetDepartmentStats(ctx)
	if err != nil {
		WriteJSONError(w, http.StatusInternalServerError, "STATS_FAILED", "Failed to get department statistics")
		return
	}

	WriteJSONResponse(w, http.StatusOK, stats)
}

// GetRecentSignups handles GET /users/recent-signups
func (h *UserHandler) GetRecentSignups(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	days := 7
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	users, err := h.userUseCase.GetRecentSignups(ctx, days)
	if err != nil {
		WriteJSONError(w, http.StatusInternalServerError, "RECENT_SIGNUPS_FAILED", "Failed to get recent signups")
		return
	}

	WriteJSONResponse(w, http.StatusOK, users)
}

// === Form Processing - Bulk Operations ===

// CreateUsersInBulk handles POST /users/bulk
func (h *UserHandler) CreateUsersInBulk(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var requests []*models.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		WriteJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
		return
	}

	users, err := h.userUseCase.CreateUsersInBulk(ctx, requests)
	if err != nil {
		if _, ok := err.(*models.ValidationError); ok {
			WriteJSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		WriteJSONError(w, http.StatusInternalServerError, "BULK_CREATE_FAILED", "Failed to create users in bulk")
		return
	}

	WriteJSONResponse(w, http.StatusCreated, users)
}

// UpdateUsersInBulk handles PUT /users/bulk
func (h *UserHandler) UpdateUsersInBulk(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var updates map[string]*models.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		WriteJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
		return
	}

	users, err := h.userUseCase.UpdateUsersInBulk(ctx, updates)
	if err != nil {
		if _, ok := err.(*models.ValidationError); ok {
			WriteJSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		WriteJSONError(w, http.StatusInternalServerError, "BULK_UPDATE_FAILED", "Failed to update users in bulk")
		return
	}

	WriteJSONResponse(w, http.StatusOK, users)
}

// DeleteUsersInBulk handles DELETE /users/bulk
func (h *UserHandler) DeleteUsersInBulk(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var request struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
		return
	}

	err := h.userUseCase.DeleteUsersInBulk(ctx, request.IDs)
	if err != nil {
		if _, ok := err.(*models.ValidationError); ok {
			WriteJSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		WriteJSONError(w, http.StatusInternalServerError, "BULK_DELETE_FAILED", "Failed to delete users in bulk")
		return
	}

	WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Users deleted successfully"})
}

// === Progressive Enhancement Features ===

// ActivateUser handles POST /users/{id}/activate
func (h *UserHandler) ActivateUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	userID := vars["id"]

	user, err := h.userUseCase.ActivateUser(ctx, userID)
	if err != nil {
		if _, ok := err.(models.NotFoundError); ok {
			WriteJSONError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}

		WriteJSONError(w, http.StatusInternalServerError, "ACTIVATION_FAILED", "Failed to activate user")
		return
	}

	WriteJSONResponse(w, http.StatusOK, user)
}

// DeactivateUser handles POST /users/{id}/deactivate
func (h *UserHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	userID := vars["id"]

	user, err := h.userUseCase.DeactivateUser(ctx, userID)
	if err != nil {
		if _, ok := err.(models.NotFoundError); ok {
			WriteJSONError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}

		WriteJSONError(w, http.StatusInternalServerError, "DEACTIVATION_FAILED", "Failed to deactivate user")
		return
	}

	WriteJSONResponse(w, http.StatusOK, user)
}

// UpdateLastLogin handles POST /users/{id}/login
func (h *UserHandler) UpdateLastLogin(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	userID := vars["id"]

	user, err := h.userUseCase.UpdateLastLogin(ctx, userID)
	if err != nil {
		if _, ok := err.(models.NotFoundError); ok {
			WriteJSONError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}

		WriteJSONError(w, http.StatusInternalServerError, "LOGIN_UPDATE_FAILED", "Failed to update last login")
		return
	}

	WriteJSONResponse(w, http.StatusOK, user)
}

// GetUserSummary handles GET /users/{id}/summary
func (h *UserHandler) GetUserSummary(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	userID := vars["id"]

	summary, err := h.userUseCase.GetUserSummary(ctx, userID)
	if err != nil {
		if _, ok := err.(models.NotFoundError); ok {
			WriteJSONError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}

		WriteJSONError(w, http.StatusInternalServerError, "SUMMARY_FAILED", "Failed to get user summary")
		return
	}

	WriteJSONResponse(w, http.StatusOK, summary)
}