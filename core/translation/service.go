package translation

import (
	"base/core/emitter"
	"base/core/logger"
	"base/core/storage"
	"base/core/types"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TranslationService struct {
	DB      *gorm.DB
	Emitter *emitter.Emitter
	Storage *storage.ActiveStorage
	Logger  logger.Logger
}

func NewTranslationService(db *gorm.DB, emitter *emitter.Emitter, storage *storage.ActiveStorage, logger logger.Logger) *TranslationService {
	return &TranslationService{
		DB:      db,
		Emitter: emitter,
		Storage: storage,
		Logger:  logger,
	}
}

func (s *TranslationService) GetAll(page *int, limit *int, model string, modelId *uint) (*types.PaginatedResponse, error) {
	// Default values for pagination
	currentPage := 1
	pageSize := 10

	if page != nil {
		currentPage = *page
	}
	if limit != nil {
		pageSize = *limit
	}

	var translations []*Translation
	var total int64

	// Build query with filters
	query := s.DB.Model(&Translation{})
	if model != "" {
		s.Logger.Info("Filtering translations by model", zap.String("model", model))
		query = query.Where("model = ?", model)
	}
	if modelId != nil {
		s.Logger.Info("Filtering translations by model_id", zap.Uint("model_id", *modelId))
		query = query.Where("model_id = ?", *modelId)
	}

	// Count total records with filters
	if err := query.Count(&total).Error; err != nil {
		s.Logger.Error("Failed to count translations", zap.Error(err))
		return nil, err
	}

	// Calculate offset
	offset := (currentPage - 1) * pageSize

	// Get translations with pagination and filters
	if err := query.Offset(offset).Limit(pageSize).Order("updated_at DESC").Find(&translations).Error; err != nil {
		s.Logger.Error("Failed to fetch translations", zap.Error(err))
		return nil, err
	}

	// Convert to response format
	responses := make([]*TranslationListResponse, len(translations))
	for i, translation := range translations {
		responses[i] = translation.ToListResponse()
	}

	// Calculate total pages
	totalPages := int(total+int64(pageSize)-1) / pageSize

	return &types.PaginatedResponse{
		Data: responses,
		Pagination: types.Pagination{
			Total:      int(total),
			Page:       currentPage,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *TranslationService) GetByID(id uint) (*TranslationResponse, error) {
	var translation Translation
	if err := s.DB.First(&translation, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("translation not found")
		}
		s.Logger.Error("Failed to fetch translation", zap.Error(err))
		return nil, err
	}

	return translation.ToResponse(), nil
}

func (s *TranslationService) Create(request *CreateTranslationRequest) (*TranslationResponse, error) {
	// Check if translation already exists for this key, model, model_id, and language
	var existing Translation
	err := s.DB.Where("`key` = ? AND model = ? AND model_id = ? AND language = ?",
		request.Key, request.Model, request.ModelId, request.Language).First(&existing).Error

	if err == nil {
		return nil, errors.New("translation already exists for this key, model, model_id, and language combination")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.Logger.Error("Failed to check existing translation", zap.Error(err))
		return nil, err
	}

	translation := &Translation{
		Key:      request.Key,
		Value:    request.Value,
		Model:    request.Model,
		ModelId:  request.ModelId,
		Language: request.Language,
	}

	if err := s.DB.Create(translation).Error; err != nil {
		s.Logger.Error("Failed to create translation", zap.Error(err))
		return nil, err
	}

	s.Logger.Info("Translation created successfully", zap.Uint("id", translation.Id))
	return translation.ToResponse(), nil
}

func (s *TranslationService) Update(request *UpdateTranslationRequest) (*TranslationResponse, error) {
	var translation Translation
	if err := s.DB.First(&translation, request.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("translation not found")
		}
		s.Logger.Error("Failed to fetch translation", zap.Error(err))
		return nil, err
	}

	// Update fields if provided
	if request.Key != "" {
		translation.Key = request.Key
	}
	if request.Value != "" {
		translation.Value = request.Value
	}
	if request.Model != "" {
		translation.Model = request.Model
	}
	if request.ModelId != 0 {
		translation.ModelId = request.ModelId
	}
	if request.Language != "" {
		translation.Language = request.Language
	}

	if err := s.DB.Save(&translation).Error; err != nil {
		s.Logger.Error("Failed to update translation", zap.Error(err))
		return nil, err
	}

	s.Logger.Info("Translation updated successfully", zap.Uint("id", translation.Id))
	return translation.ToResponse(), nil
}

func (s *TranslationService) Delete(id uint) error {
	var translation Translation
	if err := s.DB.First(&translation, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("translation not found")
		}
		s.Logger.Error("Failed to fetch translation", zap.Error(err))
		return err
	}

	if err := s.DB.Delete(&translation).Error; err != nil {
		s.Logger.Error("Failed to delete translation", zap.Error(err))
		return err
	}

	s.Logger.Info("Translation deleted successfully", zap.Uint("id", id))
	return nil
}

