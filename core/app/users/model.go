package users

import (
	"base/core/storage"
	"time"

	"gorm.io/gorm"
)

type User struct {
	Id        uint                `gorm:"column:id;primary_key;auto_increment"`
	Name      string              `gorm:"column:name;not null"`
	Username  string              `gorm:"column:username;unique;not null"`
	Email     string              `gorm:"column:email;unique;not null"`
	Avatar    *storage.Attachment `gorm:"foreignKey:ModelId;references:Id"`
	Password  string              `gorm:"column:password"`
	CreatedAt time.Time           `gorm:"column:created_at"`
	UpdatedAt time.Time           `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt      `gorm:"column:deleted_at"`
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

// Implement the Attachable interface
func (u *User) GetId() uint {
	return u.Id
}

func (u *User) GetModelName() string {
	return "users"
}

// UserResponse represents the API response structure
type UserResponse struct {
	Id       uint            `json:"id"`
	Name     string          `json:"name"`
	Username string          `json:"username"`
	Email    string          `json:"email"`
	Avatar   *AvatarResponse `json:"avatar,omitempty"`
}

// AvatarResponse represents the avatar in API responses
type AvatarResponse struct {
	Id       uint   `json:"id"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
}

// Helper function to convert User to UserResponse
func ToResponse(user *User) *UserResponse {
	response := &UserResponse{
		Id:       user.Id,
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
	}

	if user.Avatar != nil {
		response.Avatar = &AvatarResponse{
			Id:       user.Avatar.Id,
			Filename: user.Avatar.Filename,
			URL:      user.Avatar.URL,
		}
	}

	return response
}
