package module

import (
	"fmt"
	"reflect"
	"sync"

	"gorm.io/gorm"
)

// Module defines the common interface that all modules must implement.
type Module interface {
	Init() error
	Migrate() error
	GetModels() []interface{}
}

// DefaultModule provides a default implementation for the Module interface.
type DefaultModule struct{}

func (DefaultModule) Init() error {
	return nil // Default implementation does nothing
}

func (DefaultModule) Migrate() error {
	return nil // Default implementation does nothing
}

func (DefaultModule) Routes() {
	// Default implementation does nothing
}
func (DefaultModule) GetModels() []interface{} {
	return nil
}

// Seeder is an interface that modules can implement to seed the database.
type Seeder interface {
	Seed(*gorm.DB) error
}

var (
	// modulesRegistry stores all registered modules. The key is the module name.
	modulesRegistry = make(map[string]Module)
	lock            sync.RWMutex
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
	for key, value := range modulesRegistry {
		copy[key] = value
	}
	return copy
}

// HasMethod checks if a method is implemented by a module.
func HasMethod(module Module, methodName string) bool {
	moduleType := reflect.TypeOf(module)
	_, exists := moduleType.MethodByName(methodName)
	return exists
}
