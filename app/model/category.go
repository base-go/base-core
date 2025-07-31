package model

import (
	"base/core/translation"
	"time"
)

type Category struct {
	ID          uint              `json:"id" gorm:"primarykey"`
	Name        translation.Field `json:"name" gorm:"type:text"`
	Slug        string            `json:"slug" gorm:"uniqueIndex;size:255"`
	Description translation.Field `json:"description" gorm:"type:text"`
	Color       string            `json:"color" gorm:"size:7"` // Hex color
	Icon        string            `json:"icon" gorm:"size:100"`
	ParentID    *uint             `json:"parent_id" gorm:"index"`
	Parent      *Category         `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children    []Category        `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	SortOrder   int               `json:"sort_order" gorm:"default:0"`
	IsActive    bool              `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func (Category) TableName() string {
	return "categories"
}

type CreateCategoryRequest struct {
	Name        translation.Field `json:"name" binding:"required"`
	Slug        string            `json:"slug" binding:"required,min=1,max=255"`
	Description translation.Field `json:"description"`
	Color       string            `json:"color"`
	Icon        string            `json:"icon"`
	ParentID    *uint             `json:"parent_id"`
	SortOrder   int               `json:"sort_order"`
	IsActive    bool              `json:"is_active"`
}

type UpdateCategoryRequest struct {
	ID          uint              `json:"id" binding:"required"`
	Name        translation.Field `json:"name"`
	Slug        string            `json:"slug"`
	Description translation.Field `json:"description"`
	Color       string            `json:"color"`
	Icon        string            `json:"icon"`
	ParentID    *uint             `json:"parent_id"`
	SortOrder   int               `json:"sort_order"`
	IsActive    *bool             `json:"is_active"`
}

type CategoryResponse struct {
	ID          uint               `json:"id"`
	Name        translation.Field  `json:"name"`
	Slug        string             `json:"slug"`
	Description translation.Field  `json:"description"`
	Color       string             `json:"color"`
	Icon        string             `json:"icon"`
	ParentID    *uint              `json:"parent_id"`
	Parent      *CategoryResponse  `json:"parent,omitempty"`
	Children    []CategoryResponse `json:"children,omitempty"`
	SortOrder   int                `json:"sort_order"`
	IsActive    bool               `json:"is_active"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type CategoryListResponse struct {
	ID          uint              `json:"id"`
	Name        translation.Field `json:"name"`
	Slug        string            `json:"slug"`
	Description translation.Field `json:"description"`
	Color       string            `json:"color"`
	Icon        string            `json:"icon"`
	ParentID    *uint             `json:"parent_id"`
	SortOrder   int               `json:"sort_order"`
	IsActive    bool              `json:"is_active"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
