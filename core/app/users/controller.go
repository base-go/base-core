package users

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserController struct {
	UserService *UserService
}

func NewUserController(service *UserService) *UserController {
	return &UserController{
		UserService: service,
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

// CreateUser godoc
// @Summary Create a new User
// @Description Create a new User with the input payload
// @Tags Core/Users
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param users body CreateRequest true "Create User"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users [post]
func (c *UserController) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.UserService.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create item"})
		return
	}

	ctx.JSON(http.StatusCreated, item)
}

// GetUser godoc
// @Summary Get a User
// @Description Get a User by its ID
// @Tags Core/Users
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} User
// @Failure 404 {object} ErrorResponse
// @Router /users/{id} [get]
func (c *UserController) Get(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	item, err := c.UserService.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Item not found"})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// ListUser godoc
// @Summary List User
// @Description Get a list of all User
// @Tags Core/Users
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {array} User
// @Failure 500 {object} ErrorResponse
// @Router /users [get]
func (c *UserController) List(ctx *gin.Context) {
	items, err := c.UserService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch items"})
		return
	}

	ctx.JSON(http.StatusOK, items)
}

// UpdateUser godoc
// @Summary Update a User
// @Description Update a User by its ID
// @Tags Core/Users
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param users body UpdateRequest true "Update User"
// @Success 200 {object} UserResponse
// @Failure 400 {object} ErrorResponse
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
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update user: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// DeleteUser godoc
// @Summary Delete a User
// @Description Delete a User by its ID
// @Tags Core/Users
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [delete]
func (c *UserController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.UserService.Delete(uint(id)); err != nil {
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

// UpdateAvatar godoc
// @Summary Update a User's avatar
// @Description Update a User's avatar by user ID
// @Tags Core/Users
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "User ID"
// @Param avatar formData file true "Avatar file"
// @Success 200 {object} UserResponse
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

	updatedUser, err := c.UserService.UpdateAvatar(uint(id), file)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update avatar: " + err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, updatedUser)
}

// UpdatePassword godoc
// @Summary Update a User's password
// @Description Update a User's password by user ID
// @Tags Core/Users
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param password body UpdatePasswordRequest true "Password"
// @Success 200 {object} SuccessResponse
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
		fmt.Printf("Binding error: %v\n", err)
		fmt.Printf("Request body: %+v\n", ctx.Request.Body)
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input: " + err.Error()})
		return
	}

	fmt.Printf("Received UpdatePasswordRequest: %+v\n", req)

	// Additional server-side validation for password strength
	if len(req.NewPassword) < 6 {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "New password must be at least 6 characters long"})
		return
	}

	err = c.UserService.UpdatePassword(uint(id), &req)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Current password is incorrect"})
		default:
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update password: " + err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{Message: "Password updated successfully"})
}
