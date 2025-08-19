package disbursements

import (
	"net/http"
	"strconv"
	"strings"

	"base/app/models"
	"base/core/router"
	"base/core/storage"
	"base/core/types"
)

type DisbursementController struct {
	Service *DisbursementService
	Storage *storage.ActiveStorage
}

func NewDisbursementController(service *DisbursementService, storage *storage.ActiveStorage) *DisbursementController {
	return &DisbursementController{
		Service: service,
		Storage: storage,
	}
}

func (c *DisbursementController) Routes(router *router.RouterGroup) {
	// Main CRUD endpoints - specific routes MUST come before parameterized routes
	router.GET("/disbursements", c.List)          // Paginated list
	router.POST("/disbursements", c.Create)       // Create
	router.GET("/disbursements/all", c.ListAll)   // Unpaginated list - MUST be before /:id
	router.GET("/disbursements/:id", c.Get)       // Get by ID - MUST be after /all
	router.PUT("/disbursements/:id", c.Update)    // Update
	router.DELETE("/disbursements/:id", c.Delete) // Delete

	//Upload endpoints for each file field
}

// CreateDisbursement godoc
// @Summary Create a new Disbursement
// @Description Create a new Disbursement with the input payload
// @Tags App/Disbursement
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param disbursements body models.CreateDisbursementRequest true "Create Disbursement request"
// @Success 201 {object} models.DisbursementResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /api/disbursements [post]
func (c *DisbursementController) Create(ctx *router.Context) error {
	var req models.CreateDisbursementRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
	}

	item, err := c.Service.Create(&req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to create item: " + err.Error()})
	}

	return ctx.JSON(http.StatusCreated, item.ToResponse())
}

// GetDisbursement godoc
// @Summary Get a Disbursement
// @Description Get a Disbursement by its id
// @Tags App/Disbursement
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Disbursement id"
// @Success 200 {object} models.DisbursementResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Router /api/disbursements/{id} [get]
func (c *DisbursementController) Get(ctx *router.Context) error {
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

// ListDisbursements godoc
// @Summary List disbursements
// @Description Get a list of disbursements
// @Tags App/Disbursement
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param sort query string false "Sort field (id, created_at, updated_at,amount,description,)"
// @Param order query string false "Sort order (asc, desc)"
// @Success 200 {object} types.PaginatedResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /api/disbursements [get]
func (c *DisbursementController) List(ctx *router.Context) error {
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

// ListAllDisbursements godoc
// @Summary List all disbursements for select options
// @Description Get a simplified list of all disbursements with id and name only (for dropdowns/select boxes)
// @Tags App/Disbursement
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {array} models.DisbursementSelectOption
// @Failure 500 {object} types.ErrorResponse
// @Router /api/disbursements/all [get]
func (c *DisbursementController) ListAll(ctx *router.Context) error {
	items, err := c.Service.GetAllForSelect()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to fetch select options: " + err.Error()})
	}

	// Convert to select options
	var selectOptions []*models.DisbursementSelectOption
	for _, item := range items {
		selectOptions = append(selectOptions, item.ToSelectOption())
	}

	return ctx.JSON(http.StatusOK, selectOptions)
}

// UpdateDisbursement godoc
// @Summary Update a Disbursement
// @Description Update a Disbursement by its id
// @Tags App/Disbursement
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Disbursement id"
// @Param disbursements body models.UpdateDisbursementRequest true "Update Disbursement request"
// @Success 200 {object} models.DisbursementResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /api/disbursements/{id} [put]
func (c *DisbursementController) Update(ctx *router.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid id format"})
	}

	var req models.UpdateDisbursementRequest
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

// DeleteDisbursement godoc
// @Summary Delete a Disbursement
// @Description Delete a Disbursement by its id
// @Tags App/Disbursement
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Disbursement id"
// @Success 200 {object} types.SuccessResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /api/disbursements/{id} [delete]
func (c *DisbursementController) Delete(ctx *router.Context) error {
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
