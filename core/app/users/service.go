package users

import (
	"base/core/storage"
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db            *gorm.DB
	logger        *zap.Logger
	activeStorage *storage.ActiveStorage
}

func NewUserService(db *gorm.DB, logger *zap.Logger, activeStorage *storage.ActiveStorage) *UserService {
	return &UserService{
		db:            db,
		logger:        logger,
		activeStorage: activeStorage,
	}
}

// Helper method to convert user to response
func (s *UserService) toResponse(user *User) *UserResponse {
	return &UserResponse{
		Id:       user.Id,
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
	}
}

func (s *UserService) GetByID(id uint) (*UserResponse, error) {
	var user User
	if err := s.db.Preload("Avatar").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Debug("User not found", zap.Uint("user_id", id))
		} else {
			s.logger.Error("Database error while fetching user",
				zap.Error(err),
				zap.Uint("user_id", id))
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return s.toResponse(&user), nil
}

func (s *UserService) Update(id uint, req *UpdateRequest) (*UserResponse, error) {
	var user User
	if err := s.db.Preload("Avatar").First(&user, id).Error; err != nil {
		s.logger.Error("Failed to find user for update",
			zap.Error(err),
			zap.Uint("user_id", id))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update fields only if they are provided
	if req.Name != "" {
		user.Name = req.Name
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

	return s.toResponse(&user), nil
}

func (s *UserService) UpdateAvatar(ctx context.Context, id uint, avatarFile *multipart.FileHeader) (*UserResponse, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var user User
	if err := tx.Preload("Avatar").First(&user, id).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to find user for avatar update",
			zap.Error(err),
			zap.Uint("user_id", id))
		return nil, err
	}

	// Delete existing avatar if exists
	if user.Avatar != nil {
		if err := s.activeStorage.Delete(user.Avatar); err != nil {
			tx.Rollback()
			s.logger.Error("Failed to delete existing avatar",
				zap.Error(err),
				zap.Uint("user_id", id))
			return nil, fmt.Errorf("failed to delete existing avatar: %w", err)
		}
	}

	// Upload new avatar
	attachment, err := s.activeStorage.Attach(&user, "avatar", avatarFile)
	if err != nil {
		tx.Rollback()
		s.logger.Error("Failed to upload avatar",
			zap.Error(err),
			zap.Uint("user_id", id),
			zap.String("filename", avatarFile.Filename))
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	// Update user's avatar
	user.Avatar = attachment

	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to save user with new avatar",
			zap.Error(err),
			zap.Uint("user_id", id))
		return nil, fmt.Errorf("failed to update avatar: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit transaction",
			zap.Error(err),
			zap.Uint("user_id", id))
		return nil, fmt.Errorf("failed to update avatar: %w", err)
	}

	return s.toResponse(&user), nil
}

func (s *UserService) RemoveAvatar(ctx context.Context, id uint) (*UserResponse, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var user User
	if err := tx.Preload("Avatar").First(&user, id).Error; err != nil {
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

	return s.toResponse(&user), nil
}

func (s *UserService) UpdatePassword(id uint, req *UpdatePasswordRequest) error {
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
