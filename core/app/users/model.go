package users

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	Id        uint           `gorm:"column:id;primary_key;auto_increment"`
	Name      string         `gorm:"column:name;not null"`
	Username  string         `gorm:"column:username;unique;not null"`
	Email     string         `gorm:"column:email;unique;not null"`
	Avatar    string         `gorm:"column:avatar"`
	Password  string         `gorm:"column:password"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (User) TableName() string {
	return "users"
}

type CreateRequest struct {
	Name     string `json:"name" binding:"required"`
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type UpdateRequest struct {
	Name     string `form:"name"`
	Username string `form:"username"`
	Email    string `form:"email"`
}

type UpdatePasswordRequest struct {
	OldPassword string `form:"OldPassword" binding:"required"`
	NewPassword string `form:"NewPassword" binding:"required,min=6"`
}

type UserResponse struct {
	Id       uint   `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}
