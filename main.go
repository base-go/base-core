package main

import (
	"base/core"
	_ "base/docs" // Import the Swagger docs
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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
// @description API Key for authentication

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your token with the prefix "Bearer "

// DeletedAt is a type definition for GORM's soft delete functionality
type DeletedAt gorm.DeletedAt

// Time represents a time.Time
type Time time.Time

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func main() {
	// Load the .env file
	if err := godotenv.Load(); err != nil {
		log.Warn("Error loading .env file")
	}

	// Disable Gin's default logger
	gin.SetMode(gin.ReleaseMode)

	// Bootstrap the application
	app, err := core.StartApplication()
	if err != nil {
		log.Fatalf("Failed to bootstrap application: %v", err)
	}

	// Get local IP and format server address
	localIP := getLocalIP()
	addr := app.Config.ServerAddress
	if strings.HasPrefix(addr, ":") {
		addr = "0.0.0.0" + addr
	}

	fmt.Printf("\nServer is running at:\n")
	fmt.Printf("- Local:   http://localhost%s\n", strings.TrimPrefix(addr, "0.0.0.0"))
	fmt.Printf("- Network: http://%s%s\n\n", localIP, strings.TrimPrefix(addr, "0.0.0.0"))

	// Start the server
	if err := app.Router.Run(app.Config.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
