package auth

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"

	"os"
	"path/filepath"
	"sync"
	"text/template"
	"time"

	"base/core/app/users"
	"base/core/config"
	"base/core/email"
	"base/core/emitter"
	"base/core/helper"
	"base/core/types"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	emailTemplateMutex sync.RWMutex
	emailTemplateCache *template.Template
)

// loadEmailTemplate attempts to load email template from theme directory,
// falls back to hardcoded template if file doesn't exist
func loadEmailTemplate(templateName string) (*template.Template, error) {
	// Try to load from theme directory first
	themePath := filepath.Join("app", "theme", "default", "email", templateName+".html")
	if _, err := os.Stat(themePath); err == nil {
		// File exists, load it
		templateContent, err := os.ReadFile(themePath)
		if err != nil {
			return nil, fmt.Errorf("error reading theme email template %s: %w", themePath, err)
		}

		tmpl, err := template.New("email").Parse(string(templateContent))
		if err != nil {
			return nil, fmt.Errorf("error parsing theme email template %s: %w", themePath, err)
		}
		return tmpl, nil
	}

	// Fall back to hardcoded template
	tmpl, err := template.New("email").Parse(emailTemplate)
	if err != nil {
		return nil, fmt.Errorf("error parsing fallback email template: %w", err)
	}
	return tmpl, nil
}

// AuthService handles authentication related operations
type AuthService struct {
	db          *gorm.DB
	emailSender email.Sender
	emitter     *emitter.Emitter
}

// NewAuthService creates a new authentication service
func NewAuthService(db *gorm.DB, emailSender email.Sender, emitter *emitter.Emitter) *AuthService {
	return &AuthService{
		db:          db,
		emailSender: emailSender,
		emitter:     emitter,
	}
}

// validateUser checks if username or email already exists
func (s *AuthService) validateUser(email, username string) error {
	var count int64
	if err := s.db.Model(&AuthUser{}).
		Where("email = ? OR username = ?", email, username).
		Count(&count).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	if count > 0 {
		return errors.New("user already exists")
	}
	return nil
}

func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	// Validate unique constraints first
	if err := s.validateUser(req.Email, req.Username); err != nil {
		return nil, err
	}

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
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.New("user already exists")
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Generate JWT token
	token, err := helper.GenerateJWT(user.User.Id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	userData := types.UserData{
		Id:       user.Id,
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
	}

	// Emit registration event
	if s.emitter != nil {
		s.emitter.Emit("user.registered", userData)
	} else {

	}

	//Send welcome email asynchronously
	go func() {
		if err := s.sendEmail(user.Email, "Welcome to Base", "Welcome to Base", "Welcome to Base"); err != nil {

		}
	}()

	userResponse := users.ToResponse(&user.User)
	userResponse.LastLogin = now.Format(time.RFC3339)

	return &AuthResponse{
		UserResponse: *userResponse,
		AccessToken:  token,
		Exp:          now.Add(24 * time.Hour).Unix(),
	}, nil
}

func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	var user AuthUser
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Proceed with generating token and response
	now := time.Now()
	token, err := helper.GenerateJWT(user.User.Id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create the response
	userResponse := users.ToResponse(&user.User)
	if user.LastLogin != nil {
		userResponse.LastLogin = user.LastLogin.Format(time.RFC3339)
	}

	response := &AuthResponse{
		UserResponse: *userResponse,
		AccessToken:  token,
		Exp:          now.Add(24 * time.Hour).Unix(),
	}

	// Prepare the login event
	loginAllowed := true
	event := LoginEvent{
		User:         &user,
		LoginAllowed: &loginAllowed,
		Response:     response,
	}

	// Emit the login attempt event
	s.emitter.Emit("user.login_attempt", &event)

	// Check if login was allowed after event listeners have processed it
	if !loginAllowed {
		if event.Error != nil {
			return event.Response, errors.New(event.Error.Error)
		}
		return event.Response, errors.New("not authorized")
	}

	// Update last login with proper time handling
	if err := s.db.Model(&user).Update("last_login", sql.NullTime{
		Time:  now,
		Valid: true,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}

	return response, nil
}

func (s *AuthService) ForgotPassword(email string) error {
	var user AuthUser
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found: %w", err)
		}
		return fmt.Errorf("database error: %w", err)
	}

	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}
	expiry := time.Now().Add(15 * time.Minute)

	// Update reset token fields in transaction
	tx := s.db.Begin()
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
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found: %w", err)
		}
		return fmt.Errorf("database error: %w", err)
	}

	if user.ResetToken != token {
		return errors.New("invalid token")
	}

	if user.ResetTokenExpiry == nil || time.Now().After(*user.ResetTokenExpiry) {
		return errors.New("token expired")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password and clear reset token in transaction
	tx := s.db.Begin()
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

		}
	}()

	return nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return fmt.Sprintf("%x", b), nil
}

