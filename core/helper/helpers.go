package helper

import (
	"base/core/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userID uint, extend interface{}) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	cfg := config.NewConfig()

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	claims["extend"] = extend

	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateJWT(tokenString string) (uint, error) {
	cfg := config.NewConfig()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := uint(claims["user_id"].(float64))
		return userID, nil
	}

	return 0, jwt.ErrSignatureInvalid
}
