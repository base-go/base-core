package auth

import (
	"base/core/email"
	"base/core/emitter"
	"base/core/logger"
	"base/core/module"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthModule struct {
	module.DefaultModule
	DB          *gorm.DB
	Controller  *AuthController
	Service     *AuthService
	Logger      logger.Logger
	EmailSender email.Sender
	Emitter     *emitter.Emitter
}

func NewAuthModule(db *gorm.DB, router *gin.RouterGroup, emailSender email.Sender, logger logger.Logger, emitter *emitter.Emitter) module.Module {
	service := NewAuthService(db, emailSender, emitter)
	controller := NewAuthController(service, emailSender, logger)

	authModule := &AuthModule{
		DB:          db,
		Controller:  controller,
		Service:     service,
		Logger:      logger,
		EmailSender: emailSender,
		Emitter:     emitter,
	}

	return authModule
}

func (m *AuthModule) Routes(router *gin.RouterGroup) {
	// Router is already /api/auth from start.go
	m.Controller.Routes(router)
}

func (m *AuthModule) Migrate() error {
	return m.DB.AutoMigrate(&AuthUser{})
}

func (m *AuthModule) GetModels() []interface{} {
	return []interface{}{
		&AuthUser{},
	}
}
