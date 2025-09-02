package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Configuration defaults - centralized for easier maintenance
const (
	// Server defaults
	DefaultServerAddress = "localhost"
	DefaultServerPort    = ":8001"
	DefaultAppHost       = "http://localhost"
	DefaultEnvironment   = "debug"
	DefaultVersion       = "0.0.1"

	// Database defaults
	DefaultDBDriver   = "mysql"
	DefaultDBHost     = "localhost"
	DefaultDBPort     = "3306"
	DefaultDBUser     = "root"
	DefaultDBPassword = "RockeT"
	DefaultDBName     = "mydatabase"
	DefaultDBPath     = "test.db"

	// Security defaults
	DefaultJWTSecret = "secret"
	DefaultAPIKey    = "test_api_key"

	// Email defaults
	DefaultEmailProvider    = "default"
	DefaultEmailFromAddress = "no-reply@localhost"
	DefaultSMTPPort         = 587

	// Storage defaults
	DefaultStorageProvider   = "local"
	DefaultStoragePath       = "storage/uploads"
	DefaultStorageMaxSize    = 10485760 // 10MB
	DefaultStorageRegion     = "eu-central-1"
	DefaultStorageBucket     = "default"
	DefaultStorageExtensions = ".jpg,.jpeg,.png,.gif,.pdf,.doc,.docx"

	// Feature toggles defaults
	DefaultWebSocketEnabled = true
	DefaultSwaggerEnabled   = true
)

// Config holds the application configuration.
// Maintains exact same structure for backward compatibility
type Config struct {
	BaseURL              string
	CDN                  string
	Env                  string
	DBDriver             string
	DBUser               string
	DBPassword           string
	DBHost               string
	DBPort               string
	DBName               string
	DBPath               string
	DBURL                string
	ApiKey               string
	JWTSecret            string
	ServerAddress        string
	ServerPort           string
	CORSAllowedOrigins   []string
	Version              string
	EmailProvider        string
	EmailFromAddress     string
	SMTPHost             string
	SMTPPort             int
	SMTPUsername         string
	SMTPPassword         string
	SendGridAPIKey       string
	PostmarkServerToken  string
	PostmarkAccountToken string
	StorageProvider      string   `json:"storage_provider"`
	StoragePath          string   `json:"storage_path"`
	StorageBaseURL       string   `json:"storage_base_url"`
	StorageAPIKey        string   `json:"storage_api_key"`
	StorageAPISecret     string   `json:"storage_api_secret"`
	StorageAccountID     string   `json:"storage_account_id"`
	StorageEndpoint      string   `json:"storage_endpoint"`
	StorageRegion        string   `json:"storage_region"`
	StorageBucket        string   `json:"storage_bucket"`
	StoragePublicURL     string   `json:"storage_public_url"`
	StorageMaxSize       int64    `json:"storage_max_size"`
	StorageAllowedExt    []string `json:"storage_allowed_ext"`
	WebSocketEnabled     bool     `json:"websocket_enabled"`
	SwaggerEnabled       bool     `json:"swagger_enabled"`
	
	// Middleware configuration
	Middleware MiddlewareConfig `json:"middleware"`
}

// MiddlewareConfig holds middleware configuration settings
type MiddlewareConfig struct {
	// Global middleware toggles
	APIKeyEnabled     bool     `json:"api_key_enabled"`
	APIKeySkipPaths   []string `json:"api_key_skip_paths"`
	AuthEnabled       bool     `json:"auth_enabled"`
	AuthSkipPaths     []string `json:"auth_skip_paths"`
	RateLimitEnabled  bool     `json:"rate_limit_enabled"`
	RateLimitRequests int      `json:"rate_limit_requests"`
	RateLimitWindow   string   `json:"rate_limit_window"`
	RateLimitSkipPaths []string `json:"rate_limit_skip_paths"`
	LoggingEnabled    bool     `json:"logging_enabled"`
	LoggingSkipPaths  []string `json:"logging_skip_paths"`
	RecoveryEnabled   bool     `json:"recovery_enabled"`
	CORSEnabled       bool     `json:"cors_enabled"`
	
	// Webhook-specific settings
	WebhookPaths              []string `json:"webhook_paths"`
	WebhookAPIKeyEnabled      bool     `json:"webhook_api_key_enabled"`
	WebhookAuthEnabled        bool     `json:"webhook_auth_enabled"`
	WebhookSignatureEnabled   bool     `json:"webhook_signature_enabled"`
	WebhookRateLimitRequests  int      `json:"webhook_rate_limit_requests"`
	WebhookRateLimitWindow    string   `json:"webhook_rate_limit_window"`
	
	// Per-endpoint overrides
	Overrides map[string]map[string]string `json:"overrides"`
}

