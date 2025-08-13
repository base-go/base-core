package oauth

import (
	"base/core/logger"
	"base/core/module"
	"base/core/router"
	"base/core/storage"

	"gorm.io/gorm"
)

type OAuthModule struct {
	module.DefaultModule
	DB            *gorm.DB
	Controller    *OAuthController
	Service       *OAuthService
	Config        *OAuthConfig
	ActiveStorage *storage.ActiveStorage
}

func NewOAuthModule(db *gorm.DB, router *router.RouterGroup, logger logger.Logger, activeStorage *storage.ActiveStorage) module.Module {
	config := LoadConfig()
	ValidateConfig(config)

	service := NewOAuthService(db, config, activeStorage)
	controller := NewOAuthController(service, logger, config)

	oauthModule := &OAuthModule{
		DB:            db,
		Controller:    controller,
		Service:       service,
		Config:        config,
		ActiveStorage: activeStorage,
	}

	return oauthModule
}

func (m *OAuthModule) Routes(router *router.RouterGroup) {
	oauthGroup := router.Group("/oauth")
	m.Controller.Routes(oauthGroup)
}

func (m *OAuthModule) Migrate() error {
	return m.DB.AutoMigrate(&AuthProvider{})
}

func (m *OAuthModule) GetModels() []any {
	return []any{
		&AuthProvider{},
	}
}
