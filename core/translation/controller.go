package translation

import (
	"base/core/router"
	"base/core/storage"
	"net/http"
	"strconv"
)

type TranslationController struct {
	Service *TranslationService
	Storage *storage.ActiveStorage
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewTranslationController(service *TranslationService, storage *storage.ActiveStorage) *TranslationController {
	return &TranslationController{
		Service: service,
		Storage: storage,
	}
}

func (c *TranslationController) Routes(router *router.RouterGroup) {
	// CRUD operations
	router.GET("/translations", c.List)
	router.POST("/translations", c.Create)

	// Bulk operations - MUST come before parameterized routes
	router.POST("/translations/bulk", c.BulkUpdate)

	// Utility endpoints - MUST come before parameterized routes
	router.GET("/translations/languages", c.GetSupportedLanguages)

	// Model-specific operations - MUST come before parameterized routes
	router.GET("/translations/models/:model/:model_id", c.GetForModel)
	router.GET("/translations/models/:model/:model_id/:language", c.GetForModelAndLanguage)

	// CRUD operations with :id parameter - MUST come LAST
	router.GET("/translations/by-id/:id", c.Get)
	router.PUT("/translations/by-id/:id", c.Update)
	router.DELETE("/translations/by-id/:id", c.Delete)
}

// List godoc
// @Summary List translations
// @Description Get a paginated list of translations with optional filtering
// @Tags Core/Translations
// @Security ApiKeyAuth
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param model query string false "Filter by model name"
// @Param model_id query int false "Filter by model ID"
// @Success 200 {object} types.PaginatedResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /translations [get]
func (c *TranslationController) List(ctx *router.Context) error {
	var page, limit *int
	var modelId *uint

	if pageStr := ctx.Query("page"); pageStr != "" {
		if pageNum, err := strconv.Atoi(pageStr); err == nil && pageNum > 0 {
			page = &pageNum
		} else {
			return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid page number"})
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if limitNum, err := strconv.Atoi(limitStr); err == nil && limitNum > 0 {
			limit = &limitNum
		} else {
			return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid limit number"})
		}
	}

	// Handle model_id filter
	if modelIdStr := ctx.Query("model_id"); modelIdStr != "" {
		if modelIdNum, err := strconv.ParseUint(modelIdStr, 10, 32); err == nil {
			modelIdUint := uint(modelIdNum)
			modelId = &modelIdUint
		} else {
			return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid model_id"})
		}
	}

	// Get model filter
	model := ctx.Query("model")

	paginatedResponse, err := c.Service.GetAll(page, limit, model, modelId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch translations: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, paginatedResponse)
}

// Get godoc
// @Summary Get translation by ID
// @Description Get a single translation by its ID
// @Tags Core/Translations
// @Security ApiKeyAuth
// @Produce json
// @Param id path int true "Translation ID"
// @Success 200 {object} translation.TranslationResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /translations/by-id/{id} [get]
func (c *TranslationController) Get(ctx *router.Context) error {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid translation ID"})
	}

	translation, err := c.Service.GetByID(uint(id))
	if err != nil {
		if err.Error() == "translation not found" {
			return ctx.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		} else {
			return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch translation: " + err.Error()})
		}
	}

	return ctx.JSON(http.StatusOK, translation)
}

// Create godoc
// @Summary Create translation
// @Description Create a new translation
// @Tags Core/Translations
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param translation body 	translation.CreateTranslationRequest true "Translation data"
// @Success 201 {object} translation.TranslationResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /translations [post]
func (c *TranslationController) Create(ctx *router.Context) error {
	var request CreateTranslationRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request data: " + err.Error()})
	}

	translation, err := c.Service.Create(&request)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create translation: " + err.Error()})
	}

	return ctx.JSON(http.StatusCreated, translation)
}

// Update godoc
// @Summary Update translation
// @Description Update an existing translation
// @Tags Core/Translations
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "Translation ID"
// @Param translation body translation.UpdateTranslationRequest true "Translation data"
// @Success 200 {object} translation.TranslationResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /translations/by-id/{id} [put]
func (c *TranslationController) Update(ctx *router.Context) error {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid translation ID"})
	}

	var request UpdateTranslationRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request data: " + err.Error()})
	}

	request.Id = uint(id)
	translation, err := c.Service.Update(&request)
	if err != nil {
		if err.Error() == "translation not found" {
			return ctx.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		} else {
			return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update translation: " + err.Error()})
		}
	}

	return ctx.JSON(http.StatusOK, translation)
}

// Delete godoc
// @Summary Delete translation
// @Description Delete a translation by ID
// @Tags Core/Translations
// @Security ApiKeyAuth
// @Param id path int true "Translation ID"
// @Success 204
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /translations/by-id/{id} [delete]
func (c *TranslationController) Delete(ctx *router.Context) error {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid translation ID"})
	}

	err = c.Service.Delete(uint(id))
	if err != nil {
		if err.Error() == "translation not found" {
			return ctx.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		} else {
			return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete translation: " + err.Error()})
		}
	}

	ctx.Status(http.StatusNoContent)
	return nil
}

// BulkUpdate godoc
// @Summary Bulk update translations
// @Description Update multiple translations for a model at once
// @Tags Core/Translations
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param bulk body translation.BulkTranslationRequest true "Bulk translation data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /translations/bulk [post]
func (c *TranslationController) BulkUpdate(ctx *router.Context) error {
	var request BulkTranslationRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request data: " + err.Error()})
	}

	err := c.Service.BulkUpdate(&request)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update translations: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, map[string]any{"message": "Translations updated successfully"})
}

// GetForModel godoc
// @Summary Get translations for model
// @Description Get all translations for a specific model and model ID
// @Tags Core/Translations
// @Security ApiKeyAuth
// @Produce json
// @Param model path string true "Model name"
// @Param model_id path int true "Model ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /translations/models/{model}/{model_id} [get]
func (c *TranslationController) GetForModel(ctx *router.Context) error {
	model := ctx.Param("model")
	modelIdStr := ctx.Param("model_id")

	modelId, err := strconv.ParseUint(modelIdStr, 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid model ID"})
	}

	translations, err := c.Service.GetTranslationsForModel(model, uint(modelId), "")
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch translations: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, translations)
}

// GetForModelAndLanguage godoc
// @Summary Get translations for model and language
// @Description Get translations for a specific model, model ID, and language
// @Tags Core/Translations
// @Security ApiKeyAuth
// @Produce json
// @Param model path string true "Model name"
// @Param model_id path int true "Model ID"
// @Param language path string true "Language code"
// @Success 200 {object} translation.TranslationResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /translations/models/{model}/{model_id}/{language} [get]
func (c *TranslationController) GetForModelAndLanguage(ctx *router.Context) error {
	model := ctx.Param("model")
	modelIdStr := ctx.Param("model_id")
	language := ctx.Param("language")

	modelId, err := strconv.ParseUint(modelIdStr, 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid model ID"})
	}

	translations, err := c.Service.GetTranslationsForModel(model, uint(modelId), language)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch translations: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, translations)
}

// GetSupportedLanguages godoc
// @Summary Get supported languages
// @Description Get a list of all languages that have translations in the system
// @Tags Core/Translations
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} types.ErrorResponse
// @Router /translations/languages [get]
func (c *TranslationController) GetSupportedLanguages(ctx *router.Context) error {
	languages, err := c.Service.GetSupportedLanguages()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch supported languages: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, languages)
}
