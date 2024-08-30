package scenes

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SceneController struct {
	SceneService *SceneService
}

func NewSceneController(service *SceneService) *SceneController {
	return &SceneController{
		SceneService: service,
	}
}

func (c *SceneController) Routes(router *gin.RouterGroup) {
	router.GET("/scenes", c.List)
	router.GET("/scenes/:id", c.Get)
	router.POST("/scenes", c.Create)
	router.PUT("/scenes/:id", c.Update)
	router.DELETE("/scenes/:id", c.Delete)
}

// CreateScene godoc
// @Summary Create a new Scene
// @Description Create a new Scene with the input payload
// @Tags scenes
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param scenes body CreateRequest true "Create Scene"
// @Success 201 {object} CreateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /scenes [post]
func (c *SceneController) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.SceneService.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create item"})
		return
	}

	ctx.JSON(http.StatusCreated, item)
}

// GetScene godoc
// @Summary Get a Scene
// @Description Get a Scene by its ID
// @Tags scenes
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} Scene
// @Failure 404 {object} ErrorResponse
// @Router /scenes/{id} [get]
func (c *SceneController) Get(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	item, err := c.SceneService.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Item not found"})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// ListScene godoc
// @Summary List Scene
// @Description Get a list of all Scene
// @Tags scenes
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {array} Scene
// @Failure 500 {object} ErrorResponse
// @Router /scenes [get]
func (c *SceneController) List(ctx *gin.Context) {
	items, err := c.SceneService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch items"})
		return
	}

	ctx.JSON(http.StatusOK, items)
}

// UpdateScene godoc
// @Summary Update a Scene
// @Description Update a Scene by its ID
// @Tags scenes
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param scenes body UpdateRequest true "Update Scene"
// @Success 200 {object} UpdateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /scenes/{id} [put]
func (c *SceneController) Update(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.SceneService.Update(uint(id), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update item"})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// DeleteScene godoc
// @Summary Delete a Scene
// @Description Delete a Scene by its ID
// @Tags scenes
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} ErrorResponse
// @Router /scenes/{id} [delete]
func (c *SceneController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.SceneService.Delete(uint(id)); err != nil {
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