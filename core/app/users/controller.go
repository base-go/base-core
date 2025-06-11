package users

import (
	"base/core/layout"
	"base/core/logger"
	"base/core/middleware"
	"base/core/router"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserController struct {
	*layout.Controller
	service *UserService
	logger  logger.Logger
}

func NewUserController(service *UserService, logger logger.Logger, layoutEngine *layout.Engine) *UserController {
	return &UserController{
		Controller: layout.NewAppController(layoutEngine),
		service:    service,
		logger:     logger,
	}
}

func (c *UserController) Routes(ginRouter *gin.RouterGroup) {
	// Rails-style router for consistent routing pattern
	r := router.NewFromGroup(ginRouter)

	// All user-related routes require authentication
	r.Namespace("/users", func(users *router.Router) {
		users.Use(middleware.AuthMiddleware()).
			Get("/me", c.Get).
			Post("/me/edit", c.UpdateProfile).
			Post("/me/password", c.ChangePassword).
			Post("/me/avatar", c.UploadAvatar).
			Put("/me", c.Update).
			Put("/me/avatar", c.UpdateAvatar).
			Put("/me/password", c.UpdatePassword)
	})
}

func (c *UserController) Get(ctx *gin.Context) {
	id := ctx.GetUint("user_id")
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
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch user"})
		return
	}

	c.View("user/me.html").
		WithTitle("My Profile").
		WithData(gin.H{"user": item}).
		Render(ctx)
}

// UpdateProfile handles the form submission for editing user details.
func (c *UserController) UpdateProfile(ctx *gin.Context) {
	id := ctx.GetUint("user_id")
	if id == 0 {
		ctx.Redirect(http.StatusFound, "/users/me?error=Invalid+user+ID")
		return
	}

	var input struct {
		Name     string `form:"name"`
		Username string `form:"username"`
	}

	if err := ctx.ShouldBind(&input); err != nil {
		ctx.Redirect(http.StatusFound, "/users/me?error=Invalid+input")
		return
	}

	// TODO: Replace with actual service call
	// For now, we'll just log the input
	c.logger.Info("Updating profile for user", logger.Uint("user_id", id), logger.String("name", input.Name), logger.String("username", input.Username))

	ctx.Redirect(http.StatusFound, "/users/me?success=Profile+updated+successfully")
}

// ChangePassword handles the form submission for changing the user's password.
func (c *UserController) ChangePassword(ctx *gin.Context) {
	id := ctx.GetUint("user_id")
	if id == 0 {
		ctx.Redirect(http.StatusFound, "/users/me?error=Invalid+user+ID")
		return
	}

	var input struct {
		CurrentPassword  string `form:"current_password"`
		NewPassword      string `form:"new_password"`
		ConfirmPassword  string `form:"confirm_password"`
	}

	if err := ctx.ShouldBind(&input); err != nil {
		ctx.Redirect(http.StatusFound, "/users/me?error=Invalid+input")
		return
	}

	if input.NewPassword != input.ConfirmPassword {
		ctx.Redirect(http.StatusFound, "/users/me?error=New+passwords+do+not+match")
		return
	}

	// TODO: Replace with actual service call to verify current password and update to new password
	c.logger.Info("Changing password for user", logger.Uint("user_id", id))

	ctx.Redirect(http.StatusFound, "/users/me?success=Password+changed+successfully")
}

// UploadAvatar handles the form submission for uploading a new user avatar.
func (c *UserController) UploadAvatar(ctx *gin.Context) {
	id := ctx.GetUint("user_id")
	if id == 0 {
		ctx.Redirect(http.StatusFound, "/users/me?error=Invalid+user+ID")
		return
	}

	file, err := ctx.FormFile("avatar")
	if err != nil {
		ctx.Redirect(http.StatusFound, "/users/me?error=Failed+to+upload+avatar")
		return
	}

	// TODO: Add logic to save the file and update the user's avatar URL
	c.logger.Info("Avatar uploaded for user", logger.Uint("user_id", id), logger.String("filename", file.Filename))

	ctx.Redirect(http.StatusFound, "/users/me?success=Avatar+uploaded+successfully")
}

func (c *UserController) Update(ctx *gin.Context) {
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

	// Render HTML template for web interface
	c.View("user/me.html").
		WithTitle("My Profile").
		WithData(map[string]interface{}{
			"user": item,
		}).
		Render(ctx)
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
func (c *UserController) UpdateAvatar(ctx *gin.Context) {
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

	// Render HTML template for web interface
	c.View("user/me.html").
		WithTitle("My Profile").
		WithData(gin.H{
			"user": updatedUser,
		}).
		Render(ctx)
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
func (c *UserController) UpdatePassword(ctx *gin.Context) {
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
