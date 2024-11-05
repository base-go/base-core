package storage

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
)

// Attachment represents a file attachment for a model
type Attachment struct {
	Id        uint   `json:"id" gorm:"primaryKey"`
	ModelType string `json:"model_type" gorm:"index"`
	ModelId   uint   `json:"model_id" gorm:"index"`
	Field     string `json:"field" gorm:"index"`
	Filename  string `json:"filename"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	URL       string `json:"url"`
}

// AttachmentConfig defines configuration for file attachments
type AttachmentConfig struct {
	Field             string
	Path              string
	AllowedExtensions []string
	MaxFileSize       int64
	Multiple          bool
}

// Attachable interface should be implemented by models that can have attachments
type Attachable interface {
	GetId() uint
	GetModelName() string
}

// ActiveStorage handles file attachments for models
type ActiveStorage struct {
	provider    Provider
	defaultPath string
	maxFileSize int64
	allowedExts []string
	configs     map[string]AttachmentConfig
}

// NewActiveStorage creates a new ActiveStorage instance
func NewActiveStorage(config Config) (*ActiveStorage, error) {
	var provider Provider
	var err error

	switch strings.ToLower(config.Provider) {
	case "contabo":
		provider, err = NewContaboProvider(ContaboConfig{
			APIKey:    config.APIKey,
			APISecret: config.APISecret,
			Endpoint:  config.Endpoint,
			Bucket:    config.Bucket,
			BaseURL:   config.BaseURL,
		})
	case "local":
		provider, err = NewLocalProvider(LocalConfig{
			BasePath: config.Path,
			BaseURL:  config.BaseURL,
		})
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", config.Provider)
	}

	if err != nil {
		return nil, err
	}

	return &ActiveStorage{
		provider:    provider,
		defaultPath: "uploads",
		maxFileSize: 10 << 20, // 10MB default
		allowedExts: []string{".jpg", ".jpeg", ".png", ".gif", ".pdf"},
		configs:     make(map[string]AttachmentConfig),
	}, nil
}

// RegisterAttachment configures attachment settings for a model field
func (as *ActiveStorage) RegisterAttachment(modelName string, config AttachmentConfig) {
	key := fmt.Sprintf("%s.%s", modelName, config.Field)
	as.configs[key] = config
}

// Attach handles file attachment for a model
func (as *ActiveStorage) Attach(model Attachable, field string, file *multipart.FileHeader) (*Attachment, error) {
	config, err := as.getConfig(model.GetModelName(), field)
	if err != nil {
		return nil, err
	}

	if err := as.validateFile(file, config); err != nil {
		return nil, err
	}

	// Check if multiple attachments are allowed
	if !config.Multiple {
		if err := as.deleteExisting(model, field); err != nil {
			return nil, err
		}
	}

	// Upload file
	uploadPath := filepath.Join(config.Path, model.GetModelName(), field)
	result, err := as.provider.Upload(file, UploadConfig{
		AllowedExtensions: config.AllowedExtensions,
		MaxFileSize:       config.MaxFileSize,
		UploadPath:        uploadPath,
	})
	if err != nil {
		return nil, err
	}

	// Create attachment record
	attachment := &Attachment{
		ModelType: model.GetModelName(),
		ModelId:   model.GetId(),
		Field:     field,
		Filename:  result.Filename,
		Path:      result.Path,
		Size:      result.Size,
		URL:       as.provider.GetURL(result.Path),
	}

	return attachment, nil
}

// Delete removes an attachment
func (as *ActiveStorage) Delete(attachment *Attachment) error {
	if err := as.provider.Delete(attachment.Path); err != nil {
		return err
	}
	return nil
}

// GetAttachments retrieves all attachments for a model field
func (as *ActiveStorage) GetAttachments(model Attachable, field string) ([]*Attachment, error) {
	// This would typically involve a database query
	// Implementation depends on your database setup
	return nil, nil
}

// Helper functions
func (as *ActiveStorage) getConfig(modelName, field string) (AttachmentConfig, error) {
	key := fmt.Sprintf("%s.%s", modelName, field)
	config, exists := as.configs[key]
	if !exists {
		return AttachmentConfig{}, fmt.Errorf("no configuration found for %s", key)
	}
	return config, nil
}

func (as *ActiveStorage) validateFile(file *multipart.FileHeader, config AttachmentConfig) error {
	if file.Size > config.MaxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", config.MaxFileSize)
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !contains(config.AllowedExtensions, ext) {
		return fmt.Errorf("file extension %s is not allowed", ext)
	}

	return nil
}

func (as *ActiveStorage) deleteExisting(model Attachable, field string) error {
	attachments, err := as.GetAttachments(model, field)
	if err != nil {
		return err
	}

	for _, attachment := range attachments {
		if err := as.Delete(attachment); err != nil {
			return err
		}
	}

	return nil
}

// Config holds the configuration for ActiveStorage
type Config struct {
	Provider  string
	Path      string
	BaseURL   string
	APIKey    string
	APISecret string
	Endpoint  string
	Bucket    string
}

// Provider interface defines the storage provider contract
type Provider interface {
	Upload(file *multipart.FileHeader, config UploadConfig) (*UploadResult, error)
	Delete(path string) error
	GetURL(path string) string
}

// UploadConfig defines configuration for file uploads
type UploadConfig struct {
	AllowedExtensions []string
	MaxFileSize       int64
	UploadPath        string
}

// UploadResult represents the result of a file upload
type UploadResult struct {
	Filename string
	Path     string
	Size     int64
}

func contains(arr []string, val string) bool {
	for _, item := range arr {
		if item == val {
			return true
		}
	}
	return false
}
