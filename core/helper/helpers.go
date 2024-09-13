package helper

import (
	"base/app/models"
	"base/core/config"
	"reflect"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func GenerateJWT(userID uint) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	cfg := config.NewConfig()

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

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
func ToResponseSlice(items interface{}) interface{} {
    val := reflect.ValueOf(items)
    responses := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf((*models.PostResponse)(nil))), val.Len(), val.Cap())

    for i := 0; i < val.Len(); i++ {
        item := val.Index(i).Interface().(*models.Post)
        response := &models.PostResponse{
            ID:      item.ID,
            Title:   item.Title,
            Content: item.Content,
            // Map other fields
        }
        responses.Index(i).Set(reflect.ValueOf(response))
    }

    return responses.Interface()
}
