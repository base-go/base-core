package stories

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type StoryController struct {
	StoryService *StoryService
}

func NewStoryController(service *StoryService) *StoryController {
	return &StoryController{
		StoryService: service,
	}
}

func (c *StoryController) Routes(router *gin.RouterGroup) {
	router.GET("/stories", c.List)
	router.GET("/stories/:id", c.Get)
	router.POST("/stories", c.Create)
	router.PUT("/stories/:id", c.Update)
	router.DELETE("/stories/:id", c.Delete)
}

// CreateStory godoc
// @Summary Create a new Story
// @Description Create a new Story with the input payload
// @Tags stories
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param stories body CreateRequest true "Create Story"
// @Success 201 {object} CreateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /stories [post]
func (c *StoryController) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.StoryService.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create item"})
		return
	}

	ctx.JSON(http.StatusCreated, item)
}

// GetStory godoc
// @Summary Get a Story
// @Description Get a Story by its ID
// @Tags stories
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} Story
// @Failure 404 {object} ErrorResponse
// @Router /stories/{id} [get]
func (c *StoryController) Get(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	item, err := c.StoryService.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Item not found"})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// ListStory godoc
// @Summary List Story
// @Description Get a list of all Story
// @Tags stories
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {array} Story
// @Failure 500 {object} ErrorResponse
// @Router /stories [get]
func (c *StoryController) List(ctx *gin.Context) {
	items, err := c.StoryService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch items"})
		return
	}

	ctx.JSON(http.StatusOK, items)
}

// UpdateStory godoc
// @Summary Update a Story
// @Description Update a Story by its ID
// @Tags stories
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param stories body UpdateRequest true "Update Story"
// @Success 200 {object} UpdateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /stories/{id} [put]
func (c *StoryController) Update(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.StoryService.Update(uint(id), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update item"})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// DeleteStory godoc
// @Summary Delete a Story
// @Description Delete a Story by its ID
// @Tags stories
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} ErrorResponse
// @Router /stories/{id} [delete]
func (c *StoryController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.StoryService.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete item"})
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