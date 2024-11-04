package config

import (
	"os"
	"strconv"

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
	}

	// Handle SMTP_PORT as an integer
	smtpPortStr := getEnvWithLog("SMTP_PORT", "587")
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		logrus.Warnf("Invalid SMTP_PORT value: %s. Using default: 587", smtpPortStr)
		smtpPort = 587
	}
	config.SMTPPort = smtpPort

	// Debug logging
	// logrus.Infof("Loaded configuration:")
	// logrus.Infof("EMAIL_PROVIDER: %s", config.EmailProvider)
	// logrus.Infof("EMAIL_FROM_ADDRESS: %s", config.EmailFromAddress)
	// logrus.Infof("SMTP_HOST: %s", config.SMTPHost)
	// logrus.Infof("SMTP_PORT: %d", config.SMTPPort)
	// logrus.Infof("SMTP_USERNAME: %s", config.SMTPUsername)
	// logrus.Infof("SENDGRID_API_KEY: %s", maskString(config.SendGridAPIKey))
	// logrus.Infof("POSTMARK_SERVER_TOKEN: %s", maskString(config.PostmarkServerToken))

	return config
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
