package main

import (
	"time"
	"gorm.io/gorm"
)

// @title Base API
// @version 2.0.0
// @description This is the API documentation for Base
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

func main() {
	// Initialize and start the Base application
	app := New()
	
	if err := app.Start(); err != nil {
		panic(err)
	}
}
