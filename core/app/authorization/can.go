package authorization

import (
	"base/core/router"
	"fmt"
	"net/http"
	"strings"
)

// Can creates a middleware function that checks if the user has permission to perform an action on a resource
// Usage: Can('create', 'Post'), Can('update', 'User'), Can('delete', 'Comment')
func Can(action, resourceType string) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			// Get the authorization service from the context
			authorizationServiceValue, exists := c.Get("authorization_service")
			if !exists {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": "authorization service not found",
				})
				return nil
			}

			authorizationService, ok := authorizationServiceValue.(*AuthorizationService)
			if !ok {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": "invalid authorization service",
				})
				return nil
			}

			// Get user Id from context
			userId, err := GetUserIdFromContext(c)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]any{
					"error": err.Error(),
				})
				return nil
			}

			// Convert resource type to lowercase for consistency
			normalizedResourceType := strings.ToLower(resourceType)
			normalizedAction := strings.ToLower(action)

			// Check if the user has permission to perform the action on the resource type
			hasPermission, err := authorizationService.HasPermission(userId, normalizedResourceType, normalizedAction)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": fmt.Sprintf("error checking permission: %v", err),
				})
				return nil
			}

			if !hasPermission {
				c.AbortWithStatusJSON(http.StatusForbidden, map[string]any{
					"error": fmt.Sprintf("permission denied: cannot %s %s", action, resourceType),
				})
				return nil
			}

			return next(c)
		}
	}
}

// CanAccess creates a middleware function that checks if the user has permission to perform an action on a specific resource
// Usage: CanAccess('update', 'Post', 'id'), CanAccess('delete', 'User', 'userId')
func CanAccess(action, resourceType, resourceIdParam string) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			// Get the authorization service from the context
			authorizationServiceValue, exists := c.Get("authorization_service")
			if !exists {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": "authorization service not found",
				})
				return nil
			}

			authorizationService, ok := authorizationServiceValue.(*AuthorizationService)
			if !ok {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": "invalid authorization service",
				})
				return nil
			}

			// Get user Id from context
			userId, err := GetUserIdFromContext(c)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]any{
					"error": err.Error(),
				})
				return nil
			}

			// Get resource Id from URL parameters
			resourceId := c.Param(resourceIdParam)
			if resourceId == "" {
				c.AbortWithStatusJSON(http.StatusBadRequest, map[string]any{
					"error": fmt.Sprintf("missing %s parameter", resourceIdParam),
				})
				return nil
			}

			// Convert resource type to lowercase for consistency
			normalizedAction := strings.ToLower(action)

			// Check if the user has permission to access the specific resource
			hasResourcePermission, err := authorizationService.HasResourcePermission(userId, resourceType, resourceId, normalizedAction)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": fmt.Sprintf("error checking resource permission: %v", err),
				})
				return nil
			}

			if !hasResourcePermission {
				c.AbortWithStatusJSON(http.StatusForbidden, map[string]any{
					"error": fmt.Sprintf("access denied: cannot %s %s with Id %s", action, resourceType, resourceId),
				})
				return nil
			}

			return next(c)
		}
	}
}

// HasRole creates a middleware function that checks if the user has a specific role
// Usage: HasRole('Administrator'), HasRole('Owner'), HasRole('Member')
func HasRole(roleName string) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			// Get the authorization service from the context
			authorizationServiceValue, exists := c.Get("authorization_service")
			if !exists {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": "authorization service not found",
				})
				return nil
			}

			authorizationService, ok := authorizationServiceValue.(*AuthorizationService)
			if !ok {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": "invalid authorization service",
				})
				return nil
			}

			// Get user Id from context
			userId, err := GetUserIdFromContext(c)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]any{
					"error": err.Error(),
				})
				return nil
			}

			// Check if user has the required role by checking role permissions
			hasPermission, err := authorizationService.HasPermission(userId, "role", "read")
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": fmt.Sprintf("error checking role permission: %v", err),
				})
				return nil
			}

			if !hasPermission {
				c.AbortWithStatusJSON(http.StatusForbidden, map[string]any{
					"error": fmt.Sprintf("insufficient permissions: %s role required", roleName),
				})
				return nil
			}

			return next(c)
		}
	}
}

// CanAny creates a middleware function that checks if the user has ANY of the specified permissions
// Usage: CanAny([]string{"create:Post", "update:Post", "delete:Post"})
func CanAny(permissions []string) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			// Get the authorization service from the context
			authorizationServiceValue, exists := c.Get("authorization_service")
			if !exists {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": "authorization service not found",
				})
				return nil
			}

			authorizationService, ok := authorizationServiceValue.(*AuthorizationService)
			if !ok {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": "invalid authorization service",
				})
				return nil
			}

			// Get user Id from context
			userId, err := GetUserIdFromContext(c)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]any{
					"error": err.Error(),
				})
				return nil
			}

			// Check if user has any of the specified permissions
			for _, permission := range permissions {
				parts := strings.Split(permission, ":")
				if len(parts) != 2 {
					continue // Skip invalid permission format
				}

				action := strings.ToLower(strings.TrimSpace(parts[0]))
				resourceType := strings.ToLower(strings.TrimSpace(parts[1]))

				hasPermission, err := authorizationService.HasPermission(userId, resourceType, action)
				if err != nil {
					continue // Skip on error, try next permission
				}

				if hasPermission {
					return next(c) // User has at least one required permission
				}
			}

			// User doesn't have any of the required permissions
			c.AbortWithStatusJSON(http.StatusForbidden, map[string]any{
				"error": "insufficient permissions: none of the required permissions found",
			})
			return nil
		}
	}
}

// CanAll creates a middleware function that checks if the user has ALL of the specified permissions
// Usage: CanAll([]string{"read:Post", "update:Post"})
func CanAll(permissions []string) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			// Get the authorization service from the context
			authorizationServiceValue, exists := c.Get("authorization_service")
			if !exists {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": "authorization service not found",
				})
				return nil
			}

			authorizationService, ok := authorizationServiceValue.(*AuthorizationService)
			if !ok {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": "invalid authorization service",
				})
				return nil
			}

			// Get user Id from context
			userId, err := GetUserIdFromContext(c)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]any{
					"error": err.Error(),
				})
				return nil
			}

			// Check if user has all specified permissions
			for _, permission := range permissions {
				parts := strings.Split(permission, ":")
				if len(parts) != 2 {
					c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
						"error": fmt.Sprintf("invalid permission format: %s", permission),
					})
					return nil
				}

				action := strings.ToLower(strings.TrimSpace(parts[0]))
				resourceType := strings.ToLower(strings.TrimSpace(parts[1]))

				hasPermission, err := authorizationService.HasPermission(userId, resourceType, action)
				if err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
						"error": fmt.Sprintf("error checking permission %s: %v", permission, err),
					})
					return nil
				}

				if !hasPermission {
					c.AbortWithStatusJSON(http.StatusForbidden, map[string]any{
						"error": fmt.Sprintf("missing required permission: %s", permission),
					})
					return nil
				}
			}

			return next(c) // User has all required permissions
		}
	}
}
