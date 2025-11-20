package router

import (
	"net/http"
	"path"
	"strings"
	"sync"
)

// Router is a lightweight HTTP router with middleware support
type Router struct {
	trees      map[string]*node // HTTP method -> route tree
	middleware []MiddlewareFunc
	notFound   HandlerFunc
	pool       sync.Pool
	mu         sync.RWMutex
}

// New creates a new router
func New() *Router {
	r := &Router{
		trees:    make(map[string]*node),
		notFound: defaultNotFound,
	}
	r.pool.New = func() any {
		return &Context{
			params: make(Params, 0, 10),
			keys:   make(map[string]any),
		}
	}
	
	
	return r
}

// HandlerFunc defines the handler signature
type HandlerFunc func(*Context) error

// MiddlewareFunc defines the middleware signature
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Use adds global middleware
func (r *Router) Use(middleware ...MiddlewareFunc) {
	r.middleware = append(r.middleware, middleware...)
}

// GET registers a GET route
func (r *Router) GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	r.Handle(http.MethodGet, path, handler, middleware...)
}

// POST registers a POST route
func (r *Router) POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	r.Handle(http.MethodPost, path, handler, middleware...)
}

// PUT registers a PUT route
func (r *Router) PUT(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	r.Handle(http.MethodPut, path, handler, middleware...)
}

// DELETE registers a DELETE route
func (r *Router) DELETE(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	r.Handle(http.MethodDelete, path, handler, middleware...)
}

// PATCH registers a PATCH route
func (r *Router) PATCH(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	r.Handle(http.MethodPatch, path, handler, middleware...)
}

// HEAD registers a HEAD route
func (r *Router) HEAD(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	r.Handle(http.MethodHead, path, handler, middleware...)
}

// OPTIONS registers an OPTIONS route
func (r *Router) OPTIONS(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	r.Handle(http.MethodOptions, path, handler, middleware...)
}

// Handle registers a route with the given method and path
func (r *Router) Handle(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if path[0] != '/' {
		panic("path must begin with '/'")
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root
	}

	// Apply middleware in correct order: global -> route-specific
	finalHandler := handler
	for i := len(middleware) - 1; i >= 0; i-- {
		finalHandler = middleware[i](finalHandler)
	}
	for i := len(r.middleware) - 1; i >= 0; i-- {
		finalHandler = r.middleware[i](finalHandler)
	}

	root.addRoute(path, finalHandler)
}

// Group creates a new route group with prefix
func (r *Router) Group(prefix string, middleware ...MiddlewareFunc) *RouterGroup {
	return &RouterGroup{
		router:     r,
		prefix:     prefix,
		middleware: middleware,
	}
}

// ServeHTTP implements http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := r.pool.Get().(*Context)
	c.reset(w, req)
	defer r.pool.Put(c)

	r.handleRequest(c)
}

// handleRequest processes the HTTP request
func (r *Router) handleRequest(c *Context) {
	// Apply global middleware for all requests
	finalHandler := r.notFound
	for i := len(r.middleware) - 1; i >= 0; i-- {
		finalHandler = r.middleware[i](finalHandler)
	}

	r.mu.RLock()
	root := r.trees[c.Request.Method]
	r.mu.RUnlock()

	if root != nil {
		// Normalize path: remove trailing slash except for root "/"
		reqPath := c.Request.URL.Path
		if len(reqPath) > 1 {
			reqPath = strings.TrimSuffix(reqPath, "/")
		}

		if handler, params, _ := root.getValue(reqPath); handler != nil {
			c.params = params
			if err := handler(c); err != nil {
				c.Error(http.StatusInternalServerError, err)
			}
			return
		}
	}

	// Handle 404 with global middleware applied
	if err := finalHandler(c); err != nil {
		c.Error(http.StatusInternalServerError, err)
	}
}

// NotFound sets the 404 handler
func (r *Router) NotFound(handler HandlerFunc) {
	r.notFound = handler
}

// Static serves static files
func (r *Router) Static(prefix, root string) {
	// Ensure prefix starts with /
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	handler := func(c *Context) error {
		reqPath := c.Request.URL.Path

		// Remove the prefix
		file := strings.TrimPrefix(reqPath, prefix)
		file = strings.TrimPrefix(file, "/") // clean leading slash

		if file == "" || strings.HasSuffix(reqPath, "/") {
			file = "index.html"
		}

		fullPath := path.Join(root, file)
		http.ServeFile(c.Writer, c.Request, fullPath)
		return nil
	}

	// register route with wildcard
	r.GET(prefix+"/*filepath", handler)
	r.GET(prefix, handler) // also serve the exact prefix URL
}

// defaultNotFound is the default 404 handler
func defaultNotFound(c *Context) error {
	return c.String(http.StatusNotFound, "404 page not found")
}

// RouterGroup represents a group of routes with common prefix and middleware
type RouterGroup struct {
	router     *Router
	prefix     string
	middleware []MiddlewareFunc
}

// Use adds middleware to the group
func (g *RouterGroup) Use(middleware ...MiddlewareFunc) {
	g.middleware = append(g.middleware, middleware...)
}

// Group creates a sub-group
func (g *RouterGroup) Group(prefix string, middleware ...MiddlewareFunc) *RouterGroup {
	// Normalize path to avoid double slashes
	normalizedPrefix := g.prefix + prefix
	if g.prefix != "/" && prefix != "" && !strings.HasPrefix(prefix, "/") {
		normalizedPrefix = g.prefix + "/" + prefix
	}
	// Clean up double slashes
	normalizedPrefix = strings.ReplaceAll(normalizedPrefix, "//", "/")

	return &RouterGroup{
		router:     g.router,
		prefix:     normalizedPrefix,
		middleware: append(g.middleware, middleware...),
	}
}

// GET registers a GET route in the group
func (g *RouterGroup) GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	g.Handle(http.MethodGet, path, handler, middleware...)
}

// POST registers a POST route in the group
func (g *RouterGroup) POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	g.Handle(http.MethodPost, path, handler, middleware...)
}

// PUT registers a PUT route in the group
func (g *RouterGroup) PUT(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	g.Handle(http.MethodPut, path, handler, middleware...)
}

// DELETE registers a DELETE route in the group
func (g *RouterGroup) DELETE(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	g.Handle(http.MethodDelete, path, handler, middleware...)
}

// PATCH registers a PATCH route in the group
func (g *RouterGroup) PATCH(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	g.Handle(http.MethodPatch, path, handler, middleware...)
}

// OPTIONS registers an OPTIONS route in the group
func (g *RouterGroup) OPTIONS(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	g.Handle(http.MethodOptions, path, handler, middleware...)
}

// Handle registers a route in the group
func (g *RouterGroup) Handle(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	finalPath := g.prefix + path
	// Clean up double slashes
	finalPath = strings.ReplaceAll(finalPath, "//", "/")
	allMiddleware := append(g.middleware, middleware...)
	g.router.Handle(method, finalPath, handler, allMiddleware...)
}

// Static serves static files for the group
func (g *RouterGroup) Static(relativePath, root string) {
	g.router.Static(g.prefix+relativePath, root)
}

// Run starts the HTTP server
func (r *Router) Run(addr string) error {
	if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	return server.ListenAndServe()
}

// setupDefaultOptionsHandler adds a catch-all OPTIONS handler for CORS support
func (r *Router) setupDefaultOptionsHandler() {
	// Add a low-priority OPTIONS handler for all routes
	r.OPTIONS("/*filepath", func(c *Context) error {
		return c.NoContent()
	})
}
