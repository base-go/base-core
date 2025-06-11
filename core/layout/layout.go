package layout

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

// Engine is the main template engine
type Engine struct {
	templates    *template.Template
	templatesDir string
	layoutsDir   string
	sharedDir    string
	funcMap      template.FuncMap
}

// Config holds configuration for the template engine
type Config struct {
	TemplatesDir string
	LayoutsDir   string
	SharedDir    string
}

// NewEngine creates a new template engine
func NewEngine(config Config) *Engine {
	engine := &Engine{
		templatesDir: config.TemplatesDir,
		layoutsDir:   config.LayoutsDir,
		sharedDir:    config.SharedDir,
		funcMap:      make(template.FuncMap),
	}

	engine.addDefaultHelpers()
	return engine
}

func (e *Engine) addDefaultHelpers() {
	e.funcMap["yield"] = func() string {
		return "{{.Content}}"
	}

	e.funcMap["asset_path"] = func(path string) string {
		return "/static/" + path
	}

	e.funcMap["csrf_token"] = func() string {
		return "{{.CSRFToken}}"
	}
}

// AddHelper adds a template helper function
func (e *Engine) AddHelper(name string, fn any) {
	e.funcMap[name] = fn

	// If templates object already exists, recreate it with the updated funcMap
	if e.templates != nil {
		e.templates = template.New("").Funcs(e.funcMap)
	}
}

// ParseString parses a template string and registers it with the engine
func (e *Engine) ParseString(name string, content string) error {
	if e.templates == nil {
		e.templates = template.New("").Funcs(e.funcMap)
	}

	_, err := e.templates.New(name).Funcs(e.funcMap).Parse(content)
	return err
}

func (e *Engine) ReloadTemplates() error {
	return e.loadTemplates()
}

