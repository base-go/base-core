package swagger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"base/core/config"
	"base/core/router"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

// SwaggerAnnotation represents a parsed swagger annotation from controllers
type SwaggerAnnotation struct {
	Summary     string
	Description string
	Tags        string
	Method      string
	Route       string
	Security    []string
	Parameters  []ParamAnnotation
	Responses   []ResponseAnnotation
}

type ParamAnnotation struct {
	Name        string
	In          string // query, path, header, body
	Description string
	Required    bool
	Type        string
	Example     string
}

type ResponseAnnotation struct {
	Code        string
	Description string
	Schema      string
	Example     string
}

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
			"schemas": s.generateSchemas(),
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
		},
		"paths": s.generatePaths(),
	}
}

// Operation represents a single API operation
type Operation struct {
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	OperationID string                `json:"operationId,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
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
	Type       string            `json:"type,omitempty"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Items      *Schema           `json:"items,omitempty"`
	Ref        string            `json:"$ref,omitempty"`
	Example    any               `json:"example,omitempty"`
	Required   []string          `json:"required,omitempty"`
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
// Deprecated: auto-generated routes are no longer used, static files are preferred
func (s *SwaggerService) loadAutoGeneratedRoutes() {
	// No longer calling RegisterAllRoutes() - using static documentation instead
}

// generatePaths generates the paths for the OpenAPI spec
func (s *SwaggerService) generatePaths() map[string]any {
	paths := make(map[string]any)

	// Add all registered operations
	for path, operations := range s.paths {
		paths[path] = operations
	}

	// Load controller annotations automatically
	s.loadControllerAnnotations(paths)

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

// loadControllerAnnotations automatically loads controller annotations from app and core controllers
func (s *SwaggerService) loadControllerAnnotations(paths map[string]any) {
	// Scan and parse controller annotations from both app and core directories
	s.scanAndParseControllerAnnotations(paths)
}

// scanAndParseControllerAnnotations scans controller files and parses their swagger annotations
func (s *SwaggerService) scanAndParseControllerAnnotations(paths map[string]any) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		// Fallback to hardcoded endpoints if we can't get working directory
		s.addCoreAuthEndpoints(paths)
		s.addCoreProfileEndpoints(paths)
		s.addCoreMediaEndpoints(paths)
		return
	}

	// Directories to scan for controllers (absolute paths)
	controllerDirs := []string{
		filepath.Join(cwd, "app"),         // App modules: /base/app/*/controller.go
		filepath.Join(cwd, "core", "app"), // Core modules: /base/core/app/*/controller.go
	}

	for _, dir := range controllerDirs {
		s.scanDirectoryForControllers(dir, paths)
	}
}

// scanDirectoryForControllers scans a directory for controller files and parses their annotations
func (s *SwaggerService) scanDirectoryForControllers(baseDir string, paths map[string]any) {
	// Use the existing annotation parsing logic from the CLI
	controllerFiles, err := s.findControllerFiles(baseDir)
	if err != nil {
		// Fallback to hardcoded endpoints for known modules
		switch baseDir {
		case "core/app":
			s.addCoreAuthEndpoints(paths)
			s.addCoreProfileEndpoints(paths)
			s.addCoreMediaEndpoints(paths)
		}
		return
	}

	// Parse annotations from each controller file
	for _, file := range controllerFiles {
		annotations, err := s.parseSwaggerAnnotations(file)
		if err != nil {
			continue // Skip files with parse errors
		}

		// Convert annotations to OpenAPI paths
		s.convertAnnotationsToPaths(annotations, paths)
	}
}

