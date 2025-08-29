package posts

import (
	"base/app/models"
	"base/core/module"
	"base/core/router"
	"base/core/translation"

	"gorm.io/gorm"
)

type Module struct {
	module.DefaultModule
	DB                *gorm.DB
	Service           *PostService
	Controller        *PostController
	TranslationHelper *translation.Helper
}

// Init creates and initializes the Post module with all dependencies
func Init(deps module.Dependencies) module.Module {
	// Create translation service and helper
	translationService := translation.NewTranslationService(deps.DB, deps.Emitter, deps.Storage, deps.Logger)
	translationHelper := translation.NewHelper(translationService)

	// Initialize service with translation helper
	service := NewPostService(deps.DB, deps.Emitter, deps.Storage, deps.Logger, translationHelper)
	controller := NewPostController(service, deps.Storage)

	// Create module
	mod := &Module{
		DB:                deps.DB,
		Service:           service,
		Controller:        controller,
		TranslationHelper: translationHelper,
	}

	return mod
}

// Routes registers the module routes
func (m *Module) Routes(router *router.RouterGroup) {
	m.Controller.Routes(router)
}

func (m *Module) Init() error {
	return nil
}

func (m *Module) Migrate() error {
	return m.DB.AutoMigrate(&models.Post{})
}

func (m *Module) GetModels() []any {
	return []any{
		&models.Post{},
	}
}
