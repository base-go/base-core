package app

import (
	// MODULE_IMPORT_MARKER - Do not remove this comment because it's used by the CLI to add new module imports
	"base/app/posts"
	"fmt"

	"base/core/config"
	"base/core/database"
	"base/core/emitter"
	"base/core/language"
	"base/core/logger"
	"base/core/middleware"
	"base/core/module"
	"base/core/storage"
	"base/core/template"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// App represents the application instance
type App struct {
	DB           *gorm.DB
	Router       *gin.Engine
	Log          logger.Logger
	Emitter      *emitter.Emitter
	Storage      *storage.ActiveStorage
	Template     *template.Engine
	Modules      []module.Module
	Translations *language.TranslationService
}

// NewApp creates and initializes a new App instance
func NewApp(cfg *config.Config) (*App, error) {
	// Initialize logger
	logConfig := logger.Config{
		Environment: "development",
		LogPath:     "logs",
		Level:       "debug",
	}
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		return nil, err
	}
	// Initialize router
	router := gin.Default()
	// Initialize emitter
	emitter := &emitter.Emitter{}
	// Initialize database (you'll need to implement this)
	db, err := initDB(cfg)
	if err != nil {
		return nil, err
	}
	// Initialize storage
	storageConfig := storage.Config{
		Provider:  cfg.StorageProvider,
		Path:      cfg.StoragePath,
		BaseURL:   cfg.StorageBaseURL,
		APIKey:    cfg.StorageAPIKey,
		APISecret: cfg.StorageAPISecret,
		Endpoint:  cfg.StorageEndpoint,
		Bucket:    cfg.StorageBucket,
		CDN:       cfg.CDN,
	}
	activeStorage, err := storage.NewActiveStorage(db, storageConfig)
	if err != nil {
		return nil, err
	}

	// Initialize template engine
	templateConfig := template.Config{
		TemplatesDir: "app/theme/default",
		LayoutsDir:   "app/theme/default/layouts",
		SharedDir:    "app/theme/default/shared",
	}
	templateEngine := template.NewEngine(templateConfig)
	templateEngine.RegisterDefaultHelpers()

	// Initialize translation service
	translationService := language.NewTranslationService()

	// Load translations from embedded files FIRST
	if err := RegisterTranslations(translationService); err != nil {
		log.Error("Failed to load translations: " + err.Error())
	}

	// Register template helpers for translations BEFORE template registration
	fmt.Println("=== ABOUT TO REGISTER LANGUAGE HELPERS ===")
	language.RegisterTemplateHelpers(templateEngine, translationService)
	fmt.Println("=== LANGUAGE HELPERS REGISTERED ===")

	// Register all templates using our app template system
	fmt.Println("=== ABOUT TO CALL RegisterTemplates ===")
	if err := RegisterTemplates(templateEngine); err != nil {
		fmt.Printf("=== RegisterTemplates ERROR: %v ===\n", err)
		log.Error("Failed to register templates: " + err.Error())
		return nil, err // Return the error instead of continuing
	}
	fmt.Println("=== RegisterTemplates COMPLETED SUCCESSFULLY ===")

	// Set the template engine on the router
	router.HTMLRender = templateEngine

	app := &App{
		DB:           db,
		Router:       router,
		Log:          log,
		Emitter:      emitter,
		Storage:      activeStorage,
		Template:     templateEngine,
		Modules:      make([]module.Module, 0),
		Translations: translationService,
	}
	// Apply optional auth middleware to api routes
	apiRouter := router.Group("/api")
	apiRouter.Use(middleware.OptionalAuthMiddleware())

	// Apply optional auth middleware to web routes
	webRouter := router.Group("")
	webRouter.Use(middleware.OptionalAuthMiddleware())

	// Set up root routes (without language prefix)
	// This will redirect to a language-prefixed route via the middleware
	router.GET("/", language.LocalizedRoutingMiddleware(app.Translations), func(c *gin.Context) {
		c.HTML(200, "landing.html", gin.H{
			"title": "Welcome",
		})
	})

	// Create router groups for each supported language
	for _, lang := range language.GetSupportedLanguages() {
		// Create a language-specific router group
		langCode := lang.Code
		langRouter := router.Group("/" + langCode)

		// Apply middlewares to the language router group
		langRouter.Use(func(c *gin.Context) {
			// Set the language for this group
			app.Translations.SetLanguage(lang)
			c.Set("TranslationService", app.Translations)
			c.Set("CurrentLanguage", lang)
			c.Next()
		})

		// Add language-specific landing page route
		langRouter.GET("/", func(c *gin.Context) {
			c.HTML(200, "landing.html", gin.H{
				"title": "Welcome",
			})
		})

		// Create web router group within language group for module routes
		langWebRouter := langRouter.Group("/")
		langWebRouter.Use(middleware.OptionalAuthMiddleware())

		// Initialize modules with language-specific routes
		moduleInitializer := &AppModuleInitializer{
			DB:        db,
			APIRouter: apiRouter,
			WebRouter: langWebRouter,
			Logger:    log,
			Emitter:   emitter,
			Storage:   activeStorage,
			Template:  templateEngine,
		}
		modules := moduleInitializer.InitializeModules(db)
		app.Modules = append(app.Modules, modules...)
	}

	// Also register routes in the default webRouter for backward compatibility
	defaultModuleInitializer := &AppModuleInitializer{
		DB:        db,
		APIRouter: apiRouter,
		WebRouter: webRouter,
		Logger:    log,
		Emitter:   emitter,
		Storage:   activeStorage,
		Template:  templateEngine,
	}
	// Add these modules to the app as well
	app.Modules = append(app.Modules, defaultModuleInitializer.InitializeModules(db)...)
	return app, nil
}

