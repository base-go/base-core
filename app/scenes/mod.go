package scenes

import (
	"base/core/module"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SceneModule struct {
	module.DefaultModule
	DB         *gorm.DB
	Controller *SceneController
	Service    *SceneService
}

func NewSceneModule(db *gorm.DB, router *gin.RouterGroup) module.Module {
	service := NewSceneService(db)
	controller := NewSceneController(service)

	scenesModule := &SceneModule{
		DB:         db,
		Controller: controller,
		Service:    service,
	}

	scenesModule.Routes(router)
	scenesModule.Migrate()

	return scenesModule
}

func (m *SceneModule) Routes(router *gin.RouterGroup) {
	m.Controller.Routes(router)
}

func (m *SceneModule) Migrate() error {
	return m.DB.AutoMigrate(&Scene{})
}