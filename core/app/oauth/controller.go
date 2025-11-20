package oauth

import (
	"base/core/logger"
	"base/core/router"
	"net/http"
)

type OAuthController struct {
	Service *OAuthService
	Logger  logger.Logger
	Config  *OAuthConfig
}

func NewOAuthController(service *OAuthService, logger logger.Logger, config *OAuthConfig) *OAuthController {
	return &OAuthController{
		Service: service,
		Logger:  logger,
		Config:  config,
	}
}

func (c *OAuthController) Routes(router *router.RouterGroup) {
	router.POST("/google/callback", c.GoogleCallback)
	router.POST("/facebook/callback", c.FacebookCallback)
	router.POST("/apple/callback", c.AppleCallback)
}

// GoogleCallback godoc
// @Summary Google OAuth callback
// @Description Handle the OAuth callback from Google
// @Security ApiKeyAuth
// @Tags Core/OAuth
// @Accept json
// @Produce json
// @Param idToken body string true "Google Id Token"
// @Success 200 {object} profile.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /oauth/google/callback [post]
func (c *OAuthController) GoogleCallback(ctx *router.Context) error {
	var req struct {
		IdToken string `json:"idToken"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.Logger.Error("Failed to bind JSON request", logger.String("error", err.Error()))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request payload"})
		return nil
	}

	user, err := c.Service.ProcessGoogleOAuth(req.IdToken)
	if err != nil {
		c.Logger.Error("Google OAuth authentication failed", logger.String("error", err.Error()))
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return nil
	}

	ctx.JSON(http.StatusOK, user)
	return nil
}

// FacebookCallback godoc
// @Summary Facebook OAuth callback
// @Description Handle the OAuth callback from Facebook
// @Security ApiKeyAuth
// @Tags Core/OAuth
// @Accept json
// @Produce json
// @Param accessToken body string true "Facebook Access Token"
// @Success 200 {object} profile.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /oauth/facebook/callback [post]
func (c *OAuthController) FacebookCallback(ctx *router.Context) error {
	var req struct {
		AccessToken string `json:"accessToken"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.Logger.Error("Failed to bind JSON request", logger.String("error", err.Error()))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request payload"})
		return nil
	}

	user, err := c.Service.ProcessFacebookOAuth(req.AccessToken)
	if err != nil {
		c.Logger.Error("Facebook OAuth authentication failed", logger.String("error", err.Error()))
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return nil
	}

	ctx.JSON(http.StatusOK, user)
	return nil
}

// AppleCallback godoc
// @Summary Apple OAuth callback
// @Description Handle the OAuth callback from Apple
// @Security ApiKeyAuth
// @Tags Core/OAuth
// @Accept json
// @Produce json
// @Param idToken body string true "Apple Id Token"
// @Success 200 {object} profile.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /oauth/apple/callback [post]
func (c *OAuthController) AppleCallback(ctx *router.Context) error {
	var req struct {
		IdToken string `json:"idToken"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.Logger.Error("Failed to bind JSON request", logger.String("error", err.Error()))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request payload"})
		return nil
	}

	user, err := c.Service.ProcessAppleOAuth(req.IdToken)
	if err != nil {
		c.Logger.Error("Apple OAuth authentication failed", logger.String("error", err.Error()))
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return nil
	}

	ctx.JSON(http.StatusOK, user)
	return nil
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
