package model

// This file contains additional permission-related types and constants
// that extend the definitions in rbac.go
//
// The core Permission struct is defined in rbac.go:
// - ID (int)
// - Name (string)
// - Resource (string)
// - Action (string)
// - Description (string)
// - CreatedAt (time.Time)

// CreatePermissionRequest represents the request body for creating a permission
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Resource    string `json:"resource" binding:"required,min=2,max=100"`
	Action      string `json:"action" binding:"required,oneof=create read update delete list manage generate export"`
	Description string `json:"description" binding:"max=500"`
}

// PermissionListResponse represents a paginated list of permissions
type PermissionListResponse struct {
	Permissions []Permission `json:"permissions"`
	Total       int          `json:"total"`
	Page        int          `json:"page"`
	PageSize    int          `json:"page_size"`
}

// PermissionDetailResponse represents a permission with detailed information
type PermissionDetailResponse struct {
	Permission Permission `json:"permission"`
	Roles      []Role     `json:"roles_using_this_permission"`
}
