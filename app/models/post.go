package models

import (
    "time"
    "gorm.io/gorm"
    "base/core/storage"
)

// Post represents a post entity
type Post struct {
    Id        uint           `json:"id" gorm:"primaryKey"`
    Title string `json:"title" gorm:"column:title"`
    Content string `json:"content" gorm:"column:content"`
    Image *storage.Attachment `json:"image,omitempty" gorm:"type:jsonb"`
    CategoryId uint `json:"category_id_id" gorm:"column:category_id_id"`
    Category *Category `json:"category_id,omitempty" gorm:"foreignKey:CategoryId;references:Id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for the Post model
func (Post) TableName() string {
    return "posts"
}

// Implement the storage.Attachable interface
func (item *Post) GetId() uint {
    return item.Id
}

func (item *Post) GetModelName() string {
    return "post"
}

// PostListResponse represents the list view response
type PostListResponse struct {
    Id        uint      `json:"id"`
    Title string `json:"title"`
    Content string `json:"content"`
    Image *storage.Attachment `json:"image,omitempty"`
    Category *Category `json:"category_id,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// PostResponse represents the detailed view response
type PostResponse struct {
    Id        uint      `json:"id"`
    Title string `json:"title"`
    Content string `json:"content"`
    Image *storage.Attachment `json:"image,omitempty"`
    CategoryId uint `json:"category_id_id"`
    Category *CategoryResponse `json:"category_id,omitempty"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`
}

// CreatePostRequest represents the create request
type CreatePostRequest struct {
    Title string `json:"title" binding:"required"`
    Content string `json:"content" binding:"required"`
    CategoryId uint `json:"category_id_id" binding:"required"`
}

// UpdatePostRequest represents the update request
type UpdatePostRequest struct {
    Title *string `json:"title,omitempty"`
    Content *string `json:"content,omitempty"`
    CategoryId *uint `json:"category_id_id,omitempty"`
}

// ToListResponse converts the model to a list response
func (item *Post) ToListResponse() *PostListResponse {
    if item == nil {
        return nil
    }
    return &PostListResponse{
        Id: item.Id,
        Title: item.Title,
        Content: item.Content,
        Image: item.Image,
        Category: item.Category,
        CreatedAt: item.CreatedAt,
        UpdatedAt: item.UpdatedAt,
    }
}

// ToResponse converts the model to a detailed response
func (item *Post) ToResponse() *PostResponse {
    if item == nil {
        return nil
    }
    return &PostResponse{
        Id: item.Id,
        Title: item.Title,
        Content: item.Content,
        Image: item.Image,
        CategoryId: item.CategoryId,
        Category: func() *CategoryResponse {
            if item.Category == nil {
                return nil
            }
            return item.Category.ToResponse()
        }(),
        CreatedAt: item.CreatedAt,
        UpdatedAt: item.UpdatedAt,
        DeletedAt: item.DeletedAt,
    }
}

func (item *Post) Preload(db *gorm.DB) *gorm.DB {
    db = db.Preload("Category")
    return db
}
