package app

import (
	"base/core/module"
	// MODULE_IMPORT_MARKER

)

// AppModules implements module.AppModuleProvider interface
type AppModules struct{}

// GetAppModules returns the list of app-specific modules to initialize
// This is the only function that needs to be updated when adding new app modules
func (am *AppModules) GetAppModules(deps module.Dependencies) map[string]module.Module {
	modules := make(map[string]module.Module)

	modules["categories"] = categories.NewCategoryModule(deps.DB)	modules["comments"] = comments.NewCommentModule(deps.DB)	// MODULE_INITIALIZER_MARKER
	return modules
}

// NewAppModules creates a new app modules provider
func NewAppModules() *AppModules {
	return &AppModules{}
}
