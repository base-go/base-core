package profile

import (
	"base/core/logger"
	"base/core/module"
	"base/core/router"
	"base/core/storage"

	"gorm.io/gorm"
)

type UserModule struct {
	module.DefaultModule
	DB            *gorm.DB
	Controller    *ProfileController
	Service       *ProfileService
	Logger        logger.Logger
	ActiveStorage *storage.ActiveStorage
}

func NewUserModule(
	db *gorm.DB,
	router *router.RouterGroup,
	logger logger.Logger,
	activeStorage *storage.ActiveStorage,
) module.Module {
	// Initialize service with active storage
	service := NewProfileService(db, logger, activeStorage)
	controller := NewProfileController(service, logger)

	usersModule := &UserModule{
		DB:            db,
		Controller:    controller,
		Service:       service,
		Logger:        logger,
		ActiveStorage: activeStorage,
	}

	return usersModule
}

func (m *UserModule) Routes(router *router.RouterGroup) {
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

func (m *UserModule) GetModels() []any {
	return []any{
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
