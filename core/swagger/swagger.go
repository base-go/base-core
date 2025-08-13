package swagger

import (
	"net/http"
	"strings"

	"base/core/config"
	"base/core/router"
)

// SwaggerInfo holds the basic information for Swagger documentation
type SwaggerInfo struct {
	Version        string
	Host           string
	BasePath       string
	Schemes        []string
	Title          string
	Description    string
	TermsOfService string
	Contact        Contact
	License        License
}

// Contact contains the contact information for the API
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License contains the license information for the API
type License struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// SwaggerService handles swagger documentation generation
type SwaggerService struct {
	config *config.Config
	info   *SwaggerInfo
	paths  map[string]map[string]any // paths[path][method] = operation
}

// Global swagger service instance for auto-registration
var globalSwaggerService *SwaggerService

// NewSwaggerService creates a new swagger service
func NewSwaggerService(cfg *config.Config) *SwaggerService {
	service := &SwaggerService{
		config: cfg,
		paths:  make(map[string]map[string]any),
		info: &SwaggerInfo{
			Version:     cfg.Version,
			Host:        "localhost:8080", // Default, can be overridden
			BasePath:    "/",
			Schemes:     []string{"http", "https"},
			Title:       "Base Framework API",
			Description: "API documentation for Base Framework",
			Contact: Contact{
				Name:  "Base Team",
				URL:   "https://github.com/BaseTechStack",
				Email: "info@base.al",
			},
			License: License{
				Name: "MIT",
				URL:  "https://opensource.org/licenses/MIT",
			},
		},
	}
	
	// Set global instance for auto-registration
	globalSwaggerService = service
	return service
}

// GenerateSwaggerDoc generates OpenAPI 3.0 documentation for Base framework
func (s *SwaggerService) GenerateSwaggerDoc() map[string]any {
	// Build server URL from config
	serverURL := s.config.BaseURL
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":          s.info.Title,
			"description":    s.info.Description,
			"version":        s.info.Version,
			"termsOfService": s.info.TermsOfService,
			"contact": map[string]any{
				"name":  s.info.Contact.Name,
				"url":   s.info.Contact.URL,
				"email": s.info.Contact.Email,
			},
			"license": map[string]any{
				"name": s.info.License.Name,
				"url":  s.info.License.URL,
			},
		},
		"servers": []map[string]any{
			{
				"url":         serverURL,
				"description": "Base Framework API Server",
			},
		},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"ApiKeyAuth": map[string]any{
					"type": "apiKey",
					"in":   "header",
					"name": "X-Api-Key",
				},
				"BearerAuth": map[string]any{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
					"description":  "Enter the token with the `Bearer: ` prefix, e.g. \"Bearer abcde12345\"",
				},
			},
			"schemas": map[string]any{
				"Error": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"error": map[string]any{
							"type":        "string",
							"description": "Error message",
						},
					},
					"required": []string{"error"},
				},
				"Success": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"message": map[string]any{
							"type":        "string",
							"description": "Success message",
						},
					},
					"required": []string{"message"},
				},
			},
		},
		"tags": []map[string]any{
			{
				"name":        "Core/Auth",
				"description": "Authentication operations",
			},
			{
				"name":        "Core/Profile",
				"description": "User profile operations",
			},
			{
				"name":        "Core/Authorization",
				"description": "Authorization and role management",
			},
			{
				"name":        "Core/Media",
				"description": "Media file management",
			},
			{
				"name":        "Core/Translation",
				"description": "Translation and localization",
			},
			{
				"name":        "App/Categories",
				"description": "Category management",
			},
		},
		"paths": s.generatePaths(),
	}
}

// Operation represents a single API operation
type Operation struct {
	Summary     string                 `json:"summary,omitempty"`
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	OperationID string                 `json:"operationId,omitempty"`
	Parameters  []Parameter            `json:"parameters,omitempty"`
	RequestBody *RequestBody           `json:"requestBody,omitempty"`
	Responses   map[string]Response    `json:"responses"`
	Security    []map[string][]string  `json:"security,omitempty"`
}

type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Schema      Schema `json:"schema"`
}

type RequestBody struct {
	Description string             `json:"description,omitempty"`
	Required    bool               `json:"required,omitempty"`
	Content     map[string]Content `json:"content"`
}

type Response struct {
	Description string             `json:"description"`
	Content     map[string]Content `json:"content,omitempty"`
}

type Content struct {
	Schema Schema `json:"schema"`
}

