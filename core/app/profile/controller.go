package profile

import (
	"base/core/logger"
	"base/core/router"
	"base/core/types"
	"errors"
	"net/http"

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

func (c *ProfileController) Routes(router *router.RouterGroup) {
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
func (c *ProfileController) Get(ctx *router.Context) error {
	id := ctx.GetUint("user_id")
	c.logger.Debug("Getting user", logger.Uint("user_id", id))
	if id == 0 {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid user Id"})
	}

	item, err := c.service.GetById(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusNotFound, types.ErrorResponse{Error: "User not found"})
		}
		c.logger.Error("Failed to get user",
			logger.Uint("user_id", id))
		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to fetch user"})
	}

	return ctx.JSON(http.StatusOK, item)
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
func (c *ProfileController) Update(ctx *router.Context) error {
	id := ctx.GetUint("user_id")
	if id == 0 {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid Id format"})
	}

	var req UpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid input: " + err.Error()})
	}

	item, err := c.service.Update(uint(id), &req)
	if err != nil {
		c.logger.Error("Failed to update user",
			logger.Uint("user_id", id))

		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to update user: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, item)
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
func (c *ProfileController) UpdateAvatar(ctx *router.Context) error {
	id := ctx.GetUint("user_id")
	if id == 0 {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid Id format"})
	}

	file, err := ctx.FormFile("avatar")
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Failed to get avatar file: " + err.Error()})
	}

	updatedUser, err := c.service.UpdateAvatar(ctx, uint(id), file)
	if err != nil {
		c.logger.Error("Failed to update avatar",
			logger.Uint("user_id", id))

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusNotFound, types.ErrorResponse{Error: "User not found"})
		} else {
			return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to update avatar: " + err.Error()})
		}
	}

	return ctx.JSON(http.StatusOK, updatedUser)
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
func (c *ProfileController) UpdatePassword(ctx *router.Context) error {
	id := ctx.GetUint("user_id")
	if id == 0 {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid user Id"})
	}

	var req UpdatePasswordRequest
	if err := ctx.ShouldBind(&req); err != nil {
		c.logger.Error("Failed to bind password update request")
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid input: " + err.Error()})
	}

	if len(req.NewPassword) < 6 {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "New password must be at least 6 characters long"})
	}

	err := c.service.UpdatePassword(uint(id), &req)
	if err != nil {
		c.logger.Error("Failed to update password",
			logger.Uint("user_id", id))

		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ctx.JSON(http.StatusNotFound, types.ErrorResponse{Error: "User not found"})
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "Current password is incorrect"})
		default:
			return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to update password"})
		}
	}

	return ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "Password updated successfully"})
}
