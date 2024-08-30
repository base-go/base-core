package core

import (
	"base/app"
	"base/core/config"
	"base/core/database"
	"base/core/email"
	"base/core/file"
	"base/core/middleware"
	"base/core/websocket"
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"

	_ "base/docs" // Import for Swagger docs
)

type Application struct {
	Config *config.Config
	DB     *database.Database
	Router *gin.Engine
	WSHub  *websocket.Hub
}

func StartApplication() (*Application, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Warn("Error loading .env file")
	}

	// Initialize config
	cfg := config.NewConfig()

	// Initialize the database
	db, err := database.InitDB(cfg)
	if err != nil {
		return nil, err
	}

	// Initialize Logrus logger
	logger := InitializeLogger()
	log.SetOutput(logger.Out)
	log.SetFormatter(logger.Formatter)
	log.SetLevel(logger.Level)

	// Initialize email sender
	emailSender, err := email.NewSender(cfg)
	if err != nil {
		log.Errorf("Failed to initialize email sender: %v", err)
		return nil, fmt.Errorf("failed to initialize email sender: %w", err)
	}

	// Set up Gin
	gin.SetMode(cfg.Env)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LogrusLogger(logger))

	// Set up static file serving
	router.Static("/static", "./static")

	// Set up storage file serving
	router.Static("/storage", "./storage")

	// Set up CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.CORSAllowedOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Api-Key"}
	router.Use(cors.New(corsConfig))

	// Set up rate limiter
	limiter := middleware.NewIPRateLimiter(rate.Limit(1), 5) // 1 request per second with burst of 5
	router.Use(middleware.RateLimitMiddleware(limiter))

	// Set up Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create a new router group for API routes
	apiGroup := router.Group("/api/v1")
	apiGroup.Use(middleware.APIKeyMiddleware())

	// Initialize application modules
	InitializeCoreModules(database.DB, apiGroup, emailSender, logger)
	app.InitializeModules(database.DB, apiGroup)

	// Add ping route to the main router, not to app.Router
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	if err := email.Initialize(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize email sender: %w", err)
	}

	// Initialize WebSocket module
	wsHub := websocket.InitWebSocketModule(apiGroup)

	file.InitFileModule(apiGroup)

	// Create and return the Application instance
	application := &Application{
		Config: cfg,
		DB:     db,
		Router: router,
		WSHub:  wsHub,
	}

	return application, nil
}
