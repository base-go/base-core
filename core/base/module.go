package base

import (
	"base/core/config"
	"base/core/email"
	"base/core/emitter"
	"base/core/logger"
	"base/core/module"
	"base/core/storage"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module provides common functionality for all modules
type Module struct {
	module.DefaultModule
	Name        string
	DB          *gorm.DB
	Logger      logger.Logger
	Emitter     *emitter.Emitter
	Storage     *storage.ActiveStorage
	EmailSender email.Sender
	Config      *config.Config
}

// ModuleDependencies interface defines what dependencies a module needs
type ModuleDependencies interface {
	GetDB() *gorm.DB
	GetRouter() *gin.RouterGroup
	GetLogger() logger.Logger
	GetEmitter() *emitter.Emitter
	GetStorage() *storage.ActiveStorage
	GetEmailSender() email.Sender
	GetConfig() *config.Config
}

// NewModule creates a new Module instance
func NewModule(name string, deps ModuleDependencies) *Module {
	return &Module{
		Name:        name,
		DB:          deps.GetDB(),
		Logger:      deps.GetLogger(),
		Emitter:     deps.GetEmitter(),
		Storage:     deps.GetStorage(),
		EmailSender: deps.GetEmailSender(),
		Config:      deps.GetConfig(),
	}
}

// LogInfo logs an info message with module context
func (bm *Module) LogInfo(message string, fields ...logger.Field) {
	contextFields := append([]logger.Field{
		logger.String("module", bm.Name),
	}, fields...)
	
	bm.Logger.Info(message, contextFields...)
}

// LogError logs an error with module context
func (bm *Module) LogError(message string, fields ...logger.Field) {
	contextFields := append([]logger.Field{
		logger.String("module", bm.Name),
	}, fields...)
	
	bm.Logger.Error(message, contextFields...)
}

// LogWarn logs a warning with module context
func (bm *Module) LogWarn(message string, fields ...logger.Field) {
	contextFields := append([]logger.Field{
		logger.String("module", bm.Name),
	}, fields...)
	
	bm.Logger.Warn(message, contextFields...)
}

// EmitEvent emits an event with module context
func (bm *Module) EmitEvent(eventName string, data interface{}) {
	if bm.Emitter != nil {
		// Add module context to event data
		eventData := map[string]interface{}{
			"module": bm.Name,
			"data":   data,
		}
		bm.Emitter.Emit(eventName, eventData)
	}
}

// AutoMigrate performs database migration for given models
func (bm *Module) AutoMigrate(models ...interface{}) error {
	if len(models) == 0 {
		return nil
	}
	
	bm.LogInfo("Starting database migration", logger.Int("models_count", len(models)))
	
	err := bm.DB.AutoMigrate(models...)
	if err != nil {
		bm.LogError("Database migration failed", logger.String("error", err.Error()))
		return fmt.Errorf("failed to migrate %s module: %w", bm.Name, err)
	}
	
	bm.LogInfo("Database migration completed successfully")
	return nil
}

// RegisterRoutes is a helper method for route registration with logging
func (bm *Module) RegisterRoutes(router *gin.RouterGroup, routeSetup func(*gin.RouterGroup)) {
	bm.LogInfo("Registering module routes")
	routeSetup(router)
	bm.LogInfo("Module routes registered successfully")
}

// GetService returns a base service with module dependencies
func (bm *Module) GetService() *Service {
	return NewService(bm.DB, bm.Logger, bm.Emitter, bm.Storage)
}

// GetController returns a base controller with module dependencies
func (bm *Module) GetController() *Controller {
	return NewController(bm.Logger, bm.Storage)
}

// ValidateConfig validates module-specific configuration
func (bm *Module) ValidateConfig(validator func(*config.Config) error) error {
	if validator == nil {
		return nil
	}
	
	bm.LogInfo("Validating module configuration")
	
	if err := validator(bm.Config); err != nil {
		bm.LogError("Configuration validation failed", logger.String("error", err.Error()))
		return fmt.Errorf("invalid configuration for %s module: %w", bm.Name, err)
	}
	
	bm.LogInfo("Module configuration validated successfully")
	return nil
}

// SetupHooks sets up module event hooks
func (bm *Module) SetupHooks(hooks map[string][]func(interface{})) {
	if bm.Emitter == nil || len(hooks) == 0 {
		return
	}
	
	bm.LogInfo("Setting up module event hooks", logger.Int("hooks_count", len(hooks)))
	
	for eventName, handlers := range hooks {
		for _, handler := range handlers {
			bm.Emitter.On(eventName, handler)
		}
	}
	
	bm.LogInfo("Module event hooks setup completed")
}