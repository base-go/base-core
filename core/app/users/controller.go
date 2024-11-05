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
	router.GET("/users/:id", c.Get)
	router.PUT("/users/:id", c.Update)

	router.PUT("/users/:id/avatar", c.UpdateAvatar)
	router.PUT("/users/:id/password", c.UpdatePassword)
}

// @Summary Get user by ID
// @Description Get user by ID
// @Tags Core/Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [get]
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

// @Summary Update user
// @Description Update user
// @Tags Core/Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param input body UpdateRequest true "Update Request"
// @Success 200 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [put]
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

// @Summary Update user avatar
// @Description Update user avatar
// @Tags Core/Users
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "User ID"
// @Param avatar formData file true "Avatar file"
// @Success 200 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/avatar [put]
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

	updatedUser, err := c.UserService.UpdateAvatar(ctx, uint(id), file)
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

// @Summary Update user password
// @Description Update user password
// @Tags Core/Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param input body UpdatePasswordRequest true "Update Password Request"
// @Success 200 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/password [put]
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
