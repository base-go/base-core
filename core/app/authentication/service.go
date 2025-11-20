package authentication

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"text/template"
	"time"

	"base/app"
	"base/core/app/profile"
	"base/core/email"
	"base/core/emitter"
	"base/core/types"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	emailTemplateMutex sync.RWMutex
	emailTemplateCache *template.Template
)

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

func (s *AuthService) ValidateKey(key string) (any, error) {
	return nil, nil
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

	// Determine role: first user gets Owner (1), subsequent users get Member (3)
	roleId := s.determineUserRole()

	now := time.Now()

	user := AuthUser{
		User: profile.User{
			Email:     req.Email,
			Password:  string(hashedPassword),
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Username:  req.Username,
			Phone:     req.Phone,
			RoleId:    roleId,
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

	// Get extended data for JWT token
	extendData := app.Extend(user.User.Id)

	// Generate JWT token
	token, err := types.GenerateJWT(user.User.Id, extendData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	userData := types.UserData{
		Id:        user.Id,
		FirstName: user.User.FirstName,
		LastName:  user.User.LastName,
		Username:  user.Username,
		Email:     user.Email,
	}

	// Emit registration event
	if s.emitter != nil {
		s.emitter.Emit("user.registered", userData)
	} else {
		fmt.Printf("Emitter is nil in AuthService.Register; cannot emit 'user.registered' event")
	}

	// Send welcome email asynchronously
	// go func() {
	// 	if err := s.sendWelcomeEmail(&user); err != nil {
	// 		fmt.Printf("Failed to send welcome email: %v", err)
	// 	}
	// }()

	userResponse := profile.ToResponse(&user.User)
	userResponse.LastLogin = now.Format(time.RFC3339)

	return &AuthResponse{
		UserResponse: *userResponse,
		AccessToken:  token,
		Exp:          now.Add(24 * time.Hour).Unix(),
		Extend:       extendData,
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

	// Get extended data for JWT token
	extendData := app.Extend(user.User.Id)

	// Proceed with generating token and response
	now := time.Now()
	token, err := types.GenerateJWT(user.User.Id, extendData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create the response
	userResponse := profile.ToResponse(&user.User)
	if user.LastLogin != nil {
		userResponse.LastLogin = user.LastLogin.Format(time.RFC3339)
	}

	response := &AuthResponse{
		UserResponse: *userResponse,
		AccessToken:  token,
		Exp:          now.Add(24 * time.Hour).Unix(),
		Extend:       extendData,
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

	updates := map[string]any{
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

	updates := map[string]any{
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

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return fmt.Sprintf("%x", b), nil
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
	err := cachedTemplate.Execute(&body, map[string]any{
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
		<h2>%s</h2>
		<p>This code will expire in 15 minutes.</p>
		<p>If you didn't request a password reset, please ignore this email or contact support if you have concerns.</p>
	`, user.FirstName, token)
	return s.sendEmail(user.Email, title, title, content)
}

func (s *AuthService) sendPasswordChangedEmail(user *AuthUser) error {
	title := "Your Base Password Has Been Changed"
	content := fmt.Sprintf("<p>Hi %s,</p><p>Your password has been successfully changed. If you did not make this change, please contact support immediately.</p>", user.FirstName)
	return s.sendEmail(user.Email, title, title, content)
}

// determineUserRole returns the appropriate role ID for a new user
// First user gets Owner role (1), subsequent users get Member role (3)
func (s *AuthService) determineUserRole() uint {
	var userCount int64
	if err := s.db.Model(&AuthUser{}).Count(&userCount).Error; err != nil {
		// If we can't count users, default to Member role for safety
		return 3 // Member role
	}

	// First user gets Owner role, all others get Member role
	if userCount == 0 {
		return 1 // Owner role
	}
	return 3 // Member role
}
