package app

import (
	"base/core/app/auth"
	"base/core/app/media"
	"base/core/app/users"
	"base/core/email"
	"base/core/emitter"
	"base/core/logger"
	"base/core/module"
	"base/core/storage"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Core represents the core application with all its dependencies
type Core struct {
	DB          *gorm.DB
	Router      *gin.RouterGroup // Protected routes requiring auth
	AuthRouter  *gin.RouterGroup // Auth routes (login, register)
	EmailSender email.Sender
	Logger      logger.Logger
	Storage     *storage.ActiveStorage
	Emitter     *emitter.Emitter
	Modules     map[string]module.Module
}

// NewCore creates and initializes a new Core instance
func NewCore(
	db *gorm.DB,
	protectedRouter *gin.RouterGroup,
	authRouter *gin.RouterGroup,
	emailSender email.Sender,
	logger logger.Logger,
	storage *storage.ActiveStorage,
	emitter *emitter.Emitter,
) *Core {
	core := &Core{
		DB:          db,
		Router:      protectedRouter,
		AuthRouter:  authRouter,
		EmailSender: emailSender,
		Logger:      logger,
		Storage:     storage,
		Emitter:     emitter,
		Modules:     make(map[string]module.Module),
	}

	// Initialize modules
	core.Modules = core.initializeModules()
	return core
}

// initializeModules loads and initializes all core modules
func (core *Core) initializeModules() map[string]module.Module {
	modules := make(map[string]module.Module)

	if core.Emitter == nil {
		core.Logger.Error("Emitter is nil in initializeModules; cannot register event listeners")
	}

	core.Logger.Info("Starting core modules initialization")

	// Define module initializers
	moduleInitializers := map[string]func() module.Module{
		"users": func() module.Module {
			return users.NewUserModule(
				core.DB,
				core.Router,
				core.Logger,
				core.Storage,
			)
		},
		"media": func() module.Module {
			return media.NewMediaModule(
				core.DB,
				core.Router,
				core.Storage,
				core.Emitter,
				core.Logger,
			)
		},
		"auth": func() module.Module {
			return auth.NewAuthModule(
				core.DB,
				core.AuthRouter, // Use AuthRouter for auth routes
				core.EmailSender,
				core.Logger,
				core.Emitter,
			)
		},
	}

	// Initialize modules
	moduleMap := moduleInitializers

	// Register and initialize each module
	for name, initializer := range moduleMap {
		core.Logger.Info("Initializing module", zap.String("module", name))
		mod := initializer()
		if mod == nil {
			core.Logger.Error("Failed to initialize module", zap.String("module", name))
			continue
		}

		// Initialize the module
		if initializer, ok := mod.(interface{ Init() error }); ok {
			if err := initializer.Init(); err != nil {
				core.Logger.Error("Failed to initialize module",
					zap.String("module", name),
					zap.Error(err))
				continue
			}
		}

		// Migrate the module
		if migrator, ok := mod.(interface{ Migrate() error }); ok {
			if err := migrator.Migrate(); err != nil {
				core.Logger.Error("Failed to migrate module",
					zap.String("module", name),
					zap.Error(err))
				continue
			}
		}

		// Set up routes for the module
		if routeModule, ok := mod.(interface{ Routes(*gin.RouterGroup) }); ok {
			core.Logger.Info("Setting up routes for module", zap.String("module", name))
			// Use AuthRouter for auth module, protected Router for others
			if name == "auth" {
				routeModule.Routes(core.AuthRouter)
			} else {
				routeModule.Routes(core.Router)
			}
		}

		modules[name] = mod
		core.Logger.Info("Core module initialized", zap.String("module", name))
	}

	return modules
}

// runMigrations runs migrations for all modules that implement the Migrate interface
func (core *Core) runMigrations(modules map[string]module.Module) {
	for _, module := range modules {
		if migrator, ok := module.(interface{ Migrate() error }); ok {
			if err := migrator.Migrate(); err != nil {
				core.Logger.Error("Failed to migrate module",
					zap.String("error", err.Error()))
				continue
			}
		}
	}
}

// GetModuleNames returns a list of all initialized module names
func (core *Core) GetModuleNames() []string {
	names := make([]string, 0, len(core.Modules))
	for name := range core.Modules {
		names = append(names, name)
	}
	return names
}
