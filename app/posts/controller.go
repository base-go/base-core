package posts

import (
	"net/http"
	"strconv"

	"base/app/models"
	"base/core/language"
	"base/core/layout"
	"base/core/middleware"
	"base/core/router"
	"base/core/storage"

	"github.com/gin-gonic/gin"
)

type PostController struct {
	*layout.Controller
	Service *PostService
	Storage *storage.ActiveStorage
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

func NewPostController(service *PostService, storage *storage.ActiveStorage, layoutEngine *layout.Engine) *PostController {
	return &PostController{
		Controller: layout.NewAppController(layoutEngine), // Use app layout
		Service:    service,
		Storage:    storage,
	}
}

func (c *PostController) Routes(ginRouter *gin.RouterGroup) {
	// Convert to Rails-style router
	r := router.NewFromGroup(ginRouter)

	// Public routes for posts using Rails-style syntax
	r.
		Get("/posts", c.Index).   // Public post listing
		Get("/posts/:id", c.Show) // Public post viewing

	// Protected routes that require authentication
	r.Use(middleware.AuthMiddleware()).Namespace("/posts", func(authPosts *router.Router) {
		authPosts.Get("/new", c.New). // Create new post form
						Get("/:id/edit", c.Edit). // Edit post form
						Post("", c.Create).       // Create post (form submission)
						Put("/:id", c.Update).    // Update post (form submission)
						Delete("/:id", c.Delete)  // Delete post (form submission)
	})
}

// View rendering methods
func (c *PostController) Index(ctx *gin.Context) {
	paginatedResponse, err := c.Service.GetAll(nil, nil)
	if err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": getTranslation(ctx, "errors.failed_to_fetch_posts", "Failed to fetch posts: ") + err.Error(),
			"title": getTranslation(ctx, "common.error", "Error"),
		})
		return
	}

	c.View("posts/index.html").WithTitle(getTranslation(ctx, "posts.title", "Posts")).WithData(gin.H{"posts": paginatedResponse.Data}).Render(ctx)
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
		c.View("error.html").WithTitle("Error").WithError(getTranslation(ctx, "errors.invalid_post_id", "Invalid post ID")).Render(ctx)
		return
	}

	post, err := c.Service.GetById(uint(id))
	if err != nil {
		c.View("error.html").WithTitle(getTranslation(ctx, "common.error", "Error")).WithError(getTranslation(ctx, "errors.post_not_found", "Post not found")).Render(ctx)
		return
	}

	c.View("posts/show.html").WithTitle(getTranslation(ctx, "posts.details", "Post Details")).WithData(gin.H{"post": post}).Render(ctx)
}

func (c *PostController) New(ctx *gin.Context) {
	c.View("posts/new.html").WithTitle(getTranslation(ctx, "posts.new_post", "New Post")).Render(ctx)
}

func (c *PostController) Edit(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.View("error.html").WithTitle(getTranslation(ctx, "common.error", "Error")).WithError(getTranslation(ctx, "errors.invalid_post_id", "Invalid post ID")).Render(ctx)
		return
	}

	post, err := c.Service.GetById(uint(id))
	if err != nil {
		c.View("error.html").WithTitle("Error").WithError("Post not found").Render(ctx)
		return
	}

	c.View("posts/edit.html").WithTitle("Edit Post").WithData(gin.H{"post": post}).Render(ctx)
}

// Web-specific handlers that redirect after form submission
func (c *PostController) Create(ctx *gin.Context) {
	var req models.CreatePostRequest
	if err := ctx.ShouldBind(&req); err != nil {
		c.View("posts/new.html").WithTitle("New Post").WithError(err.Error()).With("form", req).Render(ctx)
		return
	}

	item, err := c.Service.Create(&req)
	if err != nil {
		c.View("posts/new.html").WithTitle("New Post").WithError("Failed to create post: " + err.Error()).WithData(gin.H{"form": req}).Render(ctx)
		return
	}

	c.Redirect("/posts/" + strconv.Itoa(int(item.Id))).Execute(ctx)
}

func (c *PostController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.View("error.html").WithTitle("Error").WithError("Invalid post ID").Render(ctx)
		return
	}

	var req models.UpdatePostRequest
	if err := ctx.ShouldBind(&req); err != nil {
		post, _ := c.Service.GetById(uint(id))
		c.View("posts/edit.html").WithTitle("Edit Post").WithError(err.Error()).WithData(gin.H{"post": post, "form": req}).Render(ctx)
		return
	}

	item, err := c.Service.Update(uint(id), &req)
	if err != nil {
		post, _ := c.Service.GetById(uint(id))
		c.View("posts/edit.html").WithTitle("Edit Post").WithError("Failed to update post: " + err.Error()).WithData(gin.H{"post": post, "form": req}).Render(ctx)
		return
	}

	c.Redirect("/posts/" + strconv.Itoa(int(item.Id))).Execute(ctx)
}

func (c *PostController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.View("error.html").WithTitle("Error").WithError("Invalid post ID").Render(ctx)
		return
	}

	if err := c.Service.Delete(uint(id)); err != nil {
		c.View("error.html").WithTitle("Error").WithError("Failed to delete post: " + err.Error()).Render(ctx)
		return
	}

	c.Redirect("/posts").Execute(ctx)
}
