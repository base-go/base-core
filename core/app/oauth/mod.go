package oauth

import (
	"base/core/module"
	"base/core/storage"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

func NewOAuthModule(db *gorm.DB, router *gin.RouterGroup, logger *logrus.Logger, activeStorage *storage.ActiveStorage) module.Module {
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

	oauthModule.Routes(router)
	return oauthModule
}

func (m *OAuthModule) Routes(router *gin.RouterGroup) {
	oauthGroup := router.Group("/oauth")
	m.Controller.Routes(oauthGroup)
}

func (m *OAuthModule) Migrate() error {
	return m.DB.AutoMigrate(&AuthProvider{})
}

func (m *OAuthModule) GetModels() []interface{} {
	return []interface{}{
		&AuthProvider{},
	}
}
