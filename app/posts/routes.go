package posts

import (
	"github.com/gin-gonic/gin"
)

func SetupPostRoutes(router *gin.Engine, postController *PostController) {
	posts := router.Group("/posts")
	{
		posts.POST("/", postController.CreatePost)
		posts.GET("/", postController.GetAllPosts)
		posts.GET("/:id", postController.GetPost)
		posts.PUT("/:id", postController.UpdatePost)
		posts.DELETE("/:id", postController.DeletePost)
	}
}
