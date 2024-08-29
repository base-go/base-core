package config

import (
	"os"
)

// Config holds the application configuration.
type Config struct {
	DBDriver      string
	DBUser        string
	DBPassword    string
	DBHost        string
	DBPort        string
	DBName        string
	DBPath        string
	DBURL         string
	JWTSecret     string
	ServerAddress string
}

// NewConfig returns a new Config instance with default values.
func NewConfig() *Config {
	config := &Config{
		DBDriver:      getEnvWithLog("DB_DRIVER", "mysql"),
		DBUser:        getEnvWithLog("DB_USER", "root"),
		DBPassword:    getEnvWithLog("DB_PASSWORD", "RockeT"),
		DBHost:        getEnvWithLog("DB_HOST", "localhost"),
		DBPort:        getEnvWithLog("DB_PORT", "3306"),
		DBName:        getEnvWithLog("DB_NAME", "mydatabase"),
		DBPath:        getEnvWithLog("DB_PATH", "test.db"),
		DBURL:         getEnvWithLog("DB_URL", ""),
		JWTSecret:     getEnvWithLog("JWT_SECRET", "secret"),
		ServerAddress: getEnvWithLog("SERVER_ADDRESS", ":8080"),
	}

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
