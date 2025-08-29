package posts

import (
	"math"
	"mime/multipart"

	"base/app/models"
	"base/core/emitter"
	"base/core/logger"
	"base/core/storage"
	"base/core/translation"
	"base/core/types"

	"gorm.io/gorm"
)

const (
	CreatePostEvent = "posts.create"
	UpdatePostEvent = "posts.update"
	DeletePostEvent = "posts.delete"
)

type PostService struct {
	DB                *gorm.DB
	Emitter           *emitter.Emitter
	Storage           *storage.ActiveStorage
	Logger            logger.Logger
	TranslationHelper *translation.Helper
}

func NewPostService(db *gorm.DB, emitter *emitter.Emitter, storage *storage.ActiveStorage, logger logger.Logger, translationHelper *translation.Helper) *PostService {
	return &PostService{
		DB:                db,
		Logger:            logger,
		Emitter:           emitter,
		Storage:           storage,
		TranslationHelper: translationHelper,
	}
}

// applySorting applies sorting to the query based on the sort and order parameters
func (s *PostService) applySorting(query *gorm.DB, sortBy *string, sortOrder *string) {
	// Valid sortable fields for Post
	validSortFields := map[string]string{
		"id":         "id",
		"created_at": "created_at",
		"updated_at": "updated_at",
		"title":      "title",
		"desc":       "desc",
		"feat":       "feat",
	}

	// Default sorting - if sort_order exists, always use it for custom ordering
	defaultSortBy := "id"
	defaultSortOrder := "desc"

	// Determine sort field
	sortField := defaultSortBy
	if sortBy != nil && *sortBy != "" {
		if field, exists := validSortFields[*sortBy]; exists {
			sortField = field
		}
	}

	// Determine sort direction (order parameter)
	sortDirection := defaultSortOrder
	if sortOrder != nil && (*sortOrder == "asc" || *sortOrder == "desc") {
		sortDirection = *sortOrder
	}

	// Apply sorting
	query.Order(sortField + " " + sortDirection)
}

