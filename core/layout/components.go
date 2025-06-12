package layout

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Component represents a reusable component with both template and reactive behavior
type Component struct {
	Name         string                 // Component name (e.g., "Card", "Button")
	Template     string                 // HTML template content
	Props        map[string]interface{} // Default props
	StateKeys    []string               // Keys for reactive state
	Methods      []string               // Method names for Petite-Vue
	StyleClasses string                 // Default CSS classes
}

// ComponentRegistry manages all available components
type ComponentRegistry struct {
	components map[string]*Component
	engine     *Engine
}

// ComponentParser handles parsing of custom component syntax
type ComponentParser struct {
	componentRegistry *ComponentRegistry
}

// NewComponentParser creates a new component parser
func NewComponentParser(registry *ComponentRegistry) *ComponentParser {
	return &ComponentParser{
		componentRegistry: registry,
	}
}

// ParseTemplate transforms custom component syntax to Go template syntax
func (p *ComponentParser) ParseTemplate(input string) (string, error) {
	// Pattern to match <{ ComponentName prop="value" prop2=value2 }> or <{ ComponentName prop->"value" prop2->value2 }>
	// This regex captures component tags with their attributes
	pattern := regexp.MustCompile(`<\{\s*(\w+)\s*([^}]*)\s*\}>`) // Escaped for string literal

	fmt.Printf("DEBUG: ParseTemplate - Input contains %d component-like patterns\n", len(pattern.FindAllString(input, -1)))
	if len(pattern.FindAllString(input, -1)) > 0 {
		fmt.Printf("DEBUG: ParseTemplate - Found patterns: %v\n", pattern.FindAllString(input, -1))
	}

	result := pattern.ReplaceAllStringFunc(input, func(match string) string {
		// Extract component name and attributes
		submatches := pattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			fmt.Printf("DEBUG: ParseTemplate - Failed to parse match: %s\n", match)
			return match // Return original if parsing fails
		}

		componentName := submatches[1]
		attributesStr := submatches[2]

		fmt.Printf("DEBUG: ParseTemplate - Processing component: %s with attributes: %s\n", componentName, attributesStr)

		// Parse attributes
		props, err := p.parseAttributes(attributesStr)
		if err != nil {
			fmt.Printf("DEBUG: ParseTemplate - Attribute parsing error for %s: %v\n", componentName, err)
			return fmt.Sprintf("<!-- Parse Error: %s -->", err.Error())
		}

		// Convert to Go template syntax
		replacement := p.buildGoTemplate(componentName, props)
		fmt.Printf("DEBUG: ParseTemplate - Replacing '%s' with '%s'\n", match, replacement)
		return replacement
	})

	if input != result {
		fmt.Printf("DEBUG: ParseTemplate - Template was modified during processing\n")
	} else {
		fmt.Printf("DEBUG: ParseTemplate - Template was NOT modified during processing\n")
	}

	return result, nil
}

// parseAttributes parses the attribute string into a map
func (p *ComponentParser) parseAttributes(attrStr string) (map[string]interface{}, error) {
	props := make(map[string]interface{})

	if strings.TrimSpace(attrStr) == "" {
		return props, nil
	}

	// Pattern to match key=value or key->value pairs
	// Handles: key="value", key='value', key=value, key=[array], key={object}
	// Also handles: key->"value", key->'value', key->value, key->[array], key->{object}
	attrPattern := regexp.MustCompile(`(\w+)\s*(=|->)\s*("[^"]*"|'[^']*'|\[[^\]]*\]|\{[^}]*\}|[^\s]+)`)

	matches := attrPattern.FindAllStringSubmatch(attrStr, -1)

	for _, match := range matches {
		if len(match) < 4 {
			continue
		}

		key := match[1]
		// match[2] contains the operator (= or ->)
		value := match[3]

		// Parse the value based on its format
		parsedValue := p.parseValue(value)
		props[key] = parsedValue
	}

	return props, nil
}

