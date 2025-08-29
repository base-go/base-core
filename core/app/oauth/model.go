package oauth

import (
	"base/core/app/profile"
	"time"

	"gorm.io/gorm"
)

type OAuthUser struct {
	profile.User   `gorm:"embedded"`
	Provider       string    `gorm:"column:provider"`
	ProviderId     string    `gorm:"column:provider_id"`
	AccessToken    string    `gorm:"column:access_token"`
	OAuthLastLogin time.Time `gorm:"column:oauth_last_login"`
}

func (OAuthUser) TableName() string {
	return "users"
}

type AuthProvider struct {
	gorm.Model
	UserId      uint
	Provider    string
	ProviderId  string
	AccessToken string
	LastLogin   time.Time
}

func (AuthProvider) TableName() string {
	return "auth_providers"
}

// You might want to add OAuth-specific request/response structs here
type OAuthLoginRequest struct {
	Provider    string `json:"provider" binding:"required"`
	AccessToken string `json:"access_token" binding:"required"`
}

type OAuthResponse struct {
	AccessToken string `json:"accessToken"`
	Exp         int64  `json:"exp"`
	Username    string `json:"username"`
	Id          uint   `json:"id"`
	Avatar      string `json:"avatar"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	LastLogin   string `json:"last_login"`
	Provider    string `json:"provider"`
}
