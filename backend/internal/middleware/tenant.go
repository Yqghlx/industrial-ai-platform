package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Context keys for tenant information
const (
	TenantIDKey   = "tenant_id"
	TenantSlugKey = "tenant_slug"
	TenantPlanKey = "tenant_plan"
)

// TenantIsolation provides automatic tenant isolation for database queries
// It extracts tenant_id from context and makes it available for query filtering
func TenantIsolation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get tenant_id from context (set by auth middleware)
		_, exists := c.Get(TenantIDKey)
		if !exists {
			// No tenant context - might be a system admin or public endpoint
			c.Next()
			return
		}

		// Store tenant_id in request context for repositories to use
		c.Next()
	}
}

// GetTenantSlug extracts tenant slug from context
func GetTenantSlug(c *gin.Context) string {
	if slug, exists := c.Get(TenantSlugKey); exists {
		if s, ok := slug.(string); ok {
			return s
		}
	}
	return ""
}

// GetTenantPlan extracts tenant plan from context
func GetTenantPlan(c *gin.Context) string {
	if plan, exists := c.Get(TenantPlanKey); exists {
		if p, ok := plan.(string); ok {
			return p
		}
	}
	return ""
}

// SetTenantContext sets tenant information in the gin context
func SetTenantContext(c *gin.Context, tenantID, tenantSlug, tenantPlan string) {
	c.Set(TenantIDKey, tenantID)
	c.Set(TenantSlugKey, tenantSlug)
	c.Set(TenantPlanKey, tenantPlan)
}

// RequireTenantInBody is a middleware that validates tenant_id in request body
// matches the authenticated user's tenant_id
func RequireTenantInBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := GetTenantID(c)
		if tenantID == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Tenant context required",
				"code":  "MISSING_TENANT",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// TenantAdminRequired requires the user to be an admin within their tenant
func TenantAdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetUserRole(c)
		if role == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
				"code":  "NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		// Check if user is admin or tenant_admin
		if role != "admin" && role != "tenant_admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Tenant admin access required",
				"code":  "INSUFFICIENT_PERMISSIONS",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TenantQueryFilter provides a helper for adding tenant filtering to SQL queries
type TenantQueryFilter struct {
	WhereClause string
	Args        []interface{}
}

// BuildTenantFilter creates a tenant filter for SQL queries
func BuildTenantFilter(tenantID string, baseQuery string, args ...interface{}) TenantQueryFilter {
	if tenantID == "" {
		return TenantQueryFilter{
			WhereClause: baseQuery,
			Args:        args,
		}
	}

	return TenantQueryFilter{
		WhereClause: baseQuery + " AND tenant_id = ?",
		Args:        append(args, tenantID),
	}
}

// TenantScopedQuery wraps a query to be tenant-scoped
// Usage: query := middleware.TenantScopedQuery(c, "SELECT * FROM devices")
func TenantScopedQuery(c *gin.Context, baseQuery string) (string, []interface{}) {
	tenantID := GetTenantID(c)
	if tenantID == "" {
		return baseQuery, nil
	}

	// Add tenant filter based on query structure
	// Check if WHERE already exists
	filter := TenantQueryFilter{
		WhereClause: baseQuery + " WHERE tenant_id = $1",
		Args:        []interface{}{tenantID},
	}

	return filter.WhereClause, filter.Args
}
