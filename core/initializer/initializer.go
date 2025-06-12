package initializer

import (
	"base/core/config"
	"base/core/database"
	"base/core/email"
	"base/core/emitter"
	"base/core/language"
	"base/core/layout"
	"base/core/logger"
	"base/core/middleware"
	"base/core/module"
	"base/core/storage"

	// Import core module packages
	"base/core/app/auth"
	"base/core/app/media"
	"base/core/app/users"
	"net/http"
	"time"

	method "github.com/bu/gin-method-override"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
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
	Layout       *layout.Engine
	Modules      []module.Module
	EmailSender  email.Sender
	Translations *language.TranslationService
}

// CoreDependencies holds all core dependencies that modules might need
type CoreDependencies struct {
	DB           *gorm.DB
	WebRouter    *gin.RouterGroup
	PublicRouter *gin.RouterGroup // Router for public routes (no auth middleware)
	APIRouter    *gin.RouterGroup
	Logger       logger.Logger
	Emitter      *emitter.Emitter
	Storage      *storage.ActiveStorage
	Translations *language.TranslationService
	EmailSender  email.Sender
	Layout       *layout.Engine
}

// ModuleRegistry holds the module definitions
type ModuleRegistry interface {
	GetModules(deps *CoreDependencies) map[string]module.Module
}

// NewApp creates and initializes a new App instance with a module registry
func NewApp(cfg *config.Config, registry ModuleRegistry) (*App, error) {
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

	// Initialize router with middleware
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(method.ProcessMethodOverride(router))
	router.Use(middleware.Logger(log))

	// Initialize session store
	store := cookie.NewStore([]byte(cfg.JWTSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 30, // 30 days default
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   cfg.Env == "production",
	})
	router.Use(sessions.Sessions("base_session", store))

	// Set up static file serving
	router.Static("/static", "./static")
	router.Static("/storage", "./storage")

	// Set up CORS
	corsConfig := cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "authorization", "X-Api-Key", "Base-Orgid", "Base-*"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))

	// Initialize core dependencies
	emitter := &emitter.Emitter{}

	db, err := initDB(cfg)
	if err != nil {
		return nil, err
	}

	activeStorage, err := initStorage(cfg, db)
	if err != nil {
		return nil, err
	}

	emailSender, err := email.NewSender(cfg)
	if err != nil {
		log.Error("Failed to initialize email sender: " + err.Error())
		// Continue without email - auth will work but no emails will be sent
	}

	templateEngine, translationService, err := initTemplateSystem(log)
	if err != nil {
		return nil, err
	}

	// Set the template engine on the router
	router.HTMLRender = templateEngine

	// Create router groups
	// Public routes (no auth middleware)
	publicRouter := router.Group("/")
	publicRouter.Use(language.CookieLanguageMiddleware(translationService))
	publicRouter.Use(middleware.TemplateContextMiddleware())

	// Main web routes with optional auth
	webRouter := router.Group("/")
	webRouter.Use(middleware.OptionalAuthMiddleware())
	webRouter.Use(language.CookieLanguageMiddleware(translationService))
	webRouter.Use(middleware.TemplateContextMiddleware())

	// Add language switching endpoint
	webRouter.GET("/switch-language/:code", func(c *gin.Context) {
		langCode := c.Param("code")
		if _, found := language.GetLanguageByCode(langCode); found {
			c.SetCookie(
				language.LanguageCookieName,
				langCode,
				60*60*24*30, // 30 days expiration
				"/",         // path
				"",          // domain
				false,       // secure
				false,       // httpOnly
			)
		}

		referer := c.GetHeader("Referer")
		if referer == "" {
			referer = "/"
		}
		c.Redirect(302, referer)
	})

	apiRouter := router.Group("/api")

	// Create core dependencies struct
	deps := &CoreDependencies{
		DB:           db,
		WebRouter:    webRouter,
		PublicRouter: publicRouter,
		APIRouter:    apiRouter,
		Logger:       log,
		Emitter:      emitter,
		Storage:      activeStorage,
		Layout:       templateEngine,
		Translations: translationService,
		EmailSender:  emailSender,
	}

	// Initialize core modules first
	coreModuleMap := initializeCoreModules(deps)

	// Initialize app modules using the registry
	appModuleMap := registry.GetModules(deps)

	// Merge core modules with app modules (app modules take precedence)
	moduleMap := make(map[string]module.Module)
	for k, v := range coreModuleMap {
		moduleMap[k] = v
	}
	for k, v := range appModuleMap {
		moduleMap[k] = v
	}
	modules := initializeModules(moduleMap, webRouter, apiRouter, log)

	app := &App{
		DB:           db,
		Router:       router,
		Log:          log,
		Emitter:      emitter,
		Storage:      activeStorage,
		Layout:       templateEngine,
		Modules:      modules,
		Translations: translationService,
	}

	return app, nil
}

// initDB initializes the database connection
func initDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := database.InitDB(cfg)
	if err != nil {
		return nil, err
	}
	return db.DB, nil
}

// initStorage initializes the storage service
func initStorage(cfg *config.Config, db *gorm.DB) (*storage.ActiveStorage, error) {
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
	return storage.NewActiveStorage(db, storageConfig)
}

