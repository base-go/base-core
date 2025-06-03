package posts

import (
	"base/app/models"
	"base/core/emitter"
	"base/core/logger"
	"base/core/module"
	"base/core/storage"
	"base/core/template"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module struct {
	module.DefaultModule
	DB         *gorm.DB
	Controller *PostController
	Service    *PostService
	Logger     *logger.Logger
	Storage    *storage.ActiveStorage
	Template   *template.Engine
	APIRouter  *gin.RouterGroup
	WebRouter  *gin.RouterGroup
}

func NewPostModule(db *gorm.DB, apiRouter, webRouter *gin.RouterGroup, log logger.Logger, emitter *emitter.Emitter, storage *storage.ActiveStorage, templateEngine *template.Engine) module.Module {

	service := NewPostService(db, emitter, storage, log)
	controller := NewPostController(service, storage, templateEngine)

	m := &Module{
		DB:         db,
		Service:    service,
		Controller: controller,
		Logger:     &log,
		Storage:    storage,
		Template:   templateEngine,
		APIRouter:  apiRouter,
		WebRouter:  webRouter,
	}

	return m
}

func (m *Module) Routes(router *gin.RouterGroup) {
	// Setup both API and Web routes
	m.Controller.SetupAPIRoutes(m.APIRouter)
	m.Controller.SetupWebRoutes(m.WebRouter)
}

func (m *Module) Migrate() error {
	return m.DB.AutoMigrate(&models.Post{})
}

func (m *Module) GetModels() []interface{} {
	return []interface{}{&models.Post{}}
}