// findControllerFiles finds controller.go files in the given directory
func (s *SwaggerService) findControllerFiles(rootDir string) ([]string, error) {
	var files []string

	// Walk through the directory to find controller.go files
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, "controller.go") || strings.Contains(path, "/controllers/") {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// parseSwaggerAnnotations parses swagger annotations from a controller file
func (s *SwaggerService) parseSwaggerAnnotations(filename string) ([]SwaggerAnnotation, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var annotations []SwaggerAnnotation
	var currentAnnotation SwaggerAnnotation
	var inAnnotationBlock bool

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if summary, found := strings.CutPrefix(line, "// @Summary "); found {
			currentAnnotation = SwaggerAnnotation{} // Reset
			currentAnnotation.Summary = summary
			inAnnotationBlock = true
		} else if inAnnotationBlock {
			if description, found := strings.CutPrefix(line, "// @Description "); found {
				currentAnnotation.Description = description
			} else if tags, found := strings.CutPrefix(line, "// @Tags "); found {
				currentAnnotation.Tags = tags
			} else if strings.HasPrefix(line, "// @Router") {
				// Parse @Router /path [method]
				routerRegex := regexp.MustCompile(`// @Router\s+(.+?)\s+\[(.+?)\]`)
				matches := routerRegex.FindStringSubmatch(line)
				if len(matches) == 3 {
					currentAnnotation.Route = matches[1]
					currentAnnotation.Method = strings.ToUpper(matches[2])
				}
			} else if strings.HasPrefix(line, "// @Param") {
				// Parse @Param name in type required description
				param := s.parseParamAnnotation(line)
				if param != nil {
					currentAnnotation.Parameters = append(currentAnnotation.Parameters, *param)
				}
			} else if strings.HasPrefix(line, "// @Success") {
				// Parse @Success code description
				resp := s.parseResponseAnnotation(line, true)
				if resp != nil {
					currentAnnotation.Responses = append(currentAnnotation.Responses, *resp)
				}
			} else if strings.HasPrefix(line, "// @Failure") {
				// Parse @Failure code description
				resp := s.parseResponseAnnotation(line, false)
				if resp != nil {
					currentAnnotation.Responses = append(currentAnnotation.Responses, *resp)
				}
			} else if security, found := strings.CutPrefix(line, "// @Security "); found {
				// Parse @Security scheme
				currentAnnotation.Security = append(currentAnnotation.Security, security)
			} else if strings.HasPrefix(line, "func ") {
				// End of annotation block - we've reached the function
				if currentAnnotation.Summary != "" && currentAnnotation.Route != "" {
					annotations = append(annotations, currentAnnotation)
				}
				inAnnotationBlock = false
			}
		} else if !strings.HasPrefix(line, "//") && line != "" {
			// Non-comment line - end annotation block
			inAnnotationBlock = false
		}
	}

	return annotations, scanner.Err()
}

// parseParamAnnotation parses @Param annotations
func (s *SwaggerService) parseParamAnnotation(line string) *ParamAnnotation {
	paramStr, _ := strings.CutPrefix(line, "// @Param ")
	parts := s.parseQuotedString(paramStr)
	if len(parts) < 4 {
		return nil
	}

	param := &ParamAnnotation{
		Name: parts[0],
		In:   parts[1],
		Type: parts[2],
	}

	if strings.ToLower(parts[3]) == "true" {
		param.Required = true
	}

	if len(parts) > 4 {
		param.Description = strings.Trim(parts[4], "\"")
	}

	if len(parts) > 5 {
		param.Example = strings.Trim(parts[5], "\"")
	}

	return param
}

// parseResponseAnnotation parses @Success and @Failure annotations
func (s *SwaggerService) parseResponseAnnotation(line string, isSuccess bool) *ResponseAnnotation {
	var prefix string
	if isSuccess {
		prefix = "// @Success "
	} else {
		prefix = "// @Failure "
	}

	respStr, _ := strings.CutPrefix(line, prefix)
	parts := s.parseQuotedString(respStr)

	if len(parts) < 2 {
		return nil
	}

	resp := &ResponseAnnotation{
		Code: parts[0],
	}

	if len(parts) > 1 {
		resp.Description = strings.Trim(parts[1], "\"")
	}

	if len(parts) > 2 {
		resp.Schema = parts[2]
	}

	if len(parts) > 3 {
		resp.Example = strings.Trim(parts[3], "\"")
	}

	return resp
}

// parseQuotedString splits a string by spaces while preserving quoted strings
func (s *SwaggerService) parseQuotedString(str string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false

	for i, char := range str {
		if char == '"' {
			inQuotes = !inQuotes
			current.WriteRune(char)
		} else if char == ' ' && !inQuotes {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(char)
		}

		if i == len(str)-1 && current.Len() > 0 {
			parts = append(parts, current.String())
		}
	}

	return parts
}

