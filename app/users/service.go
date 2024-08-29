package users

import (
	"errors"

	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		DB: db,
	}
}

func (s *UserService) Create(req *CreateRequest) (*CreateResponse, error) {
	item := User{
		Name: req.Name,
		Email: req.Email,
		Password: req.Password,
		Avatar: req.Avatar,
	}
	if err := s.DB.Create(&item).Error; err != nil {
		return nil, err
	}
	return &CreateResponse{
		Model: item.Model,
		Name: item.Name,
		Email: item.Email,
		Password: item.Password,
		Avatar: item.Avatar,
	}, nil
}

func (s *UserService) GetByID(id uint) (*User, error) {
	var item User
	if err := s.DB.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *UserService) GetAll() ([]User, error) {
	var items []User
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *UserService) Update(id uint, req *UpdateRequest) (*UpdateResponse, error) {
	item, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	item.Name = req.Name
	item.Email = req.Email
	item.Password = req.Password
	item.Avatar = req.Avatar
	if err := s.DB.Save(item).Error; err != nil {
		return nil, err
	}
	return &UpdateResponse{
		Model: item.Model,
		Name: item.Name,
		Email: item.Email,
		Password: item.Password,
		Avatar: item.Avatar,
	}, nil
}

func (s *UserService) Delete(id uint) error {
	result := s.DB.Delete(&User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("item not found")
	}
	return nil
}