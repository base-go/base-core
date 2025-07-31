package core

import (
	"base/core/config"
	"base/core/database"
	"base/core/email"
	"base/core/emitter"
	"base/core/logger"
	"base/core/middleware"
	"base/core/storage"
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// CoreApplication represents the core infrastructure
type CoreApplication struct {
	Config      *config.Config
	DB          *database.Database
	Router      *gin.Engine
	Logger      logger.Logger
	Emitter     *emitter.Emitter
	Storage     *storage.ActiveStorage
	EmailSender email.Sender
}

// StartCore initializes only the core infrastructure and returns it
// This should be called by the app layer, not the other way around
func StartCore() (*CoreApplication, error) {
	// Initialize config
	cfg := config.NewConfig()

	// Initialize logger
	logConfig := logger.Config{
		Environment: cfg.Env,
		LogPath:     "logs",
		Level:       "debug",
	}
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	log.Info("🚀 Starting Core Infrastructure",
		logger.String("version", cfg.Version),
		logger.String("environment", cfg.Env))

	// Initialize database
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Error("Failed to initialize database", logger.String("error", err.Error()))
		return nil, fmt.Errorf("database initialization failed: %w", err)
	}
	log.Info("Database initialized successfully")

	// Initialize emitter
	emitter := &emitter.Emitter{}

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
	activeStorage, err := storage.NewActiveStorage(db.DB, storageConfig)
	if err != nil {
		log.Error("Failed to initialize storage", logger.String("error", err.Error()))
		return nil, fmt.Errorf("storage initialization failed: %w", err)
	}
	log.Info("Storage initialized successfully")

	// Initialize email sender
	emailSender, err := email.NewSender(cfg)
	if err != nil {
		log.Error("Failed to initialize email sender", logger.String("error", err.Error()))
		// Continue without email functionality
		emailSender = nil
	}

	// Core services are initialized on-demand in the app layer
	log.Info("Core infrastructure ready for service initialization")

	// Setup router with core middleware
	router := setupCoreRouter(cfg, log)

	coreApp := &CoreApplication{
		Config:      cfg,
		DB:          db,
		Router:      router,
		Logger:      log,
		Emitter:     emitter,
		Storage:     activeStorage,
		EmailSender: emailSender,
	}

	log.Info("✅ Core infrastructure ready")
	return coreApp, nil
}

// setupCoreRouter initializes Gin router with core middleware only
func setupCoreRouter(cfg *config.Config, log logger.Logger) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	// Add logger middleware
	router.Use(middleware.Logger(log))

	// Setup static file serving
	router.Static("/static", "./static")
	router.Static("/storage", "./storage")

	// Setup CORS
	corsConfig := cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "authorization", "X-Api-Key", "Base-Orgid", "Base-*"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))

	// Setup Swagger
	router.GET("/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.PersistAuthorization(true)))
	log.Info("Swagger documentation enabled")

	// Add health check routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": cfg.Version,
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"version": cfg.Version,
			"swagger": "/swagger/index.html",
		})
	})

	return router
}

// Run starts the core application server
func (app *CoreApplication) Run() error {
	port := app.Config.ServerPort
	if app.Config.ServerPort != "" {
		port = app.Config.ServerPort
	}

	app.Logger.Info("🌐 Core server starting",
		logger.String("port", port),
		logger.String("swagger", app.Config.BaseURL+"/swagger/index.html"))

	return app.Router.Run(port)
}