// convertAnnotationsToPaths converts parsed annotations to OpenAPI paths
func (s *SwaggerService) convertAnnotationsToPaths(annotations []SwaggerAnnotation, paths map[string]any) {
	for _, ann := range annotations {
		route := ann.Route
		if !strings.HasPrefix(route, "/api") && !strings.HasPrefix(route, "/") {
			route = "/api/" + route
		} else if strings.HasPrefix(route, "/") && !strings.HasPrefix(route, "/api") {
			route = "/api" + route
		}

		if paths[route] == nil {
			paths[route] = make(map[string]any)
		}

		// Build operation
		operation := map[string]any{
			"summary":     ann.Summary,
			"description": ann.Description,
			"operationId": s.generateOperationID(ann.Method, route),
		}

		// Add tags
		if ann.Tags != "" {
			operation["tags"] = []string{ann.Tags}
		}

		// Add parameters
		if len(ann.Parameters) > 0 {
			var parameters []any
			for _, param := range ann.Parameters {
				paramMap := map[string]any{
					"name":        param.Name,
					"in":          param.In,
					"description": param.Description,
					"required":    param.Required,
				}

				// For body parameters, use requestBody instead of parameter schema
				if param.In == "body" {
					// Body parameters should be handled as requestBody in OpenAPI 3.0
					// Skip adding to parameters array - will be handled separately
					continue
				}

				// For other parameter types, determine if it's a model reference or primitive type
				if strings.Contains(param.Type, ".") && !strings.Contains(param.Type, "[]") {
					// This is a model reference like "models.CreatePostRequest"
					paramMap["schema"] = map[string]any{
						"$ref": "#/components/schemas/" + param.Type,
					}
				} else {
					// This is a primitive type like "string", "integer", etc.
					paramMap["schema"] = map[string]any{
						"type": param.Type,
					}
				}

				parameters = append(parameters, paramMap)
			}
			operation["parameters"] = parameters
		}

		// Handle body parameters as requestBody for OpenAPI 3.0
		for _, param := range ann.Parameters {
			if param.In == "body" {
				requestBody := map[string]any{
					"required": param.Required,
					"content": map[string]any{
						"application/json": map[string]any{},
					},
				}

				// Determine if it's a model reference or primitive type
				if strings.Contains(param.Type, ".") {
					// This is a model reference like "models.CreatePostRequest"
					requestBody["content"].(map[string]any)["application/json"].(map[string]any)["schema"] = map[string]any{
						"$ref": "#/components/schemas/" + param.Type,
					}
				} else {
					// This is a primitive type
					requestBody["content"].(map[string]any)["application/json"].(map[string]any)["schema"] = map[string]any{
						"type": param.Type,
					}
				}

				operation["requestBody"] = requestBody
				break // Only one body parameter per operation
			}
		}

		// Add responses
		responses := make(map[string]any)
		if len(ann.Responses) > 0 {
			for _, resp := range ann.Responses {
				response := map[string]any{
					"description": resp.Description,
				}

				// Add schema reference if provided
				if resp.Schema != "" {
					response["content"] = map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/" + resp.Schema,
							},
						},
					}
				}

				responses[resp.Code] = response
			}
		} else {
			responses["200"] = map[string]any{
				"description": "Success",
			}
		}
		operation["responses"] = responses

		// Add security
		if len(ann.Security) > 0 {
			var security []any
			for _, sec := range ann.Security {
				security = append(security, map[string]any{
					sec: []any{},
				})
			}
			operation["security"] = security
		}

		paths[route].(map[string]any)[strings.ToLower(ann.Method)] = operation
	}
}

// generateOperationID creates a unique operation ID from method and route
func (s *SwaggerService) generateOperationID(method, route string) string {
	parts := strings.Split(strings.Trim(route, "/"), "/")
	var opID strings.Builder

	opID.WriteString(strings.ToLower(method))

	for _, part := range parts {
		if part == "api" {
			continue
		}
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			param := strings.Trim(part, "{}")
			opID.WriteString("By")
			opID.WriteString(cases.Title(language.AmericanEnglish, cases.NoLower).String(param))
		} else {
			opID.WriteString(cases.Title(language.AmericanEnglish, cases.NoLower).String(part))
		}
	}

	return opID.String()
}

// addCoreAuthEndpoints adds Core/Auth endpoints based on controller annotations
func (s *SwaggerService) addCoreAuthEndpoints(paths map[string]any) {
	paths["/api/auth/register"] = map[string]any{
		"post": map[string]any{
			"summary":     "Register",
			"description": "Register user",
			"tags":        []string{"Core/Auth"},
			"security": []map[string][]string{
				{"ApiKeyAuth": {}},
			},
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{
							"$ref": "#/components/schemas/RegisterRequest",
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
								"$ref": "#/components/schemas/AuthResponse",
							},
						},
					},
				},
				"400": map[string]any{
					"description": "Bad request",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"409": map[string]any{
					"description": "User already exists",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"500": map[string]any{
					"description": "Internal server error",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
			},
		},
	}

	paths["/api/auth/login"] = map[string]any{
		"post": map[string]any{
			"summary":     "Login",
			"description": "Login user",
			"tags":        []string{"Core/Auth"},
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{
							"$ref": "#/components/schemas/LoginRequest",
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
								"$ref": "#/components/schemas/AuthResponse",
							},
						},
					},
				},
				"400": map[string]any{
					"description": "Bad request",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"401": map[string]any{
					"description": "Unauthorized",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"500": map[string]any{
					"description": "Internal server error",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
			},
		},
	}

	paths["/api/auth/logout"] = map[string]any{
		"post": map[string]any{
			"summary":     "Logout",
			"description": "Logout user",
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
								"$ref": "#/components/schemas/types.SuccessResponse",
							},
						},
					},
				},
				"401": map[string]any{
					"description": "Unauthorized",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
			},
		},
	}

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
									"type":        "string",
									"format":      "email",
									"description": "Email address",
									"example":     "user@example.com",
								},
							},
							"required": []string{"email"},
						},
					},
				},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Password reset email sent",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.SuccessResponse",
							},
						},
					},
				},
				"400": map[string]any{
					"description": "Bad request",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"500": map[string]any{
					"description": "Internal server error",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
			},
		},
	}

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
									"type":        "string",
									"description": "Reset token from email",
								},
								"password": map[string]any{
									"type":        "string",
									"description": "New password",
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
								"$ref": "#/components/schemas/types.SuccessResponse",
							},
						},
					},
				},
				"400": map[string]any{
					"description": "Bad request",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"500": map[string]any{
					"description": "Internal server error",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
			},
		},
	}
}

