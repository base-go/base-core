package core

import (
	"base/core/app/auth"
	"base/core/app/users"
	"base/core/email"
	"base/core/emitter"
	"base/core/logger"
	"base/core/module"
	"base/core/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ModuleInitializer is a struct to hold dependencies for module initialization
type ModuleInitializer struct {
	DB          *gorm.DB
	Router      *gin.RouterGroup
	EmailSender email.Sender
	Logger      logger.Logger
	Storage     *storage.ActiveStorage
	Emitter     *emitter.Emitter
}

// InitializeCoreModules loads and initializes all core modules
func InitializeCoreModules(init ModuleInitializer) map[string]module.Module {
	modules := make(map[string]module.Module)

	// Initialize auth module
	authModule := auth.NewAuthModule(
		init.DB,
		init.Router,
		init.EmailSender,
		init.Logger.GetZapLogger(),
		init.Emitter,
	)
	modules["auth"] = authModule

	// Initialize users module
	usersModule := users.NewUserModule(
		init.DB,
		init.Router,
		init.Logger.GetZapLogger(),
		init.Storage,
	)
	modules["users"] = usersModule

	return modules
}
