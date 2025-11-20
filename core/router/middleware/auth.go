package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"base/core/router"
)

// contextKey is an empty struct with a descriptive name tag. Using a
// pointer to a unique struct instance avoids allocations on every
// WithValue call and guarantees key uniqueness across packages.
type contextKey struct{ name string }

var (
	userContextKey = &contextKey{"user"}
)

// ContextValue retrieves a value of type T from the context using the provided key.
func ContextValue[T any](ctx context.Context, key *contextKey) (T, bool) {
	val := ctx.Value(key)
	// if nil or wrong type, return zero value
	if v, ok := val.(T); ok {
		return v, true
	}
	var zero T
	return zero, false
}

// UserFromContext is a convenience wrapper to get the user stored by the auth middlewares.
func UserFromContext[T any](ctx context.Context) (T, bool) {
	return ContextValue[T](ctx, userContextKey)
}

// AuthConfig contains authentication middleware configuration
type AuthConfig struct {
	// TokenValidator validates the token and returns user data
	TokenValidator func(token string) (any, error)

	// Context Key is used to store user data in context
	Key string

	// Header Name is to look for the token
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
		Key:        "user",
		ErrorHandler: func(c *router.Context, err error) error {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Unauthorized: " + err.Error(),
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
			// Skip authentication for CORS preflight requests
			if c.Request.Method == "OPTIONS" {
				return next(c)
			}

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

			// Store user ID with "user_id" key for authorization middleware
			// This is the essential information needed for permission checks
			if userID, ok := user.(uint); ok {
				c.Set("user_id", userID)
				c.Set(config.Key, userID) // Also store with configured key for backward compatibility
			} else if userID, ok := user.(uint64); ok {
				c.Set("user_id", userID)
				c.Set(config.Key, userID) // Also store with configured key for backward compatibility
			}

			// Also add to request context for deeper layers
			ctx := context.WithValue(c.Request.Context(), userContextKey, user)
			c.Request = c.Request.WithContext(ctx)

			return next(c)
		}
	}
}

// RequireAuth is a simple auth middleware that just checks if user is present
func RequireAuth(key string) router.MiddlewareFunc {
	if key == "" {
		key = "user"
	}

	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			// Skip authentication for CORS preflight requests
			if c.Request.Method == "OPTIONS" {
				return next(c)
			}

			if _, exists := c.Get(key); !exists {
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
			// Skip authentication for CORS preflight requests
			if c.Request.Method == "OPTIONS" {
				return next(c)
			}

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
			// Skip authentication for CORS preflight requests
			if c.Request.Method == "OPTIONS" {
				return next(c)
			}

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

			// Also add to request context for deeper layers
			ctx := context.WithValue(c.Request.Context(), userContextKey, user)
			c.Request = c.Request.WithContext(ctx)

			return next(c)
		}
	}
}
