package auth

import (
	"base/core/email"
	"base/core/logger"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthController struct {
	service     *AuthService
	emailSender email.Sender
	logger      logger.Logger
}

func NewAuthController(service *AuthService, emailSender email.Sender, logger logger.Logger) *AuthController {
	return &AuthController{
		service:     service,
		emailSender: emailSender,
		logger:      logger,
	}
}

func (c *AuthController) Routes(router *gin.RouterGroup) {
	router.POST("/register", c.Register)
	router.POST("/login", c.Login)
	router.POST("/logout", c.Logout)
	router.POST("/forgot-password", c.ForgotPassword)
	router.POST("/reset-password", c.ResetPassword)
}

// @Summary Register
// @Description Register user
// @Security ApiKeyAuth
// @Tags Core/Auth
// @Accept json
// @Produce json
// @Param body body RegisterRequest true "Register Request"
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

	user, err := c.service.Register(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to register user"})
		return
	}

	//	Send welcome email
	msg := email.Message{
		To:      []string{user.Email},
		From:    "no-reply@base.al",
		Subject: "Welcome to Base",
		Body:    c.getWelcomeEmailBody(user.Name),
		IsHTML:  true,
	}

	err = email.Send(msg)
	if err != nil {
		c.logger.Error("Failed to send welcome email",
			logger.String("error", err.Error()),
			logger.String("email", user.Email))
	} else {
		c.logger.Info("Welcome email sent",
			logger.String("email", user.Email))
	}

	ctx.JSON(http.StatusCreated, user)
}

// @Summary Login
// @Description Login user
// @Security ApiKeyAuth
// @Tags Core/Auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Login Request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	response, err := c.service.Login(&req)
	if err != nil {
		if strings.Contains(err.Error(), "access_denied") {
			// Return both the response and error when user is not an author
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
				"data":  response,
			})
			return
		}
		if strings.Contains(err.Error(), "invalid credentials") {
			ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// Logout handles user logout
// @Summary Logout
// @Description Logout user
// @Security ApiKeyAuth
// @Tags Core/Auth
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, SuccessResponse{Message: "Logout successful"})
}

// @Summary Forgot Password
// @Description Request to reset password
// @Security ApiKeyAuth
// @Tags Core/Auth
// @Accept json
// @Produce json
// @Param body body ForgotPasswordRequest true "Forgot Password Request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/forgot-password [post]
func (c *AuthController) ForgotPassword(ctx *gin.Context) {
	var req ForgotPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("Failed to bind JSON in ForgotPassword", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.logger.Info("Processing forgot password request", zap.String("email", req.Email))

	err := c.service.ForgotPassword(req.Email)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "An error occurred while processing your request"})
		}
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{Message: "Password reset email sent"})
}

// ResetPassword handles password reset requests
// @Summary Reset Password
// @Description Reset user password using token
// @Security ApiKeyAuth
// @Tags Core/Auth
// @Accept json
// @Produce json
// @Param body body ResetPasswordRequest true "Reset Password Request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/reset-password [post]
func (c *AuthController) ResetPassword(ctx *gin.Context) {
	var req ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format"})
		return
	}

	err := c.service.ResetPassword(req.Email, req.Token, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidToken):
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid or expired token"})
		case errors.Is(err, ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to reset password"})
		}
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{Message: "Password reset successful"})
}

func (c *AuthController) getWelcomeEmailBody(name string) string {
	return "<h1>Welcome to Base!</h1>" +
		"<p>Hi " + name + ",</p>" +
		"<p>Thank you for registering with our application.</p>" +
		"<p>Best regards,<br>Team</p>"
}
