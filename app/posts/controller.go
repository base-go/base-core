package posts

import (
	"net/http"
	"strconv"

	"base/app/models"
	"base/app/templates"
	"base/core/language"
	"base/core/middleware"
	"base/core/storage"
	"base/core/template"

	// Added for types.PaginatedResponse
	"github.com/gin-gonic/gin"
)

type PostController struct {
	Service  *PostService
	Storage  *storage.ActiveStorage
	Template *template.Engine
}

// getTranslation gets a translation string from the context
// It handles cases where the translation service might not be available
func getTranslation(ctx *gin.Context, key string, defaultValue string) string {
	translationService, exists := ctx.Get("TranslationService")
	if !exists {
		return defaultValue
	}

	translated := translationService.(*language.TranslationService).Translate(key)
	if translated == key {
		// Key not found, return default
		return defaultValue
	}

	return translated
}

func NewPostController(service *PostService, storage *storage.ActiveStorage, templateEngine *template.Engine) *PostController {
	return &PostController{
		Service:  service,
		Storage:  storage,
		Template: templateEngine,
	}
}

func (c *PostController) SetupAPIRoutes(router *gin.RouterGroup) {
	// JSON API endpoints
	router.GET("/posts", c.List)        // Paginated list
	router.GET("/posts/all", c.ListAll) // Unpaginated list
	router.GET("/posts/:id", c.Get)
	router.POST("/posts", c.Create)
	router.PUT("/posts/:id", c.Update)
	router.DELETE("/posts/:id", c.Delete)
}

func (c *PostController) SetupWebRoutes(router *gin.RouterGroup) {
	// Public JSON endpoint for frontend - accessible without authentication
	// This route is registered on the language-specific group (e.g., /en/posts.json)
	router.GET("/posts.json", c.ListPublic)

	// Group for protected HTML view endpoints for posts
	// This group will also be under the language-specific prefix (e.g., /en/posts)
	postsWeb := router.Group("/posts")
	postsWeb.Use(middleware.AuthMiddleware()) // Apply authentication middleware
	{
		postsWeb.GET("", c.Index)            // e.g., GET /en/posts
		postsWeb.GET("/new", c.New)          // e.g., GET /en/posts/new
		postsWeb.GET("/:id", c.Show)         // e.g., GET /en/posts/:id
		postsWeb.GET("/:id/edit", c.Edit)    // e.g., GET /en/posts/:id/edit
		postsWeb.POST("", c.CreateWeb)       // e.g., POST /en/posts (form submission)
		postsWeb.PUT("/:id", c.UpdateWeb)    // e.g., PUT /en/posts/:id (form submission)
		postsWeb.DELETE("/:id", c.DeleteWeb) // e.g., DELETE /en/posts/:id (form submission)
	}
}

// Legacy method for backwards compatibility
func (c *PostController) Routes(router *gin.RouterGroup) {
	c.SetupAPIRoutes(router)
}

// CreatePost godoc
// @Summary Create a new Post
// @Description Create a new Post with the input payload
// @Tags Post
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param posts body models.CreatePostRequest true "Create Post request"
// @Success 201 {object} models.PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts [post]
func (c *PostController) Create(ctx *gin.Context) {
	var req models.CreatePostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.Service.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create item: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, item.ToResponse())
}

// GetPost godoc
// @Summary Get a Post
// @Description Get a Post by its id
// @Tags Post
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Post id"
// @Success 200 {object} models.PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /posts/{id} [get]
func (c *PostController) Get(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid id format"})
		return
	}

	item, err := c.Service.GetById(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Item not found"})
		return
	}

	ctx.JSON(http.StatusOK, item.ToResponse())
}

// ListPosts godoc
// @Summary List posts
// @Description Get a list of posts
// @Tags Post
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Success 200 {object} types.PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts [get]
func (c *PostController) List(ctx *gin.Context) {
	var page, limit *int

	if pageStr := ctx.Query("page"); pageStr != "" {
		if pageNum, err := strconv.Atoi(pageStr); err == nil && pageNum > 0 {
			page = &pageNum
		} else {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid page number"})
			return
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if limitNum, err := strconv.Atoi(limitStr); err == nil && limitNum > 0 {
			limit = &limitNum
		} else {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid limit number"})
			return
		}
	}

	paginatedResponse, err := c.Service.GetAll(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch items: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, paginatedResponse)
}

// ListAllPosts godoc
// @Summary List all posts without pagination
// @Description Get a list of all posts without pagination
// @Tags Post
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} types.PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/all [get]
func (c *PostController) ListAll(ctx *gin.Context) {
	paginatedResponse, err := c.Service.GetAll(nil, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch all items: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, paginatedResponse)
}

// UpdatePost godoc
// @Summary Update a Post
// @Description Update a Post by its id
// @Tags Post
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Post id"
// @Param posts body models.UpdatePostRequest true "Update Post request"
// @Success 200 {object} models.PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/{id} [put]
func (c *PostController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid id format"})
		return
	}

	var req models.UpdatePostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.Service.Update(uint(id), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update item: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, item.ToResponse())
}

// DeletePost godoc
// @Summary Delete a Post
// @Description Delete a Post by its id
// @Tags Post
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Post id"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/{id} [delete]
func (c *PostController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid id format"})
		return
	}

	if err := c.Service.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete item: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{Message: "Item deleted successfully"})
}

