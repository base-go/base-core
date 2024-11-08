package auth

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"text/template"
	"time"

	"base/core/app/users"
	"base/core/email"
	"base/core/emitter"
	"base/core/helper"

	"github.com/stripe/stripe-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"pgregory.net/rand"
)

var (
	ErrInvalidToken = errors.New("invalid reset token")
	ErrTokenExpired = errors.New("reset token expired")
	ErrUserExists   = errors.New("user already exists")
	ErrInvalidTime  = errors.New("invalid time value")

	emailTemplateMutex sync.RWMutex
	emailTemplateCache *template.Template
)

type AuthService struct {
	DB          *gorm.DB
	EmailSender email.Sender
	Emitter     *emitter.Emitter
}

func NewAuthService(db *gorm.DB, emailSender email.Sender, emitter *emitter.Emitter) *AuthService {
	return &AuthService{
		DB:          db,
		EmailSender: emailSender,
		Emitter:     emitter,
	}
}

// validateUser checks if username or email already exists
func (s *AuthService) validateUser(email, username string) error {
	var count int64
	if err := s.DB.Model(&AuthUser{}).
		Where("email = ? OR username = ?", email, username).
		Count(&count).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	if count > 0 {
		return ErrUserExists
	}
	return nil
}

func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	// Validate unique constraints first
	if err := s.validateUser(req.Email, req.Username); err != nil {
		return nil, err
	}

	// Set Stripe API key
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	user := AuthUser{
		User: users.User{
			Email:    req.Email,
			Password: string(hashedPassword),
			Name:     req.Name,
			Username: req.Username,
		},
		LastLogin: &now,
	}

	// Start transaction
	tx := s.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrUserExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Generate JWT token
	token, err := helper.GenerateJWT(user.User.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	// Emit registration event
	s.Emitter.Emit("user.registered", &user.User)
	// Send welcome email asynchronously
	go func() {
		if err := s.sendWelcomeEmail(&user); err != nil {
			fmt.Printf("Failed to send welcome email: %v", err)
		}
	}()

	return &AuthResponse{
		AccessToken: token,
		Exp:         now.Add(24 * time.Hour).Unix(),
		Username:    user.Username,
		ID:          user.User.Id,
		Avatar:      user.Avatar,
		Email:       user.Email,
		Name:        user.Name,

		LastLogin: now.Format(time.RFC3339),
	}, nil
}

func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	var user AuthUser
	if err := s.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidToken
	}

	now := time.Now()
	token, err := helper.GenerateJWT(user.User.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Update last login with proper time handling
	if err := s.DB.Model(&user).Update("last_login", sql.NullTime{
		Time:  now,
		Valid: true,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}

	return &AuthResponse{
		AccessToken: token,
		Exp:         now.Add(24 * time.Hour).Unix(),
		Username:    user.Username,
		ID:          user.User.Id,
		Avatar:      user.Avatar,
		Email:       user.Email,
		Name:        user.Name,

		LastLogin: now.Format(time.RFC3339),
	}, nil
}

func (s *AuthService) ForgotPassword(email string) error {
	var user AuthUser
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found: %w", err)
		}
		return fmt.Errorf("database error: %w", err)
	}

	token := generateToken(6)
	expiry := time.Now().Add(15 * time.Minute)

	// Update reset token fields in transaction
	tx := s.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	updates := map[string]interface{}{
		"reset_token":        token,
		"reset_token_expiry": sql.NullTime{Time: expiry, Valid: true},
	}

	if err := tx.Model(&user).Updates(updates).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if err := s.sendPasswordResetEmail(&user, token); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

func (s *AuthService) ResetPassword(email, token, newPassword string) error {
	var user AuthUser
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found: %w", err)
		}
		return fmt.Errorf("database error: %w", err)
	}

	if user.ResetToken != token {
		return ErrInvalidToken
	}

	if user.ResetTokenExpiry == nil || time.Now().After(*user.ResetTokenExpiry) {
		return ErrTokenExpired
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password and clear reset token in transaction
	tx := s.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	updates := map[string]interface{}{
		"password":           string(hashedPassword),
		"reset_token":        "",
		"reset_token_expiry": nil,
	}

	if err := tx.Model(&user).Updates(updates).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update password: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Send confirmation email asynchronously
	go func() {
		if err := s.sendPasswordChangedEmail(&user); err != nil {
			fmt.Printf("Failed to send password changed email: %v\n", err)
		}
	}()

	return nil
}

func generateToken(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// Email sending functions
func (s *AuthService) sendEmail(to, subject, title, content string) error {
	var cachedTemplate *template.Template
	emailTemplateMutex.RLock()
	cachedTemplate = emailTemplateCache
	emailTemplateMutex.RUnlock()

	if cachedTemplate == nil {
		newTemplate, err := template.New("email").Parse(emailTemplate)
		if err != nil {
			return fmt.Errorf("error parsing email template: %w", err)
		}

		emailTemplateMutex.Lock()
		emailTemplateCache = newTemplate
		emailTemplateMutex.Unlock()

		cachedTemplate = newTemplate
	}

	var body bytes.Buffer
	err := cachedTemplate.Execute(&body, map[string]interface{}{
		"Title":   title,
		"Content": content,
		"Year":    time.Now().Year(),
	})
	if err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	msg := email.Message{
		To:      []string{to},
		From:    "support@albafone.app",
		Subject: subject,
		Body:    body.String(),
		IsHTML:  true,
	}
	return s.EmailSender.Send(msg)
}

func (s *AuthService) sendWelcomeEmail(user *AuthUser) error {
	title := "Welcome to Albafone"
	content := fmt.Sprintf("<p>Hi %s,</p><p>Thank you for registering with Albafone.</p>", user.Name)
	return s.sendEmail(user.Email, title, title, content)
}

func (s *AuthService) sendPasswordResetEmail(user *AuthUser, token string) error {
	title := "Reset Your Albafone Password"
	content := fmt.Sprintf(`
		<p>Hi %s,</p>
		<p>You have requested to reset your password. Use the following code to reset your password:</p>
		<h2>%s</h2>
		<p>This code will expire in 15 minutes.</p>
		<p>If you didn't request a password reset, please ignore this email or contact support if you have concerns.</p>
	`, user.Name, token)
	return s.sendEmail(user.Email, title, title, content)
}

func (s *AuthService) sendPasswordChangedEmail(user *AuthUser) error {
	title := "Your Albafone Password Has Been Changed"
	content := fmt.Sprintf("<p>Hi %s,</p><p>Your password has been successfully changed. If you did not make this change, please contact support immediately.</p>", user.Name)
	return s.sendEmail(user.Email, title, title, content)
}
