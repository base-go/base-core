package layout

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// RegisterDefaultHelpers registers essential template helper functions
func (e *Engine) RegisterDefaultHelpers() {
	// Initialize and register component system
	e.RegisterComponentHelper()
	
	// Load .bui components from theme directory
	if e.templatesDir != "" {
		componentsDir := filepath.Join(e.templatesDir, "components")
		if err := e.componentRegistry.LoadComponentsFromDirectory(componentsDir); err != nil {
			fmt.Printf("WARNING: Failed to load .bui components from %s: %v\n", componentsDir, err)
		}
	}
	
	// Note: UI components are registered separately via app/ui.RegisterUIComponents
	// Essential string helpers
	e.AddHelper("upper", strings.ToUpper)
	e.AddHelper("lower", strings.ToLower)
	e.AddHelper("title", cases.Title(language.AmericanEnglish).String)
	e.AddHelper("truncate", func(s string, length int) string {
		if len(s) <= length {
			return s
		}
		return s[:length] + "..."
	})

	// HTML safety helpers
	e.AddHelper("safe", func(s string) template.HTML {
		return template.HTML(s)
	})
	e.AddHelper("escape", template.HTMLEscapeString)

	// Basic URL helper
	e.AddHelper("url_for", func(path string) string {
		return path
	})

	// Essential form helpers
	e.AddHelper("link_to", func(text, url string) template.HTML {
		return template.HTML(fmt.Sprintf(`<a href="%s">%s</a>`, url, text))
	})

	// Time helpers
	e.AddHelper("time_ago", func(t time.Time) string {
		return time.Since(t).String() + " ago"
	})

	e.AddHelper("format_time", func(t time.Time, layout string) string {
		return t.Format(layout)
	})

	// Conditional helpers
	e.AddHelper("eq", func(a, b interface{}) bool {
		return a == b
	})

	e.AddHelper("ne", func(a, b interface{}) bool {
		return a != b
	})

	// Asset helpers
	e.AddHelper("asset_path", func(path string) string {
		return "/static/" + path
	})

	e.AddHelper("css", func(href string) template.HTML {
		return template.HTML(fmt.Sprintf(`<link rel="stylesheet" href="%s">`, href))
	})

	e.AddHelper("js", func(src string) template.HTML {
		return template.HTML(fmt.Sprintf(`<script src="%s"></script>`, src))
	})

	// Component helper functions
	e.AddHelper("dict", func(pairs ...interface{}) map[string]interface{} {
		dict := make(map[string]interface{})
		for i := 0; i < len(pairs); i += 2 {
			if i+1 < len(pairs) {
				if key, ok := pairs[i].(string); ok {
					dict[key] = pairs[i+1]
				}
			}
		}
		return dict
	})

	e.AddHelper("list", func(items ...interface{}) []interface{} {
		return items
	})
}
