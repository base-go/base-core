package stories

import (
	"errors"

	"gorm.io/gorm"
)

type StoryService struct {
	DB *gorm.DB
}

func NewStoryService(db *gorm.DB) *StoryService {
	return &StoryService{
		DB: db,
	}
}

func (s *StoryService) Create(req *CreateRequest) (*CreateResponse, error) {
	item := Story{
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

func (s *StoryService) GetByID(id uint) (*Story, error) {
	var item Story
	if err := s.DB.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *StoryService) GetAll() ([]Story, error) {
	var items []Story
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *StoryService) Update(id uint, req *UpdateRequest) (*UpdateResponse, error) {
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

func (s *StoryService) Delete(id uint) error {
	result := s.DB.Delete(&Story{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("item not found")
	}
	return nil
}