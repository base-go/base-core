package translation

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Translation represents a translation entity for any model field
type Translation struct {
	Id        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	Key       string         `json:"key" gorm:"type:varchar(255);index:idx_translation_lookup"`
	Value     string         `json:"value" gorm:"type:text"`
	Model     string         `json:"model" gorm:"type:varchar(255);index:idx_translation_lookup"`
	ModelId   uint           `json:"model_id" gorm:"type:uint;index:idx_translation_lookup"`
	Language  string         `json:"language" gorm:"type:char(5);index:idx_translation_lookup"`
}

// Field represents a field that can be translated into multiple languages
// It automatically loads and provides translations in JSON format like ActiveStorage
type Field struct {
	Original string            `json:"-"` // Internal storage only
	Values   map[string]string `json:"-"` // Internal storage for translations
}

func TranslatedField(original string) Field {
	return Field{
		Original: original,
		Values:   make(map[string]string),
	}
}

// MarshalJSON implements custom JSON marshaling for Field
func (f Field) MarshalJSON() ([]byte, error) {
	// If no translations loaded, return the original value as a simple string
	if len(f.Values) == 0 {
		return json.Marshal(f.Original)
	}

	// Create the response structure with original value + translations
	result := make(map[string]string)

	// Always include the original value if we have translations
	if f.Original != "" {
		result["original"] = f.Original
	}

	// Add all translations
	for lang, value := range f.Values {
		result[lang] = value
	}

	return json.Marshal(result)
}

// UnmarshalJSON implements custom JSON unmarshaling for Field
func (f *Field) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first (simple case)
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		f.Original = str
		f.Values = make(map[string]string)
		return nil
	}

	// Try to unmarshal as map (translations case)
	var temp map[string]string
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	f.Original = temp["original"]
	f.Values = make(map[string]string)

	for key, value := range temp {
		if key != "original" {
			f.Values[key] = value
		}
	}

	return nil
}

// Value implements driver.Valuer for database storage
func (f Field) Value() (driver.Value, error) {
	// Store only the original value in the database
	return f.Original, nil
}

// Scan implements sql.Scanner for database reading
func (f *Field) Scan(value any) error {
	if value == nil {
		f.Original = ""
		f.Values = make(map[string]string)
		return nil
	}

	switch v := value.(type) {
	case string:
		f.Original = v
	case []byte:
		f.Original = string(v)
	default:
		return fmt.Errorf("cannot scan %T into Field", value)
	}

	f.Values = make(map[string]string)
	return nil
}

// String returns the original value for string operations
func (f Field) String() string {
	return f.Original
}

// SetOriginal sets the original value
func (f *Field) SetOriginal(value string) {
	f.Original = value
	if f.Values == nil {
		f.Values = make(map[string]string)
	}
}

// SetTranslation sets a translation for a specific language
func (f *Field) SetTranslation(language, value string) {
	if f.Values == nil {
		f.Values = make(map[string]string)
	}
	f.Values[language] = value
}

// GetTranslation gets a translation for a specific language
func (f Field) GetTranslation(language string) (string, bool) {
	if len(f.Values) == 0 {
		return "", false
	}
	value, exists := f.Values[language]
	return value, exists
}

// GetTranslationOrOriginal gets a translation for a specific language, falling back to original
func (f Field) GetTranslationOrOriginal(language string) string {
	if value, exists := f.GetTranslation(language); exists && value != "" {
		return value
	}
	return f.Original
}

// HasTranslation checks if a translation exists for a specific language
func (f Field) HasTranslation(language string) bool {
	_, exists := f.GetTranslation(language)
	return exists
}

// GetAvailableLanguages returns all languages that have translations for this field
func (f Field) GetAvailableLanguages() []string {
	if len(f.Values) == 0 {
		return []string{}
	}

	languages := make([]string, 0, len(f.Values))
	for lang := range f.Values {
		languages = append(languages, lang)
	}
	return languages
}

// NewField creates a new Field with an original value
func NewField(original string) Field {
	return Field{
		Original: original,
		Values:   make(map[string]string),
	}
}

// LoadTranslations loads translations from the database using the global translation service
func (f *Field) LoadTranslations(modelName string, modelId uint, fieldName string) error {
	// This would need to be implemented with a global service instance
	// For now, we'll implement this as a hook in the GORM callbacks
	return nil
}

// AutoLoadTranslations automatically loads translations if they haven't been loaded yet
func (f *Field) AutoLoadTranslations(modelName string, modelId uint, fieldName string) error {
	// Only load if translations are not already loaded
	if len(f.Values) == 0 {
		return f.LoadTranslations(modelName, modelId, fieldName)
	}
	return nil
}

// TableName returns the table name for the Translation model
func (item *Translation) TableName() string {
	return "translations"
}

// GetId returns the Id of the model
func (item *Translation) GetId() uint {
	return item.Id
}

// GetModelName returns the model name
func (item *Translation) GetModelName() string {
	return "translation"
}

// TranslationListResponse represents the list view response
type TranslationListResponse struct {
	Id        uint      `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Model     string    `json:"model"`
	ModelId   uint      `json:"model_id"`
	Language  string    `json:"language"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TranslationResponse represents the detailed view response
type TranslationResponse struct {
	Id        uint           `json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`
	Key       string         `json:"key"`
	Value     string         `json:"value"`
	Model     string         `json:"model"`
	ModelId   uint           `json:"model_id"`
	Language  string         `json:"language"`
}

// CreateTranslationRequest represents the request payload for creating a Translation
type CreateTranslationRequest struct {
	Key      string `json:"key" binding:"required"`
	Value    string `json:"value" binding:"required"`
	Model    string `json:"model" binding:"required"`
	ModelId  uint   `json:"model_id" binding:"required"`
	Language string `json:"language" binding:"required"`
}

// UpdateTranslationRequest represents the request payload for updating a Translation
type UpdateTranslationRequest struct {
	Id       uint   `json:"id" binding:"required"`
	Key      string `json:"key,omitempty"`
	Value    string `json:"value,omitempty"`
	Model    string `json:"model,omitempty"`
	ModelId  uint   `json:"model_id,omitempty"`
	Language string `json:"language,omitempty"`
}

// BulkTranslationRequest represents a request to update multiple translations at once
type BulkTranslationRequest struct {
	Model        string            `json:"model" binding:"required"`
	ModelId      uint              `json:"model_id" binding:"required"`
	Language     string            `json:"language" binding:"required"`
	Translations map[string]string `json:"translations" binding:"required"` // key -> value mapping
}

// ToListResponse converts the model to a list response
func (item *Translation) ToListResponse() *TranslationListResponse {
	if item == nil {
		return nil
	}
	return &TranslationListResponse{
		Id:        item.Id,
		Key:       item.Key,
		Value:     item.Value,
		Model:     item.Model,
		ModelId:   item.ModelId,
		Language:  item.Language,
		UpdatedAt: item.UpdatedAt,
	}
}

// ToResponse converts the model to a detailed response
func (item *Translation) ToResponse() *TranslationResponse {
	if item == nil {
		return nil
	}
	return &TranslationResponse{
		Id:        item.Id,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
		DeletedAt: item.DeletedAt,
		Key:       item.Key,
		Value:     item.Value,
		Model:     item.Model,
		ModelId:   item.ModelId,
		Language:  item.Language,
	}
}

// Preload preloads all the model's relationships
func (item *Translation) Preload(db *gorm.DB) *gorm.DB {
	query := db
	return query
}
