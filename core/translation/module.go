package translation

import (
	"base/core/emitter"
	"base/core/logger"
	"base/core/module"
	"base/core/router"
	"base/core/storage"

	"gorm.io/gorm"
)

type Module struct {
	module.DefaultModule
	DB         *gorm.DB
	Controller *TranslationController
	Service    *TranslationService
	Logger     logger.Logger
	Storage    *storage.ActiveStorage
}

func NewTranslationModule(db *gorm.DB, router *router.RouterGroup, log logger.Logger, emitter *emitter.Emitter, storage *storage.ActiveStorage) module.Module {
	service := NewTranslationService(db, emitter, storage, log)
	controller := NewTranslationController(service, storage)

	m := &Module{
		DB:         db,
		Service:    service,
		Controller: controller,
		Logger:     log,
		Storage:    storage,
	}

	return m
}

func (m *Module) Routes(router *router.RouterGroup) {
	m.Logger.Info("Registering Translation module routes")
	m.Controller.Routes(router)
	m.Logger.Info("Translation module routes registered")
}

func (m *Module) Migrate() error {
	return m.DB.AutoMigrate(&Translation{})
}

func (m *Module) GetModels() []any {
	return []any{&Translation{}}
}
