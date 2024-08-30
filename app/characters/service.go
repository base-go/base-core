package characters

import (
	"errors"

	"gorm.io/gorm"
)

type CharacterService struct {
	DB *gorm.DB
}

func NewCharacterService(db *gorm.DB) *CharacterService {
	return &CharacterService{
		DB: db,
	}
}

func (s *CharacterService) Create(req *CreateRequest) (*CreateResponse, error) {
	item := Character{
		Name: req.Name,
		Description: req.Description,
		Gender: req.Gender,
		Main_character: req.Main_character,
	}
	if err := s.DB.Create(&item).Error; err != nil {
		return nil, err
	}
	return &CreateResponse{
		Model: item.Model,
		Name: item.Name,
		Description: item.Description,
		Gender: item.Gender,
		Main_character: item.Main_character,
	}, nil
}

func (s *CharacterService) GetByID(id uint) (*Character, error) {
	var item Character
	if err := s.DB.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *CharacterService) GetAll() ([]Character, error) {
	var items []Character
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *CharacterService) Update(id uint, req *UpdateRequest) (*UpdateResponse, error) {
	item, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	item.Name = req.Name
	item.Description = req.Description
	item.Gender = req.Gender
	item.Main_character = req.Main_character
	if err := s.DB.Save(item).Error; err != nil {
		return nil, err
	}
	return &UpdateResponse{
		Model: item.Model,
		Name: item.Name,
		Description: item.Description,
		Gender: item.Gender,
		Main_character: item.Main_character,
	}, nil
}

func (s *CharacterService) Delete(id uint) error {
	result := s.DB.Delete(&Character{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("item not found")
	}
	return nil
}