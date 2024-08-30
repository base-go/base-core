package scenes

import (
	"errors"

	"gorm.io/gorm"
)

type SceneService struct {
	DB *gorm.DB
}

func NewSceneService(db *gorm.DB) *SceneService {
	return &SceneService{
		DB: db,
	}
}

func (s *SceneService) Create(req *CreateRequest) (*CreateResponse, error) {
	item := Scene{
		Title: req.Title,
		Description: req.Description,
		Cover: req.Cover,
		Pg_rating: req.Pg_rating,
		User_id: req.User_id,
		Slug: req.Slug,
	}
	if err := s.DB.Create(&item).Error; err != nil {
		return nil, err
	}
	return &CreateResponse{
		Model: item.Model,
		Title: item.Title,
		Description: item.Description,
		Cover: item.Cover,
		Pg_rating: item.Pg_rating,
		User_id: item.User_id,
		Slug: item.Slug,
	}, nil
}

func (s *SceneService) GetByID(id uint) (*Scene, error) {
	var item Scene
	if err := s.DB.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *SceneService) GetAll() ([]Scene, error) {
	var items []Scene
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *SceneService) Update(id uint, req *UpdateRequest) (*UpdateResponse, error) {
	item, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	item.Title = req.Title
	item.Description = req.Description
	item.Cover = req.Cover
	item.Pg_rating = req.Pg_rating
	item.User_id = req.User_id
	item.Slug = req.Slug
	if err := s.DB.Save(item).Error; err != nil {
		return nil, err
	}
	return &UpdateResponse{
		Model: item.Model,
		Title: item.Title,
		Description: item.Description,
		Cover: item.Cover,
		Pg_rating: item.Pg_rating,
		User_id: item.User_id,
		Slug: item.Slug,
	}, nil
}

func (s *SceneService) Delete(id uint) error {
	result := s.DB.Delete(&Scene{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("item not found")
	}
	return nil
}