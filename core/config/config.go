package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// Config holds the application configuration.
type Config struct {
	BaseURL              string
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
	StorageAPIKey        string   `json:"storage_api_key"`
	StorageAPISecret     string   `json:"storage_api_secret"`
	StorageEndpoint      string   `json:"storage_endpoint"`
	StorageRegion        string   `json:"storage_region"`
	StorageBucket        string   `json:"storage_bucket"`
	StoragePublicURL     string   `json:"storage_public_url"`
	StorageMaxSize       int64    `json:"storage_max_size"`
	StorageAllowedExt    []string `json:"storage_allowed_ext"`
	StoragePath          string   `json:"storage_path"`
}

// NewConfig returns a new Config instance with default values.
func NewConfig() *Config {
	config := &Config{
		BaseURL:            getEnvWithLog("APPHOST", "http://localhost"),
		Env:                getEnvWithLog("ENV", "development"),
		DBDriver:           getEnvWithLog("DB_DRIVER", "mysql"),
		DBUser:             getEnvWithLog("DB_USER", "root"),
		DBPassword:         getEnvWithLog("DB_PASSWORD", "RockeT"),
		DBHost:             getEnvWithLog("DB_HOST", "localhost"),
		DBPort:             getEnvWithLog("DB_PORT", "3306"),
		DBName:             getEnvWithLog("DB_NAME", "mydatabase"),
		DBPath:             getEnvWithLog("DB_PATH", "test.db"),
		DBURL:              getEnvWithLog("DB_URL", ""),
		ApiKey:             getEnvWithLog("API_KEY", "test_api_key"),
		JWTSecret:          getEnvWithLog("JWT_SECRET", "secret"),
		ServerAddress:      getEnvWithLog("SERVER_ADDRESS", ":8080"),
		CORSAllowedOrigins: []string{"https://admin.albafone.net", "http://localhost:3000"},
		Version:            getEnvWithLog("APP_VERSION", "0.0.1"),

		EmailProvider:        getEnvWithLog("EMAIL_PROVIDER", "default"),
		EmailFromAddress:     getEnvWithLog("EMAIL_FROM_ADDRESS", "no-reply@localhost"),
		SMTPHost:             getEnvWithLog("SMTP_HOST", ""),
		SMTPUsername:         getEnvWithLog("SMTP_USERNAME", ""),
		SMTPPassword:         getEnvWithLog("SMTP_PASSWORD", ""),
		SendGridAPIKey:       getEnvWithLog("SENDGRID_API_KEY", ""),
		PostmarkServerToken:  getEnvWithLog("POSTMARK_SERVER_TOKEN", ""),
		PostmarkAccountToken: getEnvWithLog("POSTMARK_ACCOUNT_TOKEN", ""),
		StorageProvider:      getEnvWithLog("STORAGE_PROVIDER", "local"),
		StorageAPIKey:        getEnvWithLog("STORAGE_API_KEY", ""),
		StorageAPISecret:     getEnvWithLog("STORAGE_API_SECRET", ""),
		StorageEndpoint:      getEnvWithLog("STORAGE_ENDPOINT", ""),
		StorageRegion:        getEnvWithLog("STORAGE_REGION", "eu-central-1"),
		StorageBucket:        getEnvWithLog("STORAGE_BUCKET", "default"),
		StoragePublicURL:     getEnvWithLog("STORAGE_PUBLIC_URL", ""),
		StoragePath:          getEnvWithLog("STORAGE_PATH", "storage/uploads"),
		StorageAllowedExt: strings.Split(
			getEnvWithLog("STORAGE_ALLOWED_EXT", ".jpg,.jpeg,.png,.gif,.pdf,.doc,.docx"),
			",",
		),
	}

	// Handle SMTP_PORT as an integer
	smtpPortStr := getEnvWithLog("SMTP_PORT", "587")
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		logrus.Warnf("Invalid SMTP_PORT value: %s. Using default: 587", smtpPortStr)
		smtpPort = 587
	}
	config.SMTPPort = smtpPort

	storageSizeStr := getEnvWithLog("STORAGE_MAX_SIZE", "10485760")
	storageSize, err := strconv.ParseInt(storageSizeStr, 10, 64)
	if err != nil {
		logrus.Warnf("Invalid STORAGE_MAX_SIZE value: %s. Using default: 10MB", storageSizeStr)
		storageSize = 10 << 20
	}
	config.StorageMaxSize = storageSize

	return config
}
func (c *Config) GetStorageConfig() map[string]interface{} {
	return map[string]interface{}{
		"provider":    c.StorageProvider,
		"api_key":     c.StorageAPIKey,
		"api_secret":  c.StorageAPISecret,
		"endpoint":    c.StorageEndpoint,
		"region":      c.StorageRegion,
		"bucket":      c.StorageBucket,
		"public_url":  c.StoragePublicURL,
		"max_size":    c.StorageMaxSize,
		"allowed_ext": c.StorageAllowedExt,
		"path":        c.StoragePath,
		"env":         c.Env,
	}
}

// getEnvWithLog returns the value of an environment variable with a fallback default value and logs the result.
func getEnvWithLog(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}
	return fallback
}

// maskString masks a string for secure logging
// func maskString(s string) string {
// 	if len(s) <= 4 {
// 		return "****"
// 	}
// 	return s[:4] + "****"
// }
