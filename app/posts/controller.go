package posts

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "base/app/models"
)

type PostController struct {
    PostService *PostService
}

func NewPostController(service *PostService) *PostController {
    return &PostController{
        PostService: service,
    }
}

func (c *PostController) Routes(router *gin.RouterGroup) {
    router.GET("/posts", c.List)       // Paginated list
    router.GET("/posts/all", c.ListAll) // Unpaginated list
    router.GET("/posts/:id", c.Get)
    router.POST("/posts", c.Create)
    router.PUT("/posts/:id", c.Update)
    router.DELETE("/posts/:id", c.Delete)
    router.PUT("/posts/:id/image", c.UploadImage)
}

// CreatePost godoc
// @Summary Create a new Post
// @Description Create a new Post with the input payload
// @Tags Post
// @Security ApiKeyAuth
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

    item, err := c.PostService.Create(&req)
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

    item, err := c.PostService.GetById(uint(id))
    if err != nil {
        ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Item not found"})
        return
    }

    ctx.JSON(http.StatusOK, item.ToResponse())
}

// Listposts godoc
// @Summary List posts
// @Description Get a list of posts
// @Tags Post
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param search query string false "Search term for filtering results"
// @Success 200 {object} types.PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts [get]
func (c *PostController) List(ctx *gin.Context) {
    page := 1
    limit := 10

    if p, err := strconv.Atoi(ctx.DefaultQuery("page", "1")); err == nil && p > 0 {
        page = p
    } else if err != nil {
        ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid 'page' parameter"})
        return
    }

    if l, err := strconv.Atoi(ctx.DefaultQuery("limit", "10")); err == nil && l > 0 && l <= 100 {
        limit = l
    } else if err != nil {
        ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid 'limit' parameter"})
        return
    }

    var search *string
    if searchTerm := ctx.Query("search"); searchTerm != "" {
        search = &searchTerm
    }

    paginatedResponse, err := c.PostService.GetAll(&page, &limit, search)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch items: " + err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, paginatedResponse)
}

// ListAllposts godoc
// @Summary List all posts
// @Description Get a list of all posts without pagination
// @Tags Post
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {object} types.PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/all [get]
func (c *PostController) ListAll(ctx *gin.Context) {
    paginatedResponse, err := c.PostService.GetAll(nil, nil, nil)
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

    item, err := c.PostService.Update(uint(id), &req)
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

    if err := c.PostService.Delete(uint(id)); err != nil {
        ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete item: " + err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, SuccessResponse{Message: "Item deleted successfully"})
}
// UploadImage godoc
// @Summary Upload Image for a Post
// @Description Upload Image for a Post by its id
// @Tags Post
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Post id"
// @Param file formData file true "Image file to upload"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/{id}/image [put]
func (c *PostController) UploadImage(ctx *gin.Context) {
    id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid id format"})
        return
    }

    file, err := ctx.FormFile("file")
    if err != nil {
        ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to get file: " + err.Error()})
        return
    }

    if err := c.PostService.UploadImage(uint(id), file); err != nil {
        ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to upload file: " + err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, SuccessResponse{Message: "File uploaded successfully"})
}

type ErrorResponse struct {
    Error string `json:"error"`
}

type SuccessResponse struct {
    Message string `json:"message"`
}