// parseValue parses different value types
func (p *ComponentParser) parseValue(value string) interface{} {
	value = strings.TrimSpace(value)

	// String with quotes
	if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
		(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
		return value[1 : len(value)-1]
	}

	// Array
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		return p.parseArray(value)
	}

	// Object
	if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
		return p.parseObject(value)
	}

	// Boolean
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}

	// Number
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		return num
	}

	// Default to string
	return value
}

// parseArray parses array values
func (p *ComponentParser) parseArray(value string) interface{} {
	// Remove brackets and parse
	content := strings.TrimSpace(value[1 : len(value)-1])
	if content == "" {
		return []interface{}{}
	}

	// Simple JSON parsing for arrays
	var result []interface{}
	if err := json.Unmarshal([]byte(value), &result); err == nil {
		return result
	}

	// Fallback: split by comma for simple arrays
	items := strings.Split(content, ",")
	result = make([]interface{}, len(items))
	for i, item := range items {
		result[i] = p.parseValue(strings.TrimSpace(item))
	}

	return result
}

// parseObject parses object values
func (p *ComponentParser) parseObject(value string) interface{} {
	// Try to parse as JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(value), &result); err == nil {
		return result
	}

	// Fallback: parse simple key:value pairs
	content := strings.TrimSpace(value[1 : len(value)-1])
	result = make(map[string]interface{})

	// Simple parser for key:value pairs
	pairs := strings.Split(content, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			val := p.parseValue(strings.TrimSpace(kv[1]))
			result[key] = val
		}
	}

	return result
}

// buildGoTemplate converts parsed props to Go template syntax
func (p *ComponentParser) buildGoTemplate(componentName string, props map[string]interface{}) string {
	if len(props) == 0 {
		return fmt.Sprintf(`{{component "%s"}}`, componentName)
	}

	// Build dict expression
	dictParts := []string{}
	for key, value := range props {
		dictPart := p.buildDictValue(key, value)
		dictParts = append(dictParts, dictPart)
	}

	dictExpr := strings.Join(dictParts, " ")
	return fmt.Sprintf(`{{component "%s" (dict %s)}}`, componentName, dictExpr)
}

// buildDictValue converts a value to dict syntax
func (p *ComponentParser) buildDictValue(key string, value interface{}) string {
	switch v := value.(type) {
	case string:
		// Escape quotes in strings
		escaped := strings.ReplaceAll(v, `"`, `\"`) 
		return fmt.Sprintf(`"%s" "%s"`, key, escaped)

	case []interface{}:
		// Convert array to list expression
		listParts := []string{}
		for _, item := range v {
			listParts = append(listParts, p.formatValue(item))
		}
		listExpr := strings.Join(listParts, " ")
		return fmt.Sprintf(`"%s" (list %s)`, key, listExpr)

	case map[string]interface{}:
		// Convert nested object to dict
		dictParts := []string{}
		for k, val := range v {
			dictParts = append(dictParts, p.buildDictValue(k, val))
		}
		dictExpr := strings.Join(dictParts, " ")
		return fmt.Sprintf(`"%s" (dict %s)`, key, dictExpr)

	case bool:
		return fmt.Sprintf(`"%s" %t`, key, v)

	case float64:
		// Check if it's an integer
		if v == float64(int(v)) {
			return fmt.Sprintf(`"%s" %d`, key, int(v))
		}
		return fmt.Sprintf(`"%s" %f`, key, v)

	default:
		return fmt.Sprintf(`"%s" "%v"`, key, v)
	}
}

// formatValue formats a single value for template expressions
func (p *ComponentParser) formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		escaped := strings.ReplaceAll(v, `"`, `\"`) 
		return fmt.Sprintf(`"%s"`, escaped)
	case bool:
		return fmt.Sprintf(`%t`, v)
	case float64:
		if v == float64(int(v)) {
			return fmt.Sprintf(`%d`, int(v))
		}
		return fmt.Sprintf(`%f`, v)
	default:
		return fmt.Sprintf(`"%v"`, v)
	}
}

