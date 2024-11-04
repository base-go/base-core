package auth

import (
	"base/core/email"
	"base/core/event"
	"base/core/module"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthModule struct {
	module.DefaultModule
	DB          *gorm.DB
	Controller  *AuthController
	Service     *AuthService
	EmailSender email.Sender
}

func NewAuthModule(db *gorm.DB, router *gin.RouterGroup, emailSender email.Sender, logger *zap.Logger, eventService *event.EventService) module.Module {
	service := NewAuthService(db, emailSender)
	controller := NewAuthController(service, emailSender, logger, eventService)

	authModule := &AuthModule{
		DB:          db,
		Controller:  controller,
		Service:     service,
		EmailSender: emailSender,
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
	return m.DB.AutoMigrate(&AuthUser{})
}

func (m *AuthModule) GetModels() []interface{} {
	return []interface{}{
		&AuthUser{},
	}
}
