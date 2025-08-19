package countries_test

import (
	"testing"
	"time"

	"base/app/models"
	"base/test"

	"github.com/stretchr/testify/assert"
)

func TestCountryModels(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Auto-migrate Country table for testing
	err := helper.DB.AutoMigrate(&models.Country{})
	assert.NoError(t, err)

	t.Run("Country model operations comprehensive coverage", func(t *testing.T) {
		t.Run("create and basic operations", func(t *testing.T) {
			// Test Country creation
			country := &models.Country{
				Name:  "Test Name",
				Email: "Test Email",
			}
			err := helper.DB.Create(country).Error
			assert.NoError(t, err)
			assert.NotZero(t, country.Id)
			assert.NotZero(t, country.CreatedAt)
			assert.NotZero(t, country.UpdatedAt)

			// Test find by ID
			var foundCountry models.Country
			err = helper.DB.First(&foundCountry, country.Id).Error
			assert.NoError(t, err)
			assert.Equal(t, country.Id, foundCountry.Id)
			assert.Equal(t, "Test Name", foundCountry.Name)
			assert.Equal(t, "Test Email", foundCountry.Email)

		})

		t.Run("Country model methods", func(t *testing.T) {
			country := &models.Country{
				Name:  "Test Name",
				Email: "Test Email",
			}
			err := helper.DB.Create(country).Error
			assert.NoError(t, err)

			// Test TableName
			assert.Equal(t, "countries", country.TableName())

			// Test GetId
			assert.Equal(t, country.Id, country.GetId())

			// Test GetModelName
			assert.Equal(t, "country", country.GetModelName())

			// Test ToListResponse
			listResponse := country.ToListResponse()
			assert.NotNil(t, listResponse)
			assert.Equal(t, country.Id, listResponse.Id)
			assert.Equal(t, "Test Name", listResponse.Name)
			assert.Equal(t, "Test Email", listResponse.Email)

			// Test ToResponse
			response := country.ToResponse()
			assert.NotNil(t, response)
			assert.Equal(t, country.Id, response.Id)
			assert.Equal(t, "Test Name", response.Name)
			assert.Equal(t, "Test Email", response.Email)

			// Test Preload
			query := country.Preload(helper.DB)
			assert.NotNil(t, query)
		})

		t.Run("Country soft delete", func(t *testing.T) {
			country := &models.Country{
				Name:  "Test Name",
				Email: "Test Email",
			}
			err := helper.DB.Create(country).Error
			assert.NoError(t, err)

			// Test timestamps
			assert.True(t, country.CreatedAt.Before(time.Now().Add(time.Second)))
			assert.True(t, country.UpdatedAt.Before(time.Now().Add(time.Second)))

			// Test soft delete
			err = helper.DB.Delete(country).Error
			assert.NoError(t, err)

			// Should not find deleted record
			var foundCountry models.Country
			err = helper.DB.First(&foundCountry, country.Id).Error
			assert.Error(t, err)

			// Should find with Unscoped
			err = helper.DB.Unscoped().First(&foundCountry, country.Id).Error
			assert.NoError(t, err)
			assert.Equal(t, country.Id, foundCountry.Id)
			assert.NotZero(t, foundCountry.DeletedAt)
		})
	})
}
