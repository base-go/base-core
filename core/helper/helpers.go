package helper

import (
	"base/core/types"
)

// GenerateJWT is a wrapper around types.GenerateJWT for backward compatibility
func GenerateJWT(userID uint, extend interface{}) (string, error) {
	return types.GenerateJWT(userID, extend)
}

// ValidateJWT is a wrapper around types.ValidateJWT for backward compatibility
func ValidateJWT(tokenString string) (uint, error) {
	return types.ValidateJWT(tokenString)
}
