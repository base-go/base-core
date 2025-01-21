package main

import (
	"base/core"
	_ "base/docs" // Import the Swagger docs
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// @title Base API
// @version 1.0
// @description This is the API documentation for Base
// @host localhost:8001
// @BasePath /api
// @schemes http https
// @produces json
// @consumes json
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-Api-Key

// DeletedAt is a type definition for GORM's soft delete functionality
type DeletedAt gorm.DeletedAt

// Time represents a time.Time
type Time time.Time

func main() {
	// Load the .env file
	if err := godotenv.Load(); err != nil {
		log.Warn("Error loading .env file")
	}

	// Bootstrap the application
	app, err := core.StartApplication()
	if err != nil {
		log.Fatalf("Failed to bootstrap application: %v", err)
	}

	// Start the server
	if err := app.Router.Run(app.Config.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