// initTemplateSystem initializes templates and translations
func initTemplateSystem(log logger.Logger) (*layout.Engine, *language.TranslationService, error) {
	// Initialize template engine
	templateConfig := layout.Config{
		TemplatesDir: "app/theme/default",
		LayoutsDir:   "app/theme/default/layouts",
		SharedDir:    "app/theme/default/shared",
	}
	// Initialize translation service first
	translationService := language.NewTranslationService()

	// This will be implemented by the app
	// if err := app.RegisterTranslations(translationService); err != nil {
	// 	log.Error("Failed to load translations: " + err.Error())
	// }

	// Create template engine without loading templates yet
	templateEngine := layout.NewEngineWithoutLoading(templateConfig)
	templateEngine.RegisterDefaultHelpers()

	// Register template helpers BEFORE loading templates
	language.RegisterTemplateHelpers(templateEngine, translationService)

	// Now load templates with all helpers registered
	if err := templateEngine.LoadTemplates(); err != nil {
		log.Info("Failed to load templates: " + err.Error())
	}

	log.Info("Template engine initialized successfully")

	// This will be implemented by the app
	// if err := app.RegisterTemplates(templateEngine); err != nil {
	// 	log.Error("Failed to register templates: " + err.Error())
	// 	return nil, nil, err
	// }

	return templateEngine, translationService, nil
}

// initializeModules processes and initializes all modules
func initializeModules(moduleMap map[string]module.Module, webRouter *gin.RouterGroup, apiRouter *gin.RouterGroup, log logger.Logger) []module.Module {
	var modules []module.Module

	for name, mod := range moduleMap {
		if err := module.RegisterModule(name, mod); err != nil {
			log.Error("Failed to register module",
				logger.String("module", name),
				logger.String("error", err.Error()))
			continue
		}

		if err := mod.Init(); err != nil {
			log.Error("Failed to initialize module",
				logger.String("module", name),
				logger.String("error", err.Error()))
			continue
		}

		if err := mod.Migrate(); err != nil {
			log.Error("Failed to migrate module",
				logger.String("module", name),
				logger.String("error", err.Error()))
			continue
		}

		// Set up routes for the module
		if routeModule, ok := mod.(interface{ Routes(*gin.RouterGroup) }); ok {
			log.Info("Setting up routes for module", logger.String("module", name))
			routeModule.Routes(webRouter)
		}

		// Set up API routes for the module
		if apiRouteModule, ok := mod.(interface{ APIRoutes(*gin.RouterGroup) }); ok {
			log.Info("Setting up API routes for module", logger.String("module", name))
			apiRouteModule.APIRoutes(apiRouter)
		}

		modules = append(modules, mod)
	}

	return modules
}

// AppSetupCallback is called after core initialization to load app-specific resources
type AppSetupCallback func(app *App) error

// NewAppWithSetup creates and initializes a new App instance with custom setup
func NewAppWithSetup(cfg *config.Config, registry ModuleRegistry, setup AppSetupCallback) (*App, error) {
	app, err := NewApp(cfg, registry)
	if err != nil {
		return nil, err
	}

	if setup != nil {
		if err := setup(app); err != nil {
			return nil, err
		}
	}

	return app, nil
}

// NewAppWithRegistry creates an app using the provided registry and default app setup
func NewAppWithRegistry(cfg *config.Config, registry ModuleRegistry) (*App, error) {
	return NewAppWithSetup(cfg, registry, nil)
}

// AppInterface defines methods that an app package can implement for customization
type AppInterface interface {
	RegisterTranslations(translations *language.TranslationService) error
	RegisterTemplates(template *layout.Engine) error
}

// NewAppWithInterface creates an app using a registry and app interface for customization
func NewAppWithInterface(cfg *config.Config, registry ModuleRegistry, appInterface AppInterface) (*App, error) {
	setup := func(app *App) error {
		// Load app-specific translations and templates
		if err := appInterface.RegisterTranslations(app.Translations); err != nil {
			app.Log.Error("Failed to load translations: " + err.Error())
		}

		if err := appInterface.RegisterTemplates(app.Layout); err != nil {
			app.Log.Error("Failed to register templates: " + err.Error())
			return err
		}

		// Root route is handled by the home module

		return nil
	}

	return NewAppWithSetup(cfg, registry, setup)
}

// initializeCoreModules initializes all core modules automatically
func initializeCoreModules(deps *CoreDependencies) map[string]module.Module {
	// Initialize core modules
	coreModules := make(map[string]module.Module)

	// Initialize auth module
	coreModules["auth"] = auth.NewAuthModule(
		deps.DB,
		deps.PublicRouter, // Use PublicRouter for auth routes (login/register should not require authentication)
		deps.APIRouter,
		deps.EmailSender,
		deps.Logger,
		deps.Emitter,
		deps.Layout,
	)

	// Initialize users module
	coreModules["users"] = users.NewUserModule(
		deps.DB,
		deps.WebRouter,
		deps.APIRouter,
		deps.Logger,
		deps.Storage,
		deps.Layout,
	)

	// Initialize media module
	coreModules["media"] = media.NewMediaModule(
		deps.DB,
		deps.WebRouter,
		deps.APIRouter,
		deps.Storage,
		deps.Emitter,
		deps.Logger,
	)

	return coreModules
}
