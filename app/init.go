package app

import (
	"base/app/auth"
	"base/app/chapters"
	"base/app/characters"
	"base/app/customers"
	"base/app/scenes"
	"base/app/stories"
	"base/app/users"
	"base/core/module"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InitializeModules loads and initializes all modules directly
func InitializeModules(db *gorm.DB, router *gin.RouterGroup) map[string]module.Module {
	modules := make(map[string]module.Module)

	// Define the module initializers directly
	moduleInitializers := map[string]func(*gorm.DB, *gin.RouterGroup) module.Module{
		"users": func(db *gorm.DB, router *gin.RouterGroup) module.Module { return users.NewUserModule(db, router) },
		"customers": func(db *gorm.DB, router *gin.RouterGroup) module.Module {
			return customers.NewCustomerModule(db, router)
		},
		"stories": func(db *gorm.DB, router *gin.RouterGroup) module.Module {
			return stories.NewStoryModule(db, router)
		},
		"chapters": func(db *gorm.DB, router *gin.RouterGroup) module.Module {
			return chapters.NewChapterModule(db, router)
		},
		"scenes": func(db *gorm.DB, router *gin.RouterGroup) module.Module { return scenes.NewSceneModule(db, router) },
		"characters": func(db *gorm.DB, router *gin.RouterGroup) module.Module {
			return characters.NewCharacterModule(db, router)
		},
		"auth": func(db *gorm.DB, router *gin.RouterGroup) module.Module { return auth.NewAuthModule(db, router) },
		// MODULE_INITIALIZER_MARKER

	}

	// Initialize and register each module
	for name, initializer := range moduleInitializers {
		module := initializer(db, router)
		modules[name] = module
		log.Info("Initialized module: %s", name)
	}

	return modules
}
