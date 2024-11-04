package auth

import (
	"base/core/email"
	"base/core/event"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthController struct {
	AuthService  *AuthService
	EmailSender  email.Sender
	Logger       *zap.Logger
	EventService *event.EventService
}

func NewAuthController(
	service *AuthService,
	emailSender email.Sender,
	logger *zap.Logger,
	eventService *event.EventService,
) *AuthController {
	return &AuthController{
		AuthService:  service,
		EmailSender:  emailSender,
		Logger:       logger,
		EventService: eventService,
	}
}

func (c *AuthController) Routes(router *gin.RouterGroup) {
	router.POST("/register", c.Register)
	router.POST("/login", c.Login)
	router.POST("/logout", c.Logout)
	router.POST("/forgot-password", c.ForgotPassword)
	router.POST("/reset-password", c.ResetPassword)
}

func (c *AuthController) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.trackEvent(ctx, "registration", "failed", "validation_error", req.Email, nil)
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := c.AuthService.Register(&req)
	if err != nil {
		c.trackEvent(ctx, "registration", "failed", "system_error", req.Email, map[string]interface{}{
			"error": err.Error(),
		})
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to register user"})
		return
	}

	// Send welcome email
	// msg := email.Message{
	// 	To:      []string{user.Email},
	// 	From:    "support@albafone.app",
	// 	Subject: "Welcome to Our Application",
	// 	Body:    c.getWelcomeEmailBody(user.Name),
	// 	IsHTML:  true,
	// }

	// err = email.Send(msg)
	// if err != nil {
	// 	c.Logger.Error("Failed to send welcome email",
	// 		zap.Error(err),
	// 		zap.String("email", user.Email))
	// 	c.trackEvent(ctx, "welcome_email", "failed", "email_error", user.Email, map[string]interface{}{
	// 		"error": err.Error(),
	// 	})
	// } else {
	// 	c.Logger.Info("Welcome email sent",
	// 		zap.String("email", user.Email))
	// 	c.trackEvent(ctx, "welcome_email", "success", "sent", user.Email, nil)
	// }

	c.trackEvent(ctx, "registration", "success", "completed", user.Email, map[string]interface{}{
		"user_id": user.ID,
		"name":    user.Name,
	})

	ctx.JSON(http.StatusCreated, user)
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.trackEvent(ctx, "login", "failed", "validation_error", req.Email, nil)
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	response, err := c.AuthService.Login(&req)
	if err != nil {
		c.trackEvent(ctx, "login", "failed", "invalid_credentials", req.Email, map[string]interface{}{
			"ip_address": ctx.ClientIP(),
		})
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
		return
	}

	c.trackEvent(ctx, "login", "success", "authenticated", req.Email, map[string]interface{}{
		"user_id":    response.ID,
		"ip_address": ctx.ClientIP(),
	})

	ctx.JSON(http.StatusOK, response)
}

func (c *AuthController) Logout(ctx *gin.Context) {
	userID := ctx.GetString("user_id")
	userEmail := ctx.GetString("user_email")

	c.trackEvent(ctx, "logout", "success", "logged_out", userEmail, map[string]interface{}{
		"user_id":    userID,
		"ip_address": ctx.ClientIP(),
	})

	ctx.JSON(http.StatusOK, SuccessResponse{Message: "Logout successful"})
}

func (c *AuthController) ForgotPassword(ctx *gin.Context) {
	var req ForgotPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.trackEvent(ctx, "forgot_password", "failed", "validation_error", req.Email, nil)
		c.Logger.Error("Failed to bind JSON in ForgotPassword", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.Logger.Info("Processing forgot password request", zap.String("email", req.Email))

	err := c.AuthService.ForgotPassword(req.Email)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			c.trackEvent(ctx, "forgot_password", "failed", "user_not_found", req.Email, nil)
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		} else {
			c.trackEvent(ctx, "forgot_password", "failed", "system_error", req.Email, map[string]interface{}{
				"error": err.Error(),
			})
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "An error occurred while processing your request"})
		}
		return
	}

	c.trackEvent(ctx, "forgot_password", "success", "reset_requested", req.Email, map[string]interface{}{
		"ip_address": ctx.ClientIP(),
	})

	ctx.JSON(http.StatusOK, SuccessResponse{Message: "Password reset email sent"})
}

func (c *AuthController) ResetPassword(ctx *gin.Context) {
	var req ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.trackEvent(ctx, "reset_password", "failed", "validation_error", req.Email, nil)
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format"})
		return
	}

	err := c.AuthService.ResetPassword(req.Email, req.Token, req.NewPassword)
	if err != nil {
		var eventReason string
		switch {
		case errors.Is(err, ErrInvalidToken):
			eventReason = "invalid_token"
			ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid reset token"})
		case errors.Is(err, ErrTokenExpired):
			eventReason = "token_expired"
			ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Reset token has expired"})
		default:
			eventReason = "system_error"
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to reset password"})
		}

		c.trackEvent(ctx, "reset_password", "failed", eventReason, req.Email, map[string]interface{}{
			"error":      err.Error(),
			"ip_address": ctx.ClientIP(),
		})
		return
	}

	c.trackEvent(ctx, "reset_password", "success", "password_changed", req.Email, map[string]interface{}{
		"ip_address": ctx.ClientIP(),
	})

	ctx.JSON(http.StatusOK, SuccessResponse{Message: "Password reset successful"})
}

// trackEvent is a helper function to track authentication events
func (c *AuthController) trackEvent(ctx *gin.Context, eventType, status, reason, userEmail string, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	// Add common metadata
	metadata["ip_address"] = ctx.ClientIP()
	metadata["user_agent"] = ctx.Request.UserAgent()
	metadata["reason"] = reason

	_, err := c.EventService.Track(ctx.Request.Context(), event.EventOptions{
		Type:     eventType,
		Category: "authentication",
		Actor:    "user",
		ActorID:  userEmail,
		Target:   "auth_system",
		Action:   ctx.Request.Method,
		Status:   status,
		Metadata: metadata,
	})

	if err != nil {
		c.Logger.Error("Failed to track auth event",
			zap.Error(err),
			zap.String("type", eventType),
			zap.String("status", status),
			zap.String("user", userEmail))
	}
}

func (c *AuthController) getWelcomeEmailBody(name string) string {
	return "<h1>Welcome to Albafone!</h1>" +
		"<p>Hi " + name + ",</p>" +
		"<p>Thank you for registering with our application.</p>" +
		"<p>Best regards,<br>Team</p>"
}
