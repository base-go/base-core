package module

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"base/core/logger"
)

// ModuleInfo represents information about a discovered module
type ModuleInfo struct {
	Name          string
	Path          string
	PackageName   string
	ConstructorFn string
	HasInit       bool
	HasMigrate    bool
	HasRoutes     bool
}

// DynamicLoader handles dynamic discovery and loading of app modules
type DynamicLoader struct {
	logger  logger.Logger
	appDir  string
	modules map[string]*ModuleInfo
}

// NewDynamicLoader creates a new dynamic module loader
func NewDynamicLoader(logger logger.Logger, appDir string) *DynamicLoader {
	return &DynamicLoader{
		logger:  logger,
		appDir:  appDir,
		modules: make(map[string]*ModuleInfo),
	}
}

// DiscoverModules scans the app directory for modules and analyzes their structure
func (dl *DynamicLoader) DiscoverModules() error {
	dl.logger.Info("🔍 Starting dynamic module discovery", logger.String("appDir", dl.appDir))

	// Check if app directory exists
	if _, err := os.Stat(dl.appDir); os.IsNotExist(err) {
		dl.logger.Warn("App directory does not exist, skipping module discovery")
		return nil
	}

	// Read app directory entries
	entries, err := os.ReadDir(dl.appDir)
	if err != nil {
		return fmt.Errorf("failed to read app directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip special directories
		if entry.Name() == "models" || entry.Name() == "migrations" {
			continue
		}

		modulePath := filepath.Join(dl.appDir, entry.Name())
		if err := dl.analyzeModule(entry.Name(), modulePath); err != nil {
			dl.logger.Warn("Failed to analyze module",
				logger.String("module", entry.Name()),
				logger.String("error", err.Error()))
			continue
		}
	}

	dl.logger.Info("✅ Module discovery complete", logger.Int("modules", len(dl.modules)))
	return nil
}

// analyzeModule analyzes a single module directory
func (dl *DynamicLoader) analyzeModule(name, path string) error {
	// Check if module.go exists
	moduleFile := filepath.Join(path, "module.go")
	if _, err := os.Stat(moduleFile); os.IsNotExist(err) {
		return fmt.Errorf("module.go not found")
	}

	// Parse the module.go file to extract information
	info, err := dl.parseModuleFile(name, moduleFile)
	if err != nil {
		return fmt.Errorf("failed to parse module file: %w", err)
	}

	dl.modules[name] = info
	dl.logger.Debug("Discovered module",
		logger.String("name", name),
		logger.String("constructor", info.ConstructorFn),
		logger.Bool("hasInit", info.HasInit),
		logger.Bool("hasMigrate", info.HasMigrate),
		logger.Bool("hasRoutes", info.HasRoutes))

	return nil
}

// parseModuleFile parses a module.go file to extract module information
func (dl *DynamicLoader) parseModuleFile(name, filePath string) (*ModuleInfo, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse Go source
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %w", err)
	}

	info := &ModuleInfo{
		Name:        name,
		Path:        filePath,
		PackageName: node.Name.Name,
	}

	// Find constructor function and analyze module structure
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			funcName := x.Name.Name

			// Look for constructor functions (New*Module)
			if strings.HasPrefix(funcName, "New") && strings.HasSuffix(funcName, "Module") {
				info.ConstructorFn = funcName
			}

			// Check for module methods
			if x.Recv != nil && len(x.Recv.List) > 0 {
				if starExpr, ok := x.Recv.List[0].Type.(*ast.StarExpr); ok {
					if ident, ok := starExpr.X.(*ast.Ident); ok && ident.Name == "Module" {
						switch funcName {
						case "Init":
							info.HasInit = true
						case "Migrate":
							info.HasMigrate = true
						case "Routes":
							info.HasRoutes = true
						}
					}
				}
			}
		}
		return true
	})

	// If no constructor found, try to infer it from common patterns
	if info.ConstructorFn == "" {
		// Try common patterns
		patterns := []string{
			fmt.Sprintf("New%sModule", toTitle(name)),
			fmt.Sprintf("New%sModule", toTitle(strings.TrimSuffix(name, "s"))), // singular
		}

		for _, pattern := range patterns {
			if dl.functionExistsInFile(string(content), pattern) {
				info.ConstructorFn = pattern
				break
			}
		}
	}

	if info.ConstructorFn == "" {
		return nil, fmt.Errorf("no constructor function found")
	}

	return info, nil
}

