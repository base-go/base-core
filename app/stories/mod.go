package stories

import (
	"base/core/module"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StoryModule struct {
	module.DefaultModule
	DB         *gorm.DB
	Controller *StoryController
	Service    *StoryService
}

func NewStoryModule(db *gorm.DB, router *gin.RouterGroup) module.Module {
	service := NewStoryService(db)
	controller := NewStoryController(service)

	storiesModule := &StoryModule{
		DB:         db,
		Controller: controller,
		Service:    service,
	}

	storiesModule.Routes(router)
	storiesModule.Migrate()

	return storiesModule
}

func (m *StoryModule) Routes(router *gin.RouterGroup) {
	m.Controller.Routes(router)
}

func (m *StoryModule) Migrate() error {
	return m.DB.AutoMigrate(&Story{})
}