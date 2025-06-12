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
	templates         *template.Template
	templatesDir      string
	layoutsDir        string
	sharedDir         string
	funcMap           template.FuncMap
	componentRegistry *ComponentRegistry
	componentParser   *ComponentParser
}

// Config holds configuration for the template engine
type Config struct {
	TemplatesDir string
	LayoutsDir   string
	SharedDir    string
}

// NewEngine creates a new template engine and loads templates
func NewEngine(config Config) *Engine {
	engine := NewEngineWithoutLoading(config)

	// Load templates
	if err := engine.LoadTemplates(); err != nil {
		fmt.Printf("ERROR: NewEngine - Failed to load templates: %v\n", err)
	}

	return engine
}

// NewEngineWithoutLoading creates a new template engine without loading templates
func NewEngineWithoutLoading(config Config) *Engine {
	engine := &Engine{
		templatesDir: config.TemplatesDir,
		layoutsDir:   config.LayoutsDir,
		sharedDir:    config.SharedDir,
		funcMap:      make(template.FuncMap),
	}

	engine.addDefaultHelpers()

	// Initialize component registry
	engine.componentRegistry = NewComponentRegistry(engine)

	// Initialize component parser
	engine.componentParser = NewComponentParser(engine.componentRegistry)

	return engine
}

// LoadTemplates loads all templates (public method)
func (e *Engine) LoadTemplates() error {
	return e.loadTemplates()
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

	// Preprocess the template to transform custom component syntax
	processedContent, err := e.PreprocessTemplate(content)
	if err != nil {
		return fmt.Errorf("error preprocessing template %s: %w", name, err)
	}

	_, err = e.templates.New(name).Funcs(e.funcMap).Parse(processedContent)
	return err
}

// PreprocessTemplate transforms custom component syntax to standard Go template syntax
func (e *Engine) PreprocessTemplate(content string) (string, error) {
	if e.componentParser == nil {
		e.componentParser = NewComponentParser(e.componentRegistry)
	}

	return e.componentParser.ParseTemplate(content)
}

func (e *Engine) ReloadTemplates() error {
	return e.loadTemplates()
}

// GetComponentRegistry returns the component registry for this engine
func (e *Engine) GetComponentRegistry() *ComponentRegistry {
	return e.componentRegistry
}

// HasTemplate checks if a template with the given name exists
func (e *Engine) HasTemplate(name string) bool {
	if e.templates == nil {
		return false
	}
	return e.templates.Lookup(name) != nil
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
			if strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".bui") {
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
			if e.layoutsDir != "" && strings.HasPrefix(file, e.layoutsDir) {
				// For layout files, just use the base name
				templateName = filepath.Base(file)
			} else if e.sharedDir != "" && strings.HasPrefix(file, e.sharedDir) {
				// For shared files, just use the base name
				templateName = filepath.Base(file)
			} else if e.templatesDir != "" && strings.HasPrefix(file, e.templatesDir) {
				// For other files in templatesDir, use relative path from templatesDir
				relPath, err := filepath.Rel(e.templatesDir, file)
				if err == nil {
					templateName = filepath.ToSlash(relPath)
				} else {
					templateName = filepath.Base(file)
				}
			} else {
				// Fallback to base name
				templateName = filepath.Base(file)
			}

			// Debug: Show template content before preprocessing
			if strings.Contains(templateName, "landing.html") {
				previewLen := 200
				if len(content) < previewLen {
					previewLen = len(content)
				}
				fmt.Printf("DEBUG: loadTemplates - Template %s content preview: %s\n", templateName, string(content[:previewLen]))
			}

			// Preprocess the template to transform custom component syntax
			processedContent, err := e.PreprocessTemplate(string(content))
			if err != nil {
				fmt.Printf("ERROR: loadTemplates - Failed to preprocess template %s: %v\n", templateName, err)
				continue
			}

			// Debug: Show if preprocessing changed anything
			if strings.Contains(templateName, "landing.html") && string(content) != processedContent {
				fmt.Printf("DEBUG: loadTemplates - Template %s was modified during preprocessing\n", templateName)
			}

			// Parse and register with the determined name
			fmt.Printf("DEBUG: loadTemplates - Registering template: '%s' from file: %s\n", templateName, file)

			_, err = e.templates.New(templateName).Parse(processedContent)
			if err != nil {
				fmt.Printf("ERROR: loadTemplates - Failed to parse template '%s': %v\n", templateName, err)
				continue
			}

		}
	}

	fmt.Printf("DEBUG: loadTemplates - Loaded %d files so far\n", len(allFiles))
	for _, file := range allFiles {
		fmt.Printf("DEBUG: loadTemplates - File: %s\n", file)
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
					fmt.Printf("DEBUG: loadTemplates - Registering module template: '%s' from file: %s\n", templateName, viewFile)

					// Process the template content with the component parser if available
					templateContent := string(content)
					if e.componentParser != nil {
						processedContent, err := e.componentParser.ParseTemplate(templateContent)
						if err != nil {
							fmt.Printf("WARNING: loadTemplates - Component parsing for module template '%s' had issues: %v\n", templateName, err)
							// Continue with the original content if there's an error
						} else {
							templateContent = processedContent
						}
					}

					// Parse and register with the module-prefixed name
					_, parseErr := e.templates.New(templateName).Parse(templateContent)
					if parseErr != nil {
						fmt.Printf("ERROR: loadTemplates - Failed to parse module template '%s': %v\n", templateName, parseErr)
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
	fmt.Printf("DEBUG: RenderWithLayout - templateName: %s, layoutName: %s\n", templateName, layoutName)

	// Check if templates are loaded
	if e.templates == nil {
		fmt.Printf("ERROR: RenderWithLayout - templates not loaded\n")
		return fmt.Errorf("templates not loaded")
	}

	// Check if the specific template exists
	tmpl := e.templates.Lookup(templateName)
	if tmpl == nil {
		fmt.Printf("ERROR: RenderWithLayout - template '%s' not found\n", templateName)
		// List available templates for debugging
		if e.templates != nil {
			fmt.Printf("DEBUG: Available templates: ")
			for _, t := range e.templates.Templates() {
				fmt.Printf("%s ", t.Name())
			}
			fmt.Printf("\n")
		}
		return fmt.Errorf("template '%s' not found", templateName)
	}

	// Check if the layout exists
	layoutTmpl := e.templates.Lookup(layoutName)
	if layoutTmpl == nil {
		fmt.Printf("ERROR: RenderWithLayout - layout '%s' not found\n", layoutName)
		return fmt.Errorf("layout '%s' not found", layoutName)
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
