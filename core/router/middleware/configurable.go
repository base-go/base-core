package middleware

import (
	"base/core/config"
	"base/core/helper"
	"base/core/router"
	"strings"
)

// ConfigurableMiddleware creates middleware that can be conditionally applied based on configuration
type ConfigurableMiddleware struct {
	config *config.MiddlewareConfig
}

// NewConfigurableMiddleware creates a new configurable middleware instance
func NewConfigurableMiddleware(cfg *config.MiddlewareConfig) *ConfigurableMiddleware {
	return &ConfigurableMiddleware{
		config: cfg,
	}
}

// ConditionalAPIKey returns API key middleware only if required for the path
func (cm *ConfigurableMiddleware) ConditionalAPIKey() router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			path := c.Request.URL.Path
			
			if cm.config.IsAPIKeyRequired(path) {
				// Apply API key middleware
				apiKeyMiddleware := Api()
				return apiKeyMiddleware(next)(c)
			}
			
			// Skip API key middleware
			return next(c)
		}
	}
}

// ConditionalAuth returns auth middleware only if required for the path
func (cm *ConfigurableMiddleware) ConditionalAuth() router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			// Skip authentication for CORS preflight requests
			if c.Request.Method == "OPTIONS" {
				return next(c)
			}

			path := c.Request.URL.Path
			
			if cm.config.IsAuthRequired(path) {
				// Apply auth middleware
				authConfig := DefaultAuthConfig()
				authConfig.TokenValidator = func(token string) (any, error) {
					_, userID, err := helper.ValidateJWT(token)
					return userID, err
				}
				authMiddleware := Auth(authConfig)
				return authMiddleware(next)(c)
			}
			
			// Skip auth middleware
			return next(c)
		}
	}
}

// ConditionalRateLimit returns rate limit middleware only if required for the path
func (cm *ConfigurableMiddleware) ConditionalRateLimit() router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			path := c.Request.URL.Path
			
			if cm.config.IsRateLimitRequired(path) {
				// Determine rate limit settings based on path
				requests := cm.config.RateLimitRequests
				window := cm.config.GetRateLimitDuration()
				
				// Use webhook settings if it's a webhook path
				if cm.isWebhookPath(path) {
					requests = cm.config.WebhookRateLimitRequests
					window = cm.config.GetWebhookRateLimitDuration()
				}
				
				// Apply rate limit middleware
				rateLimitConfig := &RateLimitConfig{
					Limiter: NewTokenBucket(requests, window, requests),
					KeyFunc: func(c *router.Context) string {
						return c.ClientIP()
					},
					ErrorHandler: func(c *router.Context) error {
						return c.JSON(429, map[string]string{
							"error": "Rate limit exceeded",
						})
					},
				}
				rateLimitMiddleware := RateLimit(rateLimitConfig)
				return rateLimitMiddleware(next)(c)
			}
			
			// Skip rate limit middleware
			return next(c)
		}
	}
}

// ConditionalLogging returns logging middleware only if required for the path
func (cm *ConfigurableMiddleware) ConditionalLogging() router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			path := c.Request.URL.Path
			
			if cm.config.IsLoggingRequired(path) {
				// Apply logging middleware - this will be handled by main.go
				// For now, just continue to next middleware
				return next(c)
			}
			
			// Skip logging middleware
			return next(c)
		}
	}
}

// WebhookSignature creates webhook signature verification middleware
func (cm *ConfigurableMiddleware) WebhookSignature(provider string) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			path := c.Request.URL.Path
			
			// Only apply to webhook paths if signature verification is enabled
			if cm.isWebhookPath(path) && cm.config.WebhookSignatureEnabled {
				// TODO: Implement provider-specific signature verification
				// For now, just log and continue
				// This would verify HMAC signatures from Stripe, GitHub, etc.
			}
			
			return next(c)
		}
	}
}

// isWebhookPath checks if a path is configured as a webhook path
func (cm *ConfigurableMiddleware) isWebhookPath(path string) bool {
	for _, webhookPath := range cm.config.WebhookPaths {
		if cm.pathMatches(path, webhookPath) {
			return true
		}
	}
	return false
}

// pathMatches checks if a path matches a pattern (supports wildcards)
func (cm *ConfigurableMiddleware) pathMatches(path, pattern string) bool {
	if pattern == path {
		return true
	}
	
	// Handle wildcard patterns
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, prefix)
	}
	
	return false
}

// ApplyConfigurableMiddleware is a helper function to apply all configurable middleware
func ApplyConfigurableMiddleware(router *router.Router, cfg *config.MiddlewareConfig) {
	cm := NewConfigurableMiddleware(cfg)
	
	// Apply middleware in the correct order
	if cfg.RecoveryEnabled {
		router.Use(Recovery(nil)) // Recovery should be first
	}
	
	if cfg.CORSEnabled {
		// CORS middleware will be applied in main.go
	}
	
	// Apply conditional middleware
	router.Use(cm.ConditionalAPIKey())
	router.Use(cm.ConditionalAuth())
	router.Use(cm.ConditionalRateLimit())
	router.Use(cm.ConditionalLogging())
}
