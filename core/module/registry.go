package module

import (
	"context"
	"fmt"
	"sync"
)

// Registry manages module registration and lifecycle
type Registry struct {
	modules map[string]Module
	mu      sync.RWMutex
	
	// Hook functions for module lifecycle events
	onRegister []func(name string, module Module)
	onInit     []func(name string, module Module, err error)
}

// NewRegistry creates a new module registry
func NewRegistry() *Registry {
	return &Registry{
		modules:    make(map[string]Module),
		onRegister: make([]func(string, Module), 0),
		onInit:     make([]func(string, Module, error), 0),
	}
}

// Register adds a module to the registry
func (r *Registry) Register(name string, module Module) error {
	if name == "" {
		return fmt.Errorf("module name cannot be empty")
	}
	if module == nil {
		return fmt.Errorf("module cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.modules[name]; exists {
		return fmt.Errorf("module %q already registered", name)
	}

	r.modules[name] = module

	// Call registration hooks
	for _, hook := range r.onRegister {
		hook(name, module)
	}

	return nil
}

// Unregister removes a module from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.modules[name]; !exists {
		return fmt.Errorf("module %q not found", name)
	}

	delete(r.modules, name)
	return nil
}

// Get retrieves a module by name
func (r *Registry) Get(name string) (Module, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	module, exists := r.modules[name]
	if !exists {
		return nil, fmt.Errorf("module %q not found", name)
	}

	return module, nil
}

// MustGet retrieves a module by name, panicking if not found
func (r *Registry) MustGet(name string) Module {
	module, err := r.Get(name)
	if err != nil {
		panic(err)
	}
	return module
}

// Has checks if a module is registered
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.modules[name]
	return exists
}

// List returns a list of all registered module names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.modules))
	for name := range r.modules {
		names = append(names, name)
	}
	return names
}

// All returns a copy of all registered modules
func (r *Registry) All() map[string]Module {
	r.mu.RLock()
	defer r.mu.RUnlock()

	copy := make(map[string]Module, len(r.modules))
	for name, module := range r.modules {
		copy[name] = module
	}
	return copy
}

// InitAll initializes all registered modules
func (r *Registry) InitAll(ctx context.Context) error {
	modules := r.All()
	
	for name, module := range modules {
		select {
		case <-ctx.Done():
			return fmt.Errorf("initialization cancelled: %w", ctx.Err())
		default:
			err := module.Init()
			
			// Call init hooks
			for _, hook := range r.onInit {
				hook(name, module, err)
			}
			
			if err != nil {
				return fmt.Errorf("failed to initialize module %q: %w", name, err)
			}
		}
	}
	
	return nil
}

// InitAllParallel initializes all modules in parallel
func (r *Registry) InitAllParallel(ctx context.Context) error {
	modules := r.All()
	
	type result struct {
		name string
		err  error
	}
	
	results := make(chan result, len(modules))
	var wg sync.WaitGroup
	
	for name, module := range modules {
		wg.Add(1)
		go func(n string, m Module) {
			defer wg.Done()
			
			select {
			case <-ctx.Done():
				results <- result{n, ctx.Err()}
				return
			default:
				err := m.Init()
				results <- result{n, err}
				
				// Call init hooks
				for _, hook := range r.onInit {
					hook(n, m, err)
				}
			}
		}(name, module)
	}
	
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Collect results
	var errs []error
	for res := range results {
		if res.err != nil {
			errs = append(errs, fmt.Errorf("module %q: %w", res.name, res.err))
		}
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("initialization failed for %d modules: %v", len(errs), errs)
	}
	
	return nil
}

// MigrateAll runs migrations for all registered modules
func (r *Registry) MigrateAll(ctx context.Context) error {
	modules := r.All()
	
	for name, module := range modules {
		select {
		case <-ctx.Done():
			return fmt.Errorf("migration cancelled: %w", ctx.Err())
		default:
			if err := module.Migrate(); err != nil {
				return fmt.Errorf("failed to migrate module %q: %w", name, err)
			}
		}
	}
	
	return nil
}

// OnRegister adds a hook that's called when a module is registered
func (r *Registry) OnRegister(hook func(name string, module Module)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onRegister = append(r.onRegister, hook)
}

// OnInit adds a hook that's called when a module is initialized
func (r *Registry) OnInit(hook func(name string, module Module, err error)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onInit = append(r.onInit, hook)
}

// Clear removes all modules from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.modules = make(map[string]Module)
}

// Count returns the number of registered modules
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.modules)
}

// Global registry and factory registration

// ModuleFactory is a function that creates a module with dependencies
type ModuleFactory func(deps Dependencies) Module

var (
	globalAppModules = make(map[string]ModuleFactory)
	globalMu         sync.RWMutex
)

// RegisterAppModule registers a module factory for auto-discovery
// This should be called from the module's init() function
func RegisterAppModule(name string, factory ModuleFactory) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalAppModules[name] = factory
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