package categories

import (
    "math"
    "strings"
    "mime/multipart"

    "gorm.io/gorm"
    "base/core/storage"
    "base/core/types"
    "base/app/models"
)

type CategoryService struct {
    DB      *gorm.DB
    storage *storage.ActiveStorage
}

func NewCategoryService(db *gorm.DB, activeStorage *storage.ActiveStorage) *CategoryService {
    return &CategoryService{
        DB:      db,
        storage: activeStorage,
    }
}

func (s *CategoryService) Create(req *models.CreateCategoryRequest) (*models.Category, error) {
    item := models.Category{
        Title: req.Title,
        Content: req.Content,
    }

    if err := s.DB.Create(&item).Error; err != nil {
        return nil, err
    }
    return s.GetById(item.Id)
}

func (s *CategoryService) Update(id uint, req *models.UpdateCategoryRequest) (*models.Category, error) {
    // First check if item exists
    if err := s.DB.First(&models.Category{}, id).Error; err != nil {
        return nil, err
    }

    // Build updates map
    updates := make(map[string]interface{})
    if req.Title != nil {
        updates["title"] = *req.Title
    }
    if req.Content != nil {
        updates["content"] = *req.Content
    }

    if err := s.DB.Model(&models.Category{}).Where("id = ?", id).Updates(updates).Error; err != nil {
        return nil, err
    }
    return s.GetById(id)
}

func (s *CategoryService) GetById(id uint) (*models.Category, error) {
    var item models.Category
    if err := item.Preload(s.DB).First(&item, id).Error; err != nil {
        return nil, err
    }
    return &item, nil
}

func (s *CategoryService) GetAll(page *int, limit *int, search *string) (*types.PaginatedResponse, error) {
    var items []*models.Category
    var total int64
    query := s.DB.Model(&models.Category{})

    // Apply search if provided
    if search != nil && *search != "" {
        searchValue := "%" + strings.ToLower(*search) + "%"
        query = query.Where(
            
            "LOWER(title) LIKE ?", searchValue,
            
            
            ).Or("LOWER(content) LIKE ?", searchValue,
        )
    }

    // Get total count
    if err := query.Count(&total).Error; err != nil {
        return nil, err
    }

    // Apply pagination if provided
    if page != nil && limit != nil {
        offset := (*page - 1) * *limit
        query = query.Offset(offset).Limit(*limit)
    }

    // Get items with preloaded relations
    if err := (&models.Category{}).Preload(query).Find(&items).Error; err != nil {
        return nil, err
    }

    // Convert to response type
    listItems := make([]*models.CategoryListResponse, len(items))
    for i, item := range items {
        listItems[i] = item.ToListResponse()
    }

    pagination := types.Pagination{
        Total:      int(total),
        Page:       *page,
        PageSize:   *limit,
        TotalPages: int(math.Ceil(float64(total) / float64(*limit))),
    }

    response := &types.PaginatedResponse{
        Data:       listItems,
        Pagination: pagination,
    }

    return response, nil
}

func (s *CategoryService) Delete(id uint) error {
    // First check if item exists
    if err := s.DB.First(&models.Category{}, id).Error; err != nil {
        return err
    }

    return s.DB.Delete(&models.Category{}, id).Error
}







func (s *CategoryService) UploadImage(id uint, file *multipart.FileHeader) error {
    var item models.Category
    if err := s.DB.First(&item, id).Error; err != nil {
        return err
    }

    attachment, err := s.storage.Attach(&item, "image", file)
    if err != nil {
        return err
    }

    return s.DB.Model(&item).Update("image", attachment).Error
}




