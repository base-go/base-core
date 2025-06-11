package posts

import (
	"base/app/models"
	"base/core/emitter"
	"base/core/layout"
	"base/core/logger"
	"base/core/module"
	"base/core/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module struct {
	module.DefaultModule
	DB            *gorm.DB
	Controller    *PostController
	ApiController *PostApiController
	Service       *PostService
	Logger        *logger.Logger
	Storage       *storage.ActiveStorage
	Template      *layout.Engine
}

func NewPostModule(db *gorm.DB, webRouter *gin.RouterGroup, apiRouter *gin.RouterGroup, log logger.Logger, emitter *emitter.Emitter, storage *storage.ActiveStorage, templateEngine *layout.Engine) module.Module {

	service := NewPostService(db, emitter, storage, log)
	controller := NewPostController(service, storage, templateEngine)
	apiController := NewPostApiController(service, storage)
	m := &Module{
		DB:            db,
		Service:       service,
		Controller:    controller,
		ApiController: apiController,
		Logger:        &log,
		Storage:       storage,
		Template:      templateEngine,
	}

	return m
}

func (m *Module) Routes(router *gin.RouterGroup) {
	// Setup Web routes for HTML pages
	m.Controller.Routes(router)
}

func (m *Module) APIRoutes(router *gin.RouterGroup) {
	// Setup API routes with /api prefix
	m.ApiController.Routes(router)
}

func (m *Module) Migrate() error {
	return m.DB.AutoMigrate(&models.Post{})
}

func (m *Module) GetModels() []any {
	return []any{&models.Post{}}
}
