package module

import (
	"fmt"
)

// AppModuleProvider defines the interface for providing app modules
type AppModuleProvider interface {
	GetAppModules(deps Dependencies) map[string]Module
}

// AppOrchestrator handles the orchestration of app modules
type AppOrchestrator struct {
	initializer *Initializer
	provider    AppModuleProvider
}

// NewAppOrchestrator creates a new app module orchestrator
func NewAppOrchestrator(initializer *Initializer, provider AppModuleProvider) *AppOrchestrator {
	return &AppOrchestrator{
		initializer: initializer,
		provider:    provider,
	}
}

// InitializeAppModules initializes all app modules using the provider
func (ao *AppOrchestrator) InitializeAppModules(deps Dependencies) ([]Module, error) {
	deps.Logger.Info("🚀 Starting app modules initialization")

	// Get the modules from the provider (from app/init.go)
	modules := ao.provider.GetAppModules(deps)

	if len(modules) == 0 {
		deps.Logger.Info("No app modules to initialize")
		return []Module{}, nil
	}

	// Initialize them using the generic initializer
	initializedModules := ao.initializer.Initialize(modules, deps)

	deps.Logger.Info(fmt.Sprintf("✅ App modules initialization complete (%d modules)", len(initializedModules)))
	return initializedModules, nil
}
