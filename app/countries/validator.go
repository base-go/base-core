package countries

import (
	"base/app/models"
	"base/core/validator"
)

// Global validator instance using Base core validator wrapper
var validate = validator.New()

// ValidateCountryCreateRequest validates the create request
func ValidateCountryCreateRequest(req *models.CreateCountryRequest) error {
	if req == nil {
		return validator.ValidationErrors{
			{
				Field:   "request",
				Tag:     "required",
				Value:   "nil",
				Message: "request cannot be nil",
			},
		}
	}

	// Use Base core validator
	return validate.Validate(req)
}

// ValidateCountryUpdateRequest validates the update request
func ValidateCountryUpdateRequest(req *models.UpdateCountryRequest, id uint) error {
	if req == nil {
		return validator.ValidationErrors{
			{
				Field:   "request",
				Tag:     "required",
				Value:   "nil",
				Message: "request cannot be nil",
			},
		}
	}

	if id == 0 {
		return validator.ValidationErrors{
			{
				Field:   "id",
				Tag:     "required",
				Value:   "0",
				Message: "id cannot be zero",
			},
		}
	}

	// Use Base core validator
	return validate.Validate(req)
}

// ValidateCountryDeleteRequest validates the delete request
func ValidateCountryDeleteRequest(id uint) error {
	return ValidateID(id)
}

// ValidateID validates if the ID is valid
func ValidateID(id uint) error {
	if id == 0 {
		return validator.ValidationErrors{
			{
				Field:   "id",
				Tag:     "required",
				Value:   "0",
				Message: "id cannot be zero",
			},
		}
	}
	return nil
}
