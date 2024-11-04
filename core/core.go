package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"base/core/app/auth"
	"base/core/app/users"
	"base/core/email"
	"base/core/event"
	"base/core/module"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

// ModuleInitializer is a struct to hold dependencies for module initialization
type ModuleInitializer struct {
	DB           *gorm.DB
	Router       *gin.RouterGroup
	EmailSender  email.Sender
	Logger       *zap.Logger
	EventService *event.EventService
}

// InitializeLogger sets up Zap as the global logger
func InitializeLogger() *zap.Logger {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		// Only log warning as .env file might not exist in production
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "production" // Default to production if not set
	}

	logDir := "./logs"
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		panic(fmt.Sprintf("Failed to create log directory: %v", err))
	}

	// Configure encoder with detailed settings
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "func",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var cores []zapcore.Core

	if env == "debug" || env == "development" {
		// Debug mode: Log everything to console with development settings
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			zapcore.DebugLevel,
		)
		cores = append(cores, consoleCore)

		// Redirect Gin's logs
		gin.DefaultWriter = os.Stdout
		gin.DefaultErrorWriter = os.Stderr
	} else {
		// Production mode: Log to files with appropriate levels
		jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

		// Info logs
		infoFile, err := os.OpenFile(
			filepath.Join(logDir, "info.log"),
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0644,
		)
		if err != nil {
			panic(fmt.Sprintf("Failed to open info log file: %v", err))
		}

		infoCore := zapcore.NewCore(
			jsonEncoder,
			zapcore.AddSync(io.MultiWriter(infoFile, os.Stdout)),
			zapcore.InfoLevel,
		)
		cores = append(cores, infoCore)

		// Error logs
		errorFile, err := os.OpenFile(
			filepath.Join(logDir, "error.log"),
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0644,
		)
		if err != nil {
			panic(fmt.Sprintf("Failed to open error log file: %v", err))
		}

		errorCore := zapcore.NewCore(
			jsonEncoder,
			zapcore.AddSync(io.MultiWriter(errorFile, os.Stderr)),
			zapcore.ErrorLevel,
		)
		cores = append(cores, errorCore)

		// Request logs for API access
		requestFile, err := os.OpenFile(
			filepath.Join(logDir, "requests.log"),
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0644,
		)
		if err != nil {
			panic(fmt.Sprintf("Failed to open request log file: %v", err))
		}

		// Redirect Gin's logs to the request log file
		gin.DefaultWriter = io.MultiWriter(requestFile, os.Stdout)
		gin.DefaultErrorWriter = io.MultiWriter(errorFile, os.Stderr)
	}

	// Combine cores
	core := zapcore.NewTee(cores...)

	// Create logger with caller and stacktrace
	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Development(),
	)

	// Replace the global logger
	zap.ReplaceGlobals(logger)

	return logger
}

// InitializeCoreModules loads and initializes all core modules
func InitializeCoreModules(init ModuleInitializer) map[string]module.Module {
	modules := make(map[string]module.Module)
	ctx := context.Background()

	if init.Logger == nil {
		init.Logger = zap.NewNop()
	}

	if init.EventService != nil {
		// Track system startup
		init.EventService.Track(ctx, event.EventOptions{
			Type:        "system",
			Category:    "startup",
			Actor:       "system",
			Action:      "initialize_modules",
			Status:      "started",
			Description: "Starting core modules initialization",
		})
	}

	// Define module initializers with proper error handling
	moduleInitializers := map[string]func() (module.Module, error){
		"users": func() (module.Module, error) {
			m := users.NewUserModule(
				init.DB,
				init.Router,
				init.Logger,
				init.EventService,
			)
			if m == nil {
				return nil, fmt.Errorf("failed to create users module")
			}
			return m, nil
		},
		"auth": func() (module.Module, error) {
			m := auth.NewAuthModule(
				init.DB,
				init.Router,
				init.EmailSender,
				init.Logger,
				init.EventService,
			)
			if m == nil {
				return nil, fmt.Errorf("failed to create auth module")
			}
			return m, nil
		},
	}

	// Initialize each module with error handling
	for name, initializer := range moduleInitializers {
		init.Logger.Info("Initializing module", zap.String("module", name))

		module, err := initializer()
		if err != nil {
			init.Logger.Error("Module initialization failed",
				zap.String("module", name),
				zap.Error(err))
			continue
		}

		// Perform module migration
		if err := module.Migrate(); err != nil {
			init.Logger.Error("Module migration failed",
				zap.String("module", name),
				zap.Error(err))

			if init.EventService != nil {
				init.EventService.Track(ctx, event.EventOptions{
					Type:        "system",
					Category:    "startup",
					Actor:       "system",
					Target:      name,
					Action:      "migrate",
					Status:      "failed",
					Description: "Module migration failed",
					Metadata: map[string]interface{}{
						"error":  err.Error(),
						"module": name,
					},
				})
			}
			continue
		}

		modules[name] = module
		init.Logger.Info("Module initialized successfully",
			zap.String("module", name))
	}

	if init.EventService != nil {
		// Track successful module initialization
		init.EventService.Track(ctx, event.EventOptions{
			Type:        "system",
			Category:    "startup",
			Actor:       "system",
			Action:      "initialize_modules",
			Status:      "completed",
			Description: "Core modules initialization completed",
			Metadata: map[string]interface{}{
				"modules": getModuleNames(modules),
			},
		})
	}

	return modules
}

// Helper function to get module names
func getModuleNames(modules map[string]module.Module) []string {
	names := make([]string, 0, len(modules))
	for name := range modules {
		names = append(names, name)
	}
	return names
}
