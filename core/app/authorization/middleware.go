package authorization

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var (
	ErrMissingUserId        = errors.New("missing user ID in context")
	ErrMissingOrganization  = errors.New("missing organization ID in context or headers")
	ErrMissingResourceId    = errors.New("missing resource ID in request")
	ErrPermissionDenied     = errors.New("permission denied")
	ErrResourceAccessDenied = errors.New("resource access denied")
)

// GetUserIdFromContext extracts the user ID from the Gin context
func GetUserIdFromContext(c *gin.Context) (uint64, error) {
	userIdValue, exists := c.Get("user_id")
	if !exists {
		return 0, ErrMissingUserId
	}

	switch userId := userIdValue.(type) {
	case uint64:
		return userId, nil
	case uint:
		return uint64(userId), nil
	case int:
		return uint64(userId), nil
	case string:
		userIdInt, err := strconv.ParseUint(userId, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid user ID format: %w", err)
		}
		return userIdInt, nil
	default:
		return 0, fmt.Errorf("unsupported user ID type: %T", userIdValue)
	}
}

// GetOrganizationIdFromContext extracts the organization ID from the Gin context or headers
func GetOrganizationIdFromContext(c *gin.Context) (uint64, error) {
	// First try to get from context
	orgIdValue, exists := c.Get("organization_id")
	if exists {
		switch orgId := orgIdValue.(type) {
		case uint64:
			return orgId, nil
		case uint:
			return uint64(orgId), nil
		case int:
			return uint64(orgId), nil
		case string:
			orgIdInt, err := strconv.ParseUint(orgId, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid organization ID format: %w", err)
			}
			return orgIdInt, nil
		}
	}

	// Try to get from header
	orgIdHeader := c.GetHeader("base_header_orgid")
	if orgIdHeader != "" {
		orgIdInt, err := strconv.ParseUint(orgIdHeader, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid organization ID in header: %w", err)
		}
		return orgIdInt, nil
	}

	return 0, ErrMissingOrganization
}

// AuthMiddleware creates a middleware function that checks if the user has permission to access a resource
func AuthMiddleware(resourceType string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the authorization service from the Gin context
		authzServiceValue, exists := c.Get("authz_service")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "authorization service not found",
			})
			return
		}

		authzService, ok := authzServiceValue.(*AuthorizationService)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "invalid authorization service",
			})
			return
		}

		// Get user ID from context
		userId, err := GetUserIdFromContext(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Get organization ID
		orgId, err := GetOrganizationIdFromContext(c)
		if err != nil {
			// For global endpoints that don't require an organization ID
			if action == "read" && (resourceType == "auth" || resourceType == "user") {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Check if the user has permission to perform the action on the resource type
		hasPermission, err := authzService.HasPermission(userId, orgId, resourceType, action)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("error checking permission: %v", err),
			})
			return
		}

		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "permission denied",
			})
			return
		}

		// All checks passed
		c.Next()
	}
}

// ResourceAccessMiddleware checks if the user has access to a specific resource
func ResourceAccessMiddleware(resourceType string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the resource ID from the URL parameter
		resourceId := c.Param("id")
		if resourceId == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": ErrMissingResourceId.Error(),
			})
			return
		}

		// Get the authorization service
		authzServiceValue, exists := c.Get("authz_service")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "authorization service not found",
			})
			return
		}

		authzService, ok := authzServiceValue.(*AuthorizationService)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "invalid authorization service",
			})
			return
		}

		// Get user ID
		userId, err := GetUserIdFromContext(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Get organization ID
		orgId, err := GetOrganizationIdFromContext(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Check resource-specific permission
		hasPermission, err := authzService.HasResourcePermission(userId, orgId, resourceType, resourceId, action)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("error checking resource permission: %v", err),
			})
			return
		}

		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": ErrResourceAccessDenied.Error(),
			})
			return
		}

		// Store the allowed resource IDs in context for later filtering
		resourceIds, defaultScope, err := authzService.GetAccessibleResources(userId, orgId, resourceType)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("error getting accessible resources: %v", err),
			})
			return
		}

		// Store the permission scope in the context
		c.Set("permission_scope", defaultScope)
		c.Set("accessible_resources", resourceIds)

		// All checks passed
		c.Next()
	}
}

