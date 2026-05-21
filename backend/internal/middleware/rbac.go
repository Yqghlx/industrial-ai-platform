package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/service"
)

// PermissionRequired checks if user has the specified permission
func PermissionRequired(rbacSvc *service.RBACService, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by AuthRequired middleware)
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "user not authenticated",
				"code":  "NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		// Type assertion for userID
		userID, ok := userIDInterface.(int)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "invalid user id type",
				"code":  "INVALID_USER_ID",
			})
			c.Abort()
			return
		}

		// Check permission
		hasPermission, err := rbacSvc.CheckPermission(c.Request.Context(), userID, resource, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to check permission",
				"code":  "PERMISSION_CHECK_ERROR",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "permission denied",
				"code":     "INSUFFICIENT_PERMISSIONS",
				"resource": resource,
				"action":   action,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CanManageDevices returns a middleware that checks for device management permission
func CanManageDevices(rbacSvc *service.RBACService) gin.HandlerFunc {
	return PermissionRequired(rbacSvc, "devices", "manage")
}

// CanManageUsers returns a middleware that checks for user management permission
func CanManageUsers(rbacSvc *service.RBACService) gin.HandlerFunc {
	return PermissionRequired(rbacSvc, "users", "manage")
}

// CanManageRules returns a middleware that checks for rule management permission
func CanManageRules(rbacSvc *service.RBACService) gin.HandlerFunc {
	return PermissionRequired(rbacSvc, "rules", "manage")
}

// CanViewReports returns a middleware that checks for report viewing permission
func CanViewReports(rbacSvc *service.RBACService) gin.HandlerFunc {
	return PermissionRequired(rbacSvc, "reports", "read")
}

// CanGenerateReports returns a middleware that checks for report generation permission
func CanGenerateReports(rbacSvc *service.RBACService) gin.HandlerFunc {
	return PermissionRequired(rbacSvc, "reports", "generate")
}

// CanManageAlerts returns a middleware that checks for alert management permission
func CanManageAlerts(rbacSvc *service.RBACService) gin.HandlerFunc {
	return PermissionRequired(rbacSvc, "alerts", "manage")
}

// CanViewAlerts returns a middleware that checks for alert viewing permission
func CanViewAlerts(rbacSvc *service.RBACService) gin.HandlerFunc {
	return PermissionRequired(rbacSvc, "alerts", "read")
}

// CanManageWorkOrders returns a middleware that checks for work order management permission
func CanManageWorkOrders(rbacSvc *service.RBACService) gin.HandlerFunc {
	return PermissionRequired(rbacSvc, "workorders", "manage")
}

// CanViewWorkOrders returns a middleware that checks for work order viewing permission
func CanViewWorkOrders(rbacSvc *service.RBACService) gin.HandlerFunc {
	return PermissionRequired(rbacSvc, "workorders", "read")
}

// CanManageSettings returns a middleware that checks for settings management permission
func CanManageSettings(rbacSvc *service.RBACService) gin.HandlerFunc {
	return PermissionRequired(rbacSvc, "settings", "manage")
}

// CanManageTenants returns a middleware that checks for tenant management permission
func CanManageTenants(rbacSvc *service.RBACService) gin.HandlerFunc {
	return PermissionRequired(rbacSvc, "tenants", "manage")
}

// RoleRequired checks if user has one of the specified roles
func RoleRequired(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "user not authenticated",
				"code":  "NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		role := userRole.(string)
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":          "role not allowed",
			"code":           "INSUFFICIENT_ROLE",
			"required_roles": allowedRoles,
		})
		c.Abort()
	}
}

// AnyPermissionRequired checks if user has any of the specified permissions
func AnyPermissionRequired(rbacSvc *service.RBACService, permissions ...struct{ Resource, Action string }) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "user not authenticated",
				"code":  "NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		// Type assertion for userID
		userID, ok := userIDInterface.(int)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "invalid user id type",
				"code":  "INVALID_USER_ID",
			})
			c.Abort()
			return
		}

		// Check if user has any of the required permissions
		for _, perm := range permissions {
			hasPermission, err := rbacSvc.CheckPermission(c.Request.Context(), userID, perm.Resource, perm.Action)
			if err == nil && hasPermission {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "permission denied",
			"code":  "INSUFFICIENT_PERMISSIONS",
		})
		c.Abort()
	}
}
