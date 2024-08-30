package middleware

import (
	helper "base/core/helper"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warnf("Authorization header is required")
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			log.Warnf("Authorization header format must be Bearer {token}")
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		userID, err := helper.ValidateJWT(parts[1])
		if err != nil {
			log.Warnf("Invalid or expired JWT: %s", err.Error())
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired JWT"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
