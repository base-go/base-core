package core

import (
	"base/app"
	coreapp "base/core/app"
	"base/core/config"
	"base/core/database"
	"base/core/email"
	"base/core/emitter"
	"base/core/logger"
	"base/core/middleware"
	"base/core/module"
	"base/core/storage"
	"base/core/websocket"
	"time"

	"fmt"

	"github.com/gin-contrib/cors"
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

	// Register attachments configuration
	activeStorage.RegisterAttachment("users", storage.AttachmentConfig{
		Field:             "avatar",
		Path:              "users",
		AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".gif", ".webp"},
		MaxFileSize:       2 << 20, // 2MB
		Multiple:          false,
	})

	activeStorage.RegisterAttachment("users", storage.AttachmentConfig{
		Field:             "documents",
		Path:              "users",
		AllowedExtensions: []string{".pdf", ".doc", ".docx"},
		MaxFileSize:       10 << 20, // 10MB
		Multiple:          true,
	})

	appLogger.Info("Storage service initialized successfully",
		logger.String("provider", cfg.StorageProvider),
		logger.String("path", cfg.StoragePath))

	// Set up Gin
	router := gin.New()
	router.Use(gin.Recovery())

	// Set up middleware
	router.Use(middleware.Logger(appLogger))

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
	)
	modules := core.Modules
	appLogger.Info("Core modules initialized", logger.Int("count", len(modules)))

	// Register core module routes
	//core.RegisterRoutes()

	// Initialize application modules
	appLogger.Info("Initializing application modules")
	appInitializer := &app.AppModuleInitializer{
		Router:  protectedGroup, // Use protected group for app modules
		Logger:  appLogger,
		Emitter: Emitter,
		Storage: activeStorage,
	}
	appInitializer.InitializeModules(db.DB)

	// Add health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": cfg.Version,
		})
	})

	// Add ping route
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"version": cfg.Version,
			"swagger": "/swagger/index.html",
		})
	})

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