// addCoreProfileEndpoints adds Core/Profile endpoints based on controller annotations
func (s *SwaggerService) addCoreProfileEndpoints(paths map[string]any) {
	paths["/api/profile"] = map[string]any{
		"get": map[string]any{
			"summary":     "Get profile from Authenticated User Token",
			"description": "Get profile by Bearer Token",
			"tags":        []string{"Core/Profile"},
			"security": []map[string][]string{
				{"ApiKeyAuth": {}},
				{"BearerAuth": {}},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "User profile",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/UserResponse",
							},
						},
					},
				},
				"400": map[string]any{
					"description": "Bad request",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"401": map[string]any{
					"description": "Unauthorized",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"404": map[string]any{
					"description": "User not found",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"500": map[string]any{
					"description": "Internal server error",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
			},
		},
		"put": map[string]any{
			"summary":     "Update profile",
			"description": "Update user profile information",
			"tags":        []string{"Core/Profile"},
			"security": []map[string][]string{
				{"ApiKeyAuth": {}},
				{"BearerAuth": {}},
			},
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"username": map[string]any{
									"type":        "string",
									"description": "Username",
									"example":     "johndoe",
								},
								"first_name": map[string]any{
									"type":        "string",
									"description": "First name",
									"example":     "John",
								},
								"last_name": map[string]any{
									"type":        "string",
									"description": "Last name",
									"example":     "Doe",
								},
								"email": map[string]any{
									"type":        "string",
									"format":      "email",
									"description": "Email address",
									"example":     "user@example.com",
								},
							},
						},
					},
				},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Profile updated successfully",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/UserResponse",
							},
						},
					},
				},
				"400": map[string]any{
					"description": "Bad request",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"401": map[string]any{
					"description": "Unauthorized",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
			},
		},
	}

	paths["/api/profile/avatar"] = map[string]any{
		"put": map[string]any{
			"summary":     "Update avatar",
			"description": "Update user profile avatar",
			"tags":        []string{"Core/Profile"},
			"security": []map[string][]string{
				{"ApiKeyAuth": {}},
				{"BearerAuth": {}},
			},
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					"multipart/form-data": map[string]any{
						"schema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"avatar": map[string]any{
									"type":        "string",
									"format":      "binary",
									"description": "Avatar image file",
								},
							},
							"required": []string{"avatar"},
						},
					},
				},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Avatar updated successfully",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/UserResponse",
							},
						},
					},
				},
				"400": map[string]any{
					"description": "Bad request",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"401": map[string]any{
					"description": "Unauthorized",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
			},
		},
	}

	paths["/api/profile/password"] = map[string]any{
		"put": map[string]any{
			"summary":     "Update password",
			"description": "Change user password",
			"tags":        []string{"Core/Profile"},
			"security": []map[string][]string{
				{"ApiKeyAuth": {}},
				{"BearerAuth": {}},
			},
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"current_password": map[string]any{
									"type":        "string",
									"description": "Current password",
								},
								"new_password": map[string]any{
									"type":        "string",
									"description": "New password",
								},
								"confirm_password": map[string]any{
									"type":        "string",
									"description": "Confirm new password",
								},
							},
							"required": []string{"current_password", "new_password", "confirm_password"},
						},
					},
				},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Password updated successfully",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.SuccessResponse",
							},
						},
					},
				},
				"400": map[string]any{
					"description": "Bad request",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"401": map[string]any{
					"description": "Unauthorized",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
			},
		},
	}
}