func (e *Engine) loadTemplates() error {
	e.templates = template.New("").Funcs(e.funcMap)

	var allFiles []string

	// Collect all template files
	if e.sharedDir != "" {
		sharedPattern := filepath.Join(e.sharedDir, "*.html")
		if files, err := filepath.Glob(sharedPattern); err == nil {
			allFiles = append(allFiles, files...)
		}
	}

	if e.layoutsDir != "" {
		layoutPattern := filepath.Join(e.layoutsDir, "*.html")
		if files, err := filepath.Glob(layoutPattern); err == nil {
			allFiles = append(allFiles, files...)
		}
	}

	if e.templatesDir != "" {
		// Load all templates recursively from templatesDir (theme directory)
		err := filepath.Walk(e.templatesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".html") {
				allFiles = append(allFiles, path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Parse shared and layout files with proper naming
	if len(allFiles) > 0 {

		for _, file := range allFiles {

			// Read the file content
			content, readErr := os.ReadFile(file)
			if readErr != nil {

				continue
			}

			// Determine template name based on file path
			var templateName string
			if e.templatesDir != "" && strings.HasPrefix(file, e.templatesDir) {
				// For files in templatesDir, use relative path from templatesDir
				relPath, err := filepath.Rel(e.templatesDir, file)
				if err == nil {
					templateName = filepath.ToSlash(relPath)
				} else {
					templateName = filepath.Base(file)
				}
			} else {
				// For layout and shared files, just use the base name
				templateName = filepath.Base(file)
			}

			// Parse and register with the determined name
			_, parseErr := e.templates.New(templateName).Parse(string(content))
			if parseErr != nil {

				continue
			}

		}
	}

	// Automatically discover and load module-specific view templates with proper naming
	// Walk through app/ directory to find all */views directories
	err := filepath.Walk("app", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Look for directories named "views"
		if info.IsDir() && filepath.Base(path) == "views" {
			// Extract module name from path (e.g., "app/posts/views" -> "posts")
			pathParts := strings.Split(path, string(filepath.Separator))
			var moduleName string
			if len(pathParts) >= 2 {
				moduleName = pathParts[len(pathParts)-2] // Get the directory before "views"
			}

			// Load all .html files from this views directory
			if viewFiles, globErr := filepath.Glob(filepath.Join(path, "*.html")); globErr == nil {
				for _, viewFile := range viewFiles {
					// Read the file content
					content, readErr := os.ReadFile(viewFile)
					if readErr != nil {

						continue
					}

					// Create a template name like "posts/index.html"
					filename := filepath.Base(viewFile)
					templateName := filepath.Join(moduleName, filename)

					// Parse and register with the module-prefixed name
					_, parseErr := e.templates.New(templateName).Parse(string(content))
					if parseErr != nil {

						continue
					}

				}
			}
		}
		return nil
	})
	if err != nil {

	}

	return nil
}

func (e *Engine) Render(w io.Writer, name string, data any, ctx *gin.Context) error {
	// Check if templates are loaded
	if e.templates == nil {
		return fmt.Errorf("templates not loaded")
	}

	// Convert data to map if it's not already
	var templateData map[string]any

	switch d := data.(type) {
	case map[string]any:
		templateData = d
	case gin.H:
		templateData = map[string]any(d)
	default:
		templateData = map[string]any{
			"Data": data,
		}
	}

	// Add context data
	if ctx != nil {
		templateData["Request"] = ctx.Request
		templateData["Context"] = ctx

		// Make sure the TranslationService and language are available in the template data
		// This way, template helpers can access them directly without needing the context
		translationService, tsExists := ctx.Get("TranslationService")
		if tsExists {
			templateData["_ts"] = translationService // Use a simple key that's unlikely to conflict
		}

		language, langExists := ctx.Get("language")
		if langExists {
			templateData["_lang"] = language // Use a simple key that's unlikely to conflict
		}
	}

	err := e.templates.ExecuteTemplate(w, name, templateData)
	if err != nil {
		return err
	}
	return nil
}

func (e *Engine) RenderWithLayout(w io.Writer, templateName, layoutName string, data any, ctx *gin.Context) error {

	// Convert data to map if it's not already
	var templateData map[string]any

	switch d := data.(type) {
	case map[string]any:
		templateData = d
	case gin.H:
		templateData = map[string]any(d)
	default:
		templateData = map[string]any{
			"Data": data,
		}
	}

	// Add context data
	if ctx != nil {
		templateData["Request"] = ctx.Request
		templateData["Context"] = ctx

		// Make sure the TranslationService and language are available in the template data
		// This way, template helpers can access them directly without needing the context
		translationService, tsExists := ctx.Get("TranslationService")
		if tsExists {
			templateData["_ts"] = translationService // Use a simple key that's unlikely to conflict
		}

		language, langExists := ctx.Get("language")
		if langExists {
			templateData["_lang"] = language // Use a simple key that's unlikely to conflict
		}

		// Add authentication status
		session := sessions.Default(ctx)
		userID := session.Get("user_id")
		username := session.Get("username")

		if userID != nil {
			templateData["logged_in"] = true
			templateData["user_id"] = userID

			if username != nil {
				templateData["username"] = username
			}
		} else {
			// Check for token in cookie as fallback
			token, err := ctx.Cookie("auth_token")
			if err == nil && token != "" {
				// If we have a token, consider the user logged in
				templateData["logged_in"] = true
			} else {
				templateData["logged_in"] = false
			}
		}
	}

	// Render the main template to capture its content

	var contentBuilder strings.Builder
	if err := e.templates.ExecuteTemplate(&contentBuilder, templateName, templateData); err != nil {

		return err
	}

	content := contentBuilder.String()

	// Add the rendered content to template data
	templateData["Content"] = content

	// Render with layout

	err := e.templates.ExecuteTemplate(w, layoutName, templateData)
	if err != nil {

	} else {

	}
	return err
}

func (e *Engine) Instance(name string, data any) render.Render {
	return &TemplateRender{
		engine: e,
		name:   name,
		data:   data,
	}
}

// Implement render.HTMLRender interface
func (e *Engine) LoadHTMLGlob(pattern string) {
	// This is called by Gin but we load templates differently
}

func (e *Engine) LoadHTMLFiles(files ...string) {
	// This is called by Gin but we load templates differently
}

type TemplateRender struct {
	engine  *Engine
	name    string
	data    any
	context *gin.Context
}

func (r *TemplateRender) Render(w http.ResponseWriter) error {
	return r.engine.Render(w, r.name, r.data, r.context)
}

func (r *TemplateRender) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"text/html; charset=utf-8"}
	}
}
