package helper

import (
	"base/core/config"
	"base/core/types"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateJWT is a wrapper around types.GenerateJWT for backward compatibility
func GenerateJWT(userID uint, extend interface{}) (string, error) {
	return types.GenerateJWT(userID, extend)
}

func ValidateJWT(tokenString string) (interface{}, uint, error) {
	cfg := config.NewConfig()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return 0, 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := uint(claims["user_id"].(float64))
		extend := claims["extend"]

		return extend, userID, nil
	}

	return 0, 0, jwt.ErrSignatureInvalid
}
