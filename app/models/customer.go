package models

import (
	"base/core/storage"
	"time"

	"gorm.io/gorm"
)

// Customer represents a customer entity
type Customer struct {
	Id        uint                `json:"id" gorm:"primarykey"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	DeletedAt gorm.DeletedAt      `json:"deleted_at" gorm:"index"`
	Name      string              `json:"name" gorm:"size:255"`
	Email     string              `json:"email" gorm:"size:255;index"`
	CountryId uint                `json:"country_id,omitempty"`
	Country   *Country            `json:"country,omitempty" gorm:"foreignKey:CountryId"`
	Flag      *storage.Attachment `json:"flag,omitempty"`
}

// TableName returns the table name for the Customer model
func (m *Customer) TableName() string {
	return "customers"
}

// GetId returns the Id of the model
func (m *Customer) GetId() uint {
	return m.Id
}

// GetModelName returns the model name
func (m *Customer) GetModelName() string {
	return "customer"
}

// CreateCustomerRequest represents the request payload for creating a Customer
type CreateCustomerRequest struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	CountryId uint   `json:"country_id" binding:"required"`
}

// UpdateCustomerRequest represents the request payload for updating a Customer
type UpdateCustomerRequest struct {
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	CountryId uint   `json:"country_id,omitempty"`
}

// CustomerResponse represents the API response for Customer
type CustomerResponse struct {
	Id        uint                  `json:"id"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt gorm.DeletedAt        `json:"deleted_at"`
	Name      string                `json:"name"`
	Email     string                `json:"email"`
	Country   *CountryModelResponse `json:"country,omitempty"`
	Flag      *storage.Attachment   `json:"flag,omitempty"`
}

// CustomerModelResponse represents a simplified response when this model is part of other entities
type CustomerModelResponse struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

// CustomerSelectOption represents a simplified response for select boxes and dropdowns
type CustomerSelectOption struct {
	Id   uint   `json:"id"`
	Name string `json:"name"` // From Name field
}

// CustomerListResponse represents the response for list operations (optimized for performance)
type CustomerListResponse struct {
	Id        uint           `json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
}

// ToResponse converts the model to an API response
func (m *Customer) ToResponse() *CustomerResponse {
	if m == nil {
		return nil
	}
	response := &CustomerResponse{
		Id:        m.Id,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt,
		Name:      m.Name,
		Email:     m.Email,
	}
	if m.Country != nil {
		response.Country = m.Country.ToModelResponse()
	}
	if m.Flag != nil {
		response.Flag = m.Flag
	}

	return response
}

// ToModelResponse converts the model to a simplified response for when it's part of other entities
func (m *Customer) ToModelResponse() *CustomerModelResponse {
	if m == nil {
		return nil
	}
	return &CustomerModelResponse{
		Id:   m.Id,
		Name: m.Name,
	}
}

// ToSelectOption converts the model to a select option for dropdowns
func (m *Customer) ToSelectOption() *CustomerSelectOption {
	if m == nil {
		return nil
	}
	displayName := m.Name

	return &CustomerSelectOption{
		Id:   m.Id,
		Name: displayName,
	}
}

// ToListResponse converts the model to a list response (without preloaded relationships for fast listing)
func (m *Customer) ToListResponse() *CustomerListResponse {
	if m == nil {
		return nil
	}
	return &CustomerListResponse{
		Id:        m.Id,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt,
		Name:      m.Name,
		Email:     m.Email,
	}
}

// Preload preloads all the model's relationships
func (m *Customer) Preload(db *gorm.DB) *gorm.DB {
	query := db
	query = query.Preload("Country")
	query = query.Preload("Flag")
	return query
}