type Schema struct {
	Type       string                 `json:"type,omitempty"`
	Properties map[string]Schema      `json:"properties,omitempty"`
	Items      *Schema                `json:"items,omitempty"`
	Ref        string                 `json:"$ref,omitempty"`
	Example    any                    `json:"example,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

// RegisterOperation registers an API operation for swagger documentation
func (s *SwaggerService) RegisterOperation(method, path string, operation Operation) {
	if s.paths[path] == nil {
		s.paths[path] = make(map[string]any)
	}
	s.paths[path][method] = operation
}

// AutoRegisterRoute is a helper for controllers to auto-register swagger docs (simple version)
func AutoRegisterRoute(method, path, summary, description string, tags []string) {
	if globalSwaggerService == nil {
		return // Swagger not initialized yet
	}
	
	operation := Operation{
		Summary:     summary,
		Description: description,
		Tags:        tags,
		Responses: map[string]Response{
			"200": {Description: "Success"},
			"400": {Description: "Bad Request"},
			"404": {Description: "Not Found"},
			"500": {Description: "Internal Server Error"},
		},
	}
	
	globalSwaggerService.RegisterOperation(strings.ToLower(method), path, operation)
}

// AutoRegisterRouteDetailed is a helper for controllers to auto-register detailed swagger docs
func AutoRegisterRouteDetailed(method, path, summary, description string, tags []string,
	parameters []Parameter, responses map[string]Response, security []map[string][]string) {
	if globalSwaggerService == nil {
		return // Swagger not initialized yet
	}
	
	operation := Operation{
		Summary:     summary,
		Description: description,
		Tags:        tags,
		Parameters:  parameters,
		Responses:   responses,
		Security:    security,
	}
	
	// Add default responses if none provided
	if len(responses) == 0 {
		operation.Responses = map[string]Response{
			"200": {Description: "Success"},
			"400": {Description: "Bad Request"},
			"404": {Description: "Not Found"},
			"500": {Description: "Internal Server Error"},
		}
	}
	
	globalSwaggerService.RegisterOperation(strings.ToLower(method), path, operation)
}

// loadAutoGeneratedRoutes attempts to call auto-generated registration
func (s *SwaggerService) loadAutoGeneratedRoutes() {
	// Call the auto-generated registration function if it exists
	RegisterAllRoutes()
}

// generatePaths generates the paths for the OpenAPI spec
func (s *SwaggerService) generatePaths() map[string]any {
	paths := make(map[string]any)

	// Add all registered operations  
	for path, operations := range s.paths {
		paths[path] = operations
	}

	// Try to load auto-generated routes (if available)
	s.loadAutoGeneratedRoutes()

	// Manually register core routes (fallback)
	s.registerAuthRoutes(paths)
	s.registerTranslationRoutes(paths)
	s.registerMediaRoutes(paths)

	// Health check endpoint
	paths["/health"] = map[string]any{
		"get": map[string]any{
			"summary":     "Health check",
			"description": "Check the health status of the API",
			"operationId": "healthCheck",
			"tags":        []string{"System"},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "API is healthy",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"status": map[string]any{
										"type":    "string",
										"example": "ok",
									},
									"version": map[string]any{
										"type":    "string",
										"example": "1.0.0",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Root endpoint
	paths["/"] = map[string]any{
		"get": map[string]any{
			"summary":     "API information",
			"description": "Get basic API information",
			"operationId": "apiInfo",
			"tags":        []string{"System"},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "API information",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"message": map[string]any{
										"type":    "string",
										"example": "pong",
									},
									"version": map[string]any{
										"type":    "string",
										"example": "1.0.0",
									},
									"swagger": map[string]any{
										"type":    "string",
										"example": "/swagger/index.html",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return paths
}

// RegisterRoutes registers swagger routes on the router
func (s *SwaggerService) RegisterRoutes(r *router.Router) {
	// Swagger JSON endpoint
	r.GET("/swagger/doc.json", s.swaggerJSONHandler)

	// Swagger UI endpoints
	r.GET("/swagger", s.swaggerIndexHandler)
	r.GET("/swagger/", s.swaggerIndexHandler)
	r.GET("/swagger/index.html", s.swaggerIndexHandler)
}

// swaggerJSONHandler serves the OpenAPI JSON specification
func (s *SwaggerService) swaggerJSONHandler(c *router.Context) error {
	doc := s.GenerateSwaggerDoc()
	c.SetHeader("Content-Type", "application/json")
	return c.JSON(http.StatusOK, doc)
}

// swaggerIndexHandler serves the Swagger UI for /swagger root
func (s *SwaggerService) swaggerIndexHandler(c *router.Context) error {
	return c.HTML(http.StatusOK, s.generateSwaggerHTML())
}


// generateSwaggerHTML generates a Swagger UI HTML page for OpenAPI 3.0
func (s *SwaggerService) generateSwaggerHTML() string {
	return `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Base Framework API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
        .topbar {
            display: none;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/swagger/doc.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                tryItOutEnabled: true,
                requestInterceptor: (request) => {
                    // Add any custom headers if needed
                    return request;
                },
                responseInterceptor: (response) => {
                    return response;
                },
                docExpansion: "list",
                defaultModelsExpandDepth: 1,
                defaultModelExpandDepth: 1
            });
        };
    </script>
