package app

import (
	// MODULE_IMPORT_MARKER - Do not remove this comment because it's used by the CLI to add new module imports
	"base/app/category"
	"base/core/module"
)

// AppModules implements module.AppModuleProvider interface
type AppModules struct{}

// GetAppModules returns the list of app-specific modules to initialize
// This is the only function that needs to be updated when adding new app modules
func (am *AppModules) GetAppModules(deps module.Dependencies) map[string]module.Module {
	modules := make(map[string]module.Module)

	// Add app-specific modules here - the CLI will add them automatically
	// MODULE_INITIALIZER_MARKER - Do not remove this comment because it's used by the CLI to add new module initializers
	modules["category"] = category.NewModule(deps.DB)

	return modules
}

// NewAppModules creates a new app modules provider
func NewAppModules() *AppModules {
	return &AppModules{}
}
