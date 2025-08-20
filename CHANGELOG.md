# Changelog

All notable changes to the Base Framework will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v2.0.2] - 2025-08-20

### Added
- **🚀 Fully Dynamic Swagger Schema Generation** - Revolutionary automatic API documentation
  - Complete replacement of hardcoded schema generation with intelligent discovery system
  - Automatic model scanning from `/app/models/` directory for any new modules
  - Pattern-based schema generation following Base Framework naming conventions
  - Zero-maintenance swagger documentation - new modules automatically generate proper schemas
  - Smart filename-to-model conversion (`post_categor.go` → `PostCategor`)
  - Comprehensive schema set generation for any model:
    - `models.Create{Model}Request` for POST endpoints
    - `models.Update{Model}Request` for PUT endpoints  
    - `models.{Model}Response` for response objects
    - `models.{Model}SelectOption` for dropdown/select lists

### Enhanced
- **📚 Swagger Documentation System** - Complete overhaul for better API documentation
  - Fixed OpenAPI 3.0 compliance with proper `$ref` schema references
  - Enhanced requestBody handling for POST/PUT endpoints (replaced incorrect `parameters` usage)
  - Improved schema property generation with proper field types, descriptions, and examples
  - Automatic field description and example generation based on field names
  - Better error handling with graceful fallbacks to ensure swagger always works

### Fixed
- **🐛 API Documentation Issues**
  - Resolved "Unknown Type" errors in Swagger UI - now shows proper schema structures
  - Fixed parameter schema references to use correct OpenAPI 3.0 `$ref` format
  - Corrected body parameter handling to use `requestBody` instead of `parameters`
  - Fixed schema generation for dynamically created modules via `./base g` command

### Improved
- **⚡ Performance & Maintainability**
  - Eliminated manual schema maintenance - schemas now generate automatically
  - Better code organization with clear separation of concerns
  - More robust error handling in swagger service
  - Future-proof system that adapts to any new modules without code changes

## [v2.0.1] - 2025-08-19

### Added
- **🎯 Enhanced DateTime Type Support** - Comprehensive date/time handling
  - Support for MySQL, JSON, and HTML datetime formats
  - Flexible parsing with multiple format fallbacks
  - Proper timezone handling and RFC3339 compliance
  - Database scanner and valuer interface implementation

### Improved
- **📊 Database Performance Optimization**
  - Optimized field types and indexing strategies  
  - Better GORM tag support for various field types
  - Enhanced foreign key relationships and constraints

## [v2.0.0] - 2025-08-13

### Added
- **🚀 Zero-Dependency HTTP Router** - Complete Gin framework removal
  - Custom high-performance HTTP router with zero external dependencies
  - Method chaining API for clean application initialization
  - Builder pattern for application configuration and setup
  - Auto-discovery system for modules and swagger annotations
  - Simplified initialization from 4 files (368+ lines) to 1 file (~200 lines)
- **⚡ Streamlined Application Bootstrap** - Revolutionary initialization system
  - New `App` struct with fluent method chaining: `New().Start()`
  - Consolidated all initialization logic into single `base.go` file
  - Reduced main.go from 82 lines to just 7 lines for ultimate simplicity
  - Automatic environment, config, database, and router setup
  - Smart module auto-discovery and route registration

### Changed
- **BREAKING**: Complete removal of Gin framework dependency
- **BREAKING**: New application initialization API using method chaining
- **Router Architecture**: Custom tree-based router with path normalization
- **Project Structure**: Consolidated initialization files into single source
- **Startup Process**: Streamlined from complex multi-file setup to simple `.Start()` call

### Removed
- **Gin Framework**: Completely removed all Gin dependencies and imports
- **Complex Initialization**: Removed `start.go`, `app_initializer.go`, `main.go` from core (368+ lines eliminated)
- **Redundant Files**: Cleaned up all backup and temporary initialization files

### Technical Details
- Custom HTTP router with support for path parameters, wildcards, and middleware
- Zero external dependencies for core HTTP handling
- Improved performance through optimized route matching algorithms
- Enhanced path normalization and conflict detection
- Automatic OpenAPI 3.0 documentation generation

### Migration Guide
- **Existing projects**: Will automatically use new initialization system on update
- **Custom middleware**: Update to use new router middleware interface
- **Route definitions**: No changes needed - existing routes work unchanged
- **Performance**: Expect improved startup time and reduced memory footprint

 

### Added
- **Automatic Relationship Detection**: Enhanced code generation to automatically detect and create GORM relationships when field names end with `_id` and have `uint` type
- **Smart Field Processing**: Generator now creates both foreign key fields and relationship fields automatically without manual specification
- **Enhanced Templates**: Updated model, service, and request/response templates to handle auto-detected relationships properly
- **Clean Code Generation**: Eliminated duplicate field generation issues in templates
- **Proper GORM Tags**: Auto-generated relationships include correct `foreignKey` GORM tags
- **Template Consistency**: All templates now consistently handle the enhanced relationship detection system

### Changed
- **Model Template**: Updated to handle both foreign key and relationship fields generated by enhanced detection
- **Service Template**: Fixed to prevent duplicate field assignments in Create and Update operations
- **Field Processing Logic**: Enhanced `ProcessField` function in `templateutils.go` to detect `_id` suffix patterns and generate appropriate relationship structures
- **Template Logic**: Simplified template conditions to work with the new dual-field approach (foreign key + relationship)

### Fixed
- **Duplicate Fields**: Resolved issue where foreign key fields were being generated multiple times in models
- **Service Layer**: Fixed Create and Update methods to properly handle auto-detected relationship fields
- **Template Rendering**: Corrected template logic to avoid conflicts between manual and automatic relationship handling
- **Init.go Cleanup**: Previously fixed issue where destroy command wasn't properly cleaning up module registrations
- **HTTP Status Codes**: Previously corrected status codes in generated controllers
- **Directory Naming**: Previously fixed to use plural directory names (`models`, `posts`) with singular model files

### Technical Details
- Enhanced `ProcessField` function to return multiple `FieldStruct` objects when detecting `_id` patterns
- Updated all templates to distinguish between regular fields and relationship fields
- Improved GORM tag generation for automatic foreign key relationships
- Streamlined template logic to reduce complexity and improve maintainability

### Examples

Before this enhancement, you needed to manually specify relationships:
```bash
base g article title:string content:text author:belongsTo:Author category:belongsTo:Category
```

Now, relationships are automatically detected:
```bash
base g article title:string content:text author_id:uint category_id:uint
```

This automatically generates:
```go
type Article struct {
    Id         uint     `json:"id" gorm:"primarykey"`
    Title      string   `json:"title"`
    Content    string   `json:"content"`
    AuthorId   uint     `json:"author_id"`
    Author     Author   `json:"author,omitempty" gorm:"foreignKey:AuthorId"`
    CategoryId uint     `json:"category_id"`
    Category   Category `json:"category,omitempty" gorm:"foreignKey:CategoryId"`
}
```

### Migration Guide
- Existing models and modules continue to work without changes
- New modules can take advantage of automatic relationship detection by using the `_id` suffix convention
- No breaking changes to existing CLI commands or API

---

## Previous Versions

### [1.1.1] - Previous Release
- Core framework functionality
- Basic code generation
- HMVC architecture
- Manual relationship specification
- Module system with auto-registration
- Authentication and authorization
- File storage and email integration