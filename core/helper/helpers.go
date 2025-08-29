package helper

import (
	"base/core/config"
	"base/core/types"
	"errors"
	"fmt"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var PluralizeClient *pluralize.Client

func init() {
	PluralizeClient = pluralize.NewClient()
}

// GenerateJWT is a wrapper around types.GenerateJWT for backward compatibility
func GenerateJWT(userId uint) (string, error) {
	return types.GenerateJWT(userId, nil)
}

func ValidateJWT(tokenString string) (any, uint, error) {
	cfg := config.NewConfig()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return 0, 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId := uint(claims["user_id"].(float64))

		return nil, userId, nil
	}

	return nil, 0, jwt.ErrSignatureInvalid
}

// ModelRegistry holds registered model constructors for dynamic object retrieval
var ModelRegistry = make(map[string]func() any)

// RegisterModel registers a model constructor for dynamic retrieval
// Example: RegisterModel("category", func() interface{} { return &Category{} })
func RegisterModel(tableName string, constructor func() any) {
	ModelRegistry[tableName] = constructor
}

// GetObject dynamically retrieves an object by field and value
// fieldName should be in format like "category_id", "user_id", etc.
// This will automatically determine the table and model type from the field name
func GetObject(db *gorm.DB, fieldName string, fieldValue any) (any, error) {
	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	// Extract table name from field name (e.g., "category_id" -> "categories")
	tableName := extractTableNameFromField(fieldName)
	if tableName == "" {
		return nil, fmt.Errorf("cannot determine table name from field: %s", fieldName)
	}

	// Check if model is registered
	constructor, exists := ModelRegistry[tableName]
	if !exists {
		return nil, fmt.Errorf("model not registered for table: %s", tableName)
	}

	// Create new instance of the model
	model := constructor()

	// Determine the query field (usually "id" for foreign keys)
	queryField := "id"
	if strings.HasSuffix(fieldName, "_id") {
		queryField = "id"
	} else {
		queryField = fieldName
	}

	// Query the database
	result := db.Where(fmt.Sprintf("%s = ?", queryField), fieldValue).First(model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("record not found in %s with %s = %v", tableName, queryField, fieldValue)
		}
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	return model, nil
}

// GetObjectAs is a generic version that returns the object cast to the specified type
func GetObjectAs[T any](db *gorm.DB, fieldName string, fieldValue interface{}) (*T, error) {
	obj, err := GetObject(db, fieldName, fieldValue)
	if err != nil {
		return nil, err
	}

	result, ok := obj.(*T)
	if !ok {
		return nil, fmt.Errorf("object cannot be cast to requested type")
	}

	return result, nil
}

// extractTableNameFromField extracts plural table name from field name
// Examples: "category_id" -> "categories", "user_id" -> "users", "tag_id" -> "tags"
// English pluralization rules are used, add custom rules if needed
func extractTableNameFromField(fieldName string) string {
	// Remove "_id" suffix if present
	fieldName = strings.TrimSuffix(fieldName, "_id")

	return PluralizeClient.Pluralize(fieldName, 1, false)
}
