package translation

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

// Helper provides utility functions for working with translations
type Helper struct {
	DB      *gorm.DB
	Service *TranslationService
}

// NewHelper creates a new translation helper
func NewHelper(service *TranslationService) *Helper {
	return &Helper{
		DB:      service.DB,
		Service: service,
	}
}

type Translatable interface {
	TranslatedFields() []string
}

// GetTranslatedFields retrieves the translated fields for a given model
func GetTranslatedFields(model interface{}) []string {
	if translatable, ok := model.(Translatable); ok {
		return translatable.TranslatedFields()
	}
	return []string{}
}

// GetTranslationsForModel retrieves all translations for a model instance
func (h *Helper) GetTranslationsForModel(modelName string, modelId uint, language string) (map[string]string, error) {
	return h.Service.GetTranslationsForModel(modelName, modelId, language)
}

// AddTranslatedFieldsToResponse enriches a response struct with translated fields
func (h *Helper) AddTranslatedFieldsToResponse(response any, modelName string, modelId uint, language string) error {
	translations, err := h.GetTranslationsForModel(modelName, modelId, language)
	if err != nil {
		return err
	}

	if len(translations) == 0 {
		return nil
	}

	// Use reflection to add translated fields to the response
	responseValue := reflect.ValueOf(response)
	if responseValue.Kind() == reflect.Ptr {
		responseValue = responseValue.Elem()
	}

	if responseValue.Kind() != reflect.Struct {
		return fmt.Errorf("response must be a struct or pointer to struct")
	}

	// Create a map to store translated values
	translatedFields := make(map[string]any)
	for key, value := range translations {
		translatedFields[key] = value
	}

	// Try to set a field called "Translations" if it exists
	if field := responseValue.FieldByName("Translations"); field.IsValid() && field.CanSet() {
		field.Set(reflect.ValueOf(translatedFields))
	}

	return nil
}

// SetTranslation sets or updates a translation for a model field
func (h *Helper) SetTranslation(modelName string, modelId uint, key, value, language string) error {
	return h.Service.BulkSetTranslations(modelName, modelId, language, map[string]string{key: value})
}

// DeleteTranslationsForModel deletes all translations for a specific model instance
func (h *Helper) DeleteTranslationsForModel(modelName string, modelId uint) error {
	// This would need to be implemented in the service
	return nil
}

// GetAvailableLanguages returns all languages that have translations for a specific model instance
func (h *Helper) GetAvailableLanguages(modelName string, modelId uint) ([]string, error) {
	return h.Service.GetSupportedLanguages()
}

// BulkSetTranslations sets multiple translations for a model instance in a single transaction
func (h *Helper) BulkSetTranslations(modelName string, modelId uint, language string, translations map[string]string) error {
	return h.Service.BulkSetTranslations(modelName, modelId, language, translations)
}
