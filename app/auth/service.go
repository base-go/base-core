package auth

import (
	"errors"
	"time"

	"base/core/helper"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	DB *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		DB: db,
	}
}
func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if err := s.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	token, err := helper.GenerateJWT(user.ID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user.LastLogin = &now
	s.DB.Save(&user)

	return &AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   24 * 60 * 60, // 24 hours in seconds
		UserID:      user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		LastLogin:   now,
	}, nil
}

func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	var user User
	if err := s.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, err
	}

	token, err := helper.GenerateJWT(user.ID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user.LastLogin = &now
	s.DB.Save(&user)

	return &AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   24 * 60 * 60, // 24 hours in seconds
		UserID:      user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		LastLogin:   now,
	}, nil
}

func (s *AuthService) ForgotPassword(email string) error {
	var user User
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return err
	}

	token := generateResetToken()
	expiry := time.Now().Add(15 * time.Minute)

	user.ResetToken = token
	user.ResetTokenExpiry = &expiry
	if err := s.DB.Save(&user).Error; err != nil {
		return err
	}

	// TODO: Send reset token to user's email
	return nil
}

func (s *AuthService) ResetPassword(email, token, newPassword string) error {
	var user User
	if err := s.DB.Where("email = ? AND reset_token = ?", email, token).First(&user).Error; err != nil {
		return err
	}

	if user.ResetTokenExpiry == nil || time.Now().After(*user.ResetTokenExpiry) {
		return errors.New("reset token expired")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	user.ResetToken = ""
	user.ResetTokenExpiry = nil
	return s.DB.Save(&user).Error
}

func generateResetToken() string {
	// TODO: Implement a secure method to generate reset tokens
	return "reset-token"
}
