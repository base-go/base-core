package model

import (
	"base/core/translation"
	"time"
)

type Post struct {
	ID          uint              `json:"id" gorm:"primarykey"`
	Title       translation.Field `json:"title" gorm:"type:text"`
	Slug        string            `json:"slug" gorm:"uniqueIndex;size:255"`
	Content     translation.Field `json:"content" gorm:"type:longtext"`
	Excerpt     translation.Field `json:"excerpt" gorm:"type:text"`
	Status      string            `json:"status" gorm:"size:20;default:'draft'"` // draft, published, archived
	FeaturedImage string          `json:"featured_image" gorm:"size:255"`
	CategoryID  *uint             `json:"category_id" gorm:"index"`
	Category    *Category         `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Tags        []Tag             `json:"tags,omitempty" gorm:"many2many:post_tags;"`
	Comments    []Comment         `json:"comments,omitempty" gorm:"foreignKey:PostID"`
	ViewCount   int               `json:"view_count" gorm:"default:0"`
	PublishedAt *time.Time        `json:"published_at"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func (Post) TableName() string {
	return "posts"
}

type CreatePostRequest struct {
	Title         translation.Field `json:"title" binding:"required"`
	Slug          string            `json:"slug" binding:"required,min=1,max=255"`
	Content       translation.Field `json:"content" binding:"required"`
	Excerpt       translation.Field `json:"excerpt"`
	Status        string            `json:"status"`
	FeaturedImage string            `json:"featured_image"`
	CategoryID    *uint             `json:"category_id"`
	TagIDs        []uint            `json:"tag_ids"`
}

type UpdatePostRequest struct {
	ID            uint              `json:"id" binding:"required"`
	Title         translation.Field `json:"title"`
	Slug          string            `json:"slug"`
	Content       translation.Field `json:"content"`
	Excerpt       translation.Field `json:"excerpt"`
	Status        string            `json:"status"`
	FeaturedImage string            `json:"featured_image"`
	CategoryID    *uint             `json:"category_id"`
	TagIDs        []uint            `json:"tag_ids"`
}

type PostResponse struct {
	ID            uint               `json:"id"`
	Title         translation.Field  `json:"title"`
	Slug          string             `json:"slug"`
	Content       translation.Field  `json:"content"`
	Excerpt       translation.Field  `json:"excerpt"`
	Status        string             `json:"status"`
	FeaturedImage string             `json:"featured_image"`
	CategoryID    *uint              `json:"category_id"`
	Category      *CategoryResponse  `json:"category,omitempty"`
	Tags          []TagResponse      `json:"tags,omitempty"`
	Comments      []CommentResponse  `json:"comments,omitempty"`
	ViewCount     int                `json:"view_count"`
	PublishedAt   *time.Time         `json:"published_at"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}