package users

import (
	"base/core/event"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserController struct {
	UserService  *UserService
	Logger       *zap.Logger
	EventService *event.EventService
}

func NewUserController(service *UserService, logger *zap.Logger, eventService *event.EventService) *UserController {
	return &UserController{
		UserService:  service,
		Logger:       logger,
		EventService: eventService,
	}
}

func (c *UserController) Routes(router *gin.RouterGroup) {
	router.GET("/users", c.List)
	router.GET("/users/:id", c.Get)
	router.POST("/users", c.Create)
	router.PUT("/users/:id", c.Update)
	router.DELETE("/users/:id", c.Delete)
	router.PUT("/users/:id/avatar", c.UpdateAvatar)
	router.PUT("/users/:id/password", c.UpdatePassword)
}

// Create handles new user registration
func (c *UserController) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.Logger.Error("Failed to bind create user request", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.UserService.Create(&req)
	if err != nil {
		c.Logger.Error("Failed to create user",
			zap.Error(err),
			zap.String("email", req.Email))

		// Track only failed creation attempts
		c.EventService.Track(ctx.Request.Context(), event.EventOptions{
			Type:     "user_creation",
			Category: "users",
			Action:   "create",
			Status:   "failed",
			Metadata: map[string]interface{}{
				"error": err.Error(),
				"email": req.Email,
			},
		})

		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create user"})
		return
	}

	ctx.JSON(http.StatusCreated, item)
}

// Get retrieves a user by ID
func (c *UserController) Get(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	item, err := c.UserService.GetByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
			return
		}
		c.Logger.Error("Failed to get user", zap.Error(err), zap.Int("user_id", id))
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch user"})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// List retrieves all users
func (c *UserController) List(ctx *gin.Context) {
	items, err := c.UserService.GetAll()
	if err != nil {
		c.Logger.Error("Failed to fetch users", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch users"})
		return
	}

	ctx.JSON(http.StatusOK, items)
}

// Update modifies an existing user
func (c *UserController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ID format"})
		return
	}

	var req UpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input: " + err.Error()})
		return
	}

	item, err := c.UserService.Update(uint(id), &req)
	if err != nil {
		c.Logger.Error("Failed to update user",
			zap.Error(err),
			zap.Uint64("user_id", id))

		// Track failed updates
		c.EventService.Track(ctx.Request.Context(), event.EventOptions{
			Type:     "user_update",
			Category: "users",
			Action:   "update",
			Status:   "failed",
			Metadata: map[string]interface{}{
				"error":   err.Error(),
				"user_id": id,
			},
		})

		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update user: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// Delete removes a user
func (c *UserController) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ID format"})
		return
	}

	if err := c.UserService.Delete(uint(id)); err != nil {
		c.Logger.Error("Failed to delete user",
			zap.Error(err),
			zap.Int("user_id", id))

		// Track failed deletions
		c.EventService.Track(ctx.Request.Context(), event.EventOptions{
			Type:     "user_deletion",
			Category: "users",
			Action:   "delete",
			Status:   "failed",
			Metadata: map[string]interface{}{
				"error":   err.Error(),
				"user_id": id,
			},
		})

		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete user"})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{Message: "User deleted successfully"})
}

// UpdateAvatar handles user avatar updates
func (c *UserController) UpdateAvatar(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ID format"})
		return
	}

	file, err := ctx.FormFile("avatar")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to get avatar file: " + err.Error()})
		return
	}

	updatedUser, err := c.UserService.UpdateAvatar(uint(id), file)
	if err != nil {
		c.Logger.Error("Failed to update avatar",
			zap.Error(err),
			zap.Uint64("user_id", id))

		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update avatar: " + err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, updatedUser)
}

// UpdatePassword handles password changes
func (c *UserController) UpdatePassword(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	var req UpdatePasswordRequest
	if err := ctx.ShouldBind(&req); err != nil {
		c.Logger.Error("Failed to bind password update request", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input: " + err.Error()})
		return
	}

	if len(req.NewPassword) < 6 {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "New password must be at least 6 characters long"})
		return
	}

	err = c.UserService.UpdatePassword(uint(id), &req)
	if err != nil {
		c.Logger.Error("Failed to update password",
			zap.Error(err),
			zap.Uint64("user_id", id))

		// Track failed password updates
		c.EventService.Track(ctx.Request.Context(), event.EventOptions{
			Type:     "password_update",
			Category: "security",
			Action:   "update_password",
			Status:   "failed",
			Metadata: map[string]interface{}{
				"user_id": id,
				"reason":  getPasswordUpdateErrorReason(err),
			},
		})

		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Current password is incorrect"})
		default:
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update password"})
		}
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{Message: "Password updated successfully"})
}

// Helper function to categorize password update errors
func getPasswordUpdateErrorReason(err error) string {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return "user_not_found"
	case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
		return "invalid_current_password"
	default:
		return "system_error"
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
