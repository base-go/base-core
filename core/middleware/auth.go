package middleware

import (
	"fmt" // Added for Sprintf
	"base/core/helper"
	"base/core/types"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// isAPIRequest checks if the request expects JSON response
func isAPIRequest(c *gin.Context) bool {
	// Check if request is to /api/* path
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		return true
	}
	
	// Check Accept header for JSON
	accept := c.GetHeader("Accept")
	return strings.Contains(accept, "application/json")
}

// IsAuthenticated checks if the user is authenticated
func IsAuthenticated(c *gin.Context) bool {
	// Check if user_id is set in context
	userID, exists := c.Get("user_id")
	if exists && userID != nil {
		return true
	}
	
	// Check for session
	session := sessions.Default(c)
	if session.Get("user_id") != nil {
		return true
	}
	
	// Check for token in cookie
	token, err := c.Cookie("auth_token")
	if err == nil && token != "" {
		_, _, err := helper.ValidateJWT(token)
		return err == nil
	}
	
	// Check for Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			_, _, err := helper.ValidateJWT(parts[1])
			return err == nil
		}
	}
	
	return false
}

// AuthMiddleware checks for authentication using different methods:
// 1. Session for web requests (cookie-based)
// 2. JWT token in Authorization header for API requests
// 3. JWT token in cookie for web requests (fallback)
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if this is a web request with session
		if !isAPIRequest(c) {
			session := sessions.Default(c)
			userID := session.Get("user_id")
			
			if userID != nil {
				// User is authenticated via session
				c.Set("user_id", userID)
				c.Next()
				return
			}
			
			// Check for token in cookie (fallback for web)
			token, err := c.Cookie("auth_token")
			if err == nil && token != "" {
				extend, userId, err := helper.ValidateJWT(token)
				if err == nil {
					c.Set("user_id", userId)
					c.Set("extend", extend)
					c.Next()
					return
				}
			}
			
			// No valid session or cookie, redirect to login
			lang := c.GetString("lang")
			if lang == "" {
				// Fallback if lang is not found, though it should be set by LanguageMiddleware
				// log.Warnf("Language not found in context during auth redirect, defaulting to 'en'")
				lang = "en"
			}
			redirectURL := fmt.Sprintf("/%s/auth/login", lang)
			c.Redirect(http.StatusFound, redirectURL)
			c.Abort()
			return
		}
		
		// API request - check Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warnf("Authorization header is required")
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.ErrorResponse{
				Error: "Authorization header is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			log.Warnf("Authorization header format must be Bearer {token}")
			
			if isAPIRequest(c) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, types.ErrorResponse{
					Error: "Invalid authorization format",
				})
			} else {
				c.Redirect(http.StatusFound, "/auth/login")
				c.Abort()
			}
			return
		}

		extend, userId, err := helper.ValidateJWT(parts[1])
		if err != nil {
			log.Warnf("Invalid or expired JWT: %s", err.Error())
			
			if isAPIRequest(c) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, types.ErrorResponse{
					Error: "Invalid or expired token",
				})
			} else {
				c.Redirect(http.StatusFound, "/auth/login")
				c.Abort()
			}
			return
		}

		c.Set("user_id", userId)
		c.Set("extend", extend)
		c.Next()
	}
}

// OptionalAuthMiddleware for routes that work with or without authentication
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check for session (for web requests)
		if !isAPIRequest(c) {
			session := sessions.Default(c)
			userID := session.Get("user_id")
			
			if userID != nil {
				// User is authenticated via session
				c.Set("user_id", userID)
				c.Next()
				return
			}
			
			// Check for token in cookie
			token, err := c.Cookie("auth_token")
			if err == nil && token != "" {
				extend, userId, err := helper.ValidateJWT(token)
				if err == nil {
					c.Set("user_id", userId)
					c.Set("extend", extend)
					c.Next()
					return
				}
			}
		}
		
		// Check Authorization header (for API requests)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			extend, userId, err := helper.ValidateJWT(parts[1])
			if err == nil {
				c.Set("user_id", userId)
				c.Set("extend", extend)
			}
		}
		
		c.Next()
	}
}
