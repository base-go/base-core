package countries_test

import (
	"fmt"
	"testing"

	"base/app/countries"
	"base/app/models"
	"base/core/emitter"
	"base/test"

	"github.com/stretchr/testify/assert"
)

func TestCountryServices(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Auto-migrate Country table for testing
	err := helper.DB.AutoMigrate(&models.Country{})
	assert.NoError(t, err)

	// Create service with proper emitter and nil storage
	service := countries.NewCountryService(
		helper.DB,
		emitter.New(), // proper emitter
		nil,           // storage (can be nil)
		helper.Logger,
	)

	t.Run("Country service operations comprehensive coverage", func(t *testing.T) {
		t.Run("Service creation", func(t *testing.T) {
			// Test that service is created successfully
			assert.NotNil(t, service)
		})

		t.Run("Basic CRUD operations", func(t *testing.T) {
			// Test Create operation
			createReq := &models.CreateCountryRequest{
				Name:  "Test Name",
				Email: "Test Email",
			}

			country, err := service.Create(createReq)
			assert.NoError(t, err)
			assert.NotNil(t, country)
			assert.NotZero(t, country.Id)
			assert.Equal(t, "Test Name", country.Name)
			assert.Equal(t, "Test Email", country.Email)

			// Test GetById operation
			found, err := service.GetById(country.Id)
			assert.NoError(t, err)
			assert.NotNil(t, found)
			assert.Equal(t, country.Id, found.Id)

			// Test Update operation
			updateReq := &models.UpdateCountryRequest{
				Name:  "Updated Name",
				Email: "Updated Email",
			}

			updated, err := service.Update(country.Id, updateReq)
			assert.NoError(t, err)
			assert.NotNil(t, updated)
			assert.Equal(t, "Updated Name", updated.Name)
			assert.Equal(t, "Updated Email", updated.Email)

			// Test GetAll operation
			result, err := service.GetAll(nil, nil, nil, nil)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotNil(t, result.Data)

			// Test Delete operation
			err = service.Delete(country.Id)
			assert.NoError(t, err)

			// Verify deletion
			_, err = service.GetById(country.Id)
			assert.Error(t, err) // Should not find deleted record
		})

		t.Run("get operations", func(t *testing.T) {
			// Create test country
			country := &models.Country{
				Name:  "Test Name",
				Email: "Test Email",
			}
			err := helper.DB.Create(country).Error
			assert.NoError(t, err)

			// Test GetById
			found, err := service.GetById(country.Id)
			assert.NoError(t, err)
			assert.NotNil(t, found)
			assert.Equal(t, country.Id, found.Id)

			// Test GetById with invalid ID
			_, err = service.GetById(0)
			assert.Error(t, err)

			// Test GetById with non-existent ID
			_, err = service.GetById(99999)
			assert.Error(t, err)
		})

		t.Run("update operations", func(t *testing.T) {
			// Create test country
			country := &models.Country{
				Name:  "Test Name",
				Email: "Test Email",
			}
			err := helper.DB.Create(country).Error
			assert.NoError(t, err)

			// Test update
			updateReq := &models.UpdateCountryRequest{
				Name:  "Updated Name",
				Email: "Updated Email",
			}

			updated, err := service.Update(country.Id, updateReq)
			assert.NoError(t, err)
			assert.NotNil(t, updated)
			assert.Equal(t, "Updated Name", updated.Name)
			assert.Equal(t, "Updated Email", updated.Email)

			// Test update with invalid ID
			_, err = service.Update(0, updateReq)
			assert.Error(t, err)

			// Test update with non-existent ID
			_, err = service.Update(99999, updateReq)
			assert.Error(t, err)
		})

		t.Run("delete operations", func(t *testing.T) {
			// Create test country
			country := &models.Country{
				Name:  "Test Name",
				Email: "Test Email",
			}
			err := helper.DB.Create(country).Error
			assert.NoError(t, err)

			// Test delete
			err = service.Delete(country.Id)
			assert.NoError(t, err)

			// Verify deletion
			_, err = service.GetById(country.Id)
			assert.Error(t, err)

			// Test delete with invalid ID
			err = service.Delete(0)
			assert.Error(t, err)

			// Test delete with non-existent ID
			err = service.Delete(99999)
			assert.Error(t, err)
		})

		t.Run("get all operations", func(t *testing.T) {
			// Create multiple test countries
			for i := 0; i < 5; i++ {
				country := &models.Country{
					Name:  fmt.Sprintf("Test Name %d", i),
					Email: fmt.Sprintf("Test Email %d", i),
				}
				err := helper.DB.Create(country).Error
				assert.NoError(t, err)
			}

			// Test GetAll without pagination
			result, err := service.GetAll(nil, nil, nil, nil)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotEmpty(t, result.Data)

			// Test GetAll with pagination
			page := 1
			limit := 2
			result, err = service.GetAll(&page, &limit, nil, nil)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			// Type assert to slice to use len()
			if data, ok := result.Data.([]interface{}); ok {
				assert.LessOrEqual(t, len(data), limit)
			}
			assert.Equal(t, page, result.Pagination.Page)
			assert.Equal(t, limit, result.Pagination.PageSize)
		})

		t.Run("error cases and edge conditions", func(t *testing.T) {
			// Test create with duplicate unique field (if applicable)

			country1 := &models.Country{
				Name:  "Test Name",
				Email: "Test Email", // Duplicate unique value

			}
			err := helper.DB.Create(country1).Error
			assert.NoError(t, err)

			createReq := &models.CreateCountryRequest{
				Name:  "Test Name",
				Email: "Test Email", // Duplicate unique value

			}
			_, err = service.Create(createReq)
			assert.Error(t, err) // Should fail due to duplicate unique constraint

		})
	})
}
