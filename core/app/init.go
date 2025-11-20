package app

import (
	"base/core/app/authentication"
	"base/core/app/authorization"
	"base/core/app/media"
	"base/core/app/oauth"
	"base/core/app/profile"
	"base/core/module"
	"base/core/scheduler"
	"base/core/translation"
)

// CoreModules implements module.CoreModuleProvider interface
type CoreModules struct{}

// GetCoreModules returns the list of core modules to initialize
// This is the only function that needs to be updated when adding new core modules
func (cm *CoreModules) GetCoreModules(deps module.Dependencies) map[string]module.Module {
	modules := make(map[string]module.Module)

	// Core modules - essential system functionality
	modules["users"] = profile.NewUserModule(
		deps.DB,
		deps.Router,
		deps.Logger,
		deps.Storage,
	)

	modules["media"] = media.NewMediaModule(
		deps.DB,
		deps.Router,
		deps.Storage,
		deps.Emitter,
		deps.Logger,
	)

	modules["authentication"] = authentication.NewAuthenticationModule(
		deps.DB,
		deps.Router, // Will be handled by orchestrator to use AuthRouter
		deps.EmailSender,
		deps.Logger,
		deps.Emitter,
	)

	modules["oauth"] = oauth.NewOAuthModule(
		deps.DB,
		deps.Router,
		deps.Logger,
		deps.Storage,
	)

	modules["authorization"] = authorization.NewAuthorizationModule(
		deps.DB,
		deps.Router, // Will be handled by orchestrator to use AuthRouter
		deps.Logger,
	)

	modules["translation"] = translation.NewTranslationModule(
		deps.DB,
		deps.Router,
		deps.Logger,
		deps.Emitter,
		deps.Storage,
	)

	modules["scheduler"] = scheduler.NewSchedulerModule(
		deps.DB,
		deps.Router,
		deps.Logger,
		deps.Emitter,
	)

	return modules
}

// NewCoreModules creates a new core modules provider
func NewCoreModules() *CoreModules {
	return &CoreModules{}
}
