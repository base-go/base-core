package core

import (
	"base/app"
	"base/core/config"
	"base/core/database"
	"base/core/file"
	"base/core/middleware"
	"base/core/websocket"

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

	// Set up Gin
	gin.SetMode(cfg.Env)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LogrusLogger(logger))

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
	InitializeCoreModules(database.DB, apiGroup)
	app.InitializeModules(database.DB, apiGroup)

	file.InitFileModule(apiGroup)

	return &Application{
		Config: cfg,
		DB:     db,
		Router: router,
	}, nil
}
