package auth

import (
	"base/app/templates"
	"base/core/email"
	"base/core/logger"
	"base/core/template"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-contrib/sessions"
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
	// API routes for JSON responses
	router.POST("/register", c.Register)
	router.POST("/login", c.Login)
	router.POST("/logout", c.Logout)
	router.POST("/forgot-password", c.ForgotPassword)
	router.POST("/reset-password", c.ResetPassword)
}

// SetupWebRoutes registers the web authentication routes
func (c *AuthController) SetupWebRoutes(router *gin.RouterGroup, templateEngine *template.Engine) {
	// Web page routes - these render HTML templates
	authRoutes := router.Group("/auth")

	authRoutes.GET("/login", func(ctx *gin.Context) {
		success := ctx.Query("success")
		error := ctx.Query("error")

		// Use Gin's HTML renderer with template constants like the register page
		templateEngine.RenderWithLayout(ctx.Writer, "auth/login.html", templates.LayoutAuth, gin.H{
			"success": success,
			"error":   error,
		}, ctx)
	})

	authRoutes.GET("/register", func(ctx *gin.Context) {
		error := ctx.Query("error")

		// Use the template constants instead of hardcoded paths
		// The templates package defines LayoutAuth as "auth.html"
		templateEngine.RenderWithLayout(ctx.Writer, templates.PageRegister, templates.LayoutAuth, gin.H{
			"title": "Register",
			"error": error,
		}, ctx)
	})

	authRoutes.GET("/forgot-password", func(ctx *gin.Context) {
		error := ctx.Query("error")

		templateEngine.RenderWithLayout(ctx.Writer, "auth/forgot-password.html", templates.LayoutAuth, gin.H{
			"title": "Forgot Password",
			"error": error,
		}, ctx)
	})

	authRoutes.GET("/reset-password", func(ctx *gin.Context) {
		token := ctx.Query("token")
		email := ctx.Query("email")
		error := ctx.Query("error")

		templateEngine.RenderWithLayout(ctx.Writer, "auth/reset-password.html", templates.LayoutAuth, gin.H{
			"title": "Reset Password",
			"token": token,
			"email": email,
			"error": error,
		}, ctx)
	})

	// For backward compatibility, also register without the /auth prefix
	router.GET("/login", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/auth/login")
	})

	router.GET("/register", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/auth/register")
	})

	router.GET("/forgot-password", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/auth/forgot-password")
	})

	router.GET("/reset-password", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/reset-password?"+ctx.Request.URL.RawQuery)
	})

	// Web form submission routes
	authRoutes.POST("/login", c.LoginWeb)
	authRoutes.POST("/register", c.RegisterWeb)
	authRoutes.POST("/logout", c.LogoutWeb)
	authRoutes.POST("/forgot-password", c.ForgotPasswordWeb)
	authRoutes.POST("/reset-password", c.ResetPasswordWeb)
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

