package auth

import (
	"base/core/email"
	"base/core/emitter"
	"base/core/layout"
	"base/core/logger"
	"base/core/module"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthModule struct {
	module.DefaultModule
	DB            *gorm.DB
	Controller    *AuthController
	ApiController *ApiAuthController
	Service       *AuthService
	Logger        logger.Logger
	EmailSender   email.Sender
	Emitter       *emitter.Emitter
}

func NewAuthModule(db *gorm.DB, webRouter, apiRouter *gin.RouterGroup, emailSender email.Sender, logger logger.Logger, emitter *emitter.Emitter, layoutEngine *layout.Engine) module.Module {
	service := NewAuthService(db, emailSender, emitter)
	controller := NewAuthController(service, emailSender, logger, layoutEngine)
	apiController := NewApiAuthController(service, emailSender, logger)

	authModule := &AuthModule{
		DB:            db,
		Controller:    controller,
		ApiController: apiController,
		Service:       service,
		Logger:        logger,
		EmailSender:   emailSender,
		Emitter:       emitter,
	}

	return authModule
}

// Routes implements the standard module interface for web routes
func (m *AuthModule) Routes(webRouter *gin.RouterGroup) {
	// Setup web routes for auth module
	m.Controller.Routes(webRouter)
}

func (m *AuthModule) ApiRoutes(apiRouter *gin.RouterGroup) {
	m.ApiController.Routes(apiRouter)
}

func (m *AuthModule) WebRoutes(webRouter *gin.RouterGroup) {
	m.Routes(webRouter)
}

func (m *AuthModule) Migrate() error {
	return m.DB.AutoMigrate(&AuthUser{})
}

func (m *AuthModule) GetModels() []any {
	return []any{
		&AuthUser{},
	}
}