// View rendering methods
func (c *PostController) Index(ctx *gin.Context) {
	paginatedResponse, err := c.Service.GetAll(nil, nil)
	if err != nil {
		ctx.HTML(http.StatusInternalServerError, templates.PageError, gin.H{
			"error": getTranslation(ctx, "errors.failed_to_fetch_posts", "Failed to fetch posts: ") + err.Error(),
			"title": getTranslation(ctx, "common.error", "Error"),
		})
		return
	}

	c.Template.RenderWithLayout(ctx.Writer, templates.PostIndex, templates.LayoutDefault, gin.H{
		"posts": paginatedResponse.Data,
		"title": getTranslation(ctx, "posts.title", "Posts"),
	}, ctx)
}

// ListPublic provides a public JSON endpoint for posts (no authentication required)
func (c *PostController) ListPublic(ctx *gin.Context) {
	paginatedResponse, err := c.Service.GetAll(nil, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch posts: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, paginatedResponse)
}

func (c *PostController) Show(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.Template.RenderWithLayout(ctx.Writer, templates.PageError, templates.LayoutDefault, gin.H{
			"error": getTranslation(ctx, "errors.invalid_post_id", "Invalid post ID"),
			"title": getTranslation(ctx, "common.error", "Error"),
		}, ctx)
		return
	}

	post, err := c.Service.GetById(uint(id))
	if err != nil {
		c.Template.RenderWithLayout(ctx.Writer, templates.PageError, templates.LayoutDefault, gin.H{
			"error": getTranslation(ctx, "errors.post_not_found", "Post not found"),
			"title": getTranslation(ctx, "common.error", "Error"),
		}, ctx)
		return
	}

	c.Template.RenderWithLayout(ctx.Writer, templates.PostShow, templates.LayoutDefault, gin.H{
		"post":  post,
		"title": getTranslation(ctx, "posts.details", "Post Details"),
	}, ctx)
}

func (c *PostController) New(ctx *gin.Context) {
	c.Template.RenderWithLayout(ctx.Writer, "posts/new.html", "app.html", gin.H{
		"title": getTranslation(ctx, "posts.new_post", "New Post"),
	}, ctx)
}

func (c *PostController) Edit(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.Template.RenderWithLayout(ctx.Writer, templates.PageError, templates.LayoutDefault, gin.H{
			"error": "Invalid post ID",
			"title": "Error",
		}, ctx)
		return
	}

	post, err := c.Service.GetById(uint(id))
	if err != nil {
		ctx.HTML(http.StatusNotFound, templates.PageError, gin.H{
			"error": "Post not found",
		})
		return
	}

	c.Template.RenderWithLayout(ctx.Writer, templates.PostEdit, "app.html", gin.H{
		"post":  post,
		"title": "Edit Post",
	}, ctx)
}

// Web-specific handlers that redirect after form submission
func (c *PostController) CreateWeb(ctx *gin.Context) {
	var req models.CreatePostRequest
	if err := ctx.ShouldBind(&req); err != nil {
		c.Template.RenderWithLayout(ctx.Writer, "posts/new.html", "app.html", gin.H{
			"error": err.Error(),
			"title": "New Post",
		}, ctx)
		return
	}

	item, err := c.Service.Create(&req)
	if err != nil {
		c.Template.RenderWithLayout(ctx.Writer, "posts/new.html", "app.html", gin.H{
			"error": "Failed to create post: " + err.Error(),
			"title": "New Post",
		}, ctx)
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/posts/"+strconv.Itoa(int(item.Id)))
}

func (c *PostController) UpdateWeb(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.Template.RenderWithLayout(ctx.Writer, templates.PageError, templates.LayoutDefault, gin.H{
			"error": "Invalid post ID",
			"title": "Error",
		}, ctx)
		return
	}

	var req models.UpdatePostRequest
	if err := ctx.ShouldBind(&req); err != nil {
		post, _ := c.Service.GetById(uint(id))
		c.Template.RenderWithLayout(ctx.Writer, "posts/edit.html", "app.html", gin.H{
			"error": err.Error(),
			"post":  post,
			"title": "Edit Post",
		}, ctx)
		return
	}

	item, err := c.Service.Update(uint(id), &req)
	if err != nil {
		post, _ := c.Service.GetById(uint(id))
		c.Template.RenderWithLayout(ctx.Writer, "posts/edit.html", "app.html", gin.H{
			"error": "Failed to update post: " + err.Error(),
			"post":  post,
			"title": "Edit Post",
		}, ctx)
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/posts/"+strconv.Itoa(int(item.Id)))
}

func (c *PostController) DeleteWeb(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.Template.RenderWithLayout(ctx.Writer, templates.PageError, templates.LayoutDefault, gin.H{
			"error": "Invalid post ID",
			"title": "Error",
		}, ctx)
		return
	}

	if err := c.Service.Delete(uint(id)); err != nil {
		c.Template.RenderWithLayout(ctx.Writer, templates.PageError, templates.LayoutDefault, gin.H{
			"error": "Failed to delete post: " + err.Error(),
			"title": "Error",
		}, ctx)
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/posts")
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