// LoginWeb handles web login requests and sets session data
func (c *AuthController) LoginWeb(ctx *gin.Context) {
	// Check content type to determine how to parse the request
	contentType := ctx.GetHeader("Content-Type")
	var req LoginRequest
	var rememberMe bool

	if strings.Contains(contentType, "application/json") {
		// Handle JSON request
		if err := ctx.ShouldBindJSON(&req); err != nil {
			handleLoginError(ctx, "Invalid request format", http.StatusBadRequest)
			return
		}
		// Check if remember_me was included in JSON
		rememberMe = ctx.Query("remember_me") == "true"
	} else {
		// Handle form submission
		req.Email = ctx.PostForm("email")
		req.Password = ctx.PostForm("password")
		rememberMe = ctx.PostForm("remember-me") == "true"

		if req.Email == "" || req.Password == "" {
			handleLoginError(ctx, "Email and password are required", http.StatusBadRequest)
			return
		}
	}

	response, err := c.service.Login(&req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid credentials") {
			handleLoginError(ctx, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		handleLoginError(ctx, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set session data for web authentication
	session := sessions.Default(ctx)
	session.Set("user_id", response.Id)
	session.Set("username", response.Username)
	session.Set("email", response.Email)
	session.Set("logged_in", true)

	// Set session expiration based on remember me
	sessionMaxAge := 60 * 60 * 24 // 1 day by default
	if rememberMe {
		sessionMaxAge = 60 * 60 * 24 * 30 // 30 days
	}

	session.Options(sessions.Options{
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
	})
	session.Save()

	// Also set auth cookie for API access
	ctx.SetCookie(
		"auth_token",
		response.AccessToken,
		sessionMaxAge,
		"/",
		"",
		false, // Set to true in production with HTTPS
		false, // Not HTTP only to allow JavaScript access
	)

	// Store user info in context for middleware
	ctx.Set("user", response.UserResponse)

	// Return response based on accepted content type
	if acceptsJSON(ctx) {
		ctx.JSON(http.StatusOK, response)
	} else {
		// Redirect to posts page for non-API calls
		ctx.Redirect(http.StatusFound, "/posts")
	}
}

// Helper function to handle login errors
func handleLoginError(ctx *gin.Context, message string, status int) {
	if acceptsJSON(ctx) {
		ctx.JSON(status, ErrorResponse{Error: message})
	} else {
		// Redirect back to login page with error for traditional web flow
		ctx.Redirect(http.StatusFound, "/login?error="+url.QueryEscape(message))
	}
}

// RegisterWeb handles web registration requests
func (c *AuthController) RegisterWeb(ctx *gin.Context) {
	// Check content type to determine how to parse the request
	contentType := ctx.GetHeader("Content-Type")
	var req RegisterRequest
	var passwordConfirmation string

	if strings.Contains(contentType, "application/json") {
		// Handle JSON request
		if err := ctx.ShouldBindJSON(&req); err != nil {
			handleRegisterError(ctx, "Invalid request format: "+err.Error(), http.StatusBadRequest)
			return
		}
		// For JSON requests, we'll assume password was already confirmed on client side
	} else {
		// Handle form submission
		req.Name = ctx.PostForm("name")
		req.Username = ctx.PostForm("username")
		req.Email = ctx.PostForm("email")
		req.Password = ctx.PostForm("password")
		passwordConfirmation = ctx.PostForm("password_confirmation")

		// Validate required fields
		if req.Email == "" || req.Password == "" || req.Name == "" || req.Username == "" {
			handleRegisterError(ctx, "All fields are required", http.StatusBadRequest)
			return
		}

		// Check password confirmation
		if passwordConfirmation != "" && passwordConfirmation != req.Password {
			handleRegisterError(ctx, "Passwords do not match", http.StatusBadRequest)
			return
		}
	}

	// Validate password length
	if len(req.Password) < 8 {
		handleRegisterError(ctx, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// Register the user
	user, err := c.service.Register(&req)
	if err != nil {
		// Check for specific errors
		if strings.Contains(err.Error(), "user already exists") {
			handleRegisterError(ctx, "Email or username already exists", http.StatusConflict)
			return
		}

		c.logger.Error("Failed to register user",
			logger.String("email", req.Email),
			logger.String("error", err.Error()))
		handleRegisterError(ctx, "Failed to register user", http.StatusInternalServerError)
		return
	}

	// Set session data for web authentication
	session := sessions.Default(ctx)
	session.Set("user_id", user.Id)
	session.Set("username", user.Username)
	session.Set("email", user.Email)
	session.Set("logged_in", true)
	session.Options(sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 30, // 30 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
	})
	session.Save()

	// Also set auth cookie for API access
	ctx.SetCookie(
		"auth_token",
		user.AccessToken,
		60*60*24*30, // 30 days in seconds
		"/",
		"",
		false, // Set to true in production with HTTPS
		false, // Not HTTP only to allow JavaScript access
	)

	// Send welcome email asynchronously to not block the response
	go func() {
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
	}()

	// Return response based on accepted content type
	if acceptsJSON(ctx) {
		ctx.JSON(http.StatusCreated, user)
	} else {
		// Redirect to posts page for non-API calls
		ctx.Redirect(http.StatusFound, "/posts")
	}
}

// Helper function to handle registration errors
func handleRegisterError(ctx *gin.Context, message string, status int) {
	if acceptsJSON(ctx) {
		ctx.JSON(status, ErrorResponse{Error: message})
	} else {
		// Redirect back to register page with error for traditional web flow
		ctx.Redirect(http.StatusFound, "/register?error="+url.QueryEscape(message))
	}
}

// LogoutWeb handles user logout and clears cookies
func (c *AuthController) LogoutWeb(ctx *gin.Context) {
	// Clear session
	session := sessions.Default(ctx)
	session.Clear()
	session.Options(sessions.Options{
		Path:     "/",
		MaxAge:   -1, // Expire immediately
		HttpOnly: true,
	})
	session.Save()

	// Clear auth cookie as well
	ctx.SetCookie("auth_token", "", -1, "/", "", false, false)
	ctx.Set("user", nil)

	// Return response based on accepted content type
	if acceptsJSON(ctx) {
		ctx.JSON(http.StatusOK, SuccessResponse{Message: "Logged out successfully"})
	} else {
		// Redirect to login page with success message
		ctx.Redirect(http.StatusFound, "/login?success="+url.QueryEscape("Logged out successfully"))
	}
}

// ForgotPasswordWeb handles forgot password requests for web
func (c *AuthController) ForgotPasswordWeb(ctx *gin.Context) {
	// Check content type to determine how to parse the request
	contentType := ctx.GetHeader("Content-Type")
	var req ForgotPasswordRequest

	if strings.Contains(contentType, "application/json") {
		// Handle JSON request
		if err := ctx.ShouldBindJSON(&req); err != nil {
			c.logger.Error("Failed to bind JSON in ForgotPasswordWeb", zap.Error(err))
			handleForgotPasswordResponse(ctx, err.Error(), false)
			return
		}
	} else {
		// Handle form submission
		req.Email = ctx.PostForm("email")

		if req.Email == "" {
			handleForgotPasswordResponse(ctx, "Email is required", false)
			return
		}
	}

	c.logger.Info("Processing forgot password request", zap.String("email", req.Email))

	err := c.service.ForgotPassword(req.Email)
	if err != nil {
		// Log the error but don't reveal if email exists or not for security
		c.logger.Error("Error in forgot password process",
			logger.String("email", req.Email),
			logger.String("error", err.Error()))
	}

	// Always return success to prevent email enumeration attacks
	handleForgotPasswordResponse(ctx, "If your email is registered, you will receive a password reset link", true)
}

// Helper function to handle forgot password responses
func handleForgotPasswordResponse(ctx *gin.Context, message string, isSuccess bool) {
	if acceptsJSON(ctx) {
		if isSuccess {
			ctx.JSON(http.StatusOK, SuccessResponse{Message: message})
		} else {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: message})
		}
	} else {
		if isSuccess {
			// Redirect to login page with success message
			ctx.Redirect(http.StatusFound, "/login?success="+url.QueryEscape(message))
		} else {
			// Redirect back to forgot password page with error
			ctx.Redirect(http.StatusFound, "/forgot-password?error="+url.QueryEscape(message))
		}
	}
}

// Helper function to check if the client accepts JSON responses
func acceptsJSON(ctx *gin.Context) bool {
	acceptHeader := ctx.GetHeader("Accept")
	return strings.Contains(acceptHeader, "application/json") || ctx.GetHeader("Content-Type") == "application/json"
}

// ResetPasswordWeb handles password reset requests for web
func (c *AuthController) ResetPasswordWeb(ctx *gin.Context) {
	// Check content type to determine how to parse the request
	contentType := ctx.GetHeader("Content-Type")
	var req ResetPasswordRequest

	if strings.Contains(contentType, "application/json") {
		// Handle JSON request
		if err := ctx.ShouldBindJSON(&req); err != nil {
			handleResetPasswordError(ctx, "Invalid request format", http.StatusBadRequest)
			return
		}
	} else {
		// Handle form submission
		req.Email = ctx.PostForm("email")
		req.Token = ctx.PostForm("token")
		req.NewPassword = ctx.PostForm("new_password")
		confirmPassword := ctx.PostForm("confirm_password")

		// Basic validation
		if req.Email == "" || req.Token == "" || req.NewPassword == "" {
			handleResetPasswordError(ctx, "All fields are required", http.StatusBadRequest)
			return
		}

		// Check password confirmation
		if confirmPassword != "" && confirmPassword != req.NewPassword {
			handleResetPasswordError(ctx, "Passwords do not match", http.StatusBadRequest)
			return
		}
	}

	// Validate password length
	if len(req.NewPassword) < 8 {
		handleResetPasswordError(ctx, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	err := c.service.ResetPassword(req.Email, req.Token, req.NewPassword)
	if err != nil {
		message := "Failed to reset password"
		status := http.StatusInternalServerError

		switch {
		case errors.Is(err, ErrInvalidToken):
			message = "Invalid or expired token"
			status = http.StatusBadRequest
		case errors.Is(err, ErrUserNotFound):
			message = "User not found"
			status = http.StatusNotFound
		}

		c.logger.Error("Password reset failed",
			logger.String("email", req.Email),
			logger.String("error", err.Error()))

		handleResetPasswordError(ctx, message, status)
		return
	}

	// Successfully reset password
	if acceptsJSON(ctx) {
		ctx.JSON(http.StatusOK, SuccessResponse{Message: "Password reset successful"})
	} else {
		// Redirect to login page with success message
		ctx.Redirect(http.StatusFound, "/login?success="+url.QueryEscape("Password reset successful. Please login with your new password."))
	}
}

// Helper function to handle password reset errors
func handleResetPasswordError(ctx *gin.Context, message string, status int) {
	if acceptsJSON(ctx) {
		ctx.JSON(status, ErrorResponse{Error: message})
	} else {
		// Get the token and email from request to maintain state
		token := ctx.PostForm("token")
		email := ctx.PostForm("email")
		if token == "" {
			token = ctx.Query("token")
		}
		if email == "" {
			email = ctx.Query("email")
		}

		// Redirect back to reset password page with error and preserving token/email
		redirectURL := fmt.Sprintf("/reset-password?token=%s&email=%s&error=%s",
			url.QueryEscape(token),
			url.QueryEscape(email),
			url.QueryEscape(message))
		ctx.Redirect(http.StatusFound, redirectURL)
	}
}
