package logger

import (
	"os"
	"path/filepath"
)

var defaultLogger Logger

// Initialize sets up the default logger
func Initialize(env string) error {
	config := Config{
		Environment: env,
		LogPath:     filepath.Join("logs"),
		Level:       "info",
	}

	// Override log level from environment if set
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Level = level
	}

	// Override log path from environment if set
	if path := os.Getenv("LOG_PATH"); path != "" {
		config.LogPath = path
	}

	logger, err := NewLogger(config)
	if err != nil {
		return err
	}

	defaultLogger = logger
	return nil
}

// GetLogger returns the default logger instance
func GetLogger() Logger {
	if defaultLogger == nil {
		// If logger is not initialized, create a development logger
		logger, _ := NewLogger(Config{
			Environment: "development",
			Level:       "info",
		})
		defaultLogger = logger
	}
	return defaultLogger
}

// Default logger methods for convenience
func Info(msg string, fields ...Field) {
	GetLogger().Info(msg, fields...)
}

func Debug(msg string, fields ...Field) {
	GetLogger().Debug(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	GetLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	GetLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	GetLogger().Fatal(msg, fields...)
}

func With(fields ...Field) Logger {
	return GetLogger().With(fields...)
}
