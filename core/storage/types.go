package storage

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"os"

	"time"

	"gorm.io/gorm"
)

// Attachment represents a file attachment
type Attachment struct {
	Id        uint      `json:"id" gorm:"primaryKey"`
	ModelType string    `json:"model_type" gorm:"index"`
	ModelId   uint      `json:"model_id" gorm:"index"`
	Field     string    `json:"field" gorm:"index"`
	Filename  string    `json:"filename"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Value implements the driver.Valuer interface
func (a *Attachment) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface
func (a *Attachment) Scan(value any) error {
	if value == nil {
		*a = Attachment{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	return json.Unmarshal(bytes, &a)
}

// AsFileHeader converts an Attachment to a multipart.FileHeader
func (a *Attachment) AsFileHeader() (*multipart.FileHeader, error) {
	file, err := os.Open(a.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return &multipart.FileHeader{
		Filename: a.Filename,
		Size:     a.Size,
		Header:   textproto.MIMEHeader{"Content-Type": []string{"application/octet-stream"}},
	}, nil
}

// AttachmentConfig holds configuration for file attachments
type AttachmentConfig struct {
	Field             string
	Path              string
	AllowedExtensions []string
	MaxFileSize       int64
	Multiple          bool
}

// Config holds storage service configuration
type Config struct {
	Provider  string
	Path      string
	BaseURL   string
	APIKey    string
	APISecret string
	AccountID string
	Endpoint  string
	Bucket    string
	CDN       string
	Region    string
}

// Attachable interface for models that can have attachments
type Attachable interface {
	GetId() uint
	GetModelName() string
}

// Provider interface for storage providers
type Provider interface {
	Upload(file *multipart.FileHeader, config UploadConfig) (*UploadResult, error)
	Delete(path string) error
	GetURL(path string) string
}

// ActiveStorage handles file storage operations
type ActiveStorage struct {
	db          *gorm.DB
	provider    Provider
	defaultPath string
	configs     map[string]map[string]AttachmentConfig
}

// UploadConfig holds configuration for file uploads
type UploadConfig struct {
	AllowedExtensions []string
	MaxFileSize       int64
	UploadPath        string
}

// UploadResult holds the result of a file upload
type UploadResult struct {
	Filename string
	Path     string
	Size     int64
}
