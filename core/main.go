package core

import (
	"fmt"

	"base/app"
	"base/core/logger"
	"base/core/module"

	"github.com/gin-gonic/gin"
)

// Application represents the complete application
type Application struct {
	Core        *CoreApplication
	CoreModules []module.Module
	AppModules  []module.Module
}

// NewApplication creates and initializes the complete application
func NewApplication() (*Application, error) {
	// 1. Start core infrastructure
	coreApp, err := StartCore()
	if err != nil {
		return nil, fmt.Errorf("failed to start core infrastructure: %w", err)
	}

	// 2. Initialize app infrastructure and core modules
	infrastructure, err := InitializeAppInfrastructure(coreApp)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize app infrastructure: %w", err)
	}

	// 3. Initialize app modules
	appModules, err := initializeAppModules(infrastructure.Routes.Protected, coreApp)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize app modules: %w", err)
	}

	application := &Application{
		Core:        coreApp,
		CoreModules: infrastructure.CoreModules,
		AppModules:  appModules,
	}

	coreApp.Logger.Info("🎉 Application initialization complete",
		logger.Int("core_modules", len(infrastructure.CoreModules)),
		logger.Int("app_modules", len(appModules)))

	return application, nil
}

// Run starts the application server
func (app *Application) Run() error {
	return app.Core.Run()
}

// initializeAppModules initializes app modules using the orchestration system
func initializeAppModules(protectedRouter *gin.RouterGroup, coreApp *CoreApplication) ([]module.Module, error) {
	// Create dependencies for app modules
	deps := module.Dependencies{
		DB:          coreApp.DB.DB,
		Router:      protectedRouter,
		Logger:      coreApp.Logger,
		Emitter:     coreApp.Emitter,
		Storage:     coreApp.Storage,
		EmailSender: coreApp.EmailSender,
		Config:      coreApp.Config,
	}

	// Create module initializer
	initializer := module.NewInitializer(coreApp.Logger)

	// Create app modules provider
	appProvider := app.NewAppModules()

	// Create app orchestrator
	appOrchestrator := module.NewAppOrchestrator(initializer, appProvider)

	// Initialize app modules
	return appOrchestrator.InitializeAppModules(deps)
}
