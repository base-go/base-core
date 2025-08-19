package countries

import (
	"math"
	"mime/multipart"

	"base/app/models"
	"base/core/emitter"
	"base/core/logger"
	"base/core/storage"
	"base/core/types"

	"gorm.io/gorm"
)

const (
	CreateCountryEvent = "countries.create"
	UpdateCountryEvent = "countries.update"
	DeleteCountryEvent = "countries.delete"
)

type CountryService struct {
	DB      *gorm.DB
	Emitter *emitter.Emitter
	Storage *storage.ActiveStorage
	Logger  logger.Logger
}

func NewCountryService(db *gorm.DB, emitter *emitter.Emitter, storage *storage.ActiveStorage, logger logger.Logger) *CountryService {
	return &CountryService{
		DB:      db,
		Emitter: emitter,
		Storage: storage,
		Logger:  logger,
	}
}

// applySorting applies sorting to the query based on the sort and order parameters
func (s *CountryService) applySorting(query *gorm.DB, sortBy *string, sortOrder *string) {
	// Valid sortable fields for Country
	validSortFields := map[string]string{
		"id":         "id",
		"created_at": "created_at",
		"updated_at": "updated_at",
		"name":       "name",
		"flag":       "flag",
		"email":      "email",
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

func (s *CountryService) Create(req *models.CreateCountryRequest) (*models.Country, error) {
	item := &models.Country{
		Name: req.Name,
		// Flag attachment is handled via separate endpoint
		Email: req.Email,
	}

	if err := s.DB.Create(item).Error; err != nil {
		s.Logger.Error("failed to create country", logger.String("error", err.Error()))
		return nil, err
	}

	// Emit create event
	s.Emitter.Emit(CreateCountryEvent, item)

	return s.GetById(item.Id)
}

func (s *CountryService) Update(id uint, req *models.UpdateCountryRequest) (*models.Country, error) {
	item := &models.Country{}
	if err := s.DB.First(item, id).Error; err != nil {
		s.Logger.Error("failed to find country for update",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	// Build updates map
	updates := make(map[string]any)
	// For string and other fields
	if req.Name != "" {
		updates["name"] = req.Name
	}
	// Flag attachment is handled via separate endpoint
	// For string and other fields
	if req.Email != "" {
		updates["email"] = req.Email
	}

	if err := s.DB.Model(item).Updates(updates).Error; err != nil {
		s.Logger.Error("failed to update country",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	result, err := s.GetById(item.Id)
	if err != nil {
		s.Logger.Error("failed to get updated country",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	// Emit update event
	s.Emitter.Emit(UpdateCountryEvent, result)

	return result, nil
}

func (s *CountryService) Delete(id uint) error {
	item := &models.Country{}
	if err := s.DB.First(item, id).Error; err != nil {
		s.Logger.Error("failed to find country for deletion",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return err
	}

	// Delete file attachments if any
	if item.Flag != nil {
		if err := s.Storage.Delete(item.Flag); err != nil {
			s.Logger.Error("failed to delete flag",
				logger.String("error", err.Error()),
				logger.Int("id", int(id)))
			return err
		}
	}

	if err := s.DB.Delete(item).Error; err != nil {
		s.Logger.Error("failed to delete country",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return err
	}

	// Emit delete event
	s.Emitter.Emit(DeleteCountryEvent, item)

	return nil
}

func (s *CountryService) GetById(id uint) (*models.Country, error) {
	item := &models.Country{}

	query := item.Preload(s.DB)

	if err := query.First(item, id).Error; err != nil {
		s.Logger.Error("failed to get country",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	return item, nil
}

func (s *CountryService) GetAll(page *int, limit *int, sortBy *string, sortOrder *string) (*types.PaginatedResponse, error) {
	var items []*models.Country
	var total int64

	query := s.DB.Model(&models.Country{})
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
		s.Logger.Error("failed to count countries",
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
	// query = (&models.Country{}).Preload(query)

	// Execute query
	if err := query.Find(&items).Error; err != nil {
		s.Logger.Error("failed to get countries",
			logger.String("error", err.Error()))
		return nil, err
	}

	// Convert to response type
	responses := make([]*models.CountryListResponse, len(items))
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
func (s *CountryService) GetAllForSelect() ([]*models.Country, error) {
	var items []*models.Country

	query := s.DB.Model(&models.Country{})

	// Only select the necessary fields for select options
	query = query.Select("id, name")

	// Order by name/title for better UX
	query = query.Order("name ASC")

	if err := query.Find(&items).Error; err != nil {
		s.Logger.Error("Failed to fetch items for select", logger.String("error", err.Error()))
		return nil, err
	}

	return items, nil
}

// UploadFlag uploads a file for the Country's Flag field
func (s *CountryService) UploadFlag(id uint, file *multipart.FileHeader) (*models.Country, error) {
	item := &models.Country{}
	if err := s.DB.First(item, id).Error; err != nil {
		s.Logger.Error("failed to find country",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	// Delete existing file if any
	if item.Flag != nil {
		if err := s.Storage.Delete(item.Flag); err != nil {
			s.Logger.Error("failed to delete existing flag",
				logger.String("error", err.Error()),
				logger.Int("id", int(id)))
			return nil, err
		}
	}

	// Attach new file
	attachment, err := s.Storage.Attach(item, "flag", file)
	if err != nil {
		s.Logger.Error("failed to attach flag",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	// Update the model with the new attachment
	if err := s.DB.Model(item).Association("Flag").Replace(attachment); err != nil {
		s.Logger.Error("failed to associate flag",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	return s.GetById(id)
}

// RemoveFlag removes the file from the Country's Flag field
func (s *CountryService) RemoveFlag(id uint) (*models.Country, error) {
	item := &models.Country{}
	if err := s.DB.First(item, id).Error; err != nil {
		s.Logger.Error("failed to find country",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	if item.Flag == nil {
		return item, nil
	}

	if err := s.Storage.Delete(item.Flag); err != nil {
		s.Logger.Error("failed to delete flag",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	// Clear the association
	if err := s.DB.Model(item).Association("Flag").Clear(); err != nil {
		s.Logger.Error("failed to clear flag association",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, err
	}

	return s.GetById(id)
}
