package model

import (
	"base/core/translation"
	"time"
)

type Tag struct {
	ID          uint              `json:"id" gorm:"primarykey"`
	Name        translation.Field `json:"name" gorm:"type:text"`
	Slug        string            `json:"slug" gorm:"uniqueIndex;size:255"`
	Description translation.Field `json:"description" gorm:"type:text"`
	Color       string            `json:"color" gorm:"size:7"` // Hex color
	Posts       []Post            `json:"posts,omitempty" gorm:"many2many:post_tags;"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func (Tag) TableName() string {
	return "tags"
}

type CreateTagRequest struct {
	Name        translation.Field `json:"name" binding:"required"`
	Slug        string            `json:"slug" binding:"required,min=1,max=255"`
	Description translation.Field `json:"description"`
	Color       string            `json:"color"`
}

type UpdateTagRequest struct {
	ID          uint              `json:"id" binding:"required"`
	Name        translation.Field `json:"name"`
	Slug        string            `json:"slug"`
	Description translation.Field `json:"description"`
	Color       string            `json:"color"`
}

type TagResponse struct {
	ID          uint              `json:"id"`
	Name        translation.Field `json:"name"`
	Slug        string            `json:"slug"`
	Description translation.Field `json:"description"`
	Color       string            `json:"color"`
	PostCount   int               `json:"post_count,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}