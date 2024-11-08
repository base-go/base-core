package app

import (
	"base/core/emitter"
	"base/core/module"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AppModuleInitializer struct {
	Router  *gin.RouterGroup
	Logger  *zap.Logger
	Emitter *emitter.Emitter
}

func (a *AppModuleInitializer) InitializeModules(db *gorm.DB) []module.Module {
	modules := make([]module.Module, 0)

	// Define the module initializers directly
	moduleInitializers := map[string]func(*gorm.DB, *gin.RouterGroup) module.Module{
		// Example:
		// "user": user.NewUserModule,
		// "product": product.NewProductModule,

		// MODULE_INITIALIZER_MARKER - Do not remove this comment because it's used by the CLI to add new module initializers
	}

	// Initialize and register each module
	for name, initializer := range moduleInitializers {
		mod := initializer(db, a.Router)
		if err := module.RegisterModule(name, mod); err != nil {
			// Handle error (e.g., log it)
			continue
		}
		if err := mod.Init(); err != nil {
			// Handle initialization error
			continue
		}
		if err := mod.Migrate(); err != nil {
			// Handle migration error
			continue
		}
		modules = append(modules, mod)
	}

	return modules
}
