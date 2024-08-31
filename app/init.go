package app

import (
	"base/core/module"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InitializeModules loads and initializes all modules directly
func InitializeModules(db *gorm.DB, router *gin.RouterGroup) map[string]module.Module {
	modules := make(map[string]module.Module)

	// Define the module initializers directly
	moduleInitializers := map[string]func(*gorm.DB, *gin.RouterGroup) module.Module{

		// MODULE_INITIALIZER_MARKER - Do not remove this comment because it's used by the CLI to add new module initializers

	}

	// Initialize and register each module
	for name, initializer := range moduleInitializers {
		module := initializer(db, router)
		modules[name] = module
	}

	return modules
}
