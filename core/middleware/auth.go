package middleware

import (
	"base/core/types"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// AuthMiddleware checks for Bearer token authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
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
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.ErrorResponse{
				Error: "Invalid authorization format",
			})
			return
		}

		userId, err := types.ValidateJWT(parts[1])
		if err != nil {
			log.Warnf("Invalid or expired JWT: %s", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.ErrorResponse{
				Error: "Invalid or expired token",
			})
			return
		}

		c.Set("user_id", userId)
		c.Next()
	}
}
