package posts

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PostController struct {
	PostService *PostService
}

func (pc *PostController) CreatePost(c *gin.Context) {
	var post Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := pc.PostService.CreatePost(&post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	c.JSON(http.StatusCreated, post)
}

func (pc *PostController) GetPost(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	post, err := pc.PostService.GetPostByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, post)
}

func (pc *PostController) GetAllPosts(c *gin.Context) {
	posts, err := pc.PostService.GetAllPosts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		return
	}

	c.JSON(http.StatusOK, posts)
}

func (pc *PostController) UpdatePost(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var post Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post.ID = uint(id)
	if err := pc.PostService.UpdatePost(&post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}

	c.JSON(http.StatusOK, post)
}

func (pc *PostController) DeletePost(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := pc.PostService.DeletePost(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}
