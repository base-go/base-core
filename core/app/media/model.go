package media

import (
	"mime/multipart"
	"time"

	"base/core/storage"

	"gorm.io/gorm"
)

// Media represents a media entity
type Media struct {
	Id          uint                `json:"id" gorm:"primaryKey"`
	Name        string              `json:"name" gorm:"column:name"`
	Type        string              `json:"type" gorm:"column:type"`
	Description string              `json:"description" gorm:"column:description"`
	File        *storage.Attachment `json:"file,omitempty" gorm:"polymorphic:Model"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	DeletedAt   gorm.DeletedAt      `json:"deleted_at" gorm:"index"`
}

// TableName returns the table name for the Media model
func (item *Media) TableName() string {
	return "media"
}

// GetId returns the Id of the model
func (item *Media) GetId() uint {
	return item.Id
}

// GetModelName returns the model name
func (item *Media) GetModelName() string {
	return "media"
}

// Preload preloads all the model's relationships
func (item *Media) Preload(db *gorm.DB) *gorm.DB {
	return db.Preload("File")
}

// MediaListResponse represents the list view response
type MediaListResponse struct {
	Id          uint                `json:"id"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Name        string              `json:"name"`
	Type        string              `json:"type"`
	Description string              `json:"description"`
	File        *storage.Attachment `json:"file,omitempty"`
}

// MediaResponse represents the detailed view response
type MediaResponse struct {
	Id          uint                `json:"id"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	DeletedAt   gorm.DeletedAt      `json:"deleted_at,omitempty"`
	Name        string              `json:"name"`
	Type        string              `json:"type"`
	Description string              `json:"description"`
	File        *storage.Attachment `json:"file,omitempty"`
}

// MediaResponse represents the detailed view response
type MediaModelResponse struct {
	Id          uint                `json:"id"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	DeletedAt   gorm.DeletedAt      `json:"deleted_at,omitempty"`
	Name        string              `json:"name"`
	Type        string              `json:"type"`
	Description string              `json:"description"`
	File        *storage.Attachment `json:"file,omitempty"`
}

// CreateMediaRequest represents the request payload for creating a Media
type CreateMediaRequest struct {
	Name        string                `form:"name" binding:"required"`
	Type        string                `form:"type" binding:"required"`
	Description string                `form:"description"`
	File        *multipart.FileHeader `form:"file"`
}

// UpdateMediaRequest represents the request payload for updating a Media
type UpdateMediaRequest struct {
	Name        *string               `form:"name"`
	Type        *string               `form:"type"`
	Description *string               `form:"description"`
	File        *multipart.FileHeader `form:"file"`
}

// ToListResponse converts the model to a list response
func (item *Media) ToListResponse() *MediaListResponse {
	return &MediaListResponse{
		Id:          item.Id,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
		Name:        item.Name,
		Type:        item.Type,
		Description: item.Description,
		File:        item.File,
	}
}

// ToResponse converts the model to a detailed response
func (item *Media) ToResponse() *MediaResponse {
	return &MediaResponse{
		Id:          item.Id,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
		DeletedAt:   item.DeletedAt,
		Name:        item.Name,
		Type:        item.Type,
		Description: item.Description,
		File:        item.File,
	}
}

// ToResponse converts the model to a detailed response
func (item *Media) ToModelResponse() *MediaModelResponse {
	return &MediaModelResponse{
		Id:          item.Id,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
		DeletedAt:   item.DeletedAt,
		Name:        item.Name,
		Type:        item.Type,
		Description: item.Description,
		File:        item.File,
	}
}

var _ storage.Attachable = (*Media)(nil)

// GetAttachmentConfig returns the attachment configuration for the model
func (item *Media) GetAttachmentConfig() map[string]any {
	return map[string]any{
		"file": map[string]any{
			"path":       "media/:id/:filename",
			"validators": []string{"image", "audio"},
			"min_size":   1,                 // 1 byte
			"max_size":   100 * 1024 * 1024, // 100MB
		},
	}
}
