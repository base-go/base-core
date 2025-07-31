package model

import (
	"base/core/translation"
	"time"
)

type Comment struct {
	ID          uint              `json:"id" gorm:"primarykey"`
	Content     translation.Field `json:"content" gorm:"type:text"`
	AuthorName  string            `json:"author_name" gorm:"size:100"`
	AuthorEmail string            `json:"author_email" gorm:"size:100"`
	AuthorURL   string            `json:"author_url" gorm:"size:255"`
	PostID      uint              `json:"post_id" gorm:"index"`
	Post        *Post             `json:"post,omitempty" gorm:"foreignKey:PostID"`
	ParentID    *uint             `json:"parent_id" gorm:"index"`
	Parent      *Comment          `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children    []Comment         `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Status      string            `json:"status" gorm:"size:20;default:'pending'"` // pending, approved, spam, trash
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func (Comment) TableName() string {
	return "comments"
}

type CreateCommentRequest struct {
	Content     translation.Field `json:"content" binding:"required"`
	AuthorName  string            `json:"author_name" binding:"required,min=1,max=100"`
	AuthorEmail string            `json:"author_email" binding:"required,email,max=100"`
	AuthorURL   string            `json:"author_url"`
	PostID      uint              `json:"post_id" binding:"required"`
	ParentID    *uint             `json:"parent_id"`
}

type UpdateCommentRequest struct {
	ID          uint              `json:"id" binding:"required"`
	Content     translation.Field `json:"content"`
	AuthorName  string            `json:"author_name"`
	AuthorEmail string            `json:"author_email"`
	AuthorURL   string            `json:"author_url"`
	Status      string            `json:"status"`
}

type CommentResponse struct {
	ID          uint               `json:"id"`
	Content     translation.Field  `json:"content"`
	AuthorName  string             `json:"author_name"`
	AuthorEmail string             `json:"author_email"`
	AuthorURL   string             `json:"author_url"`
	PostID      uint               `json:"post_id"`
	ParentID    *uint              `json:"parent_id"`
	Parent      *CommentResponse   `json:"parent,omitempty"`
	Children    []CommentResponse  `json:"children,omitempty"`
	Status      string             `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}