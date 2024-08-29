package customers

import (
	"errors"

	"gorm.io/gorm"
)

type CustomerService struct {
	DB *gorm.DB
}

func NewCustomerService(db *gorm.DB) *CustomerService {
	return &CustomerService{
		DB: db,
	}
}

func (s *CustomerService) Create(req *CreateRequest) (*CreateResponse, error) {
	item := Customer{
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

func (s *CustomerService) GetByID(id uint) (*Customer, error) {
	var item Customer
	if err := s.DB.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *CustomerService) GetAll() ([]Customer, error) {
	var items []Customer
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *CustomerService) Update(id uint, req *UpdateRequest) (*UpdateResponse, error) {
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

func (s *CustomerService) Delete(id uint) error {
	result := s.DB.Delete(&Customer{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("item not found")
	}
	return nil
}