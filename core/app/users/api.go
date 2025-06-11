package users

import (
	"base/core/logger"
	"base/core/router"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserAPIController struct {
	service *UserService
	logger  logger.Logger
}

func NewUserAPIController(service *UserService, logger logger.Logger) *UserAPIController {
	return &UserAPIController{
		service: service,
		logger:  logger,
	}
}

func (c *UserAPIController) Routes(ginRouter *gin.RouterGroup) {
	// Convert to Rails-style router
	r := router.NewFromGroup(ginRouter)

	// API routes for JSON responses using Rails-style syntax
	r.Namespace("/api/users", func(users *router.Router) {
		users.Get("/me", c.Get).
			Put("/me", c.Update).
			Put("/me/avatar", c.UpdateAvatar).
			Put("/me/password", c.UpdatePassword)
	})
}

// @Summary Get user from Authenticated User Token
// @Description Get user by Bearer Token
// @Security ApiKeyAuth
// @Security BearerAuth
// @Tags Core/Users
// @Accept json
// @Produce json
// @Success 200 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/me [get]
func (c *UserAPIController) Get(ctx *gin.Context) {
	id := ctx.GetUint("user_id")
	c.logger.Debug("Getting user", logger.Uint("user_id", id))
	if id == 0 {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	item, err := c.service.GetByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
			return
		}
		c.logger.Error("Failed to get user",
			logger.Uint("user_id", id))
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch user"})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// @Summary Update user from Authenticated User Token
// @Description Update user by Bearer Token
// @Security ApiKeyAuth
// @Security BearerAuth
// @Tags Core/Users
// @Accept json
// @Produce json
// @Param input body UpdateRequest true "Update Request"
// @Success 200 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/me [put]
func (c *UserAPIController) Update(ctx *gin.Context) {
	id := ctx.GetUint("user_id")
	if id == 0 {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ID format"})
		return
	}

	var req UpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input: " + err.Error()})
		return
	}

	item, err := c.service.Update(uint(id), &req)
	if err != nil {
		c.logger.Error("Failed to update user",
			logger.Uint("user_id", id))

		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update user: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// @Summary Update user avatar from Authenticated User Token
// @Description Update user avatar by Bearer Token
// @Security ApiKeyAuth
// @Security BearerAuth
// @Tags Core/Users
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Avatar file"
// @Success 200 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/me/avatar [put]
func (c *UserAPIController) UpdateAvatar(ctx *gin.Context) {
	id := ctx.GetUint("user_id")
	if id == 0 {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ID format"})
		return
	}

	file, err := ctx.FormFile("avatar")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to get avatar file: " + err.Error()})
		return
	}

	updatedUser, err := c.service.UpdateAvatar(ctx, uint(id), file)
	if err != nil {
		c.logger.Error("Failed to update avatar",
			logger.Uint("user_id", id))

		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update avatar: " + err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, updatedUser)
}

// @Summary Update user password from Authenticated User Token
// @Description Update user password by Bearer Token
// @Security ApiKeyAuth
// @Security BearerAuth
// @Tags Core/Users
// @Accept json
// @Produce json
// @Param input body UpdatePasswordRequest true "Update Password Request"
// @Success 200 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/me/password [put]
func (c *UserAPIController) UpdatePassword(ctx *gin.Context) {
	id := ctx.GetUint("user_id")
	if id == 0 {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	var req UpdatePasswordRequest
	if err := ctx.ShouldBind(&req); err != nil {
		c.logger.Error("Failed to bind password update request")
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input: " + err.Error()})
		return
	}

	if len(req.NewPassword) < 6 {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "New password must be at least 6 characters long"})
		return
	}

	err := c.service.UpdatePassword(uint(id), &req)
	if err != nil {
		c.logger.Error("Failed to update password",
			logger.Uint("user_id", id))

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

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
