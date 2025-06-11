package posts

import (
	"net/http"
	"strconv"

	"base/app/models"
	"base/core/router"
	"base/core/storage"

	// Added for types.PaginatedResponse
	"github.com/gin-gonic/gin"
)

type PostApiController struct {
	Service *PostService
	Storage *storage.ActiveStorage
}

func NewPostApiController(service *PostService, storage *storage.ActiveStorage) *PostApiController {
	return &PostApiController{
		Service: service,
		Storage: storage,
	}
}

func (c *PostApiController) Routes(ginRouter *gin.RouterGroup) {
	// Convert to Rails-style router
	r := router.NewFromGroup(ginRouter)

	// Use Rails-style syntax for API endpoints
	r.Get("/posts", c.List). // Paginated list
					Get("/posts/all", c.ListAll). // Unpaginated list
					Get("/posts/:id", c.Get).
					Post("/posts", c.Create).
					Put("/posts/:id", c.Update).
					Delete("/posts/:id", c.Delete)
}

// CreatePost godoc
// @Summary Create a new Post
// @Description Create a new Post with the input payload
// @Tags Post
// @Security ApiKeyApp
// @Security BearerApp
// @Accept json
// @Produce json
// @Param posts body models.CreatePostRequest true "Create Post request"
// @Success 201 {object} models.PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts [post]
func (c *PostApiController) Create(ctx *gin.Context) {
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
// @Security ApiKeyApp
// @Security BearerApp
// @Accept json
// @Produce json
// @Param id path int true "Post id"
// @Success 200 {object} models.PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /posts/{id} [get]
func (c *PostApiController) Get(ctx *gin.Context) {
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
// @Security ApiKeyApp
// @Security BearerApp
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Success 200 {object} types.PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts [get]
func (c *PostApiController) List(ctx *gin.Context) {
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
// @Security ApiKeyApp
// @Security BearerApp
// @Accept json
// @Produce json
// @Success 200 {object} types.PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/all [get]
func (c *PostApiController) ListAll(ctx *gin.Context) {
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
// @Security ApiKeyApp
// @Security BearerApp
// @Accept json
// @Produce json
// @Param id path int true "Post id"
// @Param posts body models.UpdatePostRequest true "Update Post request"
// @Success 200 {object} models.PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/{id} [put]
func (c *PostApiController) Update(ctx *gin.Context) {
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
// @Security ApiKeyApp
// @Security BearerApp
// @Accept json
// @Produce json
// @Param id path int true "Post id"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/{id} [delete]
func (c *PostApiController) Delete(ctx *gin.Context) {
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

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