// addCoreMediaEndpoints adds Core/Media endpoints based on controller annotations
func (s *SwaggerService) addCoreMediaEndpoints(paths map[string]any) {
	paths["/api/media"] = map[string]any{
		"get": map[string]any{
			"summary":     "List media files",
			"description": "Get paginated list of media files",
			"tags":        []string{"Core/Media"},
			"security": []map[string][]string{
				{"ApiKeyAuth": {}},
				{"BearerAuth": {}},
			},
			"parameters": []map[string]any{
				{
					"name":        "page",
					"in":          "query",
					"description": "Page number",
					"required":    false,
					"schema":      map[string]any{"type": "integer"},
				},
				{
					"name":        "limit",
					"in":          "query",
					"description": "Items per page",
					"required":    false,
					"schema":      map[string]any{"type": "integer"},
				},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "List of media files",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.PaginatedResponse",
							},
						},
					},
				},
				"400": map[string]any{
					"description": "Bad request",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"500": map[string]any{
					"description": "Internal server error",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
			},
		},
		"post": map[string]any{
			"summary":     "Create a new media item",
			"description": "Create a new media item with optional file upload",
			"tags":        []string{"Core/Media"},
			"security": []map[string][]string{
				{"ApiKeyAuth": {}},
				{"BearerAuth": {}},
			},
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					"multipart/form-data": map[string]any{
						"schema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"name": map[string]any{
									"type":        "string",
									"description": "Media name",
								},
								"description": map[string]any{
									"type":        "string",
									"description": "Media description",
								},
								"file": map[string]any{
									"type":        "string",
									"format":      "binary",
									"description": "Media file",
								},
							},
							"required": []string{"name"},
						},
					},
				},
			},
			"responses": map[string]any{
				"201": map[string]any{
					"description": "Media item created successfully",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"id": map[string]any{
										"type": "integer",
									},
									"name": map[string]any{
										"type": "string",
									},
									"description": map[string]any{
										"type": "string",
									},
									"file_url": map[string]any{
										"type": "string",
									},
									"created_at": map[string]any{
										"type":   "string",
										"format": "date-time",
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
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"500": map[string]any{
					"description": "Internal server error",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
			},
		},
	}

	paths["/api/media/all"] = map[string]any{
		"get": map[string]any{
			"summary":     "List all media files",
			"description": "Get unpaginated list of all media files",
			"tags":        []string{"Core/Media"},
			"security": []map[string][]string{
				{"ApiKeyAuth": {}},
				{"BearerAuth": {}},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "List of all media files",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"type": "array",
								"items": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"id": map[string]any{
											"type": "integer",
										},
										"name": map[string]any{
											"type": "string",
										},
									},
								},
							},
						},
					},
				},
				"500": map[string]any{
					"description": "Internal server error",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
			},
		},
	}

	paths["/api/media/{id}"] = map[string]any{
		"get": map[string]any{
			"summary":     "Get media item",
			"description": "Get media item by ID",
			"tags":        []string{"Core/Media"},
			"security": []map[string][]string{
				{"ApiKeyAuth": {}},
				{"BearerAuth": {}},
			},
			"parameters": []map[string]any{
				{
					"name":        "id",
					"in":          "path",
					"description": "Media ID",
					"required":    true,
					"schema":      map[string]any{"type": "integer"},
				},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Media item details",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"id": map[string]any{
										"type": "integer",
									},
									"name": map[string]any{
										"type": "string",
									},
									"description": map[string]any{
										"type": "string",
									},
									"file_url": map[string]any{
										"type": "string",
									},
									"created_at": map[string]any{
										"type":   "string",
										"format": "date-time",
									},
								},
							},
						},
					},
				},
				"404": map[string]any{
					"description": "Media not found",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"500": map[string]any{
					"description": "Internal server error",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
			},
		},
		"put": map[string]any{
			"summary":     "Update media item",
			"description": "Update media item by ID",
			"tags":        []string{"Core/Media"},
			"security": []map[string][]string{
				{"ApiKeyAuth": {}},
				{"BearerAuth": {}},
			},
			"parameters": []map[string]any{
				{
					"name":        "id",
					"in":          "path",
					"description": "Media ID",
					"required":    true,
					"schema":      map[string]any{"type": "integer"},
				},
			},
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"name": map[string]any{
									"type":        "string",
									"description": "Media name",
								},
								"description": map[string]any{
									"type":        "string",
									"description": "Media description",
								},
							},
						},
					},
				},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Media item updated successfully",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"id": map[string]any{
										"type": "integer",
									},
									"name": map[string]any{
										"type": "string",
									},
									"description": map[string]any{
										"type": "string",
									},
									"file_url": map[string]any{
										"type": "string",
									},
									"updated_at": map[string]any{
										"type":   "string",
										"format": "date-time",
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
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"404": map[string]any{
					"description": "Media not found",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"500": map[string]any{
					"description": "Internal server error",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
			},
		},
		"delete": map[string]any{
			"summary":     "Delete media item",
			"description": "Delete media item by ID",
			"tags":        []string{"Core/Media"},
			"security": []map[string][]string{
				{"ApiKeyAuth": {}},
				{"BearerAuth": {}},
			},
			"parameters": []map[string]any{
				{
					"name":        "id",
					"in":          "path",
					"description": "Media ID",
					"required":    true,
					"schema":      map[string]any{"type": "integer"},
				},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Media item deleted successfully",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.SuccessResponse",
							},
						},
					},
				},
				"404": map[string]any{
					"description": "Media not found",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
				"500": map[string]any{
					"description": "Internal server error",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/types.ErrorResponse",
							},
						},
					},
				},
			},
		},
	}
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

								"username": map[string]any{
									"type":    "string",
									"example": "user",
								},
								"email": map[string]any{
									"type":    "string",
									"example": "user@example.com",
								},
								"password": map[string]any{
									"type":    "string",
									"example": "password123",
								},
								"first_name": map[string]any{
									"type":    "string",
									"example": "John",
								},
								"last_name": map[string]any{
									"type":    "string",
									"example": "Doe",
								},
							},
							"required": []string{"email", "password", "first_name", "last_name", "username"},
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

