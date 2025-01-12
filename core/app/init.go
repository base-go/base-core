package app

import (
	"base/core/app/auth"
	"base/core/app/users"
	"base/core/email"
	"base/core/emitter"
	"base/core/module"
	"base/core/storage"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CoreModuleInitializer holds dependencies for core module initialization
type CoreModuleInitializer struct {
	DB          *gorm.DB
	Router      *gin.RouterGroup
	EmailSender email.Sender
	Logger      *zap.Logger
	Storage     *storage.ActiveStorage
	Emitter     *emitter.Emitter
}

// InitializeCoreModules loads and initializes all core modules
func InitializeCoreModules(db *gorm.DB, router *gin.RouterGroup, emailSender email.Sender, logger *zap.Logger, storage *storage.ActiveStorage, emitter *emitter.Emitter) map[string]module.Module {
	modules := make(map[string]module.Module)

	// Check if emitter is nil
	if emitter == nil {
		logger.Error("Emitter is nil in InitializeCoreModules; cannot register event listeners")
		fmt.Println("Emitter is nil in InitializeCoreModules; cannot register event listeners")
	} else {
		fmt.Println("WE GOT EMITTER IN INITIALIZECOREMODULES")
		logger.Info("WE GOT EMITTER IN INITIALIZECOREMODULES")
	}

	// Log initialization start
	logger.Info("Starting core modules initialization")

	// Define module initializers
	moduleInitializers := map[string]func() module.Module{
		"users": func() module.Module {
			return users.NewUserModule(
				db,
				router,
				logger,
				storage,
			)
		},
		"auth": func() module.Module {
			return auth.NewAuthModule(
				db,
				router,
				emailSender,
				logger,
				emitter,
			)
		},
	}

	// Initialize and register each module
	for name, initializer := range moduleInitializers {
		logger.Info("Initializing module", zap.String("module", name))

		module := initializer()
		modules[name] = module

		logger.Info("Core module initialized", zap.String("module", name))
	}

	logger.Info("Core modules initialization completed", zap.Strings("modules", getModuleNames(modules)))
	return modules
}

// NewCoreModuleInitializer creates a new instance of CoreModuleInitializer
func NewCoreModuleInitializer(
	db *gorm.DB,
	router *gin.RouterGroup,
	emailSender email.Sender,
	logger *zap.Logger,
	storage *storage.ActiveStorage,
	emitter *emitter.Emitter,
) *CoreModuleInitializer {
	return &CoreModuleInitializer{
		DB:          db,
		Router:      router,
		EmailSender: emailSender,
		Logger:      logger,
		Storage:     storage,
		Emitter:     emitter,
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
func (i *CoreModuleInitializer) Initialize() map[string]module.Module {
	return InitializeCoreModules(
		i.DB,
		i.Router,
		i.EmailSender,
		i.Logger,
		i.Storage,
		i.Emitter,
	)
}
