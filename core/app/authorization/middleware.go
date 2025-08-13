package authorization

import (
	"base/core/router"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

var (
	ErrMissingUserId        = errors.New("missing user ID in context")
	ErrMissingOrganization  = errors.New("missing organization ID in context or headers")
	ErrMissingResourceId    = errors.New("missing resource ID in request")
	ErrPermissionDenied     = errors.New("permission denied")
	ErrResourceAccessDenied = errors.New("resource access denied")
)

// GetUserIdFromContext extracts the user ID from the context
func GetUserIdFromContext(c *router.Context) (uint64, error) {
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

// GetOrganizationIdFromContext extracts the organization ID from the context or headers
func GetOrganizationIdFromContext(c *router.Context) (uint64, error) {
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
func AuthMiddleware(resourceType string, action string) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			// Get the authorization service from the context
			authzServiceValue, exists := c.Get("authz_service")
			if !exists {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": "authorization service not found",
				})
				return nil
			}

			authzService, ok := authzServiceValue.(*AuthorizationService)
			if !ok {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": "invalid authorization service",
				})
				return nil
			}

			// Get user ID from context
			userId, err := GetUserIdFromContext(c)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]any{
					"error": err.Error(),
				})
				return nil
			}

			// Get organization ID
			orgId, err := GetOrganizationIdFromContext(c)
			if err != nil {
				// For global endpoints that don't require an organization ID
				if action == "read" && (resourceType == "auth" || resourceType == "user") {
					return next(c)
				}
				c.AbortWithStatusJSON(http.StatusBadRequest, map[string]any{
					"error": err.Error(),
				})
				return nil
			}

			// Check if the user has permission to perform the action on the resource type
			hasPermission, err := authzService.HasPermission(userId, orgId, resourceType, action)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": fmt.Sprintf("error checking permission: %v", err),
				})
				return nil
			}

			if !hasPermission {
				c.AbortWithStatusJSON(http.StatusForbidden, map[string]any{
					"error": ErrPermissionDenied.Error(),
				})
				return nil
			}

			return next(c)
		}
	}
}

// ResourceAuthMiddleware creates a middleware function that checks if the user has permission to access a specific resource
func ResourceAuthMiddleware(resourceType string, action string, resourceIdParam string) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			// Get the authorization service from the context
			authzServiceValue, exists := c.Get("authz_service")
			if !exists {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": "authorization service not found",
				})
				return nil
			}

			authzService, ok := authzServiceValue.(*AuthorizationService)
			if !ok {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": "invalid authorization service",
				})
				return nil
			}

			// Get user ID from context
			userId, err := GetUserIdFromContext(c)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]any{
					"error": err.Error(),
				})
				return nil
			}

			// Get organization ID
			orgId, err := GetOrganizationIdFromContext(c)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, map[string]any{
					"error": err.Error(),
				})
				return nil
			}

			// Get resource ID from URL parameters
			resourceId := c.Param(resourceIdParam)
			if resourceId == "" {
				c.AbortWithStatusJSON(http.StatusBadRequest, map[string]any{
					"error": ErrMissingResourceId.Error(),
				})
				return nil
			}

			// Check if the user has permission to access the specific resource
			hasPermission, err := authzService.HasResourcePermission(userId, orgId, resourceType, resourceId, action)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": fmt.Sprintf("error checking resource permission: %v", err),
				})
				return nil
			}

			if !hasPermission {
				c.AbortWithStatusJSON(http.StatusForbidden, map[string]any{
					"error": ErrResourceAccessDenied.Error(),
				})
				return nil
			}

			return next(c)
		}
	}
}

// RequireRole creates a middleware function that checks if the user has a specific role
func RequireRole(roleName string) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			// Get the authorization service from the context
			authzServiceValue, exists := c.Get("authz_service")
			if !exists {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": "authorization service not found",
				})
				return nil
			}

			authzService, ok := authzServiceValue.(*AuthorizationService)
			if !ok {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": "invalid authorization service",
				})
				return nil
			}

			// Get user ID from context
			userId, err := GetUserIdFromContext(c)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]any{
					"error": err.Error(),
				})
				return nil
			}

			// Get organization ID
			orgId, err := GetOrganizationIdFromContext(c)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, map[string]any{
					"error": err.Error(),
				})
				return nil
			}

			// TODO: Implement HasRole method in AuthorizationService or use alternative approach
			// For now, just check if user has general permission
			hasPermission, err := authzService.HasPermission(userId, orgId, "role", "read")
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": fmt.Sprintf("error checking role permission: %v", err),
				})
				return nil
			}

			if !hasPermission {
				c.AbortWithStatusJSON(http.StatusForbidden, map[string]any{
					"error": "insufficient role permissions",
				})
				return nil
			}

			return next(c)
		}
	}
}