// NewComponentRegistry creates a new component registry
func NewComponentRegistry(engine *Engine) *ComponentRegistry {
	return &ComponentRegistry{
		components: make(map[string]*Component),
		engine:     engine,
	}
}

// RegisterComponent adds a component to the registry
func (r *ComponentRegistry) RegisterComponent(component *Component) {
	r.components[component.Name] = component
}

// GetComponent retrieves a component by name
func (r *ComponentRegistry) GetComponent(name string) (*Component, bool) {
	component, exists := r.components[name]
	return component, exists
}

// RenderComponent renders a component with given props and optional override template
func (r *ComponentRegistry) RenderComponent(name string, props map[string]interface{}, overrideTemplate string) (template.HTML, error) {
	component, exists := r.GetComponent(name)
	if !exists {
		return "", fmt.Errorf("component '%s' not found", name)
	}

	// Merge default props with provided props
	finalProps := make(map[string]interface{})
	for k, v := range component.Props {
		finalProps[k] = v
	}
	for k, v := range props {
		finalProps[k] = v
	}

	// Use override template if provided, otherwise use component template
	templateContent := component.Template
	if overrideTemplate != "" {
		templateContent = overrideTemplate
	}

	// Preprocess template content to handle nested component syntax
	if r.engine.componentParser != nil {
		processedContent, err := r.engine.componentParser.ParseTemplate(templateContent)
		if err != nil {
			return "", fmt.Errorf("failed to preprocess component template '%s': %v", name, err)
		}
		templateContent = processedContent
	}

	// Create component state for Petite-Vue
	componentState := r.buildComponentState(component, finalProps)

	// Prepare template data
	templateData := map[string]interface{}{
		"Props":          finalProps,
		"ComponentState": componentState,
		"Component":      component,
	}

	// Parse and execute the template
	tmpl, err := template.New(name).Funcs(r.engine.funcMap).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse component template '%s': %v", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute component template '%s': %v", name, err)
	}

	return template.HTML(buf.String()), nil
}

// buildComponentState creates the reactive state object for Petite-Vue
func (r *ComponentRegistry) buildComponentState(component *Component, props map[string]interface{}) string {
	state := make(map[string]interface{})

	// Add state keys with default values
	for _, key := range component.StateKeys {
		if val, exists := props[key]; exists {
			state[key] = val
		} else {
			// Set sensible defaults based on type
			state[key] = r.getDefaultValue(key)
		}
	}

	// Convert to JSON for v-scope
	stateJSON, _ := json.Marshal(state)
	return string(stateJSON)
}

// getDefaultValue returns a sensible default for common state keys
func (r *ComponentRegistry) getDefaultValue(key string) interface{} {
	switch key {
	case "open", "visible", "active", "selected", "loading":
		return false
	case "items", "list", "data":
		return []interface{}{}
	case "text", "title", "content", "value":
		return ""
	case "count", "index", "page":
		return 0
	default:
		return nil
	}
}

// LoadComponentsFromDirectory loads components from a directory structure
func (r *ComponentRegistry) LoadComponentsFromDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Only process .bui files
		if !strings.HasSuffix(path, ".bui") {
			return nil
		}
		
		// Read the file content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		
		// Extract component name from template definition
		// Look for {{define "ComponentName"}}
		definePattern := regexp.MustCompile(`{{define\s+"([^"]+)"\s*}}`)
		matches := definePattern.FindStringSubmatch(string(content))
		if len(matches) < 2 {
			// Skip files without proper component definition
			return nil
		}
		
		componentName := matches[1]
		
		// Skip if it's a template path (contains slashes)
		if strings.Contains(componentName, "/") {
			return nil
		}
		
		// Create component with template content
		component := &Component{
			Name:     componentName,
			Template: string(content),
			Props:    map[string]interface{}{},
			StateKeys: []string{},
		}
		
		// Register the component
		r.RegisterComponent(component)
		fmt.Printf("DEBUG: Registered .bui component: %s from %s\n", componentName, path)
		
		return nil
	})
}

