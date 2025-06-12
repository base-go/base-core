package app

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"

	"base/ui"
	"base/core/layout"
)

// Re-export template constants from the templates package
const (
	// Layouts
	LayoutDefault = "app.html"
	LayoutLanding = "landing.html"
	LayoutAuth    = "auth.html"

	// Common pages
	PageError   = "error.html"
	PageLanding = "landing.html"

	// Auth pages
	PageRegister       = "auth/register.html"
	PageLogin          = "auth/login.html"
	PageLogout         = "auth/logout.html"
	PageForgotPassword = "auth/forgot-password.html"
	PageResetPassword  = "auth/reset-password.html"

	// Post templates
	PostIndex = "posts/index.html"
	PostShow  = "posts/show.html"
	PostEdit  = "posts/edit.html"
	PostNew   = "posts/new.html"
)

//go:embed theme/default/*.html theme/default/**/*.html
var embeddedTemplates embed.FS

// RegisterTemplates loads all templates from the embedded filesystem
// and registers them with the template engine
func RegisterTemplates(engine *layout.Engine) error {
	// Register UI components first
	if engine.GetComponentRegistry() != nil {
		ui.RegisterUIComponents(engine.GetComponentRegistry())
	}
	
	// Register template helpers
	registerHelpers(engine)

	// Initialize templates object now that all helpers are registered
	if err := engine.ParseString("_init", ""); err != nil {
		return err
	}

	// Load templates from filesystem first, which includes proper component preprocessing
	if err := engine.ReloadTemplates(); err != nil {
		return err
	}

	// Register embedded templates as fallbacks for any missing templates
	err := fs.WalkDir(embeddedTemplates, "theme/default", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && (strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".bui")) {
			rel, _ := filepath.Rel("theme/default", path)
			templateName := rel
			baseName := filepath.Base(path)

			// Only register embedded template if filesystem version doesn't exist
			if !engine.HasTemplate(templateName) && !engine.HasTemplate(baseName) {
				data, err := embeddedTemplates.ReadFile(path)
				if err != nil {
					return err
				}

				if err := engine.ParseString(templateName, string(data)); err != nil {
					return err
				}

				// Only register basename alias if it doesn't conflict with existing templates
				if baseName != templateName && !strings.Contains(baseName, "landing.html") {
					if err := engine.ParseString(baseName, string(data)); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})

	return err
}

// registerHelpers registers template helper functions
func registerHelpers(engine *layout.Engine) {
	// Add custom helpers beyond the default ones
	engine.AddHelper("partial", func(name string, data any) string {
		return "{{template \"" + name + "\" .}}"
	})

	// Add more helpers as needed
}
