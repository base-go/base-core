package core

import (
	"base/app"
	"base/core/app/auth"
	"base/core/config"
	"base/core/database"
	"base/core/email"
	"base/core/emitter"
	"base/core/logger"
	"base/core/middleware"
	"base/core/module"
	"base/core/storage"
	"base/core/websocket"

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
		Endpoint:  cfg.StorageEndpoint,
		Bucket:    cfg.StorageBucket,
		CDN:       cfg.CDN,
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
		AllowedExtensions: []string{".jpg", ".jpeg", ".png"},
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
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.CORSAllowedOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Api-Key"}
	router.Use(cors.New(corsConfig))

	// Set up Swagger
	router.GET("/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.PersistAuthorization(true)))
	appLogger.Info("Swagger documentation enabled")

	// Create API router group with API Key middleware
	apiGroup := router.Group("/api")
	apiGroup.Use(middleware.APIKeyMiddleware())

	// Create auth group that only requires API key
	authGroup := apiGroup.Group("/")

	// Create a protected group that requires Bearer token
	protectedGroup := apiGroup.Group("/")
	protectedGroup.Use(middleware.AuthMiddleware())

	// Initialize auth module with authGroup
	authModule := auth.NewAuthModule(
		db.DB,
		authGroup,
		emailSender,
		appLogger.GetZapLogger(),
		Emitter,
	)
	modules := make(map[string]module.Module)
	modules["auth"] = authModule

	// Initialize remaining core modules with protectedGroup
	moduleInit := ModuleInitializer{
		DB:          db.DB,
		Router:      protectedGroup,
		EmailSender: emailSender,
		Logger:      appLogger,
		Emitter:     Emitter,
		Storage:     activeStorage,
	}

	// Initialize remaining core modules
	remainingModules := InitializeRemainingCoreModules(moduleInit)
	for k, v := range remainingModules {
		modules[k] = v
	}

	// Initialize application modules
	appLogger.Info("Initializing application modules")
	appInitializer := &app.AppModuleInitializer{
		Router:  protectedGroup,
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
