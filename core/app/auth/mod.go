package auth

import (
	"base/core/email"
	"base/core/module"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AuthModule struct {
	module.DefaultModule
	DB         *gorm.DB
	Controller *AuthController
	Service    *AuthService
}

func NewAuthModule(db *gorm.DB, router *gin.RouterGroup, emailSender email.Sender, logger *logrus.Logger) module.Module {
	service := NewAuthService(db)
	controller := NewAuthController(service, emailSender, logger)

	authModule := &AuthModule{
		DB:         db,
		Controller: controller,
		Service:    service,
	}

	authModule.Routes(router)
	authModule.Migrate()

	return authModule
}

func (m *AuthModule) Routes(router *gin.RouterGroup) {
	authGroup := router.Group("/auth")
	m.Controller.Routes(authGroup)
}

func (m *AuthModule) Migrate() error {
	return m.DB.AutoMigrate(&User{})
}

func (m *AuthModule) GetModels() []interface{} {
	return []interface{}{
		&User{},
	}
}
