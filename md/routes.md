# Rails-Style Router Documentation

A Rails-inspired routing wrapper for Gin that provides clean, readable route definitions while maintaining all of Gin's performance and middleware capabilities.

## Table of Contents

- [Quick Start](#quick-start)
- [Basic Routes](#basic-routes)
- [Resource Routes](#resource-routes)
- [Namespaces and Route Groups](#namespaces-and-route-groups)
- [Member and Collection Routes](#member-and-collection-routes)
- [Middleware](#middleware)
- [Advanced Usage](#advanced-usage)
- [Migration from Pure Gin](#migration-from-pure-gin)

## Quick Start

```go
import (
    "base/core/router"
    "github.com/gin-gonic/gin"
)

// Create Gin engine
ginEngine := gin.New()

// Create Rails-style router wrapper
r := router.New(ginEngine)

// Define routes
r.Get("/", homeController.Index).
  Get("/about", homeController.About).
  Post("/contact", contactController.Create)
```

## Basic Routes

### HTTP Methods

```go
// Basic HTTP methods (Rails-style naming)
r.Get("/users", userController.Index)           // GET /users
r.Post("/users", userController.Create)         // POST /users
r.Put("/users/:id", userController.Update)      // PUT /users/:id
r.Patch("/users/:id", userController.Patch)     // PATCH /users/:id
r.Delete("/users/:id", userController.Destroy)  // DELETE /users/:id
```

### Method Chaining

```go
r.Get("/", homeController.Index).
  Get("/about", homeController.About).
  Get("/contact", contactController.Show).
  Post("/contact", contactController.Create)
```

### Static File Serving

```go
r.Static("/static", "./static")      // Serve files from ./static directory
r.Static("/assets", "./public")      // Serve files from ./public directory
```

## Resource Routes

Generate all RESTful routes for a resource with a single line:

```go
// Basic resource routes
r.Resources("/posts", postController)
```

This automatically creates:
- `GET /posts` → `postController.Index`
- `GET /posts/:id` → `postController.Show`  
- `POST /posts` → `postController.Create`
- `PUT /posts/:id` → `postController.Update`
- `DELETE /posts/:id` → `postController.Destroy`

### Resource Options

```go
// Only specific actions
r.Resources("/posts", postController, router.ResourceOptions{
    Only: []string{"index", "show"},
})

// Exclude specific actions  
r.Resources("/posts", postController, router.ResourceOptions{
    Except: []string{"create", "update", "destroy"},
})
```

### Controller Requirements

For Resources to work, your controller must implement the corresponding methods:

```go
type PostController struct {
    // ... fields
}

func (c *PostController) Index(ctx *gin.Context) {
    // Handle GET /posts
}

func (c *PostController) Show(ctx *gin.Context) {
    // Handle GET /posts/:id
}

func (c *PostController) Create(ctx *gin.Context) {
    // Handle POST /posts
}

func (c *PostController) Update(ctx *gin.Context) {
    // Handle PUT /posts/:id
}

func (c *PostController) Destroy(ctx *gin.Context) {
    // Handle DELETE /posts/:id
}
```

## Namespaces and Route Groups

### Basic Namespaces

```go
r.Namespace("/admin", func(admin *router.Router) {
    admin.Get("/dashboard", adminController.Dashboard)
    admin.Get("/users", adminController.Users)
    admin.Resources("/posts", adminPostController)
})
```

This creates:
- `GET /admin/dashboard`
- `GET /admin/users`
- `GET /admin/posts`
- `POST /admin/posts`
- etc.

### API Versioning

```go
r.Namespace("/api/v1", func(api *router.Router) {
    api.Resources("/posts", apiPostController)
    api.Resources("/users", apiUserController)
})

r.Namespace("/api/v2", func(api *router.Router) {
    api.Resources("/posts", apiV2PostController)
})
```

### Nested Namespaces

```go
r.Namespace("/admin", func(admin *router.Router) {
    admin.Namespace("/api", func(adminApi *router.Router) {
        adminApi.Resources("/posts", adminApiPostController)
    })
})
// Creates: GET /admin/api/posts, POST /admin/api/posts, etc.
```

## Member and Collection Routes

### Member Routes

Add routes that act on individual resource members:

```go
r.Member("/posts", func(member *router.Router) {
    member.Get("/archive", postController.Archive)     // GET /posts/:id/archive
    member.Post("/publish", postController.Publish)    // POST /posts/:id/publish
    member.Delete("/unpublish", postController.Unpublish) // DELETE /posts/:id/unpublish
})
```

### Collection Routes

Add routes that act on the entire resource collection:

```go
r.Collection("/posts", func(collection *router.Router) {
    collection.Get("/recent", postController.Recent)        // GET /posts/recent
    collection.Get("/popular", postController.Popular)      // GET /posts/popular
    collection.Post("/bulk-delete", postController.BulkDelete) // POST /posts/bulk-delete
})
```

## Middleware

### Global Middleware

```go
r.Use(gin.Recovery()).
  Use(gin.Logger()).
  Use(corsMiddleware())
```

### Route-Specific Middleware

```go
r.Use(authMiddleware()).Namespace("/admin", func(admin *router.Router) {
    admin.Get("/dashboard", adminController.Dashboard)
})
```

### Namespace Middleware

```go
r.Namespace("/api", func(api *router.Router) {
    api.Use(apiKeyMiddleware()).
        Use(rateLimitMiddleware())
    
    api.Resources("/posts", apiPostController)
})
```

## Advanced Usage

### Complex Route Setup

```go
func SetupRoutes(ginEngine *gin.Engine, deps *RouterDependencies) {
    r := router.New(ginEngine)

    // Global middleware
    r.Use(middleware.Logger()).
      Use(middleware.Recovery()).
      Use(middleware.CORS())

    // Static assets
    r.Static("/static", "./static").
      Static("/uploads", "./storage")

    // Public routes
    r.Get("/", homeController.Index).
      Get("/about", homeController.About).
      Get("/health", healthController.Check)

    // Public API
    r.Namespace("/api/v1", func(api *router.Router) {
        api.Resources("/posts", postController, router.ResourceOptions{
            Only: []string{"index", "show"},
        })
    })

    // Authenticated web routes
    r.Use(authMiddleware()).Namespace("/dashboard", func(dashboard *router.Router) {
        dashboard.Get("/", dashboardController.Index)
        dashboard.Resources("/posts", postController)
        
        // Admin section
        dashboard.Use(adminMiddleware()).Namespace("/admin", func(admin *router.Router) {
            admin.Resources("/users", userController)
            admin.Resources("/settings", settingController)
        })
    })

    // Authenticated API
    r.Use(apiAuthMiddleware()).Namespace("/api/v1", func(api *router.Router) {
        api.Resources("/posts", postController)
        api.Resources("/users", userController)
        
        // Member routes for posts
        api.Member("/posts", func(member *router.Router) {
            member.Post("/publish", postController.Publish)
            member.Post("/archive", postController.Archive)
        })
    })
}
```

### Custom Route Groups

```go
// Create reusable route groups
publicAPI := r.NewGroup("/api/v1")
publicAPI.Resources("/posts", postController, router.ResourceOptions{
    Only: []string{"index", "show"},
})

authAPI := r.NewGroup("/api/v1").Use(authMiddleware())
authAPI.Resources("/posts", postController)
```

## Migration from Pure Gin

### Before (Pure Gin)

```go
// Verbose route definitions
router := gin.New()
router.GET("/posts", postController.Index)
router.POST("/posts", postController.Create)
router.GET("/posts/:id", postController.Show)
router.PUT("/posts/:id", postController.Update)
router.DELETE("/posts/:id", postController.Destroy)

// Route groups
adminGroup := router.Group("/admin")
adminGroup.Use(authMiddleware())
adminGroup.GET("/dashboard", adminController.Dashboard)
adminGroup.GET("/users", adminController.Users)

apiGroup := router.Group("/api/v1")
apiGroup.Use(apiKeyMiddleware())
apiGroup.GET("/posts", apiPostController.Index)
apiGroup.POST("/posts", apiPostController.Create)
```

### After (Rails-Style Router)

```go
// Clean, readable route definitions
r := router.New(gin.New())

// Automatic RESTful routes
r.Resources("/posts", postController)

// Clean namespaces with middleware
r.Use(authMiddleware()).Namespace("/admin", func(admin *router.Router) {
    admin.Get("/dashboard", adminController.Dashboard)
    admin.Get("/users", adminController.Users)
})

r.Use(apiKeyMiddleware()).Namespace("/api/v1", func(api *router.Router) {
    api.Resources("/posts", apiPostController)
})
```

## Route Logging

The router automatically logs route registration with emoji indicators:

```
📍 Resources registered: /posts [index, show, create, update, destroy] on *posts.PostController
📍 Resources registered: /api/posts [index, show, create, update, destroy] on *posts.APIController
```

## Benefits

✅ **Cleaner syntax**: `router.Resources("/posts", controller)` vs multiple route definitions  
✅ **Rails conventions**: Familiar patterns for Rails developers  
✅ **Method chaining**: Fluent API for readable route definitions  
✅ **Full Gin compatibility**: All Gin features and middleware work unchanged  
✅ **Performance**: Zero overhead wrapper around Gin  
✅ **Type safety**: Reflection-based controller method binding with error checking  
✅ **Automatic logging**: Visual feedback for route registration  

## Integration with Existing Code

The Rails-style router is designed to work alongside existing Gin code:

```go
// Mix Rails-style and traditional Gin routes
r := router.New(ginEngine)

// Rails-style
r.Resources("/posts", postController)

// Traditional Gin (still works)
ginEngine.GET("/legacy", legacyHandler)

// Access underlying Gin engine when needed
underlyingGin := r.GetGinEngine()
underlyingGroup := r.GetGinGroup()
```

This allows for gradual migration and coexistence with existing route definitions.