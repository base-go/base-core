package authentication

import (
	"base/core/email"
	"base/core/emitter"
	"base/core/logger"
	"base/core/module"
	"base/core/router"

	"gorm.io/gorm"
)

type AuthenticationModule struct {
	module.DefaultModule
	DB          *gorm.DB
	Controller  *AuthController
	Service     *AuthService
	Logger      logger.Logger
	EmailSender email.Sender
	Emitter     *emitter.Emitter
}

func NewAuthenticationModule(db *gorm.DB, router *router.RouterGroup, emailSender email.Sender, logger logger.Logger, emitter *emitter.Emitter) module.Module {
	service := NewAuthService(db, emailSender, emitter)
	controller := NewAuthController(service, emailSender, logger)

	authModule := &AuthenticationModule{
		DB:          db,
		Controller:  controller,
		Service:     service,
		Logger:      logger,
		EmailSender: emailSender,
		Emitter:     emitter,
	}

	return authModule
}

func (m *AuthenticationModule) Routes(router *router.RouterGroup) {
	// Create /auth group under /api (router is already /api from main.go)
	authGroup := router.Group("/auth")

	m.Controller.Routes(authGroup)
}

func (m *AuthenticationModule) Migrate() error {
	return m.DB.AutoMigrate(&AuthUser{})
}

func (m *AuthenticationModule) GetModels() []any {
	return []any{
		&AuthUser{},
	}
}
