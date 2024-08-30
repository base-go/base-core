package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	AuthService *AuthService
}

func NewAuthController(service *AuthService) *AuthController {
	return &AuthController{
		AuthService: service,
	}
}

func (c *AuthController) Routes(router *gin.RouterGroup) {
	router.POST("/register", c.Register)
	router.POST("/login", c.Login)
	router.POST("/logout", c.Logout)
	router.POST("/forgot-password", c.ForgotPassword)
	router.POST("/reset-password", c.ResetPassword)
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with the input payload
// @Tags Core/Auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "Register User"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := c.AuthService.Register(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to register user"})
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

// Login godoc
// @Summary User login
// @Description Authenticate a user and return a token
// @Tags Core/Auth
// @Accept json
// @Produce json
// @Param user body LoginRequest true "Login User"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	response, err := c.AuthService.Login(&req)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// Logout godoc
// @Summary User logout
// @Description Logout a user
// @Tags Core/Auth
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} SuccessResponse
// @Router /auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	// In a stateless JWT-based auth system, logout is typically handled client-side
	// by removing the token. Here we just return a success message.
	ctx.JSON(http.StatusOK, SuccessResponse{Message: "Logout successful"})
}

// ForgotPassword godoc
// @Summary Request password reset
// @Description Request a password reset token
// @Tags Core/Auth
// @Accept json
// @Produce json
// @Param email body ForgotPasswordRequest true "User Email"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /auth/forgot-password [post]
func (c *AuthController) ForgotPassword(ctx *gin.Context) {
	var req ForgotPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := c.AuthService.ForgotPassword(req.Email)
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{Message: "Password reset email sent"})
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset user password using a token
// @Tags Core/Auth
// @Accept json
// @Produce json
// @Param reset body ResetPasswordRequest true "Reset Password"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/reset-password [post]
func (c *AuthController) ResetPassword(ctx *gin.Context) {
	var req ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := c.AuthService.ResetPassword(req.Email, req.Token, req.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired reset token"})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{Message: "Password reset successful"})
}
