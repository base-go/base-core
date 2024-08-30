package core // file core/core.go

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"base/core/app/auth"
	"base/core/app/users"
	"base/core/module"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"gorm.io/gorm"
)

// InitializeLogger sets up Logrus as the global logger
func InitializeLogger() *logrus.Logger {
	// Load .env file
	godotenv.Load()

	logger := logrus.New()
	env := os.Getenv("ENV")

	logDir := "./logs"
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		logger.Fatalf("Failed to create log directory: %v", err)
	}

	customFormatter := &CustomTextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		//colored:         true,
		ForceColors: true,
	}

	if env == "debug" {
		// Debug mode: Log everything to the terminal
		logger.SetOutput(os.Stdout)
		logger.SetLevel(logrus.DebugLevel)
		logger.SetFormatter(customFormatter)

		// Redirect Gin's logs to Logrus
		gin.DefaultWriter = logger.WriterLevel(logrus.InfoLevel)
		gin.DefaultErrorWriter = logger.WriterLevel(logrus.ErrorLevel)
	} else {
		// Release mode: Log to separate files
		// Setup log files
		requestFile, err := os.OpenFile(filepath.Join(logDir, "requests.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.Fatalf("Failed to open request log file: %v", err)
		}

		errorFile, err := os.OpenFile(filepath.Join(logDir, "error.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.Fatalf("Failed to open error log file: %v", err)
		}

		infoFile, err := os.OpenFile(filepath.Join(logDir, "info.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.Fatalf("Failed to open info log file: %v", err)
		}

		// Set up multi-writer for different log levels
		logger.SetOutput(io.MultiWriter(infoFile, os.Stdout))
		logger.SetLevel(logrus.InfoLevel)
		logger.SetFormatter(customFormatter)

		// Create hooks for error and request logs
		logger.AddHook(&writerHook{
			Writer:    errorFile,
			LogLevels: []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel},
			Formatter: customFormatter,
		})

		logger.AddHook(&writerHook{
			Writer:    requestFile,
			LogLevels: []logrus.Level{logrus.InfoLevel},
			Formatter: customFormatter,
		})

		// Redirect Gin's logs
		gin.DefaultWriter = io.MultiWriter(requestFile, os.Stdout)
		gin.DefaultErrorWriter = io.MultiWriter(errorFile, os.Stderr)
	}

	// Set the global logger
	logrus.SetOutput(logger.Out)
	logrus.SetFormatter(logger.Formatter)
	logrus.SetLevel(logger.Level)

	return logger
}

// CustomTextFormatter formats logs in a clean, readable text format
type CustomTextFormatter struct {
	TimestampFormat string
	ForceColors     bool
}

func (f *CustomTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format(f.TimestampFormat)
	level := strings.ToUpper(entry.Level.String())
	msg := fmt.Sprintf("%s [%s] %s", timestamp, level, entry.Message)

	for k, v := range entry.Data {
		msg += fmt.Sprintf(" %s=%v", k, v)
	}

	return []byte(msg + "\n"), nil
}

// InitializeModules loads and initializes all modules directly
func InitializeCoreModules(db *gorm.DB, router *gin.RouterGroup) map[string]module.Module {
	modules := make(map[string]module.Module)

	// Define the module initializers directly
	moduleInitializers := map[string]func(*gorm.DB, *gin.RouterGroup) module.Module{
		"users": func(db *gorm.DB, router *gin.RouterGroup) module.Module { return users.NewUserModule(db, router) },
		"auth":  func(db *gorm.DB, router *gin.RouterGroup) module.Module { return auth.NewAuthModule(db, router) },
		// MODULE_INITIALIZER_MARKER - Do not remove this comment because it's used by the CLI to add new module initializers

	}

	// Initialize and register each module
	for name, initializer := range moduleInitializers {
		module := initializer(db, router)
		modules[name] = module

	}

	return modules
}

// writerHook is a hook that writes logs of specified levels to a specified writer
type writerHook struct {
	Writer    io.Writer
	LogLevels []logrus.Level
	Formatter logrus.Formatter
}

func (hook *writerHook) Fire(entry *logrus.Entry) error {
	line, err := hook.Formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write(line)
	return err
}

func (hook *writerHook) Levels() []logrus.Level {
	return hook.LogLevels
}
