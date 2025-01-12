package main

import (
	"base/core"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"

	_ "base/docs" // Import for Swagger docs
)

// @title Base API
// @info This is the API documentation for Base
// @servers localhost:8090
// @BasePath /api
// @version 1.5
// @description This is the API documentation for Albafone
// @schemes http https
// @produces json
// @consumes json
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-Api-Key
func main() {
	// Load the .env file
	if err := godotenv.Load(); err != nil {
		log.Warn("Error loading .env file")
	}

	// if debug mode is enabled, run swag init
	if os.Getenv("ENV") == "debug" {
		log.Info("Running swag init")
		cmd := exec.Command("swag", "init", "--parseDependency", "--parseInternal", "--parseVendor")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to run swag init: %v", err)
		}
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
