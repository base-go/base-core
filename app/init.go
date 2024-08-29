package app

import (
	"log"

	"base/app/customers"
	"base/app/users"
	"base/core/module"

	"github.com/gin-gonic/gin"
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
	}

	// Initialize and register each module
	for name, initializer := range moduleInitializers {
		module := initializer(db, router)
		modules[name] = module
		log.Printf("Initialized module: %s\n", name)
	}

	return modules
}
