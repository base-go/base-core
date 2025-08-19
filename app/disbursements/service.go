package disbursements

import (
	"math"

	"base/app/models"
	"base/core/emitter"
	"base/core/logger"
	"base/core/storage"
	"base/core/types"

	"gorm.io/gorm"
)

const (
	CreateDisbursementEvent = "disbursements.create"
	UpdateDisbursementEvent = "disbursements.update"
	DeleteDisbursementEvent = "disbursements.delete"
)

type DisbursementService struct {
	DB      *gorm.DB
	Emitter *emitter.Emitter
	Storage *storage.ActiveStorage
	Logger  logger.Logger
}

func NewDisbursementService(db *gorm.DB, emitter *emitter.Emitter, storage *storage.ActiveStorage, logger logger.Logger) *DisbursementService {
	return &DisbursementService{
		DB:      db,
		Emitter: emitter,
		Storage: storage,
		Logger:  logger,
	}
}

// applySorting applies sorting to the query based on the sort and order parameters
func (s *DisbursementService) applySorting(query *gorm.DB, sortBy *string, sortOrder *string) {
	// Valid sortable fields for Disbursement
	validSortFields := map[string]string{
		"id":          "id",
		"created_at":  "created_at",
		"updated_at":  "updated_at",
		"amount":      "amount",
		"description": "description",
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

func (s *DisbursementService) Create(req *models.CreateDisbursementRequest) (*models.Disbursement, error) {
	item := &models.Disbursement{
		Amount:      req.Amount,
		Description: req.Description,
	}

	if err := s.DB.Create(item).Error; err != nil {
		s.Logger.Error("failed to create disbursement", logger.String("error", err.Error()))
		return nil, err
	}

	// Emit create event
	s.Emitter.Emit(CreateDisbursementEvent, item)

	return s.GetById(item.Id)
}

func (s *DisbursementService) Update(id uint, req *models.UpdateDisbursementRequest) (*models.Disbursement, error) {
	item := &models.Disbursement{}
	if err := s.DB.First(item, id).Error; err != nil {
		s.Logger.Error("failed to find disbursement for update",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	// Build updates map
	updates := make(map[string]any)
	// For float fields
	if req.Amount != 0 {
		updates["amount"] = req.Amount
	}
	// For string and other fields
	if req.Description != "" {
		updates["description"] = req.Description
	}

	if err := s.DB.Model(item).Updates(updates).Error; err != nil {
		s.Logger.Error("failed to update disbursement",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	result, err := s.GetById(item.Id)
	if err != nil {
		s.Logger.Error("failed to get updated disbursement",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	// Emit update event
	s.Emitter.Emit(UpdateDisbursementEvent, result)

	return result, nil
}

func (s *DisbursementService) Delete(id uint) error {
	item := &models.Disbursement{}
	if err := s.DB.First(item, id).Error; err != nil {
		s.Logger.Error("failed to find disbursement for deletion",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return err
	}

	// Delete file attachments if any

	if err := s.DB.Delete(item).Error; err != nil {
		s.Logger.Error("failed to delete disbursement",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return err
	}

	// Emit delete event
	s.Emitter.Emit(DeleteDisbursementEvent, item)

	return nil
}

func (s *DisbursementService) GetById(id uint) (*models.Disbursement, error) {
	item := &models.Disbursement{}

	query := item.Preload(s.DB)

	if err := query.First(item, id).Error; err != nil {
		s.Logger.Error("failed to get disbursement",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	return item, nil
}

func (s *DisbursementService) GetAll(page *int, limit *int, sortBy *string, sortOrder *string) (*types.PaginatedResponse, error) {
	var items []*models.Disbursement
	var total int64

	query := s.DB.Model(&models.Disbursement{})
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
		s.Logger.Error("failed to count disbursements",
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
	// query = (&models.Disbursement{}).Preload(query)

	// Execute query
	if err := query.Find(&items).Error; err != nil {
		s.Logger.Error("failed to get disbursements",
			logger.String("error", err.Error()))
		return nil, err
	}

	// Convert to response type
	responses := make([]*models.DisbursementListResponse, len(items))
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
func (s *DisbursementService) GetAllForSelect() ([]*models.Disbursement, error) {
	var items []*models.Disbursement

	query := s.DB.Model(&models.Disbursement{})

	// Only select the necessary fields for select options
	query = query.Select("id") // Only ID if no name/title field found

	// Order by name/title for better UX
	query = query.Order("id ASC")

	if err := query.Find(&items).Error; err != nil {
		s.Logger.Error("Failed to fetch items for select", logger.String("error", err.Error()))
		return nil, err
	}

	return items, nil
}
