package category

import (
	"base/app/model"
	"base/core/translation"
	"base/core/types"
	"errors"
	"math"

	"gorm.io/gorm"
)

type Service struct {
	DB                 *gorm.DB
	TranslationService *translation.TranslationService
}

func NewService(db *gorm.DB, translationService *translation.TranslationService) *Service {
	return &Service{
		DB:                 db,
		TranslationService: translationService,
	}
}

func (s *Service) Create(req *model.CreateCategoryRequest) (*model.CategoryResponse, error) {
	category := model.Category{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
	}

	if err := s.DB.Create(&category).Error; err != nil {
		return nil, err
	}

	return s.toResponse(&category), nil
}

func (s *Service) GetByID(id uint) (*model.CategoryResponse, error) {
	var category model.Category
	if err := s.DB.Preload("Parent").Preload("Children").First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}

	return s.toResponse(&category), nil
}

func (s *Service) GetBySlug(slug string) (*model.CategoryResponse, error) {
	var category model.Category
	if err := s.DB.Preload("Parent").Preload("Children").Where("slug = ?", slug).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}

	return s.toResponse(&category), nil
}

func (s *Service) GetAll(page, limit *int) (*types.PaginatedResponse, error) {
	var categories []model.Category
	var total int64

	// Count total records
	if err := s.DB.Model(&model.Category{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// Build query with pagination
	query := s.DB.Model(&model.Category{}).Preload("Parent").Preload("Children").Order("sort_order ASC, name ASC")

	// Apply pagination if provided
	currentPage := 1
	pageSize := 10
	if page != nil {
		currentPage = *page
	}
	if limit != nil {
		pageSize = *limit
	}

	offset := (currentPage - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	// Execute query
	if err := query.Find(&categories).Error; err != nil {
		return nil, err
	}

	// Convert to responses
	responses := make([]model.CategoryResponse, len(categories))
	for i, category := range categories {
		responses[i] = *s.toResponse(&category)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	return &types.PaginatedResponse{
		Data: responses,
		Pagination: types.Pagination{
			Total:      int(total),
			Page:       currentPage,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *Service) GetTree() ([]model.CategoryResponse, error) {
	var categories []model.Category
	if err := s.DB.Where("parent_id IS NULL").Preload("Children").Order("sort_order ASC, name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}

	responses := make([]model.CategoryResponse, len(categories))
	for i, category := range categories {
		responses[i] = *s.toResponse(&category)
	}

	return responses, nil
}

func (s *Service) Update(req *model.UpdateCategoryRequest) (*model.CategoryResponse, error) {
	var category model.Category
	if err := s.DB.First(&category, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}

	// Update fields
	if req.Slug != "" {
		category.Slug = req.Slug
	}
	if req.Color != "" {
		category.Color = req.Color
	}
	if req.Icon != "" {
		category.Icon = req.Icon
	}
	if req.ParentID != nil {
		category.ParentID = req.ParentID
	}
	if req.SortOrder != 0 {
		category.SortOrder = req.SortOrder
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	// Update translatable fields if provided
	if req.Name.Original != "" || len(req.Name.Values) > 0 {
		category.Name = req.Name
	}
	if req.Description.Original != "" || len(req.Description.Values) > 0 {
		category.Description = req.Description
	}

	if err := s.DB.Save(&category).Error; err != nil {
		return nil, err
	}

	return s.toResponse(&category), nil
}

func (s *Service) Delete(id uint) error {
	var category model.Category
	if err := s.DB.First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("category not found")
		}
		return err
	}

	return s.DB.Delete(&category).Error
}

func (s *Service) toResponse(category *model.Category) *model.CategoryResponse {
	response := &model.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		Color:       category.Color,
		Icon:        category.Icon,
		ParentID:    category.ParentID,
		SortOrder:   category.SortOrder,
		IsActive:    category.IsActive,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}

	if category.Parent != nil {
		response.Parent = s.toResponse(category.Parent)
	}

	if len(category.Children) > 0 {
		response.Children = make([]model.CategoryResponse, len(category.Children))
		for i, child := range category.Children {
			response.Children[i] = *s.toResponse(&child)
		}
	}

	return response
}