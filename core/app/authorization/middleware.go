package authorization

import (
	"base/core/router"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
)

var (
	ErrMissingUserId        = errors.New("missing user Id in context")
	ErrMissingOrganization  = errors.New("missing organization Id in context or headers")
	ErrMissingResourceId    = errors.New("missing resource Id in request")
	ErrPermissionDenied     = errors.New("permission denied")
	ErrResourceAccessDenied = errors.New("resource access denied")
)

// OrganizationHeader is the request header that carries the organization Id.
// It matches the name the CORS middleware allows.
const OrganizationHeader = "Base-Orgid"

// defaultService is the service the checks consult when the request context
// carries none. NewAuthorizationModule sets it.
var defaultService atomic.Pointer[AuthorizationService]

// serviceFor returns the authorization service for a request: the one stored in
// the context under "authorization_service", or else the module's. When there is
// neither it answers 500 itself and returns false.
func serviceFor(c *router.Context) (*AuthorizationService, bool) {
	if v, exists := c.Get("authorization_service"); exists {
		if s, ok := v.(*AuthorizationService); ok && s != nil {
			return s, true
		}
	}
	if s := defaultService.Load(); s != nil {
		return s, true
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
		"error": "authorization service not found: core/app/authorization must be registered before its checks run",
	})
	return nil, false
}

// GetUserIdFromContext extracts the user Id from the context
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
			return 0, fmt.Errorf("invalid user Id format: %w", err)
		}
		return userIdInt, nil
	default:
		return 0, fmt.Errorf("unsupported user Id type: %T", userIdValue)
	}
}

// GetOrganizationIdFromContext extracts the organization Id from the context or headers
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
				return 0, fmt.Errorf("invalid organization Id format: %w", err)
			}
			return orgIdInt, nil
		}
	}

	// Try to get from header. Base-Orgid is the name CORS allows; the old
	// base_header_orgid is still read for clients that send it.
	orgIdHeader := c.GetHeader(OrganizationHeader)
	if orgIdHeader == "" {
		orgIdHeader = c.GetHeader("base_header_orgid")
	}
	if orgIdHeader != "" {
		orgIdInt, err := strconv.ParseUint(orgIdHeader, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid organization Id in header: %w", err)
		}
		return orgIdInt, nil
	}

	return 0, ErrMissingOrganization
}

// AuthMiddleware creates a middleware function that checks if the user has permission to access a resource
func AuthMiddleware(resourceType string, action string) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			authorizationService, ok := serviceFor(c)
			if !ok {
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

			// Check if the user has permission to perform the action on the resource type
			hasPermission, err := authorizationService.HasPermission(userId, resourceType, action)
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

			// Get resource Id from URL parameters
			resourceId := c.Param(resourceIdParam)
			if resourceId == "" {
				c.AbortWithStatusJSON(http.StatusBadRequest, map[string]any{
					"error": ErrMissingResourceId.Error(),
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
			authorizationService, ok := serviceFor(c)
			if !ok {
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

			// Check if the user has the required role
			hasRole, err := authorizationService.HasRole(userId, roleName)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
					"error": fmt.Sprintf("error checking role: %v", err),
				})
				return nil
			}

			if !hasRole {
				c.AbortWithStatusJSON(http.StatusForbidden, map[string]any{
					"error": "insufficient role permissions",
				})
				return nil
			}

			return next(c)
		}
	}
}
