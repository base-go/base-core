package models

import (
    "time"
    "gorm.io/gorm"
    "base/core/storage"
)

// Category represents a category entity
type Category struct {
    Id        uint           `json:"id" gorm:"primaryKey"`
    Title string `json:"title" gorm:"column:title"`
    Content string `json:"content" gorm:"column:content"`
    Image *storage.Attachment `json:"image,omitempty" gorm:"type:jsonb"`
    Posts []*Post `json:"posts,omitempty" gorm:"foreignKey:CategoryId"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for the Category model
func (Category) TableName() string {
    return "categories"
}

// Implement the storage.Attachable interface
func (item *Category) GetId() uint {
    return item.Id
}

func (item *Category) GetModelName() string {
    return "category"
}

// CategoryListResponse represents the list view response
type CategoryListResponse struct {
    Id        uint      `json:"id"`
    Title string `json:"title"`
    Content string `json:"content"`
    Image *storage.Attachment `json:"image,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// CategoryResponse represents the detailed view response
type CategoryResponse struct {
    Id        uint      `json:"id"`
    Title string `json:"title"`
    Content string `json:"content"`
    Image *storage.Attachment `json:"image,omitempty"`
    Posts []*PostResponse `json:"posts,omitempty"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`
}

// CreateCategoryRequest represents the create request
type CreateCategoryRequest struct {
    Title string `json:"title" binding:"required"`
    Content string `json:"content" binding:"required"`
}

// UpdateCategoryRequest represents the update request
type UpdateCategoryRequest struct {
    Title *string `json:"title,omitempty"`
    Content *string `json:"content,omitempty"`
}

// ToListResponse converts the model to a list response
func (item *Category) ToListResponse() *CategoryListResponse {
    if item == nil {
        return nil
    }
    return &CategoryListResponse{
        Id: item.Id,
        Title: item.Title,
        Content: item.Content,
        Image: item.Image,
        CreatedAt: item.CreatedAt,
        UpdatedAt: item.UpdatedAt,
    }
}

// ToResponse converts the model to a detailed response
func (item *Category) ToResponse() *CategoryResponse {
    if item == nil {
        return nil
    }
    return &CategoryResponse{
        Id: item.Id,
        Title: item.Title,
        Content: item.Content,
        Image: item.Image,
        Posts: func() []*PostResponse {
            if item.Posts == nil {
                return nil
            }
            responses := make([]*PostResponse, len(item.Posts))
            for i, v := range item.Posts {
                responses[i] = v.ToResponse()
            }
            return responses
        }(),
        CreatedAt: item.CreatedAt,
        UpdatedAt: item.UpdatedAt,
        DeletedAt: item.DeletedAt,
    }
}

func (item *Category) Preload(db *gorm.DB) *gorm.DB {
    db = db.Preload("Posts")
    return db
}
