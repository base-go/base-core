package characters

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CharacterController struct {
	CharacterService *CharacterService
}

func NewCharacterController(service *CharacterService) *CharacterController {
	return &CharacterController{
		CharacterService: service,
	}
}

func (c *CharacterController) Routes(router *gin.RouterGroup) {
	router.GET("/characters", c.List)
	router.GET("/characters/:id", c.Get)
	router.POST("/characters", c.Create)
	router.PUT("/characters/:id", c.Update)
	router.DELETE("/characters/:id", c.Delete)
}

// CreateCharacter godoc
// @Summary Create a new Character
// @Description Create a new Character with the input payload
// @Tags characters
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param characters body CreateRequest true "Create Character"
// @Success 201 {object} CreateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /characters [post]
func (c *CharacterController) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.CharacterService.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create item"})
		return
	}

	ctx.JSON(http.StatusCreated, item)
}

// GetCharacter godoc
// @Summary Get a Character
// @Description Get a Character by its ID
// @Tags characters
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} Character
// @Failure 404 {object} ErrorResponse
// @Router /characters/{id} [get]
func (c *CharacterController) Get(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	item, err := c.CharacterService.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Item not found"})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// ListCharacter godoc
// @Summary List Character
// @Description Get a list of all Character
// @Tags characters
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {array} Character
// @Failure 500 {object} ErrorResponse
// @Router /characters [get]
func (c *CharacterController) List(ctx *gin.Context) {
	items, err := c.CharacterService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch items"})
		return
	}

	ctx.JSON(http.StatusOK, items)
}

// UpdateCharacter godoc
// @Summary Update a Character
// @Description Update a Character by its ID
// @Tags characters
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param characters body UpdateRequest true "Update Character"
// @Success 200 {object} UpdateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /characters/{id} [put]
func (c *CharacterController) Update(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.CharacterService.Update(uint(id), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update item"})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// DeleteCharacter godoc
// @Summary Delete a Character
// @Description Delete a Character by its ID
// @Tags characters
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} ErrorResponse
// @Router /characters/{id} [delete]
func (c *CharacterController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.CharacterService.Delete(uint(id)); err != nil {
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