</body>
</html>`
}

// SetHost sets the host for swagger documentation
func (s *SwaggerService) SetHost(host string) {
	s.info.Host = host
}

// SetBasePath sets the base path for swagger documentation
func (s *SwaggerService) SetBasePath(basePath string) {
	s.info.BasePath = basePath
}

// SetSchemes sets the supported schemes (http, https)
func (s *SwaggerService) SetSchemes(schemes []string) {
	s.info.Schemes = schemes
}

// registerAuthRoutes registers authentication endpoints
func (s *SwaggerService) registerAuthRoutes(paths map[string]any) {
	// Register endpoint
	paths["/api/auth/register"] = map[string]any{
		"post": map[string]any{
			"summary":     "Register user",
			"description": "Register a new user account",
			"tags":        []string{"Core/Auth"},
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"email": map[string]any{
									"type":    "string",
									"example": "user@example.com",
								},
								"password": map[string]any{
									"type":    "string",
									"example": "password123",
								},
								"name": map[string]any{
									"type":    "string",
									"example": "John Doe",
								},
							},
							"required": []string{"email", "password", "name"},
						},
					},
				},
			},
			"responses": map[string]any{
				"201": map[string]any{
					"description": "User registered successfully",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"user": map[string]any{
										"type": "object",
									},
									"token": map[string]any{
										"type": "string",
									},
								},
							},
						},
					},
				},
				"400": map[string]any{
					"description": "Bad request",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/Error",
							},
						},
					},
				},
			},
		},
	}

	// Login endpoint
	paths["/api/auth/login"] = map[string]any{
		"post": map[string]any{
			"summary":     "User login",
			"description": "Authenticate user and return JWT token",
			"tags":        []string{"Core/Auth"},
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"email": map[string]any{
									"type":    "string",
									"example": "user@example.com",
								},
								"password": map[string]any{
									"type":    "string",
									"example": "password123",
								},
							},
							"required": []string{"email", "password"},
						},
					},
				},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Login successful",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"user": map[string]any{
										"type": "object",
									},
									"token": map[string]any{
										"type": "string",
									},
								},
							},
						},
					},
				},
				"401": map[string]any{
					"description": "Invalid credentials",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/Error",
							},
						},
					},
				},
			},
		},
	}

	// Logout endpoint
	paths["/api/auth/logout"] = map[string]any{
		"post": map[string]any{
			"summary":     "User logout",
			"description": "Logout user and invalidate token",
			"tags":        []string{"Core/Auth"},
			"security": []map[string][]string{
				{"BearerAuth": {}},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Logout successful",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/Success",
							},
						},
					},
				},
			},
		},
	}

	// Forgot password endpoint
	paths["/api/auth/forgot-password"] = map[string]any{
		"post": map[string]any{
			"summary":     "Forgot password",
			"description": "Send password reset email",
			"tags":        []string{"Core/Auth"},
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"email": map[string]any{
									"type":    "string",
									"example": "user@example.com",
								},
							},
							"required": []string{"email"},
						},
					},
				},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Reset email sent",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/Success",
							},
						},
					},
				},
			},
		},
	}

	// Reset password endpoint
	paths["/api/auth/reset-password"] = map[string]any{
		"post": map[string]any{
			"summary":     "Reset password",
			"description": "Reset user password with token",
			"tags":        []string{"Core/Auth"},
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"token": map[string]any{
									"type": "string",
								},
								"password": map[string]any{
									"type": "string",
								},
							},
							"required": []string{"token", "password"},
						},
					},
				},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Password reset successful",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/Success",
							},
						},
					},
				},
			},
		},
	}
}

// registerTranslationRoutes registers translation endpoints  
func (s *SwaggerService) registerTranslationRoutes(paths map[string]any) {
	// Placeholder for translation routes
	paths["/api/translations"] = map[string]any{
		"get": map[string]any{
			"summary": "List translations",
			"tags":    []string{"Core/Translation"},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "List of translations",
				},
			},
		},
	}
}

// registerMediaRoutes registers media endpoints
func (s *SwaggerService) registerMediaRoutes(paths map[string]any) {
	// Placeholder for media routes
	paths["/api/media"] = map[string]any{
		"get": map[string]any{
			"summary": "List media files",
			"tags":    []string{"Core/Media"},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "List of media files",
				},
			},
		},
	}
}
