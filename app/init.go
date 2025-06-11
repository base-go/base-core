package app

import (
	// MODULE_IMPORT_MARKER - Do not remove this comment because it's used by the CLI to add new module imports
	"base/app/home"
	"base/app/posts"

	// Core modules are automatically initialized in core/initializer/initializer.go

	"base/core/initializer"
	"base/core/language"
	"base/core/layout"
	"base/core/module"
)

// Registry implements both ModuleRegistry and AppInterface
type Registry struct{}

// GetModules returns a map of module name to module instance
func (r *Registry) GetModules(deps *initializer.CoreDependencies) map[string]module.Module {
	// Only initialize app-specific modules here
	// Core modules are automatically initialized in core/initializer/initializer.go
	return map[string]module.Module{
		"home":  home.NewHomeModule(deps.Layout, deps.Logger),
		"posts": posts.NewPostModule(deps.DB, deps.WebRouter, deps.APIRouter, deps.Logger, deps.Emitter, deps.Storage, deps.Layout),
		// MODULE_INITIALIZER_MARKER - Do not remove this comment because it's used by the CLI to add new module initializers
	}
}

// RegisterTranslations implements AppInterface
func (r *Registry) RegisterTranslations(translations *language.TranslationService) error {
	return RegisterTranslations(translations)
}

// RegisterTemplates implements AppInterface
func (r *Registry) RegisterTemplates(layoutEngine *layout.Engine) error {
	return RegisterTemplates(layoutEngine)
}