func (s *TranslationService) GetTranslationsForModel(model string, modelId uint, language string) (map[string]string, error) {
	s.Logger.Info("Fetching translations for model", zap.String("model", model), zap.Uint("model_id", modelId), zap.String("language", language))

	var translations []Translation
	query := s.DB.Where("model = ? AND model_id = ?", model, modelId)

	if language != "" {
		query = query.Where("language = ?", language)
	}

	if err := query.Find(&translations).Error; err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, t := range translations {
		key := t.Key
		if language == "" {
			key = fmt.Sprintf("%s_%s", t.Key, t.Language)
		}
		result[key] = t.Value
	}

	return result, nil
}

// BulkUpdate updates multiple translations for a model at once
func (s *TranslationService) BulkUpdate(request *BulkTranslationRequest) error {
	s.Logger.Info("Starting bulk translation update",
		zap.String("model", request.Model),
		zap.Uint("model_id", request.ModelId),
		zap.String("language", request.Language),
		zap.Int("count", len(request.Translations)))

	err := s.BulkSetTranslations(request.Model, request.ModelId, request.Language, request.Translations)
	if err != nil {
		s.Logger.Error("Failed to bulk update translations", zap.Error(err))
		return err
	}

	s.Logger.Info("Bulk translation update completed successfully")
	return nil
}

// BulkSetTranslations sets multiple translations for a model instance in a single transaction
func (s *TranslationService) BulkSetTranslations(modelName string, modelId uint, language string, translations map[string]string) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for key, value := range translations {
		var translation Translation
		err := tx.Where("model = ? AND model_id = ? AND `key` = ? AND language = ?",
			modelName, modelId, key, language).First(&translation).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new translation
			translation = Translation{
				Model:    modelName,
				ModelId:  modelId,
				Key:      key,
				Value:    value,
				Language: language,
			}
			if err := tx.Create(&translation).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			// Update existing translation
			translation.Value = value
			if err := tx.Save(&translation).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

// GetSupportedLanguages returns a list of languages that have translations in the system
func (s *TranslationService) GetSupportedLanguages() ([]string, error) {
	s.Logger.Info("Fetching supported languages")
	var languages []string
	if err := s.DB.Model(&Translation{}).Distinct("language").Pluck("language", &languages).Error; err != nil {
		return nil, err
	}
	return languages, nil
}

// LoadTranslationsForField loads translations from the database for a specific field
func (s *TranslationService) LoadTranslationsForField(field *Field, modelName string, modelId uint, fieldName string) error {
	// Query translations for this specific field
	var translations []Translation
	err := s.DB.Where("model = ? AND model_id = ? AND `key` = ?", modelName, modelId, fieldName).Find(&translations).Error

	if err != nil {
		return err
	}

	// Initialize Values map if needed
	if field.Values == nil {
		field.Values = make(map[string]string)
	}

	// Load all translations
	for _, translation := range translations {
		field.Values[translation.Language] = translation.Value
	}

	return nil
}
