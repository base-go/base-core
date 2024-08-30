package chapters

import (
	"base/core/module"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ChapterModule struct {
	module.DefaultModule
	DB         *gorm.DB
	Controller *ChapterController
	Service    *ChapterService
}

func NewChapterModule(db *gorm.DB, router *gin.RouterGroup) module.Module {
	service := NewChapterService(db)
	controller := NewChapterController(service)

	chaptersModule := &ChapterModule{
		DB:         db,
		Controller: controller,
		Service:    service,
	}

	chaptersModule.Routes(router)
	chaptersModule.Migrate()

	return chaptersModule
}

func (m *ChapterModule) Routes(router *gin.RouterGroup) {
	m.Controller.Routes(router)
}

func (m *ChapterModule) Migrate() error {
	return m.DB.AutoMigrate(&Chapter{})
}