package module

import (
	"base/core/config"
	"base/core/email"
	"base/core/emitter"
	"base/core/logger"
	"base/core/router"
	"base/core/storage"

	"gorm.io/gorm"
)

// Dependencies contains all dependencies that can be injected into modules
type Dependencies struct {
	DB          *gorm.DB
	Router      *router.RouterGroup
	Logger      logger.Logger
	Emitter     *emitter.Emitter
	Storage     *storage.ActiveStorage
	EmailSender email.Sender
	Config      *config.Config
}

// Initializer handles module initialization logic
type Initializer struct {
	logger logger.Logger
}

// NewInitializer creates a new module initializer
func NewInitializer(logger logger.Logger) *Initializer {
	return &Initializer{
		logger: logger,
	}
}

// Initialize initializes a map of modules with dependencies
func (mi *Initializer) Initialize(modules map[string]Module, deps Dependencies) []Module {
	var initializedModules []Module

	for name, mod := range modules {
		mi.logger.Info("Initializing module", logger.String("module", name))

		// Register module
		if err := RegisterModule(name, mod); err != nil {
			mi.logger.Error("Failed to register module",
				logger.String("module", name),
				logger.String("error", err.Error()))
			continue
		}

		// Initialize
		if initModule, ok := mod.(interface{ Init() error }); ok {
			if err := initModule.Init(); err != nil {
				mi.logger.Error("Failed to initialize module",
					logger.String("module", name),
					logger.String("error", err.Error()))
				continue
			}
		}

		// Migrate
		if migrator, ok := mod.(interface{ Migrate() error }); ok {
			if err := migrator.Migrate(); err != nil {
				mi.logger.Error("Failed to migrate module",
					logger.String("module", name),
					logger.String("error", err.Error()))
				continue
			}
		}

		// Setup routes
		if routeModule, ok := mod.(interface{ Routes(*router.RouterGroup) }); ok {
			routeModule.Routes(deps.Router)
		}

		initializedModules = append(initializedModules, mod)
		mi.logger.Info("Module initialized successfully", logger.String("module", name))
	}

	return initializedModules
}
