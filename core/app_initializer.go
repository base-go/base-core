package core

import (
	"fmt"

	"base/core/app"
	"base/core/middleware"
	"base/core/module"

	"github.com/gin-gonic/gin"
)

// AppRoutes contains the different route groups for the application
type AppRoutes struct {
	API       *gin.RouterGroup
	Protected *gin.RouterGroup
	Auth      *gin.RouterGroup
	Public    *gin.RouterGroup
}

// AppInfrastructure contains the initialized app infrastructure
type AppInfrastructure struct {
	Routes      *AppRoutes
	CoreModules []module.Module
}

// InitializeAppInfrastructure sets up the application infrastructure on top of core
func InitializeAppInfrastructure(coreApp *CoreApplication) (*AppInfrastructure, error) {
	coreApp.Logger.Info("🏗️  Initializing App Infrastructure")

	// Setup API routes with middleware
	apiGroup := coreApp.Router.Group("/api")
	apiGroup.Use(middleware.APIKeyMiddleware())
	apiGroup.Use(middleware.HeaderMiddleware(coreApp.Logger))

	// Protected routes require authentication
	protectedGroup := apiGroup.Group("/")
	protectedGroup.Use(middleware.AuthMiddleware(coreApp.Logger))

	// Auth routes (login, register) - only require API key
	authGroup := apiGroup.Group("/auth")

	// Public routes (no authentication required)
	publicGroup := coreApp.Router.Group("/")

	routes := &AppRoutes{
		API:       apiGroup,
		Protected: protectedGroup,
		Auth:      authGroup,
		Public:    publicGroup,
	}

	// Initialize core modules using the new orchestration system
	coreModules, err := initializeCoreModules(routes, coreApp)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize core modules: %w", err)
	}

	infrastructure := &AppInfrastructure{
		Routes:      routes,
		CoreModules: coreModules,
	}

	return infrastructure, nil
}

// initializeCoreModules initializes core modules using the orchestration system
func initializeCoreModules(routes *AppRoutes, coreApp *CoreApplication) ([]module.Module, error) {
	// Create dependencies for core modules
	deps := module.Dependencies{
		DB:          coreApp.DB.DB,
		Router:      routes.Protected,
		Logger:      coreApp.Logger,
		Emitter:     coreApp.Emitter,
		Storage:     coreApp.Storage,
		EmailSender: coreApp.EmailSender,
		Config:      coreApp.Config,
	}

	// Create module initializer
	initializer := module.NewInitializer(coreApp.Logger)

	// Create core modules provider
	coreProvider := app.NewCoreModules()

	// Create core orchestrator
	coreOrchestrator := module.NewCoreOrchestrator(initializer, coreProvider, routes.Auth)

	// Initialize core modules
	return coreOrchestrator.InitializeCoreModules(deps)
}