package home

import (
	"base/core/layout"
	"base/core/logger"
	"base/core/module"

	"github.com/gin-gonic/gin"
)

type HomeModule struct {
	module.DefaultModule
	Controller *HomeController
	Logger     logger.Logger
}

func NewHomeModule(layoutEngine *layout.Engine, logger logger.Logger) *HomeModule {
	return &HomeModule{
		Controller: NewHomeController(layoutEngine),
		Logger:     logger,
	}
}

// Init implements the Module interface
func (m *HomeModule) Init() error {
	m.Logger.Info("Initializing home module")
	return nil
}

// Migrate implements the Module interface
func (m *HomeModule) Migrate() error {
	// Home module doesn't need database migrations
	return nil
}

// GetModels implements the Module interface
func (m *HomeModule) GetModels() []any {
	// Home module doesn't have database models
	return nil
}

// Routes sets up the module routes
func (m *HomeModule) Routes(router *gin.RouterGroup) {
	m.Logger.Info("Setting up home module routes")

	// Use the controller's Routes method for consistency
	m.Controller.Routes(router)

	m.Logger.Info("Home module routes registered")
}

// GetController returns the home controller for external use
func (m *HomeModule) GetController() *HomeController {
	return m.Controller
}
