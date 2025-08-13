package categories_test

import (
	"testing"
	"time"

	"base/app/models"
	"base/test"

	"github.com/stretchr/testify/assert"
)

func TestCategoryModels(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Auto-migrate Category table for testing
	err := helper.DB.AutoMigrate(&models.Category{})
	assert.NoError(t, err)

	t.Run("Category model operations comprehensive coverage", func(t *testing.T) {
		t.Run("create and basic operations", func(t *testing.T) {
			// Test Category creation
			category := &models.Category{
				Name: "Test Name",
				Description: "Test Description",
				Translation: nil,
				IsActive: true,
				
			}
			err := helper.DB.Create(category).Error
			assert.NoError(t, err)
			assert.NotZero(t, category.Id)
			assert.NotZero(t, category.CreatedAt)
			assert.NotZero(t, category.UpdatedAt)

			// Test find by ID
			var foundCategory models.Category
			err = helper.DB.First(&foundCategory, category.Id).Error
			assert.NoError(t, err)
			assert.Equal(t, category.Id, foundCategory.Id)
			assert.Equal(t, "Test Name", foundCategory.Name)
			assert.Equal(t, "Test Description", foundCategory.Description)
			assert.Equal(t, nil, foundCategory.Translation)
			assert.Equal(t, true, foundCategory.IsActive)
			
		})

		t.Run("Category model methods", func(t *testing.T) {
			category := &models.Category{
				Name: "Test Name",
				Description: "Test Description",
				Translation: nil,
				IsActive: true,
				
			}
			err := helper.DB.Create(category).Error
			assert.NoError(t, err)

			// Test TableName
			assert.Equal(t, "categories", category.TableName())

			// Test GetId
			assert.Equal(t, category.Id, category.GetId())

			// Test GetModelName
			assert.Equal(t, "category", category.GetModelName())

			// Test ToListResponse
			listResponse := category.ToListResponse()
			assert.NotNil(t, listResponse)
			assert.Equal(t, category.Id, listResponse.Id)
			assert.Equal(t, "Test Name", listResponse.Name)
			assert.Equal(t, "Test Description", listResponse.Description)
			assert.Equal(t, nil, listResponse.Translation)
			assert.Equal(t, true, listResponse.IsActive)
			

			// Test ToResponse
			response := category.ToResponse()
			assert.NotNil(t, response)
			assert.Equal(t, category.Id, response.Id)
			assert.Equal(t, "Test Name", response.Name)
			assert.Equal(t, "Test Description", response.Description)
			assert.Equal(t, nil, response.Translation)
			assert.Equal(t, true, response.IsActive)
			

			// Test Preload
			query := category.Preload(helper.DB)
			assert.NotNil(t, query)
		})

		t.Run("Category soft delete", func(t *testing.T) {
			category := &models.Category{
				Name: "Test Name",
				Description: "Test Description",
				Translation: nil,
				IsActive: true,
				
			}
			err := helper.DB.Create(category).Error
			assert.NoError(t, err)

			// Test timestamps
			assert.True(t, category.CreatedAt.Before(time.Now().Add(time.Second)))
			assert.True(t, category.UpdatedAt.Before(time.Now().Add(time.Second)))

			// Test soft delete
			err = helper.DB.Delete(category).Error
			assert.NoError(t, err)

			// Should not find deleted record
			var foundCategory models.Category
			err = helper.DB.First(&foundCategory, category.Id).Error
			assert.Error(t, err)

			// Should find with Unscoped
			err = helper.DB.Unscoped().First(&foundCategory, category.Id).Error
			assert.NoError(t, err)
			assert.Equal(t, category.Id, foundCategory.Id)
			assert.NotZero(t, foundCategory.DeletedAt)
		})

		t.Run("validation of required fields", func(t *testing.T) {
			category := &models.Category{}
			err := helper.DB.Create(category).Error
			