// GenerateStaticFiles generates static Swagger documentation files to the specified directory
func (s *SwaggerService) GenerateStaticFiles(docsDir string) error {
	// Create docs directory if it doesn't exist
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	// Generate swagger.json
	doc := s.GenerateSwaggerDoc()
	jsonData, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal swagger doc: %w", err)
	}

	jsonPath := filepath.Join(docsDir, "swagger.json")
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write swagger.json: %w", err)
	}

	// Generate docs.go file for swaggo compatibility
	docsGoContent := s.generateDocsGo(doc)
	docsGoPath := filepath.Join(docsDir, "docs.go")
	if err := os.WriteFile(docsGoPath, []byte(docsGoContent), 0644); err != nil {
		return fmt.Errorf("failed to write docs.go: %w", err)
	}

	// Generate swagger.yaml
	yamlContent := s.generateSwaggerYAML(doc)
	yamlPath := filepath.Join(docsDir, "swagger.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		return fmt.Errorf("failed to write swagger.yaml: %w", err)
	}

	fmt.Printf("Generated static Swagger documentation in %s/\n", docsDir)
	fmt.Printf("- swagger.json\n")
	fmt.Printf("- swagger.yaml\n")
	fmt.Printf("- docs.go\n")

	return nil
}

// generateDocsGo generates a docs.go file compatible with swaggo
func (s *SwaggerService) generateDocsGo(doc map[string]any) string {
	jsonBytes, _ := json.Marshal(doc)
	jsonStr := string(jsonBytes)

	return fmt.Sprintf(`// Package docs GENERATED BY BASE FRAMEWORK'S SWAGGER. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = %[1]q

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          %[2]q,
	Host:             %[3]q,
	BasePath:         %[4]q,
	Schemes:          []string{%[5]s},
	Title:            %[6]q,
	Description:      %[7]q,
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
`,
		jsonStr,
		s.info.Version,
		s.info.Host,
		s.info.BasePath,
		formatSchemes(s.info.Schemes),
		s.info.Title,
		s.info.Description,
	)
}

// generateSwaggerYAML generates YAML format of the OpenAPI spec
func (s *SwaggerService) generateSwaggerYAML(doc map[string]any) string {
	// Simple YAML generation - in production you might want to use a proper YAML library
	yamlBuilder := strings.Builder{}
	yamlBuilder.WriteString("openapi: \"3.0.3\"\n")

	if info, ok := doc["info"].(map[string]any); ok {
		yamlBuilder.WriteString("info:\n")
		if title, ok := info["title"].(string); ok {
			yamlBuilder.WriteString(fmt.Sprintf("  title: \"%s\"\n", title))
		}
		if description, ok := info["description"].(string); ok {
			yamlBuilder.WriteString(fmt.Sprintf("  description: \"%s\"\n", description))
		}
		if version, ok := info["version"].(string); ok {
			yamlBuilder.WriteString(fmt.Sprintf("  version: \"%s\"\n", version))
		}
	}

	if servers, ok := doc["servers"].([]map[string]any); ok && len(servers) > 0 {
		yamlBuilder.WriteString("servers:\n")
		for _, server := range servers {
			if url, ok := server["url"].(string); ok {
				yamlBuilder.WriteString(fmt.Sprintf("  - url: \"%s\"\n", url))
			}
			if desc, ok := server["description"].(string); ok {
				yamlBuilder.WriteString(fmt.Sprintf("    description: \"%s\"\n", desc))
			}
		}
	}

	yamlBuilder.WriteString("# Full OpenAPI spec available in swagger.json\n")
	return yamlBuilder.String()
}

// generateSchemas generates all the schema definitions for the OpenAPI spec
func (s *SwaggerService) generateSchemas() map[string]any {
	schemas := map[string]any{
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
		"types.ErrorResponse": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"error": map[string]any{
					"type":        "string",
					"description": "Error message",
				},
			},
			"required": []string{"error"},
		},
		"types.SuccessResponse": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"message": map[string]any{
					"type":        "string",
					"description": "Success message",
				},
			},
			"required": []string{"message"},
		},
		"types.PaginatedResponse": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"data": map[string]any{
					"type":        "array",
					"description": "List of items",
					"items": map[string]any{
						"type": "object",
					},
				},
				"pagination": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"total": map[string]any{
							"type":        "integer",
							"description": "Total number of items",
						},
						"page": map[string]any{
							"type":        "integer",
							"description": "Current page number",
						},
						"limit": map[string]any{
							"type":        "integer",
							"description": "Items per page",
						},
						"total_pages": map[string]any{
							"type":        "integer",
							"description": "Total number of pages",
						},
					},
				},
			},
			"required": []string{"data", "pagination"},
		},
	}

	// Dynamically generate schemas from models directory
	s.generateDynamicSchemas(schemas)

	return schemas
}

