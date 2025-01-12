package users

import (
	"base/core/module"
	"base/core/storage"

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
	ActiveStorage *storage.ActiveStorage
}

func NewUserModule(
	db *gorm.DB,
	router *gin.RouterGroup,
	logger *zap.Logger,
	activeStorage *storage.ActiveStorage,
) module.Module {
	// Initialize service with active storage
	service := NewUserService(db, logger, activeStorage)

	// Initialize controller
	controller := NewUserController(service, logger)

	usersModule := &UserModule{
		DB:            db,
		Controller:    controller,
		Service:       service,
		Logger:        logger,
		ActiveStorage: activeStorage,
	}

	// Set up routes
	usersModule.Routes(router)

	// Perform database migration
	if err := usersModule.Migrate(); err != nil {
		logger.Error("Failed to migrate user module",
			zap.Error(err))
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
