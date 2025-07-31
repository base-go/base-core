package base

import (
	"base/core/logger"
	"base/core/storage"
	"base/core/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Controller provides common functionality for all controllers
type Controller struct {
	Logger  logger.Logger
	Storage *storage.ActiveStorage
}

// NewController creates a new Controller instance
func NewController(logger logger.Logger, storage *storage.ActiveStorage) *Controller {
	return &Controller{
		Logger:  logger,
		Storage: storage,
	}
}

// RespondSuccess sends a successful JSON response
func (bc *Controller) RespondSuccess(c *gin.Context, data interface{}, message ...string) {
	response := types.SuccessResponse{
		Success: true,
		Data:    data,
	}
	
	if len(message) > 0 {
		response.Message = message[0]
	}
	
	c.JSON(http.StatusOK, response)
}

// RespondCreated sends a created JSON response
func (bc *Controller) RespondCreated(c *gin.Context, data interface{}, message ...string) {
	response := types.SuccessResponse{
		Success: true,
		Data:    data,
	}
	
	if len(message) > 0 {
		response.Message = message[0]
	} else {
		response.Message = "Resource created successfully"
	}
	
	c.JSON(http.StatusCreated, response)
}

// RespondError sends an error JSON response
func (bc *Controller) RespondError(c *gin.Context, statusCode int, message string, details ...interface{}) {
	bc.Logger.Error("Controller error", 
		logger.String("message", message),
		logger.Int("status_code", statusCode),
		logger.String("path", c.Request.URL.Path),
		logger.String("method", c.Request.Method),
	)
	
	response := types.ErrorResponse{
		Error:   message,
		Success: false,
	}
	
	if len(details) > 0 {
		response.Details = details[0]
	}
	
	c.JSON(statusCode, response)
}

// RespondValidationError sends a validation error response
func (bc *Controller) RespondValidationError(c *gin.Context, err error) {
	bc.RespondError(c, http.StatusBadRequest, "Validation failed", err.Error())
}

// RespondNotFound sends a not found error response
func (bc *Controller) RespondNotFound(c *gin.Context, resource string) {
	bc.RespondError(c, http.StatusNotFound, resource+" not found")
}

// RespondInternalError sends an internal server error response
func (bc *Controller) RespondInternalError(c *gin.Context, message string) {
	bc.RespondError(c, http.StatusInternalServerError, message)
}

// RespondUnauthorized sends an unauthorized error response
func (bc *Controller) RespondUnauthorized(c *gin.Context, message ...string) {
	msg := "Unauthorized"
	if len(message) > 0 {
		msg = message[0]
	}
	bc.RespondError(c, http.StatusUnauthorized, msg)
}

// RespondForbidden sends a forbidden error response
func (bc *Controller) RespondForbidden(c *gin.Context, message ...string) {
	msg := "Forbidden"
	if len(message) > 0 {
		msg = message[0]
	}
	bc.RespondError(c, http.StatusForbidden, msg)
}

// GetPaginationParams extracts page and limit from query parameters
func (bc *Controller) GetPaginationParams(c *gin.Context) (page int, limit int) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	
	limit, err = strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}
	
	return page, limit
}

// GetIDParam extracts ID parameter from URL path
func (bc *Controller) GetIDParam(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}