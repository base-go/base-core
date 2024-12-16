package storage

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Constants and Patterns
var (
	illegalCharsPattern = regexp.MustCompile(`[^a-zA-Z0-9\-\.]`)
	multiDashPattern    = regexp.MustCompile(`-+`)
)

// Core Types
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

type AttachmentConfig struct {
	Field             string
	Path              string
	AllowedExtensions []string
	MaxFileSize       int64
	Multiple          bool
}

type Config struct {
	Provider  string
	Path      string
	BaseURL   string
	APIKey    string
	APISecret string
	Endpoint  string
	Bucket    string
	CDN       string
}

// Interfaces
type Attachable interface {
	GetId() uint
	GetModelName() string
}

type Provider interface {
	Upload(file *multipart.FileHeader, config UploadConfig) (*UploadResult, error)
	Delete(path string) error
	GetURL(path string) string
}

// Upload Types
type UploadConfig struct {
	AllowedExtensions []string
	MaxFileSize       int64
	UploadPath        string
}

type UploadResult struct {
	Filename string
	Path     string
	Size     int64
}

// ActiveStorage Implementation
type ActiveStorage struct {
	db          *gorm.DB
	provider    Provider
	defaultPath string
	maxFileSize int64
	allowedExts []string
	configs     map[string]AttachmentConfig
}

func NewActiveStorage(db *gorm.DB, config Config) (*ActiveStorage, error) {
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
	case "r2":
		provider, err = NewR2Provider(R2Config{
			AccessKeyID:     config.APIKey,
			AccessKeySecret: config.APISecret,
			AccountID:       config.Endpoint,
			Bucket:          config.Bucket,
			BaseURL:         config.BaseURL,
			CDN:             config.CDN,
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

	if err := db.AutoMigrate(&Attachment{}); err != nil {
		return nil, fmt.Errorf("failed to migrate attachments table: %w", err)
	}

	return &ActiveStorage{
		db:          db,
		provider:    provider,
		defaultPath: "uploads",
		maxFileSize: 10 << 20,
		allowedExts: []string{".jpg", ".jpeg", ".png", ".gif", ".pdf"},
		configs:     make(map[string]AttachmentConfig),
	}, nil
}

func (as *ActiveStorage) RegisterAttachment(modelName string, config AttachmentConfig) {
	key := fmt.Sprintf("%s.%s", modelName, config.Field)
	as.configs[key] = config
}

// Simplified Attach method
func (as *ActiveStorage) Attach(model Attachable, field string, file *multipart.FileHeader) (*Attachment, error) {
	config, err := as.getConfig(model.GetModelName(), field)
	if err != nil {
		return nil, err
	}

	if err := as.validateFile(file, config); err != nil {
		return nil, err
	}

	// First, delete any existing attachments
	var existingAttachments []Attachment
	if err := as.db.Where(&Attachment{
		ModelType: model.GetModelName(),
		ModelId:   model.GetId(),
		Field:     field,
	}).Find(&existingAttachments).Error; err != nil {
		return nil, fmt.Errorf("failed to query existing attachments: %w", err)
	}

	// Delete existing attachments
	for _, existing := range existingAttachments {
		// Delete from storage
		_ = as.provider.Delete(existing.Path) // Ignore storage delete errors
		// Delete from database
		if err := as.db.Unscoped().Delete(&existing).Error; err != nil {
			return nil, fmt.Errorf("failed to delete existing attachment: %w", err)
		}
	}

	// Upload new file
	uploadPath := filepath.Join(config.Path, model.GetModelName(), field)
	result, err := as.provider.Upload(file, UploadConfig{
		AllowedExtensions: config.AllowedExtensions,
		MaxFileSize:       config.MaxFileSize,
		UploadPath:        uploadPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Create new attachment record
	attachment := &Attachment{
		ModelType: model.GetModelName(),
		ModelId:   model.GetId(),
		Field:     field,
		Filename:  result.Filename,
		Path:      result.Path,
		Size:      result.Size,
		URL:       as.provider.GetURL(result.Path),
	}

	if err := as.db.Create(attachment).Error; err != nil {
		// If we fail to create the record, clean up the uploaded file
		_ = as.provider.Delete(result.Path)
		return nil, fmt.Errorf("failed to save attachment: %w", err)
	}

	return attachment, nil
}

func (as *ActiveStorage) Delete(attachment *Attachment) error {
	if err := as.provider.Delete(attachment.Path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	if err := as.db.Delete(attachment).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	return nil
}

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

func slugify(s string) string {
	s = strings.ToLower(s)
	ext := filepath.Ext(s)
	name := strings.TrimSuffix(s, ext)
	name = illegalCharsPattern.ReplaceAllString(name, "-")
	name = multiDashPattern.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	return name + ext
}

func generateUniqueFilename(originalName string) string {
	sluggedName := slugify(originalName)
	ext := filepath.Ext(sluggedName)
	name := sluggedName[:len(sluggedName)-len(ext)]
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
	return fmt.Sprintf("%s-%s%s", name, timestamp, ext)
}

func contains(arr []string, val string) bool {
	for _, item := range arr {
		if item == val {
			return true
		}
	}
	return false
}