func (s *PostService) Create(req *models.CreatePostRequest) (*models.Post, error) {
	item := &models.Post{
		Title: translation.NewField(req.Title),
		Desc:  req.Desc,
		// handled separately
	}

	if err := s.DB.Create(item).Error; err != nil {
		s.Logger.Error("failed to create post", logger.String("error", err.Error()))
		return nil, err
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
		return nil, err
	}

	// Validate request
	if err := ValidatePostUpdateRequest(req, id); err != nil {
		return nil, err
	}

	// Build updates map
	updates := make(map[string]any)
	if req.Title != "" {
		item.Title = translation.NewField(req.Title)
		updates["title"] = item.Title.Original
	}
	// For non-pointer string fields
	if req.Desc != "" {
		updates["desc"] = req.Desc
	}
	// Feat attachment is handled via separate endpoint

	if err := s.DB.Model(item).Updates(updates).Error; err != nil {
		s.Logger.Error("failed to update post",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	result, err := s.GetById(item.Id)
	if err != nil {
		s.Logger.Error("failed to get updated post",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
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
		return err
	}

	// Delete file attachments if any
	if item.Feat != nil {
		if err := s.Storage.Delete(item.Feat); err != nil {
			s.Logger.Error("failed to delete feat",
				logger.String("error", err.Error()),
				logger.Int("id", int(id)))
			return err
		}
	}

	if err := s.DB.Delete(item).Error; err != nil {
		s.Logger.Error("failed to delete post",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return err
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
		return nil, err
	}

	// Load translations for all translatable fields
	if err := s.loadTranslationsForItem(item); err != nil {
		s.Logger.Error("Failed to load translations", logger.String("error", err.Error()))
		// Continue without translations rather than failing
	}

	return item, nil
}

func (s *PostService) GetAll(page *int, limit *int, sortBy *string, sortOrder *string) (*types.PaginatedResponse, error) {
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
		return nil, err
	}

	// Apply pagination if provided
	if page != nil && limit != nil {
		offset := (*page - 1) * *limit
		query = query.Offset(offset).Limit(*limit)
	}

	// Apply sorting
	s.applySorting(query, sortBy, sortOrder)

	// Don't preload relationships for list response (faster)
	// query = (&models.Post{}).Preload(query)

	// Execute query
	if err := query.Find(&items).Error; err != nil {
		s.Logger.Error("failed to get posts",
			logger.String("error", err.Error()))
		return nil, err
	}

	// Load translations for all items
	if err := s.loadTranslationsForItems(items); err != nil {
		s.Logger.Error("Failed to load translations for items", logger.String("error", err.Error()))
		// Continue without translations rather than failing
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

// GetAllForSelect gets all items for select box/dropdown options (simplified response)
func (s *PostService) GetAllForSelect() ([]*models.Post, error) {
	var items []*models.Post

	query := s.DB.Model(&models.Post{})

	// Only select the necessary fields for select options
	query = query.Select("id, title")

	// Order by name/title for better UX
	query = query.Order("title ASC")

	if err := query.Find(&items).Error; err != nil {
		s.Logger.Error("Failed to fetch items for select", logger.String("error", err.Error()))
		return nil, err
	}

	// Load translations for all items
	if err := s.loadTranslationsForItems(items); err != nil {
		s.Logger.Error("Failed to load translations for items", logger.String("error", err.Error()))
		// Continue without translations rather than failing
	}

	return items, nil
}

// loadTranslationsForItem loads translations for all translatable fields in a single item
func (s *PostService) loadTranslationsForItem(item *models.Post) error {
	if item == nil {
		return nil
	}

	modelName := item.GetModelName()
	modelId := item.GetId()

	// Load translations for each translatable field
	if err := s.TranslationHelper.Service.LoadTranslationsForField(&item.Title, modelName, modelId, "title"); err != nil {
		s.Logger.Error("Failed to load translations for Title", logger.String("error", err.Error()))
		// Don't fail the request, just log the error
	}

	return nil
}

// loadTranslationsForItems loads translations for all translatable fields in multiple items
func (s *PostService) loadTranslationsForItems(items []*models.Post) error {
	for _, item := range items {
		if err := s.loadTranslationsForItem(item); err != nil {
			return err
		}
	}
	return nil
}

// UploadFeat uploads a file for the Post's Feat field
func (s *PostService) UploadFeat(id uint, file *multipart.FileHeader) (*models.Post, error) {
	item := &models.Post{}
	if err := s.DB.First(item, id).Error; err != nil {
		s.Logger.Error("failed to find post",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	// Delete existing file if any
	if item.Feat != nil {
		if err := s.Storage.Delete(item.Feat); err != nil {
			s.Logger.Error("failed to delete existing feat",
				logger.String("error", err.Error()),
				logger.Int("id", int(id)))
			return nil, err
		}
	}

	// Attach new file
	attachment, err := s.Storage.Attach(item, "feat", file)
	if err != nil {
		s.Logger.Error("failed to attach feat",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	// Update the model with the new attachment
	if err := s.DB.Model(item).Association("Feat").Replace(attachment); err != nil {
		s.Logger.Error("failed to associate feat",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	return s.GetById(id)
}

// RemoveFeat removes the file from the Post's Feat field
func (s *PostService) RemoveFeat(id uint) (*models.Post, error) {
	item := &models.Post{}
	if err := s.DB.First(item, id).Error; err != nil {
		s.Logger.Error("failed to find post",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	if item.Feat == nil {
		return item, nil
	}

	if err := s.Storage.Delete(item.Feat); err != nil {
		s.Logger.Error("failed to delete feat",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	// Clear the association
	if err := s.DB.Model(item).Association("Feat").Clear(); err != nil {
		s.Logger.Error("failed to clear feat association",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	return s.GetById(id)
}
