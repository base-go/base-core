package users

import (
	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(router *gin.Engine, userController *UserController) {
	users := router.Group("/users")
	{
		users.POST("/", userController.CreateUser)
		users.GET("/", userController.GetAllUsers)
		users.GET("/:id", userController.GetUser)
		users.PUT("/:id", userController.UpdateUser)
		users.DELETE("/:id", userController.DeleteUser)
	}
}
