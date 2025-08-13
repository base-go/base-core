package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"base/core/router"
)

// AuthConfig contains authentication middleware configuration
type AuthConfig struct {
	// TokenValidator validates the token and returns user data
	TokenValidator func(token string) (any, error)

	// ContextKey is the key used to store user data in context
	ContextKey string

	// HeaderName is the header name to look for the token
	HeaderName string

	// Scheme is the authentication scheme (e.g., "Bearer")
	Scheme string

	// ErrorHandler handles authentication errors
	ErrorHandler func(*router.Context, error) error

	// SkipPaths lists paths that don't require authentication
	SkipPaths []string
}

// DefaultAuthConfig returns default auth configuration
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		HeaderName: "Authorization",
		Scheme:     "Bearer",
		ContextKey: "user",
		ErrorHandler: func(c *router.Context, err error) error {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Unauthorized",
			})
		},
	}
}

// Auth creates authentication middleware
func Auth(config *AuthConfig) router.MiddlewareFunc {
	if config == nil {
		config = DefaultAuthConfig()
	}

	if config.TokenValidator == nil {
		panic("TokenValidator is required for auth middleware")
	}

	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			// Check if path should be skipped
			for _, path := range config.SkipPaths {
				if c.Request.URL.Path == path {
					return next(c)
				}
			}

			// Get token from header
			authHeader := c.Header(config.HeaderName)
			if authHeader == "" {
				return config.ErrorHandler(c, errors.New("missing authorization header"))
			}

			// Extract token from scheme
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != config.Scheme {
				return config.ErrorHandler(c, errors.New("invalid authorization format"))
			}

			token := parts[1]

			// Validate token
			user, err := config.TokenValidator(token)
			if err != nil {
				return config.ErrorHandler(c, err)
			}

			// Store user in context
			c.Set(config.ContextKey, user)

			// Also add to request context for deeper layers
			ctx := context.WithValue(c.Request.Context(), config.ContextKey, user)
			c.Request = c.Request.WithContext(ctx)

			return next(c)
		}
	}
}

// RequireAuth is a simple auth middleware that just checks if user is present
func RequireAuth(contextKey string) router.MiddlewareFunc {
	if contextKey == "" {
		contextKey = "user"
	}

	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			if _, exists := c.Get(contextKey); !exists {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Authentication required",
				})
			}
			return next(c)
		}
	}
}

// APIKeyAuth creates API key authentication middleware
func APIKeyAuth(validateKey func(string) (any, error)) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			// Check header first
			apiKey := c.Header("X-API-Key")

			// Fall back to query parameter
			if apiKey == "" {
				apiKey = c.Query("api_key")
			}

			if apiKey == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "API key required",
				})
			}

			// Validate API key
			data, err := validateKey(apiKey)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid API key",
				})
			}

			// Store API key data in context
			c.Set("api_key_data", data)

			return next(c)
		}
	}
}

// BasicAuth creates basic authentication middleware
func BasicAuth(validateCredentials func(username, password string) (any, error)) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			username, password, hasAuth := c.Request.BasicAuth()
			if !hasAuth {
				c.SetHeader("WWW-Authenticate", `Basic realm="Restricted"`)
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Authorization required",
				})
			}

			user, err := validateCredentials(username, password)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid credentials",
				})
			}

			c.Set("user", user)
			return next(c)
		}
	}
}
