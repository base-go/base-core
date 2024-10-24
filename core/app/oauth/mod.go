package oauth

import (
	"base/core/module"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type OAuthModule struct {
	module.DefaultModule
	DB         *gorm.DB
	Controller *OAuthController
	Service    *OAuthService
	Config     *OAuthConfig
}

func NewOAuthModule(db *gorm.DB, router *gin.RouterGroup, logger *logrus.Logger) module.Module {
	config := LoadConfig()
	ValidateConfig(config)

	service := NewOAuthService(db, config)
	controller := NewOAuthController(service, logger, config)

	oauthModule := &OAuthModule{
		DB:         db,
		Controller: controller,
		Service:    service,
		Config:     config,
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
