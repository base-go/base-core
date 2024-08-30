package characters

import (
	"base/core/module"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CharacterModule struct {
	module.DefaultModule
	DB         *gorm.DB
	Controller *CharacterController
	Service    *CharacterService
}

func NewCharacterModule(db *gorm.DB, router *gin.RouterGroup) module.Module {
	service := NewCharacterService(db)
	controller := NewCharacterController(service)

	charactersModule := &CharacterModule{
		DB:         db,
		Controller: controller,
		Service:    service,
	}

	charactersModule.Routes(router)
	charactersModule.Migrate()

	return charactersModule
}

func (m *CharacterModule) Routes(router *gin.RouterGroup) {
	m.Controller.Routes(router)
}

func (m *CharacterModule) Migrate() error {
	return m.DB.AutoMigrate(&Character{})
}