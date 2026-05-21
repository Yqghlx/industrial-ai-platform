package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

const (
	// MaxPageSize is the maximum allowed page size
	MaxPageSize = 100
	// DefaultPageSize is the default page size
	DefaultPageSize = 20
	// MaxLimit is the maximum allowed limit for data queries
	MaxLimit = 10000
	// DefaultLimit is the default limit for data queries
	DefaultLimit = 1000
)

// PaginationParams holds validated pagination parameters
type PaginationParams struct {
	Page     int
	PageSize int
}

// ValidatePagination validates and normalizes pagination parameters
// Returns validated page and pageSize values
// - page: minimum 1
// - pageSize: minimum 1, maximum MaxPageSize (100)
func ValidatePagination(page, pageSize int) PaginationParams {
	// Ensure page is at least 1
	if page < 1 {
		page = 1
	}

	// Ensure pageSize is at least 1
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}

	// Ensure pageSize does not exceed maximum
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}

// GetPagination extracts and validates pagination from query parameters
func GetPagination(c *gin.Context) PaginationParams {
	page := 1
	pageSize := DefaultPageSize

	if p := c.Query("page"); p != "" {
		var parsed int
		if _, err := fmt.Sscanf(p, "%d", &parsed); err == nil {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		var parsed int
		if _, err := fmt.Sscanf(ps, "%d", &parsed); err == nil {
			pageSize = parsed
		}
	}

	return ValidatePagination(page, pageSize)
}

// ValidateLimit validates and normalizes a limit parameter
// Returns validated limit value (minimum 1, maximum MaxLimit)
func ValidateLimit(limit int) int {
	// Ensure limit is at least 1
	if limit < 1 {
		return DefaultLimit
	}

	// Ensure limit does not exceed maximum
	if limit > MaxLimit {
		return MaxLimit
	}

	return limit
}

// GetLimit extracts and validates limit from query parameters
func GetLimit(c *gin.Context) int {
	limit := DefaultLimit

	if l := c.Query("limit"); l != "" {
		var parsed int
		if _, err := fmt.Sscanf(l, "%d", &parsed); err == nil {
			limit = parsed
		}
	}

	return ValidateLimit(limit)
}
