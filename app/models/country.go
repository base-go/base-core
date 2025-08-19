package models

import (
	"base/core/storage"
	"time"

	"gorm.io/gorm"
)

// Country represents a country entity
type Country struct {
	Id        uint                `json:"id" gorm:"primarykey"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	DeletedAt gorm.DeletedAt      `json:"deleted_at,omitempty" gorm:"index"`
	Name      string              `json:"name" gorm:"size:255"`
	Email     string              `json:"email" gorm:"size:255;index"`
	Flag      *storage.Attachment `json:"flag,omitempty"`
}

// TableName returns the table name for the Country model
func (m *Country) TableName() string {
	return "countries"
}

// GetId returns the Id of the model
func (m *Country) GetId() uint {
	return m.Id
}

// GetModelName returns the model name
func (m *Country) GetModelName() string {
	return "country"
}

// CreateCountryRequest represents the request payload for creating a Country
type CreateCountryRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UpdateCountryRequest represents the request payload for updating a Country
type UpdateCountryRequest struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// CountryResponse represents the API response for Country
type CountryResponse struct {
	Id        uint                `json:"id"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	DeletedAt gorm.DeletedAt      `json:"deleted_at,omitempty"`
	Name      string              `json:"name"`
	Email     string              `json:"email"`
	Flag      *storage.Attachment `json:"flag,omitempty"`
}

// CountryModelResponse represents a simplified response when this model is part of other entities
type CountryModelResponse struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

// CountrySelectOption represents a simplified response for select boxes and dropdowns
type CountrySelectOption struct {
	Id   uint   `json:"id"`
	Name string `json:"name"` // From Name field
}

// CountryListResponse represents the response for list operations (optimized for performance)
type CountryListResponse struct {
	Id        uint           `json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
}

// ToResponse converts the model to an API response
func (m *Country) ToResponse() *CountryResponse {
	if m == nil {
		return nil
	}
	response := &CountryResponse{
		Id:        m.Id,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt,
		Name:      m.Name,
		Email:     m.Email,
	}
	if m.Flag != nil {
		response.Flag = m.Flag
	}

	return response
}

// ToModelResponse converts the model to a simplified response for when it's part of other entities
func (m *Country) ToModelResponse() *CountryModelResponse {
	if m == nil {
		return nil
	}
	return &CountryModelResponse{
		Id:   m.Id,
		Name: m.Name,
	}
}

// ToSelectOption converts the model to a select option for dropdowns
func (m *Country) ToSelectOption() *CountrySelectOption {
	if m == nil {
		return nil
	}
	displayName := m.Name

	return &CountrySelectOption{
		Id:   m.Id,
		Name: displayName,
	}
}

// ToListResponse converts the model to a list response (without preloaded relationships for fast listing)
func (m *Country) ToListResponse() *CountryListResponse {
	if m == nil {
		return nil
	}
	return &CountryListResponse{
		Id:        m.Id,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt,
		Name:      m.Name,
		Email:     m.Email,
	}
}

// Preload preloads all the model's relationships
func (m *Country) Preload(db *gorm.DB) *gorm.DB {
	query := db
	return query
}
