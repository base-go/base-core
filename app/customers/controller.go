package customers

import (
	"net/http"
	"strconv"
	"strings"

	"base/app/models"
	"base/core/router"
	"base/core/storage"
	"base/core/types"
)

type CustomerController struct {
	Service *CustomerService
	Storage *storage.ActiveStorage
}

func NewCustomerController(service *CustomerService, storage *storage.ActiveStorage) *CustomerController {
	return &CustomerController{
		Service: service,
		Storage: storage,
	}
}

func (c *CustomerController) Routes(router *router.RouterGroup) {
	// Main CRUD endpoints - specific routes MUST come before parameterized routes
	router.GET("/customers", c.List)          // Paginated list
	router.POST("/customers", c.Create)       // Create
	router.GET("/customers/all", c.ListAll)   // Unpaginated list - MUST be before /:id
	router.GET("/customers/:id", c.Get)       // Get by ID - MUST be after /all
	router.PUT("/customers/:id", c.Update)    // Update
	router.DELETE("/customers/:id", c.Delete) // Delete

	//Upload endpoints for each file field
	router.POST("/customers/:id/flag", c.UploadFlag)
	router.DELETE("/customers/:id/flag", c.RemoveFlag)
}

// CreateCustomer godoc
// @Summary Create a new Customer
// @Description Create a new Customer with the input payload
// @Tags App/Customer
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param customers body models.CreateCustomerRequest true "Create Customer request"
// @Success 201 {object} models.CustomerResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /api/customers [post]
func (c *CustomerController) Create(ctx *router.Context) error {
	var req models.CreateCustomerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
	}

	item, err := c.Service.Create(&req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to create item: " + err.Error()})
	}

	return ctx.JSON(http.StatusCreated, item.ToResponse())
}

// GetCustomer godoc
// @Summary Get a Customer
// @Description Get a Customer by its id
// @Tags App/Customer
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Customer id"
// @Success 200 {object} models.CustomerResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Router /api/customers/{id} [get]
func (c *CustomerController) Get(ctx *router.Context) error {
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

// ListCustomers godoc
// @Summary List customers
// @Description Get a list of customers
// @Tags App/Customer
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
// @Router /api/customers [get]
func (c *CustomerController) List(ctx *router.Context) error {
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

// ListAllCustomers godoc
// @Summary List all customers for select options
// @Description Get a simplified list of all customers with id and name only (for dropdowns/select boxes)
// @Tags App/Customer
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {array} models.CustomerSelectOption
// @Failure 500 {object} types.ErrorResponse
// @Router /api/customers/all [get]
func (c *CustomerController) ListAll(ctx *router.Context) error {
	items, err := c.Service.GetAllForSelect()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to fetch select options: " + err.Error()})
	}

	// Convert to select options
	var selectOptions []*models.CustomerSelectOption
	for _, item := range items {
		selectOptions = append(selectOptions, item.ToSelectOption())
	}

	return ctx.JSON(http.StatusOK, selectOptions)
}

// UpdateCustomer godoc
// @Summary Update a Customer
// @Description Update a Customer by its id
// @Tags App/Customer
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Customer id"
// @Param customers body models.UpdateCustomerRequest true "Update Customer request"
// @Success 200 {object} models.CustomerResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /api/customers/{id} [put]
func (c *CustomerController) Update(ctx *router.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid id format"})
	}

	var req models.UpdateCustomerRequest
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

// DeleteCustomer godoc
// @Summary Delete a Customer
// @Description Delete a Customer by its id
// @Tags App/Customer
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Customer id"
// @Success 200 {object} types.SuccessResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /api/customers/{id} [delete]
func (c *CustomerController) Delete(ctx *router.Context) error {
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
// @Summary Upload flag for Customer
// @Description Upload a file for the Customer's flag field
// @Tags App/Customer
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Customer id"
// @Param file formData file true "Flag file"
// @Success 200 {object} models.CustomerResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /api/customers/{id}/flag [post]
func (c *CustomerController) UploadFlag(ctx *router.Context) error {
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
// @Summary Remove flag from Customer
// @Description Remove the file from the Customer's flag field
// @Tags App/Customer
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Customer id"
// @Success 200 {object} models.CustomerResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /api/customers/{id}/flag [delete]
func (c *CustomerController) RemoveFlag(ctx *router.Context) error {
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
