package posts

import (
    "math"
    "strings"
    "mime/multipart"

    "gorm.io/gorm"
    "base/core/storage"
    "base/core/types"
    "base/app/models"
)

type PostService struct {
    DB      *gorm.DB
    storage *storage.ActiveStorage
}

func NewPostService(db *gorm.DB, activeStorage *storage.ActiveStorage) *PostService {
    return &PostService{
        DB:      db,
        storage: activeStorage,
    }
}

func (s *PostService) Create(req *models.CreatePostRequest) (*models.Post, error) {
    item := models.Post{
        Title: req.Title,
        Content: req.Content,
        CategoryId: req.CategoryId,
    }

    if err := s.DB.Create(&item).Error; err != nil {
        return nil, err
    }
    return s.GetById(item.Id)
}

func (s *PostService) Update(id uint, req *models.UpdatePostRequest) (*models.Post, error) {
    // First check if item exists
    if err := s.DB.First(&models.Post{}, id).Error; err != nil {
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
    if req.CategoryId != nil {
        updates["category_id"] = *req.CategoryId
    }

    if err := s.DB.Model(&models.Post{}).Where("id = ?", id).Updates(updates).Error; err != nil {
        return nil, err
    }
    return s.GetById(id)
}

func (s *PostService) GetById(id uint) (*models.Post, error) {
    var item models.Post
    if err := item.Preload(s.DB).First(&item, id).Error; err != nil {
        return nil, err
    }
    return &item, nil
}

func (s *PostService) GetAll(page *int, limit *int, search *string) (*types.PaginatedResponse, error) {
    var items []*models.Post
    var total int64
    query := s.DB.Model(&models.Post{})

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
    if err := (&models.Post{}).Preload(query).Find(&items).Error; err != nil {
        return nil, err
    }

    // Convert to response type
    listItems := make([]*models.PostListResponse, len(items))
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

func (s *PostService) Delete(id uint) error {
    // First check if item exists
    if err := s.DB.First(&models.Post{}, id).Error; err != nil {
        return err
    }

    return s.DB.Delete(&models.Post{}, id).Error
}







func (s *PostService) UploadImage(id uint, file *multipart.FileHeader) error {
    var item models.Post
    if err := s.DB.First(&item, id).Error; err != nil {
        return err
    }

    attachment, err := s.storage.Attach(&item, "image", file)
    if err != nil {
        return err
    }

    return s.DB.Model(&item).Update("image", attachment).Error
}