// generateDynamicSchemas automatically discovers and generates schemas from models
func (s *SwaggerService) generateDynamicSchemas(schemas map[string]any) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		// Fallback to hardcoded schemas if we can't get working directory
		s.addPostSchemas(schemas)
		s.addUserSchemas(schemas)
		return
	}

	modelsDir := filepath.Join(cwd, "app", "models")

	// Read all .go files in the models directory
	files, err := filepath.Glob(filepath.Join(modelsDir, "*.go"))
	if err != nil {
		// Fallback to hardcoded schemas
		s.addPostSchemas(schemas)
		s.addUserSchemas(schemas)
		return
	}

	// Parse each model file and extract struct definitions
	for _, file := range files {
		s.parseModelFileSimple(file, schemas)
	}
}

// parseModelFileSimple extracts model names from Go files and generates standard schemas
func (s *SwaggerService) parseModelFileSimple(filename string, schemas map[string]any) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	fileContent := string(content)

	// Extract base model name from filename (e.g., post.go -> Post, post_categor.go -> PostCategor)
	baseName := filepath.Base(filename)
	baseName = strings.TrimSuffix(baseName, ".go")

	// Convert snake_case to PascalCase (post_categor -> PostCategor)
	modelName := s.snakeCaseToPascalCase(baseName)

	// Look for Request/Response struct definitions in the file
	if strings.Contains(fileContent, "Request struct") {
		s.generateStandardSchemas(modelName, schemas)
	}
}

// snakeCaseToPascalCase converts snake_case to PascalCase
func (s *SwaggerService) snakeCaseToPascalCase(input string) string {
	parts := strings.Split(input, "_")
	result := ""
	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return result
}

// generateStandardSchemas generates standard CRUD schemas for a model
func (s *SwaggerService) generateStandardSchemas(modelName string, schemas map[string]any) {
	// Generate CreateRequest schema
	schemas["models.Create"+modelName+"Request"] = map[string]any{
		"type":       "object",
		"properties": s.generateModelProperties(),
		"required":   []string{"tile", "desc", "date"},
	}

	// Generate UpdateRequest schema
	schemas["models.Update"+modelName+"Request"] = map[string]any{
		"type":       "object",
		"properties": s.generateModelProperties(),
	}

	// Generate Response schema
	schemas["models."+modelName+"Response"] = map[string]any{
		"type":       "object",
		"properties": s.generateResponseProperties(),
		"required":   []string{"id", "created_at", "updated_at"},
	}

	// Generate SelectOption schema
	schemas["models."+modelName+"SelectOption"] = map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id": map[string]any{
				"type":        "integer",
				"description": "Unique identifier",
				"example":     1,
			},
			"name": map[string]any{
				"type":        "string",
				"description": "Display name",
				"example":     "Sample Name",
			},
		},
		"required": []string{"id", "name"},
	}
}

// generateModelProperties generates standard model properties
func (s *SwaggerService) generateModelProperties() map[string]any {
	return map[string]any{
		"tile": map[string]any{
			"type":        "string",
			"description": "Title",
			"example":     "Sample Title",
		},
		"desc": map[string]any{
			"type":        "string",
			"description": "Description",
			"example":     "Sample description",
		},
		"date": map[string]any{
			"type":        "string",
			"description": "Date",
			"example":     "2024-01-15",
		},
		"nite": map[string]any{
			"type":        "string",
			"description": "Notes",
			"example":     "Additional notes",
		},
	}
}

// generateResponseProperties generates standard response properties
func (s *SwaggerService) generateResponseProperties() map[string]any {
	properties := s.generateModelProperties()

	// Add standard response fields
	properties["id"] = map[string]any{
		"type":        "integer",
		"description": "Unique identifier",
		"example":     1,
	}
	properties["created_at"] = map[string]any{
		"type":        "string",
		"format":      "date-time",
		"description": "Creation timestamp",
		"example":     "2024-01-15T10:30:00Z",
	}
	properties["updated_at"] = map[string]any{
		"type":        "string",
		"format":      "date-time",
		"description": "Last update timestamp",
		"example":     "2024-01-15T10:30:00Z",
	}
	properties["deleted_at"] = map[string]any{
		"type":        "string",
		"format":      "date-time",
		"description": "Deletion timestamp",
		"example":     nil,
	}

	return properties
}

