package auth

import (
	"base/core/email"
	"base/core/emitter"
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
	Logger      *zap.Logger
	EmailSender email.Sender
	Emitter     *emitter.Emitter
}

func NewAuthModule(db *gorm.DB, router *gin.RouterGroup, emailSender email.Sender, logger *zap.Logger, emitter *emitter.Emitter) module.Module {
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
	authModule.Migrate()
	authModule.Routes(router)
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