// AppModuleInitializer holds all dependencies needed for app module initialization
type AppModuleInitializer struct {
	DB        *gorm.DB
	APIRouter *gin.RouterGroup
	WebRouter *gin.RouterGroup
	Logger    logger.Logger
	Emitter   *emitter.Emitter
	Storage   *storage.ActiveStorage
	Template  *template.Engine
}

// InitializeModules initializes all application modules
func (a *AppModuleInitializer) InitializeModules(db *gorm.DB) []module.Module {
	var modules []module.Module
	// Initialize modules
	moduleMap := a.getModules(db)
	// Register and initialize each module
	for name, mod := range moduleMap {

		if err := module.RegisterModule(name, mod); err != nil {
			a.Logger.Error("Failed to register module",
				logger.String("module", name),
				logger.String("error", err.Error()))
			continue
		}
		// Initialize the module
		if err := mod.Init(); err != nil {
			a.Logger.Error("Failed to initialize module",
				logger.String("module", name),
				logger.String("error", err.Error()))
			continue
		}
		// Migrate the module
		if err := mod.Migrate(); err != nil {
			a.Logger.Error("Failed to migrate module",
				logger.String("module", name),
				logger.String("error", err.Error()))
			continue
		}

		// Set up routes for the module
		if routeModule, ok := mod.(interface{ Routes(*gin.RouterGroup) }); ok {
			a.Logger.Info("Setting up routes for module", logger.String("module", name))
			routeModule.Routes(a.WebRouter)
		}

		modules = append(modules, mod)
	}
	return modules
}

// getModules returns a map of module name to module instance
func (a *AppModuleInitializer) getModules(db *gorm.DB) map[string]module.Module {
	modules := make(map[string]module.Module)
	// Define the module initializers directly
	moduleInitializers := map[string]func(*gorm.DB, *gin.RouterGroup, *gin.RouterGroup, logger.Logger, *emitter.Emitter, *storage.ActiveStorage, *template.Engine) module.Module{
		"posts": func(db *gorm.DB, apiRouter, webRouter *gin.RouterGroup, log logger.Logger, emitter *emitter.Emitter, activeStorage *storage.ActiveStorage, templateEngine *template.Engine) module.Module {
			return posts.NewPostModule(db, apiRouter, webRouter, log, emitter, activeStorage, templateEngine)
		},

		// MODULE_INITIALIZER_MARKER - Do not remove this comment because it's used by the CLI to add new module initializers
	}

	// Initialize and register each module
	for name, initializer := range moduleInitializers {
		modules[name] = initializer(db, a.APIRouter, a.WebRouter, a.Logger, a.Emitter, a.Storage, a.Template)
	}

	return modules
}

// initDB initializes the database connection
func initDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := database.InitDB(cfg)
	if err != nil {
		return nil, err
	}
	return db.DB, nil
}

/*
 This function returns an extension of the user's context with a company ID
 If no company is found, returns just the user ID
 This is useful for contexts where you want to extend the user context with additional information
 For example, you could use this to pass the user ID and the company ID to the handler function

 Example:

	Let's find the company for the user
	since we have access to the database, we can do it here instead of in the handler function

	var company models.Company
	if err := database.DB.Where("user_id = ?", user_id).First(&company).Error; err != nil {
		// If no company found, return just the user ID
		return map[string]interface{}{
			"user_id": user_id,
		}
	}

	Then return the context with both user ID and company ID
	return map[string]interface{}{
		"user_id": user_id,
		"company_id": company.Id,
	}
*/

func Extend(user_id uint) interface{} {
	// Get database instance
	if database.DB == nil {
		return map[string]interface{}{
			"user_id": user_id,
		}
	}

	// Get company for the user Here like in Examle above

	return map[string]interface{}{
		"user_id": user_id,
	}
}
