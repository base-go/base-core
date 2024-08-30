package main

import (
	"os"

	"base/app"
	"base/core"
	"base/core/config"
	"base/core/database"
	"base/core/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "base/docs" // Import for Swagger docs
)

// @title Base API
// @version 1.0
// @description This is the API documentation for Base
// @host localhost:8080
// @BasePath /api/v1
// @schemes http
// @produces json
// @consumes json
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-Api-Key
func main() {
	// Load .env file before initializing the logger to use the GIN_MODE variable
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize the database
	cfg := config.NewConfig()
	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize the database: %v", err)
	}

	// Initialize Logrus logger based on the environment
	logger := core.InitializeLogger()
	log.SetOutput(logger.Out)
	log.SetFormatter(logger.Formatter)
	log.SetLevel(logger.Level)

	log.Info("Starting the application")
	log.Info("Database connected, driver:  " + cfg.DBDriver)

	log.Info("Core initialized mode: ", os.Getenv("GIN_MODE"))
	gin.SetMode(os.Getenv("GIN_MODE"))

	// Set up Gin
	router := gin.New()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Use(gin.Recovery())
	router.Use(middleware.LogrusLogger(logger))
	// cors.Default() setup the middleware with default options being
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Api-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}

	router.Use(cors.New(corsConfig))

	// Setup Swagger

	// Create a new router group for API routes with authentication
	apiGroup := router.Group("/api/v1")
	//	apiGroup.Use(middleware.APIKeyMiddleware())
	//	apiGroup.Use(middleware.AuthMiddleware())
	// Initialize application modules with the authenticated API group
	app.InitializeModules(database.DB, apiGroup)

	// Start the server
	port := cfg.ServerAddress
	log.Infof("Server starting on %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
