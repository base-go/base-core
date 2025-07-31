package middleware

import (
	"base/core/helper"
	"base/core/logger"
	"base/core/types"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware checks for Bearer token authentication
func AuthMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("Authorization header is required")
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.ErrorResponse{
				Error: "Authorization header is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			log.Warn("Authorization header format must be Bearer {token}")
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.ErrorResponse{
				Error: "Invalid authorization format",
			})
			return
		}

		extend, userId, err := helper.ValidateJWT(parts[1])
		if err != nil {
			log.Warn("Invalid or expired JWT", logger.String("error", err.Error()))
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.ErrorResponse{
				Error: "Invalid or expired token",
			})
			return
		}

		c.Set("user_id", userId)
		c.Set("extend", extend)
		c.Next()
	}
}
