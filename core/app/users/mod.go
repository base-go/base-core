package users

import (
	"base/core/module"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserModule struct {
	module.DefaultModule
	DB         *gorm.DB
	Controller *UserController
	Service    *UserService
}

func NewUserModule(db *gorm.DB, router *gin.RouterGroup) module.Module {
	service := NewUserService(db)
	controller := NewUserController(service)

	usersModule := &UserModule{
		DB:         db,
		Controller: controller,
		Service:    service,
	}

	usersModule.Routes(router)
	usersModule.Migrate()

	return usersModule
}

func (m *UserModule) Routes(router *gin.RouterGroup) {
	m.Controller.Routes(router)
}

func (m *UserModule) Migrate() error {
	return m.DB.AutoMigrate(&User{})
}

func (m *UserModule) GetModels() []interface{} {
	return []interface{}{
		&User{},
	}
}
