package helper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	// ContextKeyPrefix is the prefix for context keys
	ContextKeyPrefix = "base_"
)

// GetContextString retrieves a string value from the context with the given key
// The key should be provided without the base_ prefix
func GetContextString(c *gin.Context, key string) string {
	// If the key doesn't have the prefix, add it
	if !strings.HasPrefix(key, ContextKeyPrefix) {
		key = ContextKeyPrefix + strings.ToLower(key)
	}

	// Get the value from the context
	value, exists := c.Get(key)
	if !exists {
		return ""
	}

	// Convert to string
	strValue, ok := value.(string)
	if !ok {
		return ""
	}

	return strValue
}

// GetContextUint retrieves a uint value from the context with the given key
// The key should be provided without the base_ prefix
// Returns 0 if not found or invalid
func GetContextUint(c *gin.Context, key string) uint {
	// Get the string value
	strValue := GetContextString(c, key)
	if strValue == "" {
		return 0
	}

	// Convert to uint
	var uintValue uint
	_, err := fmt.Sscanf(strValue, "%d", &uintValue)
	if err != nil {
		log.Warnf("Failed to convert context value %s to uint: %v", strValue, err)
		return 0
	}

	return uintValue
}

// GetContextInt retrieves an int value from the context with the given key
// The key should be provided without the base_ prefix
// Returns 0 if not found or invalid
func GetContextInt(c *gin.Context, key string) int {
	strValue := GetContextString(c, key)
	if strValue == "" {
		return 0
	}

	intValue, err := strconv.Atoi(strValue)
	if err != nil {
		log.Warnf("Failed to convert context value %s to int: %v", strValue, err)
		return 0
	}

	return intValue
}

// GetContextBool retrieves a boolean value from the context with the given key
// The key should be provided without the base_ prefix
// Returns false if not found or invalid
func GetContextBool(c *gin.Context, key string) bool {
	strValue := GetContextString(c, key)
	if strValue == "" {
		return false
	}

	strValue = strings.ToLower(strValue)
	return strValue == "true" || strValue == "1" || strValue == "yes"
}

// GetContextFloat retrieves a float64 value from the context with the given key
// The key should be provided without the base_ prefix
// Returns 0 if not found or invalid
func GetContextFloat(c *gin.Context, key string) float64 {
	strValue := GetContextString(c, key)
	if strValue == "" {
		return 0
	}

	floatValue, err := strconv.ParseFloat(strValue, 64)
	if err != nil {
		log.Warnf("Failed to convert context value %s to float: %v", strValue, err)
		return 0
	}

	return floatValue
}
