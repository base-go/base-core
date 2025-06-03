package models

import (
	"time"

	"gorm.io/gorm"
)

// Post represents a post entity
type Post struct {
	Id        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	Title     string         `json:"title"`
	Content   string         `json:"content"`
	Status    int            `json:"status"`
}

// TableName returns the table name for the Post model
func (item *Post) TableName() string {
	return "posts"
}

// GetId returns the Id of the model
func (item *Post) GetId() uint {
	return item.Id
}

// GetModelName returns the model name
func (item *Post) GetModelName() string {
	return "post"
}

// PostListResponse represents the list view response
type PostListResponse struct {
	Id        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    int       `json:"status"`
}

// PostResponse represents the detailed view response
type PostResponse struct {
	Id        uint           `json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`
	Title     string         `json:"title"`
	Content   string         `json:"content"`
	Status    int            `json:"status"`
}

// CreatePostRequest represents the request payload for creating a Post
type CreatePostRequest struct {
	Title   string `json:"title" form:"title" binding:"required"`
	Content string `json:"content" form:"content" binding:"required"`
	Status  int    `json:"status" form:"status" binding:"required"`
}

// UpdatePostRequest represents the request payload for updating a Post
type UpdatePostRequest struct {
	Title   string `json:"title,omitempty" form:"title" binding:"required"`
	Content string `json:"content,omitempty" form:"content" binding:"required"`
	Status  int    `json:"status,omitempty" form:"status" binding:"required"`
}

// ToListResponse converts the model to a list response
func (item *Post) ToListResponse() *PostListResponse {
	if item == nil {
		return nil
	}
	return &PostListResponse{
		Id:        item.Id,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
		Title:     item.Title,
		Content:   item.Content,
		Status:    item.Status,
	}
}

// ToResponse converts the model to a detailed response
func (item *Post) ToResponse() *PostResponse {
	if item == nil {
		return nil
	}
	return &PostResponse{
		Id:        item.Id,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
		DeletedAt: item.DeletedAt,
		Title:     item.Title,
		Content:   item.Content,
		Status:    item.Status,
	}
}

// Preload preloads all the model's relationships
func (item *Post) Preload(db *gorm.DB) *gorm.DB {
	query := db
	return query
}
