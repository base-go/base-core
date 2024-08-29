package main

import (
	"basego/base/app/posts"
	"basego/base/app/users"
	"basego/base/core/config"
	"basego/base/core/database"
	"basego/base/core/middleware"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize configuration
	config.Init()

	// Initialize database
	database.InitDB()

	// Set up Gin
	router := gin.Default()

	// Setup Posts
	postService := &posts.PostService{DB: database.DB}
	postController := &posts.PostController{PostService: postService}
	posts.SetupPostRoutes(router, postController)

	// Setup Users
	userService := &users.UserService{DB: database.DB}
	userController := &users.UserController{UserService: userService}
	users.SetupUserRoutes(router, userController)

	// Apply auth middleware to protected routes
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// Add protected routes here
		// Example: protected.POST("/posts", postController.CreatePost)
	}

	// Start the server
	log.Printf("Server starting on %s", config.AppConfig.ServerAddress)
	log.Fatal(router.Run(config.AppConfig.ServerAddress))
}
