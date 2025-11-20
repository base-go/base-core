package module

// ConfigurableModule extends the base Module interface with middleware configuration
type ConfigurableModule interface {
	Module
	
	// MiddlewareConfig returns middleware configuration overrides for this module
	// This allows modules to override global middleware settings for their specific routes
	MiddlewareConfig() *MiddlewareOverrides
}

// MiddlewareOverrides defines middleware configuration overrides for specific paths
type MiddlewareOverrides struct {
	// PathRules maps URL paths to middleware settings
	// Supports wildcards: "/api/webhooks/*" matches all webhook endpoints
	PathRules map[string]MiddlewareSettings
	
	// Global overrides apply to all routes in this module
	Global *MiddlewareSettings
}

// MiddlewareSettings defines middleware configuration for a specific path or module
type MiddlewareSettings struct {
	// APIKey controls API key requirement
	// nil = use global setting, true = require, false = skip
	APIKey *bool `json:"api_key,omitempty"`
	
	// Auth controls authentication requirement
	// nil = use global setting, true = require, false = skip
	Auth *bool `json:"auth,omitempty"`
	
	// RateLimit controls rate limiting
	// nil = use global setting, config = custom rate limit
	RateLimit *RateLimitConfig `json:"rate_limit,omitempty"`
	
	// Logging controls request logging
	// nil = use global setting, true = enable, false = disable
	Logging *bool `json:"logging,omitempty"`
	
	// CORS controls CORS headers
	// nil = use global setting, true = enable, false = disable
	CORS *bool `json:"cors,omitempty"`
	
	// WebhookSignature controls webhook signature verification
	// nil = use global setting, config = custom webhook config
	WebhookSignature *WebhookSignatureConfig `json:"webhook_signature,omitempty"`
}

// RateLimitConfig defines custom rate limiting configuration
type RateLimitConfig struct {
	// Requests per window
	Requests int `json:"requests"`
	
	// Window duration (e.g., "1m", "1h")
	Window string `json:"window"`
	
	// KeyFunc determines how to extract the rate limit key
	// Options: "ip", "user", "api_key", or custom function name
	KeyFunc string `json:"key_func,omitempty"`
}

// WebhookSignatureConfig defines webhook signature verification configuration
type WebhookSignatureConfig struct {
	// Provider name (e.g., "stripe", "github", "paypal")
	Provider string `json:"provider"`
	
	// Header name containing the signature
	Header string `json:"header"`
	
	// Secret environment variable name
	SecretEnvVar string `json:"secret_env_var"`
	
	// Algorithm (e.g., "sha256", "sha1")
	Algorithm string `json:"algorithm,omitempty"`
}

// Helper functions for creating middleware settings

// DisableAPIKey creates a setting that disables API key requirement
func DisableAPIKey() *MiddlewareSettings {
	disabled := false
	return &MiddlewareSettings{APIKey: &disabled}
}

// RequireAPIKey creates a setting that requires API key
func RequireAPIKey() *MiddlewareSettings {
	enabled := true
	return &MiddlewareSettings{APIKey: &enabled}
}

// DisableAuth creates a setting that disables authentication requirement
func DisableAuth() *MiddlewareSettings {
	disabled := false
	return &MiddlewareSettings{Auth: &disabled}
}

// RequireAuth creates a setting that requires authentication
func RequireAuth() *MiddlewareSettings {
	enabled := true
	return &MiddlewareSettings{Auth: &enabled}
}

// DisableAuthAndAPIKey creates a setting that disables both auth and API key (useful for webhooks)
func DisableAuthAndAPIKey() *MiddlewareSettings {
	disabled := false
	return &MiddlewareSettings{
		APIKey: &disabled,
		Auth:   &disabled,
	}
}

// CustomRateLimit creates a setting with custom rate limiting
func CustomRateLimit(requests int, window string) *MiddlewareSettings {
	return &MiddlewareSettings{
		RateLimit: &RateLimitConfig{
			Requests: requests,
			Window:   window,
		},
	}
}

// WebhookSignature creates a setting for webhook signature verification
func WebhookSignature(provider, header, secretEnvVar string) *MiddlewareSettings {
	return &MiddlewareSettings{
		WebhookSignature: &WebhookSignatureConfig{
			Provider:     provider,
			Header:       header,
			SecretEnvVar: secretEnvVar,
		},
	}
}
