package chapters

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChapterController struct {
	ChapterService *ChapterService
}

func NewChapterController(service *ChapterService) *ChapterController {
	return &ChapterController{
		ChapterService: service,
	}
}

func (c *ChapterController) Routes(router *gin.RouterGroup) {
	router.GET("/chapters", c.List)
	router.GET("/chapters/:id", c.Get)
	router.POST("/chapters", c.Create)
	router.PUT("/chapters/:id", c.Update)
	router.DELETE("/chapters/:id", c.Delete)
}

// CreateChapter godoc
// @Summary Create a new Chapter
// @Description Create a new Chapter with the input payload
// @Tags chapters
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param chapters body CreateRequest true "Create Chapter"
// @Success 201 {object} CreateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chapters [post]
func (c *ChapterController) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.ChapterService.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create item"})
		return
	}

	ctx.JSON(http.StatusCreated, item)
}

// GetChapter godoc
// @Summary Get a Chapter
// @Description Get a Chapter by its ID
// @Tags chapters
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} Chapter
// @Failure 404 {object} ErrorResponse
// @Router /chapters/{id} [get]
func (c *ChapterController) Get(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	item, err := c.ChapterService.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Item not found"})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// ListChapter godoc
// @Summary List Chapter
// @Description Get a list of all Chapter
// @Tags chapters
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {array} Chapter
// @Failure 500 {object} ErrorResponse
// @Router /chapters [get]
func (c *ChapterController) List(ctx *gin.Context) {
	items, err := c.ChapterService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch items"})
		return
	}

	ctx.JSON(http.StatusOK, items)
}

// UpdateChapter godoc
// @Summary Update a Chapter
// @Description Update a Chapter by its ID
// @Tags chapters
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param chapters body UpdateRequest true "Update Chapter"
// @Success 200 {object} UpdateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chapters/{id} [put]
func (c *ChapterController) Update(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.ChapterService.Update(uint(id), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update item"})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// DeleteChapter godoc
// @Summary Delete a Chapter
// @Description Delete a Chapter by its ID
// @Tags chapters
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} ErrorResponse
// @Router /chapters/{id} [delete]
func (c *ChapterController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.ChapterService.Delete(uint(id)); err != nil {
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