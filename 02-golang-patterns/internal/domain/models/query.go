package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// UserFilter represents filtering criteria for users
type UserFilter struct {
	Name      string    `json:"name,omitempty"`
	Email     string    `json:"email,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	IsActive  *bool     `json:"is_active,omitempty"`
}

// Validate validates the user filter
func (f *UserFilter) Validate() error {
	// Email format validation if provided
	if f.Email != "" && !strings.Contains(f.Email, "@") {
		return NewValidationError("invalid email format")
	}
	return nil
}

// Matches checks if a user matches the filter criteria
func (f *UserFilter) Matches(user *User) bool {
	if f.Name != "" && !strings.Contains(strings.ToLower(user.Name), strings.ToLower(f.Name)) {
		return false
	}
	if f.Email != "" && !strings.Contains(strings.ToLower(user.Email), strings.ToLower(f.Email)) {
		return false
	}
	if !f.CreatedAt.IsZero() && user.CreatedAt.Before(f.CreatedAt) {
		return false
	}
	return true
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Offset   int `json:"offset"`
}

// NewPaginationParams creates pagination parameters with validation
func NewPaginationParams(page, pageSize int) *PaginationParams {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	
	return &PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Offset:   (page - 1) * pageSize,
	}
}

// Validate validates pagination parameters
func (p *PaginationParams) Validate() error {
	if p.Page < 1 {
		return NewValidationError("page must be greater than 0")
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		return NewValidationError("page_size must be between 1 and 100")
	}
	return nil
}

// SortParams represents sorting parameters
type SortParams struct {
	Field string `json:"field"`
	Order string `json:"order"` // "asc" or "desc"
}

// NewSortParams creates sort parameters with validation
func NewSortParams(field, order string) *SortParams {
	if field == "" {
		field = "created_at"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	
	return &SortParams{
		Field: field,
		Order: order,
	}
}

// Validate validates sort parameters
func (s *SortParams) Validate() error {
	validFields := []string{"id", "name", "email", "created_at", "updated_at"}
	validField := false
	for _, field := range validFields {
		if s.Field == field {
			validField = true
			break
		}
	}
	
	if !validField {
		return NewValidationError(fmt.Sprintf("invalid sort field: %s. Valid fields: %s", s.Field, strings.Join(validFields, ", ")))
	}
	
	if s.Order != "asc" && s.Order != "desc" {
		return NewValidationError("sort order must be 'asc' or 'desc'")
	}
	
	return nil
}

// QueryParams combines all query parameters
type QueryParams struct {
	Filter     *UserFilter       `json:"filter,omitempty"`
	Pagination *PaginationParams `json:"pagination,omitempty"`
	Sort       *SortParams       `json:"sort,omitempty"`
}

// NewQueryParamsFromRequest creates QueryParams from HTTP request parameters
func NewQueryParamsFromRequest(params map[string]string) (*QueryParams, error) {
	qp := &QueryParams{
		Filter:     &UserFilter{},
		Pagination: &PaginationParams{},
		Sort:       &SortParams{},
	}
	
	// Parse filter parameters
	if name := params["name"]; name != "" {
		qp.Filter.Name = name
	}
	if email := params["email"]; email != "" {
		qp.Filter.Email = email
	}
	
	// Parse pagination parameters
	page := 1
	if pageStr := params["page"]; pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	pageSize := 10
	if pageSizeStr := params["page_size"]; pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}
	
	qp.Pagination = NewPaginationParams(page, pageSize)
	
	// Parse sort parameters
	field := params["sort_field"]
	order := params["sort_order"]
	qp.Sort = NewSortParams(field, order)
	
	// Validate all parameters
	if err := qp.Filter.Validate(); err != nil {
		return nil, err
	}
	if err := qp.Pagination.Validate(); err != nil {
		return nil, err
	}
	if err := qp.Sort.Validate(); err != nil {
		return nil, err
	}
	
	return qp, nil
}

// PaginatedResult represents a paginated result set
type PaginatedResult struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// NewPaginatedResult creates a new paginated result
func NewPaginatedResult(data interface{}, total int, pagination *PaginationParams) *PaginatedResult {
	totalPages := (total + pagination.PageSize - 1) / pagination.PageSize
	if totalPages < 1 {
		totalPages = 1
	}
	
	return &PaginatedResult{
		Data:       data,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
	}
}

// ProgressiveLoadParams represents parameters for progressive loading
type ProgressiveLoadParams struct {
	BatchSize int    `json:"batch_size"`
	Cursor    string `json:"cursor,omitempty"`
	Direction string `json:"direction"` // "forward" or "backward"
}

// NewProgressiveLoadParams creates progressive load parameters with validation
func NewProgressiveLoadParams(batchSize int, cursor, direction string) *ProgressiveLoadParams {
	if batchSize < 1 || batchSize > 50 {
		batchSize = 20
	}
	if direction != "forward" && direction != "backward" {
		direction = "forward"
	}
	
	return &ProgressiveLoadParams{
		BatchSize: batchSize,
		Cursor:    cursor,
		Direction: direction,
	}
}

// Validate validates progressive load parameters
func (p *ProgressiveLoadParams) Validate() error {
	if p.BatchSize < 1 || p.BatchSize > 50 {
		return NewValidationError("batch_size must be between 1 and 50")
	}
	if p.Direction != "forward" && p.Direction != "backward" {
		return NewValidationError("direction must be 'forward' or 'backward'")
	}
	return nil
}

// ProgressiveResult represents a progressive loading result
type ProgressiveResult struct {
	Data      interface{} `json:"data"`
	NextCursor string     `json:"next_cursor,omitempty"`
	PrevCursor string     `json:"prev_cursor,omitempty"`
	HasMore   bool        `json:"has_more"`
}