// Email sending functions
func (s *AuthService) sendEmail(to, subject, title, content string) error {
	return s.sendEmailWithTemplate(to, subject, title, content, "register")
}

// sendEmailWithTemplate sends email using specified template name
func (s *AuthService) sendEmailWithTemplate(to, subject, title, content, templateName string) error {
	var cachedTemplate *template.Template
	emailTemplateMutex.RLock()
	cachedTemplate = emailTemplateCache
	emailTemplateMutex.RUnlock()

	if cachedTemplate == nil {
		newTemplate, err := loadEmailTemplate(templateName)
		if err != nil {
			return fmt.Errorf("error loading email template: %w", err)
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
		From:    "no-reply@base.al",
		Subject: subject,
		Body:    body.String(),
		IsHTML:  true,
	}
	return s.emailSender.Send(msg)
}

func (s *AuthService) sendPasswordResetEmail(user *AuthUser, token string) error {
	title := "Reset Your Base Password"
	content := fmt.Sprintf(`
		<p>Hi %s,</p>
		<p>You have requested to reset your password. Use the following code to reset your password:</p>
		<h2 style="color: #667eea; font-size: 32px; margin: 20px 0; text-align: center; padding: 15px; background-color: #f8f9fa; border-radius: 8px; border: 2px dashed #667eea;">%s</h2>
		<p>This code will expire in 15 minutes.</p>
		<p>If you didn't request a password reset, please ignore this email or contact support if you have concerns.</p>
	`, user.Name, token)
	return s.sendEmailWithTemplate(user.Email, title, title, content, "password-reset")
}

func (s *AuthService) sendPasswordChangedEmail(user *AuthUser) error {
	title := "Password Changed Successfully"
	content := fmt.Sprintf(`
		<p>Hi %s,</p>
		<p>Your password has been successfully changed.</p>
		<p>If you did not make this change, please contact our support team immediately.</p>
		<p>For your security, we recommend:</p>
		<ul>
			<li>Use a unique, strong password</li>
			<li>Enable two-factor authentication if available</li>
			<li>Regularly review your account activity</li>
		</ul>
	`, user.Name)
	return s.sendEmailWithTemplate(user.Email, title, title, content, "password-changed")
}

// GetUserByID retrieves a user by their ID
func (s *AuthService) GetUserByID(userID uint) (*users.UserResponse, error) {
	var user users.User
	result := s.db.First(&user, userID)
	if result.Error != nil {
		return nil, result.Error
	}

	return users.ToResponse(&user), nil
}

// ValidateToken validates a JWT token and returns the user information
func (s *AuthService) ValidateToken(tokenString string) (*users.UserResponse, error) {
	if tokenString == "" {
		return nil, errors.New("empty token")
	}

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Return the secret key used for signing
		cfg := config.NewConfig()
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Check token expiration
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, errors.New("invalid expiration claim")
	}

	if time.Unix(int64(exp), 0).Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	// Extract user ID
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("invalid subject claim")
	}

	// Find user in database
	var user AuthUser
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Return user response
	return &users.UserResponse{
		Id:       user.Id,
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}
