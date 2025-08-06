package profile

import (
	"base/core/logger"
	"base/core/types"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ProfileController struct {
	service *ProfileService
	logger  logger.Logger
}

func NewProfileController(service *ProfileService, logger logger.Logger) *ProfileController {
	return &ProfileController{
		service: service,
		logger:  logger,
	}
}

func (c *ProfileController) Routes(router *gin.RouterGroup) {
	router.GET("/profile", c.Get)
	router.PUT("/profile", c.Update)
	router.PUT("/profile/avatar", c.UpdateAvatar)
	router.PUT("/profile/password", c.UpdatePassword)
}

// @Summary Get profile from Authenticated User Token
// @Description Get profile by Bearer Token
// @Security ApiKeyAuth
// @Security BearerAuth
// @Tags Core/Profile
// @Accept json
// @Produce json
// @Success 200 {object} User
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /profile [get]
func (c *ProfileController) Get(ctx *gin.Context) {
	id := ctx.GetUint("user_id")
	c.logger.Debug("Getting user", logger.Uint("user_id", id))
	if id == 0 {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid user ID"})
		return
	}

	item, err := c.service.GetByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, types.ErrorResponse{Error: "User not found"})
			return
		}
		c.logger.Error("Failed to get user",
			logger.Uint("user_id", id))
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to fetch user"})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// @Summary Update profile from Authenticated User Token
// @Description Update profile by Bearer Token
// @Security ApiKeyAuth
// @Security BearerAuth
// @Tags Core/Profile
// @Accept json
// @Produce json
// @Param input body UpdateRequest true "Update Request"
// @Success 200 {object} User
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /profile [put]
func (c *ProfileController) Update(ctx *gin.Context) {
	id := ctx.GetUint("user_id")
	if id == 0 {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid ID format"})
		return
	}

	var req UpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid input: " + err.Error()})
		return
	}

	item, err := c.service.Update(uint(id), &req)
	if err != nil {
		c.logger.Error("Failed to update user",
			logger.Uint("user_id", id))

		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to update user: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// @Summary Update profile avatar from Authenticated User Token
// @Description Update profile avatar by Bearer Token
// @Security ApiKeyAuth
// @Security BearerAuth
// @Tags Core/Profile
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Avatar file"
// @Success 200 {object} User
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /profile/avatar [put]
func (c *ProfileController) UpdateAvatar(ctx *gin.Context) {
	id := ctx.GetUint("user_id")
	if id == 0 {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid ID format"})
		return
	}

	file, err := ctx.FormFile("avatar")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Failed to get avatar file: " + err.Error()})
		return
	}

	updatedUser, err := c.service.UpdateAvatar(ctx, uint(id), file)
	if err != nil {
		c.logger.Error("Failed to update avatar",
			logger.Uint("user_id", id))

		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, types.ErrorResponse{Error: "User not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to update avatar: " + err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, updatedUser)
}

// @Summary Update profile password from Authenticated User Token
// @Description Update profile password by Bearer Token
// @Security ApiKeyAuth
// @Security BearerAuth
// @Tags Core/Profile
// @Accept json
// @Produce json
// @Param input body UpdatePasswordRequest true "Update Password Request"
// @Success 200 {object} User
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /profile/password [put]
func (c *ProfileController) UpdatePassword(ctx *gin.Context) {
	id := ctx.GetUint("user_id")
	if id == 0 {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid user ID"})
		return
	}

	var req UpdatePasswordRequest
	if err := ctx.ShouldBind(&req); err != nil {
		c.logger.Error("Failed to bind password update request")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid input: " + err.Error()})
		return
	}

	if len(req.NewPassword) < 6 {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "New password must be at least 6 characters long"})
		return
	}

	err := c.service.UpdatePassword(uint(id), &req)
	if err != nil {
		c.logger.Error("Failed to update password",
			logger.Uint("user_id", id))

		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			ctx.JSON(http.StatusNotFound, types.ErrorResponse{Error: "User not found"})
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "Current password is incorrect"})
		default:
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to update password"})
		}
		return
	}

	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "Password updated successfully"})
}
