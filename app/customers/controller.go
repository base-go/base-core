package customers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CustomerController struct {
	CustomerService *CustomerService
}

func NewCustomerController(service *CustomerService) *CustomerController {
	return &CustomerController{
		CustomerService: service,
	}
}

func (c *CustomerController) Routes(router *gin.RouterGroup) {
	router.GET("/customers", c.List)
	router.GET("/customers/:id", c.Get)
	router.POST("/customers", c.Create)
	router.PUT("/customers/:id", c.Update)
	router.DELETE("/customers/:id", c.Delete)
}

// CreateCustomer godoc
// @Summary Create a new Customer
// @Description Create a new Customer with the input payload
// @Tags customers
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param customers body CreateRequest true "Create Customer"
// @Success 201 {object} CreateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers [post]
func (c *CustomerController) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.CustomerService.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create item"})
		return
	}

	ctx.JSON(http.StatusCreated, item)
}

// GetCustomer godoc
// @Summary Get a Customer
// @Description Get a Customer by its ID
// @Tags customers
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} Customer
// @Failure 404 {object} ErrorResponse
// @Router /customers/{id} [get]
func (c *CustomerController) Get(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	item, err := c.CustomerService.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Item not found"})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// ListCustomer godoc
// @Summary List Customer
// @Description Get a list of all Customer
// @Tags customers
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {array} Customer
// @Failure 500 {object} ErrorResponse
// @Router /customers [get]
func (c *CustomerController) List(ctx *gin.Context) {
	items, err := c.CustomerService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch items"})
		return
	}

	ctx.JSON(http.StatusOK, items)
}

// UpdateCustomer godoc
// @Summary Update a Customer
// @Description Update a Customer by its ID
// @Tags customers
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param customers body UpdateRequest true "Update Customer"
// @Success 200 {object} UpdateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/{id} [put]
func (c *CustomerController) Update(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.CustomerService.Update(uint(id), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update item"})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// DeleteCustomer godoc
// @Summary Delete a Customer
// @Description Delete a Customer by its ID
// @Tags customers
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/{id} [delete]
func (c *CustomerController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.CustomerService.Delete(uint(id)); err != nil {
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
