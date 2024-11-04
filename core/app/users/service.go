package users

import (
	"base/core/config"
	"base/core/file"
	"errors"
	"fmt"
	"mime/multipart"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	DB     *gorm.DB
	Logger *zap.Logger
}

func NewUserService(db *gorm.DB, logger *zap.Logger) *UserService {
	return &UserService{
		DB:     db,
		Logger: logger,
	}
}

func (s *UserService) Create(req *CreateRequest) (*UserResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.Logger.Error("Failed to hash password",
			zap.Error(err),
			zap.String("email", req.Email))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := User{
		Name:     req.Name,
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.DB.Create(&user).Error; err != nil {
		s.Logger.Error("Failed to create user in database",
			zap.Error(err),
			zap.String("email", req.Email))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &UserResponse{
		Id:       user.Id,
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
	}, nil
}

func (s *UserService) GetByID(id uint) (*UserResponse, error) {
	var user User
	if err := s.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.Logger.Debug("User not found",
				zap.Uint("user_id", id))
		} else {
			s.Logger.Error("Database error while fetching user",
				zap.Error(err),
				zap.Uint("user_id", id))
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &UserResponse{
		Id:       user.Id,
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
	}, nil
}

func (s *UserService) GetAll() ([]UserResponse, error) {
	var users []User
	if err := s.DB.Find(&users).Error; err != nil {
		s.Logger.Error("Failed to fetch users from database", zap.Error(err))
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = UserResponse{
			Id:       user.Id,
			Name:     user.Name,
			Username: user.Username,
			Email:    user.Email,
			Avatar:   user.Avatar,
		}
	}

	return userResponses, nil
}

func (s *UserService) Update(id uint, req *UpdateRequest) (*UserResponse, error) {
	var user User
	if err := s.DB.First(&user, id).Error; err != nil {
		s.Logger.Error("Failed to find user for update",
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

	if err := s.DB.Save(&user).Error; err != nil {
		s.Logger.Error("Failed to save user updates",
			zap.Error(err),
			zap.Uint("user_id", id))
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &UserResponse{
		Id:       user.Id,
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
	}, nil
}

func (s *UserService) Delete(id uint) error {
	result := s.DB.Delete(&User{}, id)
	if result.Error != nil {
		s.Logger.Error("Failed to delete user",
			zap.Error(result.Error),
			zap.Uint("user_id", id))
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		s.Logger.Debug("No user found to delete",
			zap.Uint("user_id", id))
		return errors.New("user not found")
	}
	return nil
}

func (s *UserService) UpdateAvatar(id uint, avatarFile *multipart.FileHeader) (*UserResponse, error) {
	var user User
	if err := s.DB.First(&user, id).Error; err != nil {
		s.Logger.Error("Failed to find user for avatar update",
			zap.Error(err),
			zap.Uint("user_id", id))
		return nil, err
	}

	config := config.NewConfig()
	customConfig := file.UploadConfig{
		AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".gif"},
		MaxFileSize:       5 << 20, // 5 MB
		UploadPath:        "/storage/avatars",
	}

	result, err := file.Upload(avatarFile, customConfig)
	if err != nil {
		s.Logger.Error("Failed to upload avatar file",
			zap.Error(err),
			zap.Uint("user_id", id),
			zap.String("filename", avatarFile.Filename))
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	user.Avatar = config.BaseURL + result.Path

	if err := s.DB.Save(&user).Error; err != nil {
		s.Logger.Error("Failed to save user with new avatar",
			zap.Error(err),
			zap.Uint("user_id", id))
		return nil, fmt.Errorf("failed to update user avatar: %w", err)
	}

	return &UserResponse{
		Id:       user.Id,
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
	}, nil
}

func (s *UserService) UpdatePassword(id uint, req *UpdatePasswordRequest) error {
	var user User
	if err := s.DB.First(&user, id).Error; err != nil {
		s.Logger.Error("Failed to find user for password update",
			zap.Error(err),
			zap.Uint("user_id", id))
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		s.Logger.Info("Invalid old password provided",
			zap.Uint("user_id", id))
		return bcrypt.ErrMismatchedHashAndPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.Logger.Error("Failed to hash new password",
			zap.Error(err),
			zap.Uint("user_id", id))
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.Password = string(hashedPassword)
	if err := s.DB.Save(&user).Error; err != nil {
		s.Logger.Error("Failed to save new password",
			zap.Error(err),
			zap.Uint("user_id", id))
		return fmt.Errorf("failed to update user password: %w", err)
	}

	return nil
}
