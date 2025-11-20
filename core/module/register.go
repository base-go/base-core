package module

import (
	"fmt"
	"maps"
	"reflect"
	"sync"

	"base/core/router"
	"gorm.io/gorm"
)

// Module defines the common interface that all modules must implement.
type Module interface {
	Init() error
	Migrate() error
	GetModels() []any
	Routes(*router.RouterGroup)
}

// DefaultModule provides a default implementation for the Module interface.
type DefaultModule struct{}

// Translatable is an interface that modules can implement to define translatable fields
type Translatable interface {
	TranslatedFields() []string
}

func (DefaultModule) Init() error {
	return nil // Default implementation does nothing
}

func (DefaultModule) Migrate() error {
	return nil // Default implementation does nothing
}

func (DefaultModule) Routes(router *router.RouterGroup) {
	// Default implementation does nothing
}
func (DefaultModule) GetModels() []any {
	return nil
}

// Seeder is an interface that modules can implement to seed the database.
type Seeder interface {
	Seed(*gorm.DB) error
}

// ModuleFactory is a function that creates a module with dependencies
type ModuleFactory func(deps Dependencies) Module

var (
	// modulesRegistry stores all registered modules. The key is the module name.
	modulesRegistry = make(map[string]Module)

	// globalAppModules stores factory functions for app modules (used by auto-discovery)
	globalAppModules = make(map[string]ModuleFactory)

	lock     sync.RWMutex
	globalMu sync.RWMutex
)

// RegisterModule registers a module under a unique name. It returns an error
// if the module is already registered under that name.
func RegisterModule(name string, module Module) error {
	lock.Lock()
	defer lock.Unlock()
	if _, exists := modulesRegistry[name]; exists {
		return fmt.Errorf("error: Module already registered: %s", name)
	}
	modulesRegistry[name] = module
	fmt.Printf("Successfully registered module: %s\n", name)
	return nil
}

// GetModule retrieves a module by its name.
func GetModule(name string) (Module, error) {
	lock.RLock()
	defer lock.RUnlock()
	module, exists := modulesRegistry[name]
	if !exists {
		return nil, fmt.Errorf("error: Module not found: %s", name)
	}
	return module, nil
}

// GetAllModules retrieves a copy of the registry map, protecting it from modifications.
func GetAllModules() map[string]Module {
	lock.RLock()
	defer lock.RUnlock()
	copy := make(map[string]Module, len(modulesRegistry))
	maps.Copy(copy, modulesRegistry)
	return copy
}

// HasMethod checks if a method is implemented by a module.
func HasMethod(module Module, methodName string) bool {
	moduleType := reflect.TypeOf(module)
	_, exists := moduleType.MethodByName(methodName)
	return exists
}

// RegisterAppModule registers a module factory for auto-discovery
// This should be called from the module's init() function
func RegisterAppModule(name string, factory ModuleFactory) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalAppModules[name] = factory
	fmt.Printf("Successfully registered app module factory: %s\n", name)
}

// GetAppModule retrieves a registered module factory
func GetAppModule(name string) ModuleFactory {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalAppModules[name]
}

// GetAllAppModules returns all registered app module factories
func GetAllAppModules() map[string]ModuleFactory {
	globalMu.RLock()
	defer globalMu.RUnlock()

	copy := make(map[string]ModuleFactory)
	for k, v := range globalAppModules {
		copy[k] = v
	}
	return copy
}