// ScopeFilterMiddleware applies access scope filtering to queries
func ScopeFilterMiddleware(resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First execute the request handler
		c.Next()

		// After handler execution, check if we need to filter results
		permissionScope, exists := c.Get("permission_scope")
		if !exists {
			return // No filtering needed
		}

		scope, ok := permissionScope.(string)
		if !ok || scope == AccessScopeAll {
			return // All access or invalid scope format, no filtering needed
		}

		// Check if we need to filter based on the response
		// This is a simple example and would need to be adapted to your response structure
		var response interface{}
		if v, exists := c.Get("response_data"); exists {
			response = v
		} else {
			return // No response data to filter
		}

		// Apply filtering based on scope and response type
		// This would depend on your specific application structure
		filteredResponse := applyFilterByScope(response, resourceType, scope, c)
		c.Set("response_data", filteredResponse)
	}
}

// applyFilterByScope filters response data based on permission scope
// This is a placeholder implementation that would need to be customized
func applyFilterByScope(data interface{}, resourceType, scope string, c *gin.Context) interface{} {
	// Example implementation - would need to be customized for your data structures
	userId, _ := GetUserIdFromContext(c)

	switch scope {
	case AccessScopeOwn:
		// Filter for only resources owned by the current user
		return filterOwnResources(data, resourceType, userId)

	case AccessScopeTeam:
		// Filter for resources within the user's team/department
		department := getUserDepartment(c)
		return filterTeamResources(data, resourceType, department)

	default:
		// Default to returning everything
		return data
	}
}

func filterOwnResources(data interface{}, resourceType string, userId uint64) interface{} {
	// Handle different resource types
	switch resourceType {
	case "project":
		if projects, ok := data.([]map[string]interface{}); ok {
			filtered := make([]map[string]interface{}, 0)
			for _, project := range projects {
				if ownerID, exists := project["owner_id"].(uint64); exists && ownerID == userId {
					filtered = append(filtered, project)
				}
			}
			return filtered
		}
	case "employee":
		if employees, ok := data.([]map[string]interface{}); ok {
			filtered := make([]map[string]interface{}, 0)
			for _, employee := range employees {
				if empID, exists := employee["user_id"].(uint64); exists && empID == userId {
					filtered = append(filtered, employee)
				}
			}
			return filtered
		}
	case "absence":
		if absences, ok := data.([]map[string]interface{}); ok {
			filtered := make([]map[string]interface{}, 0)
			for _, absence := range absences {
				if absenceUserID, exists := absence["user_id"].(uint64); exists && absenceUserID == userId {
					filtered = append(filtered, absence)
				}
			}
			return filtered
		}
	}
	return data
}

func filterTeamResources(data interface{}, resourceType string, department string) interface{} {
	// Handle different resource types
	switch resourceType {
	case "project":
		if projects, ok := data.([]map[string]interface{}); ok {
			filtered := make([]map[string]interface{}, 0)
			for _, project := range projects {
				if projectDept, exists := project["department"].(string); exists && projectDept == department {
					filtered = append(filtered, project)
				}
			}
			return filtered
		}
	case "employee":
		if employees, ok := data.([]map[string]interface{}); ok {
			filtered := make([]map[string]interface{}, 0)
			for _, employee := range employees {
				if empDept, exists := employee["department"].(string); exists && empDept == department {
					filtered = append(filtered, employee)
				}
			}
			return filtered
		}
	case "absence":
		if absences, ok := data.([]map[string]interface{}); ok {
			filtered := make([]map[string]interface{}, 0)
			for _, absence := range absences {
				if absenceDept, exists := absence["department"].(string); exists && absenceDept == department {
					filtered = append(filtered, absence)
				}
			}
			return filtered
		}
	}
	return data
}

func getUserDepartment(c *gin.Context) string {
	// Get department from context if available
	if dept, exists := c.Get("user_department"); exists {
		if department, ok := dept.(string); ok {
			return department
		}
	}
	return ""
}
