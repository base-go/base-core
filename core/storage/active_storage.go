package storage

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

func NewActiveStorage(db *gorm.DB, config Config) (*ActiveStorage, error) {
	var provider Provider
	var err error

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	// If path is relative, make it absolute using cwd
	storagePath := config.Path
	if !filepath.IsAbs(storagePath) {
		storagePath = filepath.Join(cwd, storagePath)
	}

	switch strings.ToLower(config.Provider) {
	case "local":
		provider, err = NewLocalProvider(LocalConfig{
			BasePath: storagePath,
			BaseURL:  config.BaseURL,
		})
	case "s3":
		provider, err = NewS3Provider(S3Config{
			APIKey:          config.APIKey,
			APISecret:       config.APISecret,
			AccessKeyID:     config.APIKey,
			AccessKeySecret: config.APISecret,
			AccountID:       config.AccountID,
			Endpoint:        config.Endpoint,
			Bucket:          config.Bucket,
			BaseURL:         config.BaseURL,
			Region:          config.Region,
		})
	case "r2":
		provider, err = NewR2Provider(R2Config{
			AccessKeyID:     config.APIKey,
			AccessKeySecret: config.APISecret,
			AccountID:       config.AccountID,
			Bucket:          config.Bucket,
			BaseURL:         config.BaseURL,
			CDN:             config.CDN,
		})
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", config.Provider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage provider: %w", err)
	}

	as := &ActiveStorage{
		db:          db,
		provider:    provider,
		defaultPath: storagePath,
		configs:     make(map[string]map[string]AttachmentConfig),
	}

	// Auto-migrate the Attachment model
	if err := db.AutoMigrate(&Attachment{}); err != nil {
		return nil, fmt.Errorf("failed to migrate attachments table: %w", err)
	}

	return as, nil
}

func (as *ActiveStorage) RegisterAttachment(modelName string, config AttachmentConfig) {
	if as.configs[modelName] == nil {
		as.configs[modelName] = make(map[string]AttachmentConfig)
	}
	as.configs[modelName][config.Field] = config
}

func (as *ActiveStorage) Attach(model Attachable, field string, file *multipart.FileHeader) (*Attachment, error) {
	// Get config for model
	config, err := as.getConfig(model.GetModelName(), field)
	if err != nil {
		return nil, err
	}

	// Validate file
	if err := as.validateFile(file, config); err != nil {
		return nil, err
	}

	// Create attachment record
	attachment := &Attachment{
		ModelType: model.GetModelName(),
		ModelId:   model.GetId(),
		Field:     field,
		Filename:  file.Filename,
		Size:      file.Size,
	}

	// Upload file using provider
	result, err := as.provider.Upload(file, UploadConfig{
		AllowedExtensions: config.AllowedExtensions,
		MaxFileSize:       config.MaxFileSize,
		UploadPath:        filepath.Join(config.Path, model.GetModelName(), field),
	})
	if err != nil {
		return nil, err
	}

	// Update attachment with upload result
	attachment.Path = result.Path
	attachment.URL = as.provider.GetURL(result.Path)

	// Save attachment record
	if err := as.db.Create(attachment).Error; err != nil {
		// Try to delete uploaded file if record creation fails
		_ = as.provider.Delete(result.Path)
		return nil, err
	}

	return attachment, nil
}

func (as *ActiveStorage) Delete(attachment *Attachment) error {
	if err := as.provider.Delete(attachment.Path); err != nil {
		return err
	}
	return as.db.Delete(attachment).Error
}

func (as *ActiveStorage) getConfig(modelName, field string) (AttachmentConfig, error) {
	modelConfigs, ok := as.configs[modelName]
	if !ok {
		return AttachmentConfig{}, fmt.Errorf("no attachment config found for model %s", modelName)
	}

	config, ok := modelConfigs[field]
	if !ok {
		return AttachmentConfig{}, fmt.Errorf("no attachment config found for field %s in model %s", field, modelName)
	}

	return config, nil
}

func (as *ActiveStorage) validateFile(file *multipart.FileHeader, config AttachmentConfig) error {
	if file.Size > config.MaxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", config.MaxFileSize)
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if len(config.AllowedExtensions) > 0 && !strings.Contains(strings.Join(config.AllowedExtensions, ","), ext) {
		return fmt.Errorf("file extension %s is not allowed", ext)
	}

	return nil
}
