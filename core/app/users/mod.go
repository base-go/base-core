package users

import (
	"base/core/logger"
	"base/core/module"
	"base/core/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserModule struct {
	module.DefaultModule
	DB            *gorm.DB
	Controller    *UserController
	Service       *UserService
	Logger        logger.Logger
	ActiveStorage *storage.ActiveStorage
}

func NewUserModule(
	db *gorm.DB,
	router *gin.RouterGroup,
	logger logger.Logger,
	activeStorage *storage.ActiveStorage,
) module.Module {
	// Initialize service with active storage
	service := NewUserService(db, logger, activeStorage)
	controller := NewUserController(service, logger)

	usersModule := &UserModule{
		DB:            db,
		Controller:    controller,
		Service:       service,
		Logger:        logger,
		ActiveStorage: activeStorage,
	}

	return usersModule
}

func (m *UserModule) Routes(router *gin.RouterGroup) {
	m.Controller.Routes(router)
}

func (m *UserModule) Migrate() error {
	err := m.DB.AutoMigrate(&User{})
	if err != nil {
		m.Logger.Error("Migration failed", logger.String("error", err.Error()))
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
