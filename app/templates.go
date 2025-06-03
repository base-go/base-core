package app

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"

	"base/app/templates"
	"base/core/template"
)

// Re-export template constants from the templates package
const (
	// Layouts
	LayoutDefault = templates.LayoutDefault
	LayoutLanding = templates.LayoutLanding
	LayoutAuth    = templates.LayoutAuth

	// Common pages
	PageError   = templates.PageError
	PageLanding = templates.PageLanding

	// Auth pages
	PageRegister       = templates.PageRegister
	PageLogin          = templates.PageLogin
	PageLogout         = templates.PageLogout
	PageForgotPassword = templates.PageForgotPassword
	PageResetPassword  = templates.PageResetPassword

	// Post templates
	PostIndex = templates.PostIndex
	PostShow  = templates.PostShow
	PostEdit  = templates.PostEdit
	PostNew   = templates.PostNew
)

//go:embed theme/default/*.html theme/default/**/*.html
var embeddedTemplates embed.FS

// RegisterTemplates loads all templates from the embedded filesystem
// and registers them with the template engine
func RegisterTemplates(engine *template.Engine) error {

	// Register template helpers
	registerHelpers(engine)

	// Don't initialize templates yet - wait until all helpers are registered

	// Debug: Test if embed is working at all

	// Try to read a specific file
	if _, err := embeddedTemplates.ReadFile("theme/default/layouts/landing.html"); err != nil {

	}
	// Debug: List all embedded files first

	walkErr := fs.WalkDir(embeddedTemplates, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {

			return err
		}

		return nil
	})
	if walkErr != nil {

		return walkErr
	}

	// Initialize templates object now that all helpers are registered

	if err := engine.ParseString("_init", ""); err != nil {

		return err
	}

	// Register all templates now that helpers are available

	return fs.WalkDir(embeddedTemplates, "theme/default", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".html") {
			data, err := embeddedTemplates.ReadFile(path)
			if err != nil {
				return err
			}

			rel, _ := filepath.Rel("theme/default", path)
			templateName := rel
			baseName := filepath.Base(path)

			if err := engine.ParseString(templateName, string(data)); err != nil {

				return err
			}

			// Only register basename alias if it doesn't conflict with existing full templates
			if baseName != templateName && !strings.Contains(baseName, "landing.html") {
				if err := engine.ParseString(baseName, string(data)); err != nil {

					return err
				}
			}
		}
		return nil
	})
}

// registerHelpers registers template helper functions
func registerHelpers(engine *template.Engine) {
	// Add custom helpers beyond the default ones
	engine.AddHelper("partial", func(name string, data any) string {
		return "{{template \"" + name + "\" .}}"
	})

	// Add more helpers as needed
}
