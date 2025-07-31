package category

import (
	"base/app/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	Service *Service
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewController(service *Service) *Controller {
	return &Controller{
		Service: service,
	}
}

func (c *Controller) Routes(router *gin.RouterGroup) {

	router.GET("/categories", c.List)
	router.POST("/categories", c.Create)
	router.GET("/categories/tree", c.GetTree)
	router.GET("/categories/:id", c.Get)
	router.GET("/categories/slug/:slug", c.GetBySlug)
	router.PUT("/categories/:id", c.Update)
	router.DELETE("/categories/:id", c.Delete)
}

// List godoc
// @Summary List categories
// @Description Get a paginated list of categories
// @Tags Categories
// @Security ApiKeyAuth
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Success 200 {object} types.PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories [get]
func (c *Controller) List(ctx *gin.Context) {
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

	result, err := c.Service.GetAll(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch categories: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// GetTree godoc
// @Summary Get category tree
// @Description Get categories as a hierarchical tree structure
// @Tags Categories
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} model.CategoryResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/tree [get]
func (c *Controller) GetTree(ctx *gin.Context) {
	result, err := c.Service.GetTree()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch category tree: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// Get godoc
// @Summary Get category by ID
// @Description Get a single category by its ID
// @Tags Categories
// @Security ApiKeyAuth
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} model.CategoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/{id} [get]
func (c *Controller) Get(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid category ID"})
		return
	}

	category, err := c.Service.GetByID(uint(id))
	if err != nil {
		if err.Error() == "category not found" {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch category: " + err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, category)
}

// GetBySlug godoc
// @Summary Get category by slug
// @Description Get a single category by its slug
// @Tags Categories
// @Security ApiKeyAuth
// @Produce json
// @Param slug path string true "Category slug"
// @Success 200 {object} model.CategoryResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/slug/{slug} [get]
func (c *Controller) GetBySlug(ctx *gin.Context) {
	slug := ctx.Param("slug")

	category, err := c.Service.GetBySlug(slug)
	if err != nil {
		if err.Error() == "category not found" {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch category: " + err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, category)
}

// Create godoc
// @Summary Create category
// @Description Create a new category
// @Tags Categories
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param category body model.CreateCategoryRequest true "Category data"
// @Success 201 {object} model.CategoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories [post]
func (c *Controller) Create(ctx *gin.Context) {
	var request model.CreateCategoryRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request data: " + err.Error()})
		return
	}

	category, err := c.Service.Create(&request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create category: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, category)
}

// Update godoc
// @Summary Update category
// @Description Update an existing category
// @Tags Categories
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body model.UpdateCategoryRequest true "Category data"
// @Success 200 {object} model.CategoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/{id} [put]
func (c *Controller) Update(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid category ID"})
		return
	}

	var request model.UpdateCategoryRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request data: " + err.Error()})
		return
	}

	request.ID = uint(id)
	category, err := c.Service.Update(&request)
	if err != nil {
		if err.Error() == "category not found" {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update category: " + err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, category)
}

// Delete godoc
// @Summary Delete category
// @Description Delete a category by ID
// @Tags Categories
// @Security ApiKeyAuth
// @Param id path int true "Category ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/{id} [delete]
func (c *Controller) Delete(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid category ID"})
		return
	}

	err = c.Service.Delete(uint(id))
	if err != nil {
		if err.Error() == "category not found" {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete category: " + err.Error()})
		}
		return
	}

	ctx.Status(http.StatusNoContent)
}
