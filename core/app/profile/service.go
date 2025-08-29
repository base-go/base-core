package profile

import (
	"base/core/logger"
	"base/core/storage"
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ProfileService struct {
	db            *gorm.DB
	logger        logger.Logger
	activeStorage *storage.ActiveStorage
}

func NewProfileService(db *gorm.DB, logger logger.Logger, activeStorage *storage.ActiveStorage) *ProfileService {
	if db == nil {
		panic("db is required")
	}
	if logger == nil {
		panic("logger is required")
	}
	if activeStorage == nil {
		panic("activeStorage is required")
	}

	// Register avatar attachment configuration
	activeStorage.RegisterAttachment("users", storage.AttachmentConfig{
		Field:             "avatar",
		Path:              "avatars",
		AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".gif"},
		MaxFileSize:       5 << 20, // 5MB
		Multiple:          false,
	})

	return &ProfileService{
		db:            db,
		logger:        logger,
		activeStorage: activeStorage,
	}
}

// Helper method to convert user to response
func (s *ProfileService) ToResponse(user *User) *UserResponse {
	return ToResponse(user)
}

func (s *ProfileService) GetById(id uint) (*UserResponse, error) {
	var user User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("User not found",
				logger.Uint("user_id", id))
		} else {
			s.logger.Error("Database error while fetching user",

				logger.Uint("user_id", id))
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return s.ToResponse(&user), nil
}

func (s *ProfileService) Update(id uint, req *UpdateRequest) (*UserResponse, error) {
	var user User
	if err := s.db.First(&user, id).Error; err != nil {
		s.logger.Error("Failed to find user for update",
			zap.Error(err),
			zap.Uint("user_id", id))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := s.db.Save(&user).Error; err != nil {
		s.logger.Error("Failed to save user updates",
			zap.Error(err),
			zap.Uint("user_id", id))
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.ToResponse(&user), nil
}

func (s *ProfileService) UpdateAvatar(ctx context.Context, id uint, avatarFile *multipart.FileHeader) (*UserResponse, error) {
	var user User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, err
	}

	// Just attach the new file - cleanup is handled inside Attach
	attachment, err := s.activeStorage.Attach(&user, "avatar", avatarFile)
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	// Update user's avatar
	user.Avatar = attachment
	if err := s.db.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.ToResponse(&user), nil
}

func (s *ProfileService) RemoveAvatar(ctx context.Context, id uint) (*UserResponse, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var user User
	if err := tx.First(&user, id).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if user.Avatar != nil {
		if err := s.activeStorage.Delete(user.Avatar); err != nil {
			tx.Rollback()
			s.logger.Error("Failed to delete avatar",
				zap.Error(err),
				zap.Uint("user_id", id))
			return nil, fmt.Errorf("failed to delete avatar: %w", err)
		}
		user.Avatar = nil
		if err := tx.Save(&user).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return s.ToResponse(&user), nil
}

func (s *ProfileService) UpdatePassword(id uint, req *UpdatePasswordRequest) error {
	var user User
	if err := s.db.First(&user, id).Error; err != nil {
		s.logger.Error("Failed to find user for password update",
			zap.Error(err),
			zap.Uint("user_id", id))
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		s.logger.Info("Invalid old password provided",
			zap.Uint("user_id", id))
		return bcrypt.ErrMismatchedHashAndPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash new password",
			zap.Error(err),
			zap.Uint("user_id", id))
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.Password = string(hashedPassword)
	if err := s.db.Save(&user).Error; err != nil {
		s.logger.Error("Failed to save new password",
			zap.Error(err),
			zap.Uint("user_id", id))
		return fmt.Errorf("failed to update user password: %w", err)
	}

	return nil
}
