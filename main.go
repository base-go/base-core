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
	log.Infof("Server starting on %s", app.Config.ServerAddress)
	if err := app.Router.Run(app.Config.ServerAddress); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
