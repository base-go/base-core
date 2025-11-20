package module

import (
	"fmt"

	"base/core/logger"
	"base/core/router"
)

// CoreModuleProvider defines the interface for providing core modules
type CoreModuleProvider interface {
	GetCoreModules(deps Dependencies) map[string]Module
}

// CoreOrchestrator handles the orchestration of core modules
type CoreOrchestrator struct {
	initializer *Initializer
	provider    CoreModuleProvider
}

// NewCoreOrchestrator creates a new core module orchestrator
func NewCoreOrchestrator(initializer *Initializer, provider CoreModuleProvider) *CoreOrchestrator {
	return &CoreOrchestrator{
		initializer: initializer,
		provider:    provider,
	}
}

// InitializeCoreModules initializes all core modules using the provider
func (co *CoreOrchestrator) InitializeCoreModules(deps Dependencies) ([]Module, error) {
	deps.Logger.Info("🏗️  Starting core modules initialization")

	// Get the modules from the provider (from core/app/init.go)
	modules := co.provider.GetCoreModules(deps)

	if len(modules) == 0 {
		deps.Logger.Info("No core modules to initialize")
		return []Module{}, nil
	}

	// Initialize them using a custom core initializer that handles auth routing
	initializedModules := co.initializeCoreModules(modules, deps)

	deps.Logger.Info(fmt.Sprintf("✅ Core modules initialization complete (%d modules)", len(initializedModules)))
	return initializedModules, nil
}

// initializeCoreModules initializes core modules with special handling for auth modules
func (co *CoreOrchestrator) initializeCoreModules(modules map[string]Module, deps Dependencies) []Module {
	var initializedModules []Module

	for name, mod := range modules {
		deps.Logger.Info("Initializing core module", logger.String("module", name))

		// Register module
		if err := RegisterModule(name, mod); err != nil {
			deps.Logger.Error("Failed to register core module",
				logger.String("module", name),
				logger.String("error", err.Error()))
			continue
		}

		// Initialize
		if initModule, ok := mod.(interface{ Init() error }); ok {
			if err := initModule.Init(); err != nil {
				deps.Logger.Error("Failed to initialize core module",
					logger.String("module", name),
					logger.String("error", err.Error()))
				continue
			}
		}

		// Migrate
		if migrator, ok := mod.(interface{ Migrate() error }); ok {
			if err := migrator.Migrate(); err != nil {
				deps.Logger.Error("Failed to migrate core module",
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
		deps.Logger.Info("Core module initialized successfully", logger.String("module", name))
	}

	return initializedModules
}