// addPostSchemas adds Post-related schemas to the schema map
func (s *SwaggerService) addPostSchemas(schemas map[string]any) {
	// CreatePostRequest schema
	schemas["models.CreatePostRequest"] = map[string]any{
		"type": "object",
		"properties": map[string]any{
			"tile": map[string]any{
				"type":        "string",
				"description": "Title of the post",
				"example":     "My Blog Post",
			},
			"desc": map[string]any{
				"type":        "string",
				"description": "Description of the post",
				"example":     "This is a sample blog post description",
			},
			"date": map[string]any{
				"type":        "string",
				"description": "Date of the post",
				"example":     "2024-01-15",
			},
		},
		"required": []string{"tile", "desc", "date"},
	}

	// UpdatePostRequest schema
	schemas["models.UpdatePostRequest"] = map[string]any{
		"type": "object",
		"properties": map[string]any{
			"tile": map[string]any{
				"type":        "string",
				"description": "Title of the post",
				"example":     "Updated Blog Post",
			},
			"desc": map[string]any{
				"type":        "string",
				"description": "Description of the post",
				"example":     "This is an updated blog post description",
			},
			"date": map[string]any{
				"type":        "string",
				"description": "Date of the post",
				"example":     "2024-02-20",
			},
		},
	}

	// PostResponse schema
	schemas["models.PostResponse"] = map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id": map[string]any{
				"type":        "integer",
				"description": "Unique identifier",
				"example":     1,
			},
			"created_at": map[string]any{
				"type":        "string",
				"format":      "date-time",
				"description": "Creation timestamp",
				"example":     "2024-01-15T10:30:00Z",
			},
			"updated_at": map[string]any{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2024-01-15T10:30:00Z",
			},
			"deleted_at": map[string]any{
				"type":        "string",
				"format":      "date-time",
				"description": "Deletion timestamp (null if not deleted)",
				"example":     nil,
				"nullable":    true,
			},
			"tile": map[string]any{
				"type":        "string",
				"description": "Title of the post",
				"example":     "My Blog Post",
			},
			"desc": map[string]any{
				"type":        "string",
				"description": "Description of the post",
				"example":     "This is a sample blog post description",
			},
			"date": map[string]any{
				"type":        "string",
				"description": "Date of the post",
				"example":     "2024-01-15",
			},
		},
		"required": []string{"id", "created_at", "updated_at", "tile", "desc", "date"},
	}

	// PostSelectOption schema
	schemas["models.PostSelectOption"] = map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id": map[string]any{
				"type":        "integer",
				"description": "Unique identifier",
				"example":     1,
			},
			"name": map[string]any{
				"type":        "string",
				"description": "Display name for select options",
				"example":     "My Blog Post",
			},
		},
		"required": []string{"id", "name"},
	}
}

// addUserSchemas adds User-related schemas for authentication
func (s *SwaggerService) addUserSchemas(schemas map[string]any) {
	// LoginRequest schema (used by Auth controller)
	schemas["LoginRequest"] = map[string]any{
		"type": "object",
		"properties": map[string]any{
			"email": map[string]any{
				"type":        "string",
				"format":      "email",
				"description": "User email address",
				"example":     "user@example.com",
			},
			"password": map[string]any{
				"type":        "string",
				"description": "User password",
				"example":     "password123",
			},
		},
		"required": []string{"email", "password"},
	}

	// RegisterRequest schema
	schemas["RegisterRequest"] = map[string]any{
		"type": "object",
		"properties": map[string]any{
			"username": map[string]any{
				"type":        "string",
				"description": "Username",
				"example":     "johndoe",
			},
			"email": map[string]any{
				"type":        "string",
				"format":      "email",
				"description": "User email address",
				"example":     "user@example.com",
			},
			"password": map[string]any{
				"type":        "string",
				"description": "User password",
				"example":     "password123",
			},
			"first_name": map[string]any{
				"type":        "string",
				"description": "First name",
				"example":     "John",
			},
			"last_name": map[string]any{
				"type":        "string",
				"description": "Last name",
				"example":     "Doe",
			},
		},
		"required": []string{"email", "password", "first_name", "last_name", "username"},
	}

	// AuthResponse schema
	schemas["AuthResponse"] = map[string]any{
		"type": "object",
		"properties": map[string]any{
			"user": map[string]any{
				"$ref": "#/components/schemas/UserResponse",
			},
			"token": map[string]any{
				"type":        "string",
				"description": "JWT authentication token",
				"example":     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			},
		},
		"required": []string{"user", "token"},
	}

	// UserResponse schema
	schemas["UserResponse"] = map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id": map[string]any{
				"type":        "integer",
				"description": "Unique identifier",
				"example":     1,
			},
			"username": map[string]any{
				"type":        "string",
				"description": "Username",
				"example":     "johndoe",
			},
			"email": map[string]any{
				"type":        "string",
				"format":      "email",
				"description": "User email address",
				"example":     "user@example.com",
			},
			"first_name": map[string]any{
				"type":        "string",
				"description": "First name",
				"example":     "John",
			},
			"last_name": map[string]any{
				"type":        "string",
				"description": "Last name",
				"example":     "Doe",
			},
			"created_at": map[string]any{
				"type":        "string",
				"format":      "date-time",
				"description": "Creation timestamp",
				"example":     "2024-01-15T10:30:00Z",
			},
			"updated_at": map[string]any{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2024-01-15T10:30:00Z",
			},
		},
		"required": []string{"id", "username", "email", "first_name", "last_name", "created_at", "updated_at"},
	}
}

// formatSchemes formats schemes array for Go code generation
func formatSchemes(schemes []string) string {
	if len(schemes) == 0 {
		return ""
	}
	quoted := make([]string, len(schemes))
	for i, scheme := range schemes {
		quoted[i] = fmt.Sprintf("\"%s\"", scheme)
	}
	return strings.Join(quoted, ", ")
}