// RegisterComponentHelper registers the template helper function for rendering components
func (e *Engine) RegisterComponentHelper() {
	if e.componentRegistry == nil {
		e.componentRegistry = NewComponentRegistry(e)
	}

	// Register the main component helper
	e.AddHelper("component", func(name string, props ...interface{}) template.HTML {
		fmt.Printf("DEBUG: component helper called for: %s with props: %v\n", name, props)
		var propsMap map[string]interface{}

		if len(props) > 0 {
			if p, ok := props[0].(map[string]interface{}); ok {
				propsMap = p
			} else {
				propsMap = map[string]interface{}{"data": props[0]}
			}
		} else {
			propsMap = map[string]interface{}{}
		}

		html, err := e.componentRegistry.RenderComponent(name, propsMap, "")
		if err != nil {
			fmt.Printf("DEBUG: component helper error for %s: %v\n", name, err)
			return template.HTML(fmt.Sprintf("<!-- Component Error: %s -->", err.Error()))
		}
		fmt.Printf("DEBUG: component helper rendered %s successfully\n", name)
		return html
	})

	// Register helper for component with custom template
	e.AddHelper("componentWith", func(name string, customTemplate string, props ...interface{}) template.HTML {
		var propsMap map[string]interface{}

		if len(props) > 0 {
			if p, ok := props[0].(map[string]interface{}); ok {
				propsMap = p
			} else {
				propsMap = map[string]interface{}{"data": props[0]}
			}
		} else {
			propsMap = map[string]interface{}{}
		}

		html, err := e.componentRegistry.RenderComponent(name, propsMap, customTemplate)
		if err != nil {
			return template.HTML(fmt.Sprintf("<!-- Component Error: %s -->", err.Error()))
		}
		return html
	})

	// Register helper for inline component creation
	e.AddHelper("defineComponent", func(templateStr string, state string) template.HTML {
		// For quick inline components
		scopeAttr := ""
		if state != "" {
			scopeAttr = fmt.Sprintf(` v-scope="%s"`, state)
		}

		// Wrap template with v-scope if state is provided
		if strings.Contains(templateStr, "v-scope") || state == "" {
			return template.HTML(templateStr)
		}

		// Add v-scope to the root element
		if strings.HasPrefix(strings.TrimSpace(templateStr), "<") {
			// Find the first > and insert v-scope before it
			firstGt := strings.Index(templateStr, ">")
			if firstGt > 0 {
				return template.HTML(templateStr[:firstGt] + scopeAttr + templateStr[firstGt:])
			}
		}

		// Fallback: wrap in a div
		return template.HTML(fmt.Sprintf(`<div%s>%s</div>`, scopeAttr, templateStr))
	})

	// Register a shorthand helper for more intuitive component syntax
	e.AddHelper("c", func(name string, args ...interface{}) template.HTML {
		// This helper allows for a more intuitive syntax:
		// {{c "Dropdown" buttonContent="Click Me" content="content html" list=.listGoObject}}

		if len(args) == 0 {
			return template.HTML(fmt.Sprintf("<!-- Component Error: No props provided for %s -->", name))
		}

		// Create props map from named arguments
		props := make(map[string]interface{})

		// Process arguments in pairs (key, value)
		for i := 0; i < len(args); i += 2 {
			if i+1 >= len(args) {
				break // Avoid index out of range
			}

			// The key should be a string
			if key, ok := args[i].(string); ok {
				props[key] = args[i+1]
			}
		}

		// Render the component
		html, err := e.componentRegistry.RenderComponent(name, props, "")
		if err != nil {
			return template.HTML(fmt.Sprintf("<!-- Component Error: %s -->", err.Error()))
		}
		return html
	})
}
