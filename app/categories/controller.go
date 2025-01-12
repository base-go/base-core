package categories

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "base/app/models"
)

type CategoryController struct {
    CategoryService *CategoryService
}

func NewCategoryController(service *CategoryService) *CategoryController {
    return &CategoryController{
        CategoryService: service,
    }
}

func (c *CategoryController) Routes(router *gin.RouterGroup) {
    router.GET("/categories", c.List)       // Paginated list
    router.GET("/categories/all", c.ListAll) // Unpaginated list
    router.GET("/categories/:id", c.Get)
    router.POST("/categories", c.Create)
    router.PUT("/categories/:id", c.Update)
    router.DELETE("/categories/:id", c.Delete)
    router.PUT("/categories/:id/image", c.UploadImage)
}

// CreateCategory godoc
// @Summary Create a new Category
// @Description Create a new Category with the input payload
// @Tags Category
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param categories body models.CreateCategoryRequest true "Create Category request"
// @Success 201 {object} models.CategoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories [post]
func (c *CategoryController) Create(ctx *gin.Context) {
    var req models.CreateCategoryRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }

    item, err := c.CategoryService.Create(&req)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create item: " + err.Error()})
        return
    }

    ctx.JSON(http.StatusCreated, item.ToResponse())
}

// GetCategory godoc
// @Summary Get a Category
// @Description Get a Category by its id
// @Tags Category
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "Category id"
// @Success 200 {object} models.CategoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /categories/{id} [get]
func (c *CategoryController) Get(ctx *gin.Context) {
    id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid id format"})
        return
    }

    item, err := c.CategoryService.GetById(uint(id))
    if err != nil {
        ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Item not found"})
        return
    }

    ctx.JSON(http.StatusOK, item.ToResponse())
}

// Listcategories godoc
// @Summary List categories
// @Description Get a list of categories
// @Tags Category
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param search query string false "Search term for filtering results"
// @Success 200 {object} types.PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories [get]
func (c *CategoryController) List(ctx *gin.Context) {
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

    paginatedResponse, err := c.CategoryService.GetAll(&page, &limit, search)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch items: " + err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, paginatedResponse)
}

// ListAllcategories godoc
// @Summary List all categories
// @Description Get a list of all categories without pagination
// @Tags Category
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {object} types.PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/all [get]
func (c *CategoryController) ListAll(ctx *gin.Context) {
    paginatedResponse, err := c.CategoryService.GetAll(nil, nil, nil)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch all items: " + err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, paginatedResponse)
}

// UpdateCategory godoc
// @Summary Update a Category
// @Description Update a Category by its id
// @Tags Category
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "Category id"
// @Param categories body models.UpdateCategoryRequest true "Update Category request"
// @Success 200 {object} models.CategoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/{id} [put]
func (c *CategoryController) Update(ctx *gin.Context) {
    id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid id format"})
        return
    }

    var req models.UpdateCategoryRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }

    item, err := c.CategoryService.Update(uint(id), &req)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update item: " + err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, item.ToResponse())
}

// DeleteCategory godoc
// @Summary Delete a Category
// @Description Delete a Category by its id
// @Tags Category
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "Category id"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/{id} [delete]
func (c *CategoryController) Delete(ctx *gin.Context) {
    id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid id format"})
        return
    }

    if err := c.CategoryService.Delete(uint(id)); err != nil {
        ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete item: " + err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, SuccessResponse{Message: "Item deleted successfully"})
}
// UploadImage godoc
// @Summary Upload Image for a Category
// @Description Upload Image for a Category by its id
// @Tags Category
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Category id"
// @Param file formData file true "Image file to upload"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/{id}/image [put]
func (c *CategoryController) UploadImage(ctx *gin.Context) {
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

    if err := c.CategoryService.UploadImage(uint(id), file); err != nil {
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
