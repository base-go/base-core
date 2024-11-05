package users

import (
	"base/core/event"
	"base/core/module"
	"base/core/storage"
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserModule struct {
	module.DefaultModule
	DB            *gorm.DB
	Controller    *UserController
	Service       *UserService
	Logger        *zap.Logger
	EventService  *event.EventService
	ActiveStorage *storage.ActiveStorage
}

func NewUserModule(
	db *gorm.DB,
	router *gin.RouterGroup,
	logger *zap.Logger,
	eventService *event.EventService,
	activeStorage *storage.ActiveStorage,
) module.Module {
	// Initialize service with active storage
	service := NewUserService(db, logger, activeStorage)

	// Initialize controller with event tracking
	controller := NewUserController(service, logger, eventService)

	usersModule := &UserModule{
		DB:            db,
		Controller:    controller,
		Service:       service,
		Logger:        logger,
		EventService:  eventService,
		ActiveStorage: activeStorage,
	}

	// Set up routes
	usersModule.Routes(router)

	// Perform database migration
	if err := usersModule.Migrate(); err != nil {
		logger.Error("Failed to migrate user module",
			zap.Error(err))
		// Track only critical failures
		eventService.Track(context.Background(), event.EventOptions{
			Type:        "system_event",
			Category:    "migration",
			Actor:       "system",
			Target:      "user_module",
			Action:      "migrate",
			Status:      "failed",
			Description: "Failed to migrate user module",
			Metadata: map[string]interface{}{
				"error": err.Error(),
			},
		})
	}

	return usersModule
}

func (m *UserModule) Routes(router *gin.RouterGroup) {
	m.Controller.Routes(router)
}

func (m *UserModule) Migrate() error {
	err := m.DB.AutoMigrate(&User{})
	if err != nil {
		m.Logger.Error("Migration failed", zap.Error(err))
		return err
	}
	return nil
}

func (m *UserModule) GetModels() []interface{} {
	return []interface{}{
		&User{},
	}
}

func (m *UserModule) GetModelNames() []string {
	models := m.GetModels()
	names := make([]string, len(models))
	for i, model := range models {
		names[i] = m.DB.Model(model).Statement.Table
	}
	return names
}
