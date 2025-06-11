package router

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// Router provides a Rails-style routing interface over Gin
type Router struct {
	ginEngine *gin.Engine
	ginGroup  *gin.RouterGroup
	namespace string
}

// HandlerFunc represents a controller action
type HandlerFunc func(*gin.Context)

// Controller interface for Rails-style controllers
type Controller interface{}

// ResourceOptions defines which actions to include/exclude for resources
type ResourceOptions struct {
	Only   []string // Only include these actions
	Except []string // Exclude these actions
}

// New creates a new Rails-style router wrapping a Gin engine
func New(ginEngine *gin.Engine) *Router {
	return &Router{
		ginEngine: ginEngine,
		ginGroup:  &ginEngine.RouterGroup,
		namespace: "",
	}
}

// NewGroup creates a new router group (like Rails namespaces)
func (r *Router) NewGroup(path string) *Router {
	return &Router{
		ginEngine: r.ginEngine,
		ginGroup:  r.ginGroup.Group(path),
		namespace: r.namespace + path,
	}
}

// Use adds middleware to the router (chainable)
func (r *Router) Use(middleware ...gin.HandlerFunc) *Router {
	r.ginGroup.Use(middleware...)
	return r
}

// Basic HTTP methods with Rails-style naming
func (r *Router) Get(path string, handler HandlerFunc) *Router {
	r.ginGroup.GET(path, gin.HandlerFunc(handler))
	return r
}

func (r *Router) Post(path string, handler HandlerFunc) *Router {
	r.ginGroup.POST(path, gin.HandlerFunc(handler))
	return r
}

func (r *Router) Put(path string, handler HandlerFunc) *Router {
	r.ginGroup.PUT(path, gin.HandlerFunc(handler))
	return r
}

func (r *Router) Patch(path string, handler HandlerFunc) *Router {
	r.ginGroup.PATCH(path, gin.HandlerFunc(handler))
	return r
}

func (r *Router) Delete(path string, handler HandlerFunc) *Router {
	r.ginGroup.DELETE(path, gin.HandlerFunc(handler))
	return r
}

// Resources creates RESTful routes for a controller (Rails-style)
func (r *Router) Resources(path string, controller Controller, opts ...ResourceOptions) *Router {
	// Default REST actions
	actions := []string{"index", "show", "create", "update", "destroy"}

	// Apply options if provided
	if len(opts) > 0 {
		opt := opts[0]
		if len(opt.Only) > 0 {
			actions = opt.Only
		} else if len(opt.Except) > 0 {
			actions = filterActions(actions, opt.Except)
		}
	}

	// Get controller type for reflection
	controllerValue := reflect.ValueOf(controller)

	// Map actions to HTTP methods and paths
	for _, action := range actions {
		switch action {
		case "index":
			if method := r.getControllerMethod(controllerValue, "Index"); method.IsValid() {
				r.Get(path, r.wrapMethod(method))
			}
		case "show":
			if method := r.getControllerMethod(controllerValue, "Show"); method.IsValid() {
				r.Get(path+"/:id", r.wrapMethod(method))
			}
		case "create":
			if method := r.getControllerMethod(controllerValue, "Create"); method.IsValid() {
				r.Post(path, r.wrapMethod(method))
			}
		case "update":
			if method := r.getControllerMethod(controllerValue, "Update"); method.IsValid() {
				r.Put(path+"/:id", r.wrapMethod(method))
			}
		case "destroy":
			if method := r.getControllerMethod(controllerValue, "Destroy"); method.IsValid() {
				r.Delete(path+"/:id", r.wrapMethod(method))
			}
		}
	}

	return r
}

// Namespace creates a group with a path prefix (Rails-style)
func (r *Router) Namespace(path string, configure func(*Router)) *Router {
	group := r.NewGroup(path)
	configure(group)
	return r
}

// Member adds routes for individual resource members
func (r *Router) Member(path string, configure func(*Router)) *Router {
	memberRouter := r.NewGroup(path + "/:id")
	configure(memberRouter)
	return r
}

// Collection adds routes for the entire resource collection
func (r *Router) Collection(path string, configure func(*Router)) *Router {
	collectionRouter := r.NewGroup(path)
	configure(collectionRouter)
	return r
}

// Static serves static files (like Rails assets)
func (r *Router) Static(path, directory string) *Router {
	r.ginGroup.Static(path, directory)
	return r
}

// Helper methods

func (r *Router) getControllerMethod(controllerValue reflect.Value, methodName string) reflect.Value {
	method := controllerValue.MethodByName(methodName)
	if !method.IsValid() {
		// Try with different casing
		method = controllerValue.MethodByName(strings.ToLower(methodName))
	}
	return method
}

func (r *Router) wrapMethod(method reflect.Value) HandlerFunc {
	return func(ctx *gin.Context) {
		// Call the controller method with gin.Context
		args := []reflect.Value{reflect.ValueOf(ctx)}
		method.Call(args)
	}
}

func filterActions(actions []string, except []string) []string {
	var filtered []string
	exceptMap := make(map[string]bool)
	for _, e := range except {
		exceptMap[e] = true
	}

	for _, action := range actions {
		if !exceptMap[action] {
			filtered = append(filtered, action)
		}
	}
	return filtered
}

// GetGinEngine returns the underlying Gin engine for advanced usage
func (r *Router) GetGinEngine() *gin.Engine {
	return r.ginEngine
}

// GetGinGroup returns the underlying Gin router group
func (r *Router) GetGinGroup() *gin.RouterGroup {
	return r.ginGroup
}

// NewFromGroup creates a Rails-style router from an existing Gin RouterGroup
func NewFromGroup(ginGroup *gin.RouterGroup) *Router {
	return &Router{
		ginEngine: nil, // Won't be used when wrapping a group
		ginGroup:  ginGroup,
		namespace: "",
	}
}