// GetRateLimitDuration returns the rate limit window as time.Duration
func (m *MiddlewareConfig) GetRateLimitDuration() time.Duration {
	duration, err := time.ParseDuration(m.RateLimitWindow)
	if err != nil {
		return time.Minute // default to 1 minute
	}
	return duration
}

// GetWebhookRateLimitDuration returns the webhook rate limit window as time.Duration
func (m *MiddlewareConfig) GetWebhookRateLimitDuration() time.Duration {
	duration, err := time.ParseDuration(m.WebhookRateLimitWindow)
	if err != nil {
		return time.Hour // default to 1 hour
	}
	return duration
}

// IsAPIKeyRequired checks if API key is required for a given path
func (m *MiddlewareConfig) IsAPIKeyRequired(path string) bool {
	if !m.APIKeyEnabled {
		return false
	}
	
	// Check if it's a webhook path
	if m.isWebhookPath(path) {
		return m.WebhookAPIKeyEnabled
	}
	
	// Check global skip paths
	for _, skipPath := range m.APIKeySkipPaths {
		if m.pathMatches(path, skipPath) {
			return false
		}
	}
	
	// Check per-endpoint overrides
	for overridePath, settings := range m.Overrides {
		if m.pathMatches(path, overridePath) {
			if apiKeySetting, exists := settings["api_key"]; exists {
				return apiKeySetting != "disabled"
			}
		}
	}
	
	return true
}

// IsAuthRequired checks if authentication is required for a given path
func (m *MiddlewareConfig) IsAuthRequired(path string) bool {
	if !m.AuthEnabled {
		return false
	}
	
	// Check if it's a webhook path
	if m.isWebhookPath(path) {
		return m.WebhookAuthEnabled
	}
	
	// Check global skip paths
	for _, skipPath := range m.AuthSkipPaths {
		if m.pathMatches(path, skipPath) {
			return false
		}
	}
	
	// Check per-endpoint overrides
	for overridePath, settings := range m.Overrides {
		if m.pathMatches(path, overridePath) {
			if authSetting, exists := settings["auth"]; exists {
				return authSetting != "disabled"
			}
		}
	}
	
	return true
}

// IsRateLimitRequired checks if rate limiting is required for a given path
func (m *MiddlewareConfig) IsRateLimitRequired(path string) bool {
	if !m.RateLimitEnabled {
		return false
	}
	
	// Check global skip paths
	for _, skipPath := range m.RateLimitSkipPaths {
		if m.pathMatches(path, skipPath) {
			return false
		}
	}
	
	return true
}

// IsLoggingRequired checks if logging is required for a given path
func (m *MiddlewareConfig) IsLoggingRequired(path string) bool {
	if !m.LoggingEnabled {
		return false
	}
	
	// Check global skip paths
	for _, skipPath := range m.LoggingSkipPaths {
		if m.pathMatches(path, skipPath) {
			return false
		}
	}
	
	return true
}

// isWebhookPath checks if a path is configured as a webhook path
func (m *MiddlewareConfig) isWebhookPath(path string) bool {
	for _, webhookPath := range m.WebhookPaths {
		if m.pathMatches(path, webhookPath) {
			return true
		}
	}
	return false
}

// pathMatches checks if a path matches a pattern (supports wildcards)
func (m *MiddlewareConfig) pathMatches(path, pattern string) bool {
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

// NewConfig returns a new Config instance with default values.
// Improved version with better organization and error handling
func NewConfig() *Config {
	// Server configuration
	serverAddr := getEnvWithLog("SERVER_ADDRESS", DefaultServerAddress)
	serverPort := normalizePort(getEnvWithLog("SERVER_PORT", DefaultServerPort))
	baseURL := buildBaseURL(getEnvWithLog("APPHOST", DefaultAppHost), serverPort)

	// Create config with all basic string/simple values
	config := &Config{
		// Server settings
		BaseURL:       baseURL,
		CDN:           getEnvWithLog("CDN", ""),
		Env:           getEnvWithLog("ENV", DefaultEnvironment),
		ServerAddress: serverAddr,
		ServerPort:    serverPort,
		Version:       getEnvWithLog("APP_VERSION", DefaultVersion),

		// Database settings
		DBDriver:   getEnvWithLog("DB_DRIVER", DefaultDBDriver),
		DBUser:     getEnvWithLog("DB_USER", DefaultDBUser),
		DBPassword: getEnvWithLog("DB_PASSWORD", DefaultDBPassword),
		DBHost:     getEnvWithLog("DB_HOST", DefaultDBHost),
		DBPort:     getEnvWithLog("DB_PORT", DefaultDBPort),
		DBName:     getEnvWithLog("DB_NAME", DefaultDBName),
		DBPath:     getEnvWithLog("DB_PATH", DefaultDBPath),
		DBURL:      getEnvWithLog("DB_URL", ""),

		// Security settings
		ApiKey:    getEnvWithLog("API_KEY", DefaultAPIKey),
		JWTSecret: getEnvWithLog("JWT_SECRET", DefaultJWTSecret),

		// Email settings
		EmailProvider:        getEnvWithLog("EMAIL_PROVIDER", DefaultEmailProvider),
		EmailFromAddress:     getEnvWithLog("EMAIL_FROM_ADDRESS", DefaultEmailFromAddress),
		SMTPHost:             getEnvWithLog("SMTP_HOST", ""),
		SMTPUsername:         getEnvWithLog("SMTP_USERNAME", ""),
		SMTPPassword:         getEnvWithLog("SMTP_PASSWORD", ""),
		SendGridAPIKey:       getEnvWithLog("SENDGRID_API_KEY", ""),
		PostmarkServerToken:  getEnvWithLog("POSTMARK_SERVER_TOKEN", ""),
		PostmarkAccountToken: getEnvWithLog("POSTMARK_ACCOUNT_TOKEN", ""),

		// Storage settings
		StorageProvider:  getEnvWithLog("STORAGE_PROVIDER", DefaultStorageProvider),
		StoragePath:      getEnvWithLog("STORAGE_PATH", DefaultStoragePath),
		StorageBaseURL:   getEnvWithLog("STORAGE_BASE_URL", ""),
		StorageAPIKey:    getEnvWithLog("STORAGE_API_KEY", ""),
		StorageAPISecret: getEnvWithLog("STORAGE_API_SECRET", ""),
		StorageAccountID: getEnvWithLog("STORAGE_ACCOUNT_ID", ""),
		StorageEndpoint:  getEnvWithLog("STORAGE_ENDPOINT", ""),
		StorageRegion:    getEnvWithLog("STORAGE_REGION", DefaultStorageRegion),
		StorageBucket:    getEnvWithLog("STORAGE_BUCKET", DefaultStorageBucket),
		StoragePublicURL: getEnvWithLog("STORAGE_PUBLIC_URL", ""),
	}

	// Parse complex values with proper error handling
	parseCORSOrigins(config)
	parseStorageExtensions(config)
	parseIntegerValues(config)
	parseBooleanValues(config)
	parseMiddlewareConfig(config)

	return config
}

// parseCORSOrigins parses and cleans CORS origins
func parseCORSOrigins(config *Config) {
	corsOriginsStr := getEnvWithLog("CORS_ALLOWED_ORIGINS", "")
	if corsOriginsStr != "" {
		origins := strings.Split(corsOriginsStr, ",")
		// Clean up whitespace
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
		config.CORSAllowedOrigins = origins
	}
}

// parseStorageExtensions parses allowed storage extensions
func parseStorageExtensions(config *Config) {
	extensionsStr := getEnvWithLog("STORAGE_ALLOWED_EXT", DefaultStorageExtensions)
	if extensionsStr != "" {
		extensions := strings.Split(extensionsStr, ",")
		// Clean up whitespace
		for i, ext := range extensions {
			extensions[i] = strings.TrimSpace(ext)
		}
		config.StorageAllowedExt = extensions
	}
}

// parseIntegerValues parses all integer configuration values
func parseIntegerValues(config *Config) {
	// SMTP Port
	config.SMTPPort = parseIntWithDefault("SMTP_PORT", DefaultSMTPPort)

	// Storage Max Size
	config.StorageMaxSize = parseInt64WithDefault("STORAGE_MAX_SIZE", DefaultStorageMaxSize)
}

// parseBooleanValues parses all boolean configuration values
func parseBooleanValues(config *Config) {
	// WebSocket enabled
	config.WebSocketEnabled = parseBoolWithDefault("WS_ENABLED", DefaultWebSocketEnabled)

	// Swagger enabled
	config.SwaggerEnabled = parseBoolWithDefault("SWAGGER_ENABLED", DefaultSwaggerEnabled)
}

// parseMiddlewareConfig parses middleware configuration from environment variables
func parseMiddlewareConfig(config *Config) {
	// Parse middleware overrides JSON if provided
	overridesStr := getEnvWithLog("MIDDLEWARE_OVERRIDES", "{}")
	var overrides map[string]map[string]string
	if err := json.Unmarshal([]byte(overridesStr), &overrides); err != nil {
		logConfigError("Invalid MIDDLEWARE_OVERRIDES JSON: %s. Using empty overrides", overridesStr)
		overrides = make(map[string]map[string]string)
	}
	
	// Parse webhook paths
	webhookPathsStr := getEnvWithLog("MIDDLEWARE_WEBHOOK_PATHS", "/api/webhooks/*,/webhooks/*")
	webhookPaths := []string{}
	if webhookPathsStr != "" {
		paths := strings.Split(webhookPathsStr, ",")
		for _, path := range paths {
			webhookPaths = append(webhookPaths, strings.TrimSpace(path))
		}
	}
	
	config.Middleware = MiddlewareConfig{
		// Global middleware settings
		APIKeyEnabled:     parseBoolWithDefault("MIDDLEWARE_API_KEY_ENABLED", true),
		APIKeySkipPaths:   parsePathList("MIDDLEWARE_API_KEY_SKIP_PATHS", "/health,/,/docs,/swagger"),
		AuthEnabled:       parseBoolWithDefault("MIDDLEWARE_AUTH_ENABLED", false),
		AuthSkipPaths:     parsePathList("MIDDLEWARE_AUTH_SKIP_PATHS", "/api/auth/login,/api/auth/register,/api/auth/forgot-password"),
		RateLimitEnabled:  parseBoolWithDefault("MIDDLEWARE_RATE_LIMIT_ENABLED", true),
		RateLimitRequests: parseIntWithDefault("MIDDLEWARE_RATE_LIMIT_REQUESTS", 60),
		RateLimitWindow:   getEnvWithLog("MIDDLEWARE_RATE_LIMIT_WINDOW", "1m"),
		RateLimitSkipPaths: parsePathList("MIDDLEWARE_RATE_LIMIT_SKIP_PATHS", "/health,/"),
		LoggingEnabled:    parseBoolWithDefault("MIDDLEWARE_LOGGING_ENABLED", true),
		LoggingSkipPaths:  parsePathList("MIDDLEWARE_LOGGING_SKIP_PATHS", ""),
		RecoveryEnabled:   parseBoolWithDefault("MIDDLEWARE_RECOVERY_ENABLED", true),
		CORSEnabled:       parseBoolWithDefault("MIDDLEWARE_CORS_ENABLED", true),
		
		// Webhook-specific settings
		WebhookPaths:              webhookPaths,
		WebhookAPIKeyEnabled:      parseBoolWithDefault("MIDDLEWARE_WEBHOOK_API_KEY_ENABLED", false),
		WebhookAuthEnabled:        parseBoolWithDefault("MIDDLEWARE_WEBHOOK_AUTH_ENABLED", false),
		WebhookSignatureEnabled:   parseBoolWithDefault("MIDDLEWARE_WEBHOOK_SIGNATURE_ENABLED", true),
		WebhookRateLimitRequests:  parseIntWithDefault("MIDDLEWARE_WEBHOOK_RATE_LIMIT_REQUESTS", 1000),
		WebhookRateLimitWindow:    getEnvWithLog("MIDDLEWARE_WEBHOOK_RATE_LIMIT_WINDOW", "1h"),
		
		// Per-endpoint overrides
		Overrides: overrides,
	}
}

// parsePathList parses a comma-separated list of paths
func parsePathList(key, defaultValue string) []string {
	pathsStr := getEnvWithLog(key, defaultValue)
	if pathsStr == "" {
		return []string{}
	}
	
	paths := strings.Split(pathsStr, ",")
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		trimmed := strings.TrimSpace(path)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Helper functions for type parsing with error handling

// parseIntWithDefault parses an integer environment variable with default fallback
func parseIntWithDefault(key string, defaultValue int) int {
	valueStr := getEnvWithLog(key, fmt.Sprintf("%d", defaultValue))
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		logConfigError("Invalid %s value: %s. Using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}

// parseInt64WithDefault parses an int64 environment variable with default fallback
func parseInt64WithDefault(key string, defaultValue int64) int64 {
	valueStr := getEnvWithLog(key, fmt.Sprintf("%d", defaultValue))
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		logConfigError("Invalid %s value: %s. Using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}

// parseBoolWithDefault parses a boolean environment variable with default fallback
func parseBoolWithDefault(key string, defaultValue bool) bool {
	valueStr := getEnvWithLog(key, fmt.Sprintf("%t", defaultValue))
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		logConfigError("Invalid %s value: %s. Using default: %t", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}

// normalizePort ensures port starts with ":"
func normalizePort(port string) string {
	if port != "" && port[0] != ':' {
		return ":" + port
	}
	return port
}

// buildBaseURL constructs the base URL with port if needed
func buildBaseURL(baseURL, port string) string {
	if !strings.Contains(baseURL, ":") || strings.HasSuffix(baseURL, "localhost") {
		return baseURL + port
	}
	return baseURL
}

// logConfigError logs configuration errors in a consistent format
func logConfigError(format string, args ...any) {
	fmt.Printf("[CONFIG ERROR] "+format+"\n", args...)
}

// GetStorageConfig returns storage configuration as a map for backward compatibility
func (c *Config) GetStorageConfig() map[string]any {
	return map[string]any{
		"provider":    c.StorageProvider,
		"api_key":     c.StorageAPIKey,
		"api_secret":  c.StorageAPISecret,
		"endpoint":    c.StorageEndpoint,
		"region":      c.StorageRegion,
		"bucket":      c.StorageBucket,
		"public_url":  c.StoragePublicURL,
		"base_url":    c.StorageBaseURL,
		"max_size":    c.StorageMaxSize,
		"allowed_ext": c.StorageAllowedExt,
		"path":        c.StoragePath,
		"env":         c.Env,
	}
}

// getEnvWithLog returns the value of an environment variable with a fallback default value
func getEnvWithLog(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}
	return fallback
}

// Validation methods for production use

// Validate validates the configuration and returns any errors
func (c *Config) Validate() []error {
	var errors []error

	// Validate database configuration
	if c.DBDriver != "sqlite" && c.DBURL == "" {
		if c.DBHost == "" {
			errors = append(errors, fmt.Errorf("DB_HOST is required for database driver: %s", c.DBDriver))
		}
		if c.DBName == "" {
			errors = append(errors, fmt.Errorf("DB_NAME is required for database driver: %s", c.DBDriver))
		}
	}

	if c.DBDriver == "sqlite" && c.DBPath == "" {
		errors = append(errors, fmt.Errorf("DB_PATH is required for SQLite driver"))
	}

	// Validate storage configuration
	if c.StorageProvider == "s3" || c.StorageProvider == "r2" {
		if c.StorageAPIKey == "" {
			errors = append(errors, fmt.Errorf("STORAGE_API_KEY is required for %s provider", c.StorageProvider))
		}
		if c.StorageAPISecret == "" {
			errors = append(errors, fmt.Errorf("STORAGE_API_SECRET is required for %s provider", c.StorageProvider))
		}
		if c.StorageBucket == "" {
			errors = append(errors, fmt.Errorf("STORAGE_BUCKET is required for %s provider", c.StorageProvider))
		}
	}

	// Validate email configuration
	if c.EmailProvider == "smtp" && c.SMTPHost == "" {
		errors = append(errors, fmt.Errorf("SMTP_HOST is required for SMTP email provider"))
	}

	// Security validations for production
	if c.Env == "production" {
		if c.JWTSecret == DefaultJWTSecret {
			errors = append(errors, fmt.Errorf("JWT_SECRET must be changed from default value in production"))
		}
		if c.ApiKey == DefaultAPIKey {
			errors = append(errors, fmt.Errorf("API_KEY must be changed from default value in production"))
		}
	}

	return errors
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// IsDevelopment returns true if the environment is development/debug
func (c *Config) IsDevelopment() bool {
	return c.Env == "debug" || c.Env == "development"
}

// GetDatabaseDSN builds a database connection string based on the driver
func (c *Config) GetDatabaseDSN() string {
	if c.DBURL != "" {
		return c.DBURL
	}

	switch c.DBDriver {
	case "sqlite":
		return c.DBPath
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
	case "postgres":
		return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
			c.DBHost, c.DBPort, c.DBUser, c.DBName, c.DBPassword)
	default:
		return ""
	}
}
