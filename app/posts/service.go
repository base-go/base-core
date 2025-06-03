package posts

import (
	"fmt"
	"math"

	"base/app/models"
	"base/core/emitter"
	"base/core/logger"
	"base/core/storage"
	"base/core/types"

	"gorm.io/gorm"
)

const (
	CreatePostEvent = "posts.create"
	UpdatePostEvent = "posts.update"
	DeletePostEvent = "posts.delete"
)

type PostService struct {
	DB      *gorm.DB
	Emitter *emitter.Emitter
	Storage *storage.ActiveStorage
	Logger  logger.Logger
}

func NewPostService(db *gorm.DB, emitter *emitter.Emitter, storage *storage.ActiveStorage, logger logger.Logger) *PostService {
	return &PostService{
		DB:      db,
		Emitter: emitter,
		Storage: storage,
		Logger:  logger,
	}
}

func (s *PostService) Create(req *models.CreatePostRequest) (*models.Post, error) {
	item := &models.Post{
		Title:   req.Title,
		Content: req.Content,
		Status:  req.Status,
	}

	if err := s.DB.Create(item).Error; err != nil {
		s.Logger.Error("failed to create post", logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Emit create event
	s.Emitter.Emit(CreatePostEvent, item)

	return s.GetById(item.Id)
}

func (s *PostService) Update(id uint, req *models.UpdatePostRequest) (*models.Post, error) {
	item := &models.Post{}
	if err := s.DB.First(item, id).Error; err != nil {
		s.Logger.Error("failed to find post for update",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, fmt.Errorf("failed to find post: %w", err)
	}

	// Build updates map
	updates := make(map[string]interface{})
	// For string and other fields
	if req.Title != "" {
		updates["title"] = req.Title
	}
	// For string and other fields
	if req.Content != "" {
		updates["content"] = req.Content
	}
	// For integer fields
	if req.Status != 0 {
		updates["status"] = req.Status
	}

	if err := s.DB.Model(item).Updates(updates).Error; err != nil {
		s.Logger.Error("failed to update post",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	result, err := s.GetById(item.Id)
	if err != nil {
		s.Logger.Error("failed to get updated post",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, fmt.Errorf("failed to get updated post: %w", err)
	}

	// Emit update event
	s.Emitter.Emit(UpdatePostEvent, result)

	return result, nil
}

func (s *PostService) Delete(id uint) error {
	item := &models.Post{}
	if err := s.DB.First(item, id).Error; err != nil {
		s.Logger.Error("failed to find post for deletion",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return fmt.Errorf("failed to find post: %w", err)
	}

	// Delete file attachments if any

	if err := s.DB.Delete(item).Error; err != nil {
		s.Logger.Error("failed to delete post",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return fmt.Errorf("failed to delete post: %w", err)
	}

	// Emit delete event
	s.Emitter.Emit(DeletePostEvent, item)

	return nil
}

func (s *PostService) GetById(id uint) (*models.Post, error) {
	item := &models.Post{}

	query := item.Preload(s.DB)

	if err := query.First(item, id).Error; err != nil {
		s.Logger.Error("failed to get post",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	return item, nil
}

func (s *PostService) GetAll(page *int, limit *int) (*types.PaginatedResponse, error) {
	var items []*models.Post
	var total int64
	query := s.DB.Model(&models.Post{})
	// Set default values if nil
	defaultPage := 1
	defaultLimit := 10
	if page == nil {
		page = &defaultPage
	}
	if limit == nil {
		limit = &defaultLimit
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		s.Logger.Error("failed to count posts",
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to count posts: %w", err)
	}

	// Apply pagination if provided
	if page != nil && limit != nil {
		offset := (*page - 1) * *limit
		query = query.Offset(offset).Limit(*limit)
	}

	// Preload relationships
	query = (&models.Post{}).Preload(query)

	// Execute query
	if err := query.Find(&items).Error; err != nil {
		s.Logger.Error("failed to get posts",
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}

	// Convert to response type
	responses := make([]*models.PostListResponse, len(items))
	for i, item := range items {
		responses[i] = item.ToListResponse()
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(*limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	return &types.PaginatedResponse{
		Data: responses,
		Pagination: types.Pagination{
			Total:      int(total),
			Page:       *page,
			PageSize:   *limit,
			TotalPages: totalPages,
		},
	}, nil
}
