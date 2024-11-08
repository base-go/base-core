package app

import (
	"base/core"
	"base/core/app/auth"
	"base/core/app/users"
	"base/core/email"
	"base/core/event"
	"base/core/module"
	"base/core/storage"
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CoreModuleInitializer holds dependencies for core module initialization
type CoreModuleInitializer struct {
	DB           *gorm.DB
	Router       *gin.RouterGroup
	EmailSender  email.Sender
	Logger       *zap.Logger
	EventService *event.EventService
}

// InitializeCoreModules loads and initializes all core modules
func InitializeCoreModules(db *gorm.DB, router *gin.RouterGroup, emailSender email.Sender, logger *zap.Logger, eventService *event.EventService) map[string]module.Module {
	modules := make(map[string]module.Module)
	ctx := context.Background()

	// Track module initialization start
	eventService.Track(ctx, event.EventOptions{
		Type:        "system_event",
		Category:    "initialization",
		Actor:       "system",
		ActorID:     "system",
		Target:      "core_modules",
		Action:      "initialize",
		Status:      "started",
		Description: "Starting core modules initialization",
	})

	// Define module initializers
	moduleInitializers := map[string]func() module.Module{
		"users": func() module.Module {
			return users.NewUserModule(
				db,
				router,
				logger,
				eventService,
				&storage.ActiveStorage{}, // Ensure this instance is correctly configured
			)
		},
		"auth": func() module.Module {
			return auth.NewAuthModule(
				db,
				router,
				emailSender,
				logger,
				eventService,
				core.Emitter,
			)
		},
	}

	// Initialize and register each module
	for name, initializer := range moduleInitializers {
		// Track individual module initialization
		eventService.Track(ctx, event.EventOptions{
			Type:        "system_event",
			Category:    "initialization",
			Actor:       "system",
			ActorID:     "system",
			Target:      "module",
			TargetID:    name,
			Action:      "initialize",
			Status:      "started",
			Description: "Initializing module: " + name,
		})

		module := initializer()
		modules[name] = module

		logger.Info("Core module initialized", zap.String("module", name))

		// Track successful module initialization
		eventService.Track(ctx, event.EventOptions{
			Type:        "system_event",
			Category:    "initialization",
			Actor:       "system",
			ActorID:     "system",
			Target:      "module",
			TargetID:    name,
			Action:      "initialize",
			Status:      "completed",
			Description: "Successfully initialized module: " + name,
		})
	}

	// Track module initialization completion
	eventService.Track(ctx, event.EventOptions{
		Type:        "system_event",
		Category:    "initialization",
		Actor:       "system",
		ActorID:     "system",
		Target:      "core_modules",
		Action:      "initialize",
		Status:      "completed",
		Description: "Core modules initialization completed",
		Metadata: map[string]interface{}{
			"module_count": len(modules),
			"modules":      getModuleNames(modules),
		},
	})

	return modules
}

// NewCoreModuleInitializer creates a new instance of CoreModuleInitializer
func NewCoreModuleInitializer(
	db *gorm.DB,
	router *gin.RouterGroup,
	emailSender email.Sender,
	logger *zap.Logger,
	eventService *event.EventService,
) *CoreModuleInitializer {
	return &CoreModuleInitializer{
		DB:           db,
		Router:       router,
		EmailSender:  emailSender,
		Logger:       logger,
		EventService: eventService,
	}
}

// getModuleNames extracts module names from the modules map
func getModuleNames(modules map[string]module.Module) []string {
	names := make([]string, 0, len(modules))
	for name := range modules {
		names = append(names, name)
	}
	return names
}

// Initialize initializes all core modules using the initializer's dependencies
func (cmi *CoreModuleInitializer) Initialize() map[string]module.Module {
	return InitializeCoreModules(
		cmi.DB,
		cmi.Router,
		cmi.EmailSender,
		cmi.Logger,
		cmi.EventService,
	)
}
