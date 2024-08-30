package auth

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email            string     `json:"email" gorm:"uniqueIndex"`
	Password         string     `json:"-" gorm:"column:password"`
	FirstName        string     `json:"first_name" gorm:"column:first_name"`
	LastName         string     `json:"last_name" gorm:"column:last_name"`
	StripeID         string     `json:"stripe_id" gorm:"column:stripe_id"`
	ResetToken       string     `json:"-" gorm:"column:reset_token"`
	ResetTokenExpiry *time.Time `json:"-" gorm:"column:reset_token_expiry"`
	LastLogin        *time.Time `json:"last_login" gorm:"column:last_login"`
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type AuthResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
	UserID      uint      `json:"user_id"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	LastLogin   time.Time `json:"last_login"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
