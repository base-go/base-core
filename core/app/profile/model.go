package profile

import (
	"base/core/app/authorization"
	"base/core/storage"
	"time"

	"gorm.io/gorm"
)

type User struct {
	Id        uint                `gorm:"column:id;primary_key;auto_increment"`
	FirstName string              `gorm:"column:first_name;not null;size:255"`
	LastName  string              `gorm:"column:last_name;not null;size:255"`
	Username  string              `gorm:"column:username;unique;not null;size:255"`
	Phone     string              `gorm:"column:phone;unique;size:255"`
	Email     string              `gorm:"column:email;unique;not null;size:255"`
	RoleId    uint                `gorm:"column:role_id;default:3"`
	Role      *authorization.Role `gorm:"foreignKey:RoleId"`
	Avatar    *storage.Attachment `gorm:"foreignKey:ModelId;references:Id"`
	Password  string              `gorm:"column:password;size:255"`
	LastLogin *time.Time          `gorm:"column:last_login"`
	CreatedAt time.Time           `gorm:"column:created_at"`
	UpdatedAt time.Time           `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt      `gorm:"column:deleted_at"`
}

func (User) TableName() string {
	return "users"
}

type CreateRequest struct {
	FirstName string `json:"first_name" binding:"required,max=255"`
	LastName  string `json:"last_name" binding:"required,max=255"`
	Username  string `json:"username" binding:"required,max=255"`
	Phone     string `json:"phone" binding:"max=255"`
	Email     string `json:"email" binding:"required,email,max=255"`
	Password  string `json:"password" binding:"required,min=8,max=255"`
}

type UpdateRequest struct {
	FirstName string `form:"first_name" binding:"max=255"`
	LastName  string `form:"last_name" binding:"max=255"`
	Username  string `form:"username" binding:"max=255"`
	Phone     string `form:"phone" binding:"max=255"`
	Email     string `form:"email" binding:"email,max=255"`
}

type UpdatePasswordRequest struct {
	OldPassword string `form:"OldPassword" binding:"required,max=255"`
	NewPassword string `form:"NewPassword" binding:"required,min=6,max=255"`
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
	Id        uint   `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	RoleId    uint   `json:"role_id"`
	RoleName  string `json:"role_name"`
	AvatarURL string `json:"avatar_url"`
	LastLogin string `json:"last_login"`
}

// AvatarResponse represents the avatar in API responses
type AvatarResponse struct {
	Id       uint   `json:"id"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
}

// ToResponse converts the User to a UserResponse
func (u *User) ToResponse() *UserResponse {
	if u == nil {
		return nil
	}
	response := &UserResponse{
		Id:        u.Id,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Username:  u.Username,
		Phone:     u.Phone,
		Email:     u.Email,
		RoleId:    u.RoleId,
	}

	// Include role name if role relationship is loaded
	if u.Role != nil {
		response.RoleName = u.Role.Name
	}

	if u.Avatar != nil {
		response.AvatarURL = u.Avatar.URL
	}

	if u.LastLogin != nil {
		response.LastLogin = u.LastLogin.Format(time.RFC3339)
	}

	return response
}

// UserModelResponse represents a simplified response when User is part of other entities
type UserModelResponse struct {
	Id        uint   `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
}

// ToModelResponse converts the model to a simplified response for when it's part of other entities
func (u *User) ToModelResponse() *UserModelResponse {
	if u == nil {
		return nil
	}
	return &UserModelResponse{
		Id:        u.Id,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Username:  u.Username,
	}
}

// Helper function to convert User to UserResponse (deprecated - use ToResponse method)
func ToResponse(user *User) *UserResponse {
	if user == nil {
		return nil
	}
	return user.ToResponse()
}
