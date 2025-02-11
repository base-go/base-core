package media

import (
	"net/http"
	"strconv"

	"base/core/logger"
	"base/core/storage"

	"github.com/gin-gonic/gin"
)

type MediaController struct {
	Service *MediaService
	Storage *storage.ActiveStorage
	Logger  logger.Logger
}

func NewMediaController(service *MediaService, storage *storage.ActiveStorage, logger logger.Logger) *MediaController {
	return &MediaController{
		Service: service,
		Storage: storage,
		Logger:  logger,
	}
}

func (c *MediaController) Routes(router *gin.RouterGroup) {
	// Main CRUD endpoints
	router.GET("", c.List)        // Paginated list
	router.GET("/all", c.ListAll) // Unpaginated list
	router.GET("/:id", c.Get)
	router.POST("", c.Create)
	router.PUT("/:id", c.Update)
	router.DELETE("/:id", c.Delete)

	// File management endpoints
	router.PUT("/:id/file", c.UpdateFile)
	router.DELETE("/:id/file", c.RemoveFile)
}

// Create godoc
// @Summary Create a new media item
// @Description Create a new media item with optional file upload
// @Tags Core/Media
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Media name"
// @Param type formData string true "Media type"
// @Param description formData string false "Media description"
// @Param file formData file false "Media file"
// @Success 201 {object} MediaResponse
// @Router /media [post]
// @Security ApiKeyAuth
func (c *MediaController) Create(ctx *gin.Context) {
	var req CreateMediaRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Handle file upload
	if file, err := ctx.FormFile("file"); err == nil {
		req.File = file
	}

	item, err := c.Service.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, item.ToResponse())
}

// UpdateFile godoc
// @Summary Update media file
// @Description Update the file attached to a media item
// @Tags Core/Media
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Media ID"
// @Param file formData file true "Media file"
// @Success 200 {object} MediaResponse
// @Router /media/{id}/file [put]
// @Security ApiKeyAuth
func (c *MediaController) UpdateFile(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id parameter"})
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}

	item, err := c.Service.UpdateFile(ctx, uint(id), file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, item.ToResponse())
}

// RemoveFile godoc
// @Summary Remove media file
// @Description Remove the file attached to a media item
// @Tags Core/Media
// @Produce json
// @Param id path int true "Media ID"
// @Success 200 {object} MediaResponse
// @Router /media/{id}/file [delete]
// @Security ApiKeyAuth
func (c *MediaController) RemoveFile(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id parameter"})
		return
	}

	item, err := c.Service.RemoveFile(ctx, uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, item.ToResponse())
}

// Update godoc
// @Summary Update a media item
// @Description Update a media item's details and optionally its file
// @Tags Core/Media
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Media ID"
// @Param name formData string false "Media name"
// @Param type formData string false "Media type"
// @Param description formData string false "Media description"
// @Param file formData file false "Media file"
// @Success 200 {object} MediaResponse
// @Router /media/{id} [put]
// @Security ApiKeyAuth
func (c *MediaController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id parameter"})
		return
	}

	var req UpdateMediaRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Handle file upload
	if file, err := ctx.FormFile("file"); err == nil {
		req.File = file
	}

	item, err := c.Service.Update(uint(id), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, item.ToResponse())
}

// Delete godoc
// @Summary Delete a media item
// @Description Delete a media item and its associated file
// @Tags Core/Media
// @Produce json
// @Param id path int true "Media ID"
// @Success 204 "No Content"
// @Router /media/{id} [delete]
// @Security ApiKeyAuth
func (c *MediaController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id parameter"})
		return
	}

	if err := c.Service.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Get godoc
// @Summary Get a media item
// @Description Get a media item by ID
// @Tags Core/Media
// @Produce json
// @Param id path int true "Media ID"
// @Success 200 {object} MediaResponse
// @Router /media/{id} [get]
// @Security ApiKeyAuth
func (c *MediaController) Get(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id parameter"})
		return
	}

	item, err := c.Service.GetById(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "media not found"})
		return
	}

	ctx.JSON(http.StatusOK, item.ToResponse())
}

// List godoc
// @Summary List media items
// @Description Get a paginated list of media items
// @Tags Core/Media
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} types.PaginatedResponse
// @Router /media [get]
// @Security ApiKeyAuth
func (c *MediaController) List(ctx *gin.Context) {
	page := 1
	limit := 10

	if pageStr := ctx.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	result, err := c.Service.GetAll(&page, &limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// ListAll godoc
// @Summary List all media items
// @Description Get an unpaginated list of all media items
// @Tags Core/Media
// @Produce json
// @Success 200 {array} MediaListResponse
// @Router /media/all [get]
// @Security ApiKeyAuth
func (c *MediaController) ListAll(ctx *gin.Context) {
	result, err := c.Service.GetAll(nil, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

type ErrorResponse struct {
	Error string `json:"error"`
}
