package main

import (
	"base/core"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"

	_ "base/docs" // Import for Swagger docs
)

func main() {
	// Load the .env file
	if err := godotenv.Load(); err != nil {
		log.Warn("Error loading .env file")
	}

	// If there are command line arguments, execute the command
	if len(os.Args) > 1 {
		core.ExecuteCommand(os.Args)
		return
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
	log.Infof("Server starting on http://localhost%s", app.Config.ServerAddress)
	if err := app.Router.Run(app.Config.ServerAddress); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
