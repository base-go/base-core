package main

import (
	"base/core"

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
