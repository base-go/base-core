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
	return &Config{
		DBDriver:   getEnv("DB_DRIVER", "mysql"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBName:     getEnv("DB_NAME", "mydatabase"),
		DBPath:     getEnv("DB_PATH", "test.db"), // Default path for SQLite
		DBURL:      getEnv("DB_URL", ""),

		JWTSecret:     getEnv("JWT_SECRET", "secret"),
		ServerAddress: getEnv("SERVER_ADDRESS", ":8080"),
	}
}

// getEnv returns the value of an environment variable with a fallback default value.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
