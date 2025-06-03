package template

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

type Engine struct {
	templates    *template.Template
	templatesDir string
	layoutsDir   string
	sharedDir    string
	funcMap      template.FuncMap
}

type Config struct {
	TemplatesDir string
	LayoutsDir   string
	SharedDir    string
}

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

func (e *Engine) AddHelper(name string, fn any) {

	e.funcMap[name] = fn

	// If templates object already exists, recreate it with the updated funcMap
	if e.templates != nil {
		e.templates = template.New("").Funcs(e.funcMap)
	}
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
		// Load templates from main directory
		mainPattern := filepath.Join(e.templatesDir, "*.html")
		if files, err := filepath.Glob(mainPattern); err == nil {
			allFiles = append(allFiles, files...)
		}

		// Load templates from subdirectories
		postsPattern := filepath.Join(e.templatesDir, "posts", "*.html")
		if files, err := filepath.Glob(postsPattern); err == nil {
			allFiles = append(allFiles, files...)
		}
	}

	// Parse all files at once
	if len(allFiles) > 0 {
		if _, err := e.templates.ParseFiles(allFiles...); err != nil {
			return err
		}
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
	}

	// Render the main template to capture its content
	var contentBuilder strings.Builder
	if err := e.templates.ExecuteTemplate(&contentBuilder, templateName, templateData); err != nil {
		return err
	}

	// Add the rendered content to template data
	templateData["Content"] = contentBuilder.String()

	// Render with layout
	return e.templates.ExecuteTemplate(w, layoutName, templateData)
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
