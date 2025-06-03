package template

import (
	"html/template"
)

// ParseString parses a template string with the given name and registers it with the engine
func (e *Engine) ParseString(name string, content string) error {

	if e.templates == nil {
		e.templates = template.New("").Funcs(e.funcMap)
	}

	// Always apply the current funcMap when creating new templates
	// This ensures that functions added after initialization are available
	_, err := e.templates.New(name).Funcs(e.funcMap).Parse(content)
	return err
}

// ParseStringTemplate parses a template string with the given name
// This is an alias for ParseString for backward compatibility
func (e *Engine) ParseStringTemplate(name string, content string) error {
	return e.ParseString(name, content)
}