// functionExistsInFile checks if a function exists in the file content using regex
func (dl *DynamicLoader) functionExistsInFile(content, funcName string) bool {
	pattern := fmt.Sprintf(`func\s+%s\s*\(`, regexp.QuoteMeta(funcName))
	matched, _ := regexp.MatchString(pattern, content)
	return matched
}

// GetDiscoveredModules returns all discovered modules
func (dl *DynamicLoader) GetDiscoveredModules() map[string]*ModuleInfo {
	return dl.modules
}

// GenerateImportFile generates a Go file that imports all discovered modules
func (dl *DynamicLoader) GenerateImportFile(outputPath string) error {
	dl.logger.Info("📝 Generating module import file", logger.String("output", outputPath))

	if len(dl.modules) == 0 {
		dl.logger.Info("No modules to generate imports for")
		return nil
	}

	// Generate import file content
	var builder strings.Builder

	builder.WriteString("// Code generated by dynamic module loader. DO NOT EDIT.\n\n")
	builder.WriteString("package main\n\n")
	builder.WriteString("import (\n")

	// Add imports for each module
	for name := range dl.modules {
		builder.WriteString(fmt.Sprintf("\t_ \"base/app/%s\"\n", name))
	}

	builder.WriteString(")\n")

	// Write to file
	if err := os.WriteFile(outputPath, []byte(builder.String()), 0644); err != nil {
		return fmt.Errorf("failed to write import file: %w", err)
	}

	dl.logger.Info("✅ Import file generated successfully")
	return nil
}

// LoadModules loads all discovered modules using their registered factories
func (dl *DynamicLoader) LoadModules(deps Dependencies) (map[string]Module, error) {
	dl.logger.Info("🔄 Loading discovered modules", logger.Int("count", len(dl.modules)))

	modules := make(map[string]Module)

	for name, info := range dl.modules {
		// Try to get the module factory from the global registry
		factory := GetAppModule(name)
		if factory == nil {
			dl.logger.Warn("Module factory not found in registry",
				logger.String("module", name),
				logger.String("constructor", info.ConstructorFn))
			continue
		}

		// Create module instance using the factory
		module := factory(deps)
		if module == nil {
			dl.logger.Error("Factory returned nil module", logger.String("module", name))
			continue
		}

		modules[name] = module
		dl.logger.Debug("Module loaded successfully", logger.String("module", name))
	}

	dl.logger.Info("✅ Module loading complete", logger.Int("loaded", len(modules)))
	return modules, nil
}

// GetModuleInfo returns information about a specific module
func (dl *DynamicLoader) GetModuleInfo(name string) (*ModuleInfo, bool) {
	info, exists := dl.modules[name]
	return info, exists
}

// ListModuleNames returns a list of all discovered module names
func (dl *DynamicLoader) ListModuleNames() []string {
	names := make([]string, 0, len(dl.modules))
	for name := range dl.modules {
		names = append(names, name)
	}
	return names
}

// ValidateModule checks if a module meets the basic requirements
func (dl *DynamicLoader) ValidateModule(name string) error {
	info, exists := dl.modules[name]
	if !exists {
		return fmt.Errorf("module %s not found", name)
	}

	if info.ConstructorFn == "" {
		return fmt.Errorf("module %s has no constructor function", name)
	}

	// Additional validation can be added here
	return nil
}

// ScanModuleFile scans a module file for specific patterns or content
func (dl *DynamicLoader) ScanModuleFile(filePath string, pattern string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if regex.MatchString(line) {
			matches = append(matches, fmt.Sprintf("Line %d: %s", lineNum, strings.TrimSpace(line)))
		}
	}

	return matches, scanner.Err()
}

// toTitle converts a string to title case (replacement for deprecated strings.Title)
func toTitle(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
