package core

import (
	"base/app"
	"base/core/config"
	"base/core/database"
	"base/core/emitter"
	"base/core/initializer"
	"base/core/logger"
	"base/core/module"
	"base/core/websocket"
	"fmt"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Application struct {
	Config  *config.Config
	DB      *database.Database
	Router  *gin.Engine
	WSHub   *websocket.Hub
	Modules []module.Module
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

	// Email and storage services are initialized in the initializer package

	// Initialize main application using initializer with app registry
	appLogger.Info("Initializing main application")
	registry := &app.Registry{}
	mainApp, err := initializer.NewAppWithInterface(cfg, registry, registry)
	if err != nil {
		appLogger.Error("Failed to initialize main application",
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("main application initialization failed: %w", err)
	}

	router := mainApp.Router

	// Set up Swagger
	router.GET("/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.PersistAuthorization(true)))
	appLogger.Info("Swagger documentation enabled")

	// Core modules are now handled by the initializer

	// Add health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": cfg.Version,
		})
	})

	// Initialize WebSocket module
	appLogger.Info("Initializing WebSocket module")
	WsGroup := mainApp.Router.Group("/ws")
	wsHub := websocket.InitWebSocketModule(WsGroup)

	// Create application instance
	application := &Application{
		Config:  cfg,
		DB:      db,
		Router:  router,
		WSHub:   wsHub,
		Modules: mainApp.Modules,
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
		"📦 Modules:      " + fmt.Sprintf("%d", len(mainApp.Modules)) + "\n" +
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return application, nil
}
