package countries

import (
	"net/http"
	"strconv"
	"strings"

	"base/app/models"
	"base/core/router"
	"base/core/storage"
	"base/core/types"
)

type CountryController struct {
	Service *CountryService
	Storage *storage.ActiveStorage
}

func NewCountryController(service *CountryService, storage *storage.ActiveStorage) *CountryController {
	return &CountryController{
		Service: service,
		Storage: storage,
	}
}

func (c *CountryController) Routes(router *router.RouterGroup) {
	// Main CRUD endpoints - specific routes MUST come before parameterized routes
	router.GET("/countries", c.List)          // Paginated list
	router.POST("/countries", c.Create)       // Create
	router.GET("/countries/all", c.ListAll)   // Unpaginated list - MUST be before /:id
	router.GET("/countries/:id", c.Get)       // Get by ID - MUST be after /all
	router.PUT("/countries/:id", c.Update)    // Update
	router.DELETE("/countries/:id", c.Delete) // Delete

	//Upload endpoints for each file field
	router.POST("/countries/:id/flag", c.UploadFlag)
	router.DELETE("/countries/:id/flag", c.RemoveFlag)
}

// CreateCountry godoc
// @Summary Create a new Country
// @Description Create a new Country with the input payload
// @Tags App/Country
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param countries body models.CreateCountryRequest true "Create Country request"
// @Success 201 {object} models.CountryResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /api/countries [post]
func (c *CountryController) Create(ctx *router.Context) error {
	var req models.CreateCountryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
	}

	item, err := c.Service.Create(&req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to create item: " + err.Error()})
	}

	return ctx.JSON(http.StatusCreated, item.ToResponse())
}

// GetCountry godoc
// @Summary Get a Country
// @Description Get a Country by its id
// @Tags App/Country
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Country id"
// @Success 200 {object} models.CountryResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Router /api/countries/{id} [get]
func (c *CountryController) Get(ctx *router.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid id format"})
	}

	item, err := c.Service.GetById(uint(id))
	if err != nil {
		return ctx.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Item not found"})
	}

	return ctx.JSON(http.StatusOK, item.ToResponse())
}

// ListCountries godoc
// @Summary List countries
// @Description Get a list of countries
// @Tags App/Country
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param sort query string false "Sort field (id, created_at, updated_at,name,flag,email,)"
// @Param order query string false "Sort order (asc, desc)"
// @Success 200 {object} types.PaginatedResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /api/countries [get]
func (c *CountryController) List(ctx *router.Context) error {
	var page, limit *int
	var sortBy, sortOrder *string

	// Parse page parameter
	if pageStr := ctx.Query("page"); pageStr != "" {
		if pageNum, err := strconv.Atoi(pageStr); err == nil && pageNum > 0 {
			page = &pageNum
		} else {
			return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid page number"})
		}
	}

	// Parse limit parameter
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if limitNum, err := strconv.Atoi(limitStr); err == nil && limitNum > 0 {
			limit = &limitNum
		} else {
			return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid limit number"})
		}
	}

	// Parse sort parameters
	if sortStr := ctx.Query("sort"); sortStr != "" {
		sortBy = &sortStr
	}

	if orderStr := ctx.Query("order"); orderStr != "" {
		if orderStr == "asc" || orderStr == "desc" {
			sortOrder = &orderStr
		} else {
			return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid sort order. Use 'asc' or 'desc'"})
		}
	}

	paginatedResponse, err := c.Service.GetAll(page, limit, sortBy, sortOrder)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to fetch items: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, paginatedResponse)
}

// ListAllCountries godoc
// @Summary List all countries for select options
// @Description Get a simplified list of all countries with id and name only (for dropdowns/select boxes)
// @Tags App/Country
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {array} models.CountrySelectOption
// @Failure 500 {object} types.ErrorResponse
// @Router /api/countries/all [get]
func (c *CountryController) ListAll(ctx *router.Context) error {
	items, err := c.Service.GetAllForSelect()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to fetch select options: " + err.Error()})
	}

	// Convert to select options
	var selectOptions []*models.CountrySelectOption
	for _, item := range items {
		selectOptions = append(selectOptions, item.ToSelectOption())
	}

	return ctx.JSON(http.StatusOK, selectOptions)
}

// UpdateCountry godoc
// @Summary Update a Country
// @Description Update a Country by its id
// @Tags App/Country
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Country id"
// @Param countries body models.UpdateCountryRequest true "Update Country request"
// @Success 200 {object} models.CountryResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /api/countries/{id} [put]
func (c *CountryController) Update(ctx *router.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid id format"})
	}

	var req models.UpdateCountryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
	}

	item, err := c.Service.Update(uint(id), &req)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return ctx.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Item not found"})
		}
		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to update item: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, item.ToResponse())
}

// DeleteCountry godoc
// @Summary Delete a Country
// @Description Delete a Country by its id
// @Tags App/Country
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Country id"
// @Success 200 {object} types.SuccessResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /api/countries/{id} [delete]
func (c *CountryController) Delete(ctx *router.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid id format"})
	}

	if err := c.Service.Delete(uint(id)); err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return ctx.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Item not found"})
		}
		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to delete item: " + err.Error()})
	}

	ctx.Status(http.StatusNoContent)
	return nil
}

// UploadFlag godoc
// @Summary Upload flag for Country
// @Description Upload a file for the Country's flag field
// @Tags App/Country
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Country id"
// @Param file formData file true "Flag file"
// @Success 200 {object} models.CountryResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /api/countries/{id}/flag [post]
func (c *CountryController) UploadFlag(ctx *router.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid id format"})
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "No file uploaded"})
	}

	item, err := c.Service.UploadFlag(uint(id), file)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to upload flag: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, item.ToResponse())
}

// RemoveFlag godoc
// @Summary Remove flag from Country
// @Description Remove the file from the Country's flag field
// @Tags App/Country
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Country id"
// @Success 200 {object} models.CountryResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /api/countries/{id}/flag [delete]
func (c *CountryController) RemoveFlag(ctx *router.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid id format"})
	}

	item, err := c.Service.RemoveFlag(uint(id))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to remove flag: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, item.ToResponse())
}
