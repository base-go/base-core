package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Disbursement represents a disbursement entity
type Disbursement struct {
	Id          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	Amount      float64        `json:"amount" gorm:"type:decimal(10,2)"`
	Description string         `json:"description" gorm:"size:255"`
}

// TableName returns the table name for the Disbursement model
func (m *Disbursement) TableName() string {
	return "disbursements"
}

// GetId returns the Id of the model
func (m *Disbursement) GetId() uint {
	return m.Id
}

// GetModelName returns the model name
func (m *Disbursement) GetModelName() string {
	return "disbursement"
}

// DisbursementResponse represents the API response for Disbursement
type DisbursementResponse struct {
	Id          uint           `json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty"`
	Amount      float64        `json:"amount"`
	Description string         `json:"description"`
}

// DisbursementModelResponse represents a simplified response when this model is part of other entities
type DisbursementModelResponse struct {
	Id   uint   `json:"id"`
	Name string `json:"name"` // Display name
}

// DisbursementSelectOption represents a simplified response for select boxes and dropdowns
type DisbursementSelectOption struct {
	Id   uint   `json:"id"`
	Name string `json:"name"` // Display name
}

// DisbursementListResponse represents the response for list operations (optimized for performance)
type DisbursementListResponse struct {
	Id          uint           `json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty"`
	Amount      float64        `json:"amount"`
	Description string         `json:"description"`
}

// CreateDisbursementRequest represents the request payload for creating a Disbursement
type CreateDisbursementRequest struct {
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

// UpdateDisbursementRequest represents the request payload for updating a Disbursement
type UpdateDisbursementRequest struct {
	Amount      float64 `json:"amount,omitempty"`
	Description string  `json:"description,omitempty"`
}

// ToResponse converts the model to an API response
func (m *Disbursement) ToResponse() *DisbursementResponse {
	if m == nil {
		return nil
	}
	response := &DisbursementResponse{
		Id:          m.Id,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt,
		Amount:      m.Amount,
		Description: m.Description,
	}

	return response
}

// ToModelResponse converts the model to a simplified response for when it's part of other entities
func (m *Disbursement) ToModelResponse() *DisbursementModelResponse {
	if m == nil {
		return nil
	}
	return &DisbursementModelResponse{
		Id:   m.Id,
		Name: fmt.Sprintf("Disbursement #%d", m.Id), // Fallback to ID-based display
	}
}

// ToSelectOption converts the model to a select option for dropdowns
func (m *Disbursement) ToSelectOption() *DisbursementSelectOption {
	if m == nil {
		return nil
	}
	displayName := m.Description // Using first string field as display name

	return &DisbursementSelectOption{
		Id:   m.Id,
		Name: displayName,
	}
}

// ToListResponse converts the model to a list response (without preloaded relationships for fast listing)
func (m *Disbursement) ToListResponse() *DisbursementListResponse {
	if m == nil {
		return nil
	}
	return &DisbursementListResponse{
		Id:          m.Id,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt,
		Amount:      m.Amount,
		Description: m.Description,
	}
}

// Preload preloads all the model's relationships
func (m *Disbursement) Preload(db *gorm.DB) *gorm.DB {
	query := db
	return query
}
