package core

import (
	"base/app"
	"base/app/posts"
	"base/app/templates"
	coreapp "base/core/app"
	"base/core/app/auth"
	"base/core/config"
	"base/core/database"
	"base/core/email"
	"base/core/emitter"
	"base/core/language"
	"base/core/logger"
	"base/core/middleware"
	"base/core/module"
	"base/core/storage"
	"base/core/template"
	"base/core/websocket"
	"fmt"
	"net/http"
	"time"

	method "github.com/bu/gin-method-override"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Application struct {
	Config  *config.Config
	DB      *database.Database
	Router  *gin.Engine
	WSHub   *websocket.Hub
	Modules map[string]module.Module
	Logger  logger.Logger
	Emitter *emitter.Emitter
}

var Emitter = emitter.New() // This ensures Emitter is created once

// StartApplication initializes and starts the application
func StartApplication() (*Application, error) {
	// Initialize config
	cfg := config.NewConfig()

	// Initialize logger first
	logConfig := logger.Config{
		Environment: cfg.Env,
		LogPath:     "logs",
		Level:       "debug",
	}
	appLogger, err := logger.NewLogger(logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	appLogger.Info("Starting application initialization",
		logger.String("version", cfg.Version),
		logger.String("environment", cfg.Env))

	// Initialize the database
	db, err := database.InitDB(cfg)
	if err != nil {
		appLogger.Error("Failed to initialize database",
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("database initialization failed: %w", err)
	}
	appLogger.Info("Database initialized successfully")

	// Initialize email sender
	emailSender, err := email.NewSender(cfg)
	if err != nil {
		appLogger.Error("Failed to initialize email sender",
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("email sender initialization failed: %w", err)
	}
	appLogger.Info("Email sender initialized successfully")

	// Initialize storage service
	appLogger.Info("Initializing storage service")
	storageConfig := storage.Config{
		Provider:  cfg.StorageProvider,
		Path:      cfg.StoragePath,
		BaseURL:   cfg.StorageBaseURL,
		APIKey:    cfg.StorageAPIKey,
		APISecret: cfg.StorageAPISecret,
		AccountID: cfg.StorageAccountID,
		Endpoint:  cfg.StorageEndpoint,
		Bucket:    cfg.StorageBucket,
		CDN:       cfg.CDN,
		Region:    cfg.StorageRegion,
	}

	activeStorage, err := storage.NewActiveStorage(db.DB, storageConfig)
	if err != nil {
		appLogger.Error("Failed to initialize storage service",
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("storage service initialization failed: %w", err)
	}

	appLogger.Info("Storage service initialized successfully",
		logger.String("provider", cfg.StorageProvider),
		logger.String("path", cfg.StoragePath))

	// Set up Gin
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(method.ProcessMethodOverride(router))

	// Set up middleware
	router.Use(middleware.Logger(appLogger))

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

	// Set up Swagger
	router.GET("/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.PersistAuthorization(true)))
	appLogger.Info("Swagger documentation enabled")

	// Create API router group with API key requirement for all routes
	apiGroup := router.Group("/api")
	apiGroup.Use(middleware.APIKeyMiddleware())
	// Add header middleware to extract and process Base- prefixed headers
	apiGroup.Use(middleware.HeaderMiddleware())

	// Create auth group for login/register (only requires API key)
	authGroup := apiGroup.Group("/auth")
	webGroup := router.Group("")
	// Create protected group that requires Bearer token
	protectedGroup := apiGroup.Group("/")
	protectedGroup.Use(middleware.AuthMiddleware())

	// Initialize core modules with all dependencies
	appLogger.Info("Initializing core modules")

	core := coreapp.NewCore(
		db.DB,
		protectedGroup, // Use protected group for most routes
		authGroup,      // Pass auth group for auth routes
		emailSender,
		appLogger,
		activeStorage,
		Emitter,
		webGroup,
	)
	modules := core.Modules
	appLogger.Info("Core modules initialized", logger.Int("count", len(modules)))

	// Register core module routes
	//core.RegisterRoutes()

	// Initialize translation service for language support
	appLogger.Info("Initializing translation service and loading translations")
	translationService := language.NewTranslationService()

	// Load translations
	if err := app.RegisterTranslations(translationService); err != nil {
		appLogger.Error("Failed to load translations: " + err.Error())
		// Translations are critical, consider returning an error to halt startup
		return nil, fmt.Errorf("failed to load translations during app start: %w", err)
	}
	appLogger.Info("Translations loaded successfully")

	appLogger.Info("Initializing application modules and template engine")

	// Set up template engine and routes manually, passing the populated translationService
	templateEngine := setupTemplateRoutes(router, appLogger, translationService)

	// Apply optional auth middleware to web routes
	webGroup.Use(middleware.OptionalAuthMiddleware())

	// Register language-specific router groups for each supported language
	appLogger.Info("Setting up language-specific routes")
	for _, lang := range language.GetSupportedLanguages() {
		langCode := lang.Code
		langRouter := router.Group("/" + langCode)

		appLogger.Info("Registering language routes", logger.String("language", langCode))

		// Apply language middleware to set the language context
		langRouter.Use(func(c *gin.Context) {
			translationService.SetLanguage(lang)
			c.Set("TranslationService", translationService)
			c.Set("CurrentLanguage", lang)
			c.Next()
		})

		// Add language-specific landing page route
		langRouter.GET("/", func(c *gin.Context) {
			c.HTML(200, "landing.html", gin.H{
				"title": "Welcome",
			})
		})

		// Create language-specific web router group for module routes
		langWebRouter := langRouter.Group("/")
		langWebRouter.Use(middleware.OptionalAuthMiddleware())

		// Set up auth routes directly for this language
		appLogger.Info("Setting up auth routes for language", logger.String("language", langCode))
		authService := auth.NewAuthService(db.DB, emailSender, Emitter)
		authController := auth.NewAuthController(authService, emailSender, appLogger)
		authController.SetupWebRoutes(langWebRouter, templateEngine)
		appLogger.Info("Language-specific auth routes initialized", logger.String("language", langCode))

		// Set up posts routes directly for this language without module registration
		appLogger.Info("Setting up posts routes for language", logger.String("language", langCode))
		postsService := posts.NewPostService(db.DB, Emitter, activeStorage, appLogger)
		postsController := posts.NewPostController(postsService, activeStorage, templateEngine)
		postsController.SetupWebRoutes(langWebRouter)
		appLogger.Info("Language-specific posts routes initialized", logger.String("language", langCode))
	}

	// Also initialize default (non-language-prefixed) routes for backward compatibility
	appInitializer := &app.AppModuleInitializer{
		DB:        db.DB,
		APIRouter: protectedGroup, // Use protected group for API
		WebRouter: webGroup,       // Use web group for HTML
		Logger:    appLogger,
		Emitter:   Emitter,
		Storage:   activeStorage,
		Template:  templateEngine,
	}
	appModules := appInitializer.InitializeModules(db.DB)
	appLogger.Info("Default application modules initialized", logger.Int("count", len(appModules)))

	// Add health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": cfg.Version,
		})
	})

	// Landing page route is handled in setupTemplateRoutes()

	// Initialize WebSocket module
	appLogger.Info("Initializing WebSocket module")
	wsHub := websocket.InitWebSocketModule(apiGroup)

	// Create application instance
	application := &Application{
		Config:  cfg,
		DB:      db,
		Router:  router,
		WSHub:   wsHub,
		Modules: modules,
		Logger:  appLogger,
		Emitter: Emitter,
	}

	appLogger.Info("Application Started Successfully!\n" +
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n" +
		"🔖 Version:      " + cfg.Version + "\n" +
		"🌍 Environment:  " + cfg.Env + "\n" +
		"🔌 Server:       " + cfg.ServerAddress + "\n" +
		"🌐 App URL:      " + cfg.BaseURL + "\n" +
		"🔗 API URL:      " + cfg.BaseURL + "/api\n" +
		"📚 Swagger Docs: " + cfg.BaseURL + "/swagger/index.html\n" +
		"📦 Modules:      " + fmt.Sprintf("%d", len(modules)) + "\n" +
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return application, nil
}

// setupTemplateRoutes initializes the template engine and sets up the landing page route
func setupTemplateRoutes(router *gin.Engine, logger logger.Logger, ts *language.TranslationService) *template.Engine {
	// Initialize template engine
	templateConfig := template.Config{
		TemplatesDir: "app/theme/default",
		LayoutsDir:   "app/theme/default/layouts",
		SharedDir:    "app/theme/default/shared",
	}
	templateEngine := template.NewEngine(templateConfig)
	templateEngine.RegisterDefaultHelpers()

	// Initialize and register language helpers
	logger.Info("Initializing language helpers for template engine")
	language.RegisterTemplateHelpers(templateEngine, ts) // Use the passed-in translationService (ts)

	// Now register templates with all helpers available
	if err := app.RegisterTemplates(templateEngine); err != nil {
		logger.Error("Failed to register templates: " + err.Error())
	} else {
		logger.Info("Templates registered successfully")
	}

	// Set the template engine on the router
	router.HTMLRender = templateEngine

	// Add landing page route
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, templates.PageLanding, gin.H{
			"title": "Welcome",
		})
	})

	logger.Info("Template engine and landing page route initialized")
	return templateEngine
}
