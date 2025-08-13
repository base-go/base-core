package categories_test

import (
	"testing"

	"base/app/categories"
	"base/app/models"
	"base/core/emitter"
	"base/test"

	"github.com/stretchr/testify/assert"
)

func TestCategoryServices(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Auto-migrate Category table for testing
	err := helper.DB.AutoMigrate(&models.Category{})
	assert.NoError(t, err)

	// Create service with proper emitter and nil storage
	service := categories.NewCategoryService(
		helper.DB,
		emitter.New(), // proper emitter
		nil,            // storage (can be nil)
		helper.Logger,
	)

	t.Run("Category service operations comprehensive coverage", func(t *testing.T) {
		t.Run("Service creation", func(t *testing.T) {
			// Test that service is created successfully
			assert.NotNil(t, service)
		})

		t.Run("Basic CRUD operations", func(t *testing.T) {
			// Test Create operation
			createReq := &models.CreateCategoryRequest{
				Name: "Test Name",
				Description: "Test Description",
				Translation: nil,
				IsActive: true,
				
			}

			category, err := service.Create(createReq)
			assert.NoError(t, err)
			assert.NotNil(t, category)
			assert.NotZero(t, category.Id)
			assert.Equal(t, "Test Name", category.Name)
			assert.Equal(t, "Test Description", category.Description)
			assert.Equal(t, nil, category.Translation)
			assert.Equal(t, true, category.IsActive)
			

			// Test GetById operation
			found, err := service.GetById(category.Id)
			assert.NoError(t, err)
			assert.NotNil(t, found)
			assert.Equal(t, category.Id, found.Id)

			// Test Update operation
			updateReq := &models.UpdateCategoryRequest{
				Name: "Updated Name",
				Description: "Updated Description",
				Translation: nil,
				IsActive: false,
				
			}

			updated, err := service.Update(category.Id, updateReq)
			assert.NoError(t, err)
			assert.NotNil(t, updated)
			assert.Equal(t, "Updated Name", updated.Name)
			assert.Equal(t, "Updated Description", updated.Description)
			assert.Equal(t, nil, updated.Translation)
			assert.Equal(t, false, updated.IsActive)
			

			// Test GetAll operation
			result, err := service.GetAll(nil, nil)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotNil(t, result.Data)

			// Test Delete operation
			err = service.Delete(category.Id)
			assert.NoError(t, err)

			// Verify deletion
			_, err = service.GetById(category.Id)
			assert.Error(t, err) // Should not find deleted record
		})

		t.Run("get operations", func(t *testing.T) {
			// Create test category
			category := &models.Category{
				Name: "Test Name",
				Description: "Test Description",
				Translation: nil,
				IsActive: true,
				
			}
			err := helper.DB.Create(category).Error
			assert.NoError(t, err)

			// Test GetById
			found, err := service.GetById(category.Id)
			assert.NoError(t, err)
			assert.NotNil(t, found)
			assert.Equal(t, category.Id, found.Id)

			// Test GetById with invalid ID
			_, err = service.GetById(0)
			assert.Error(t, err)

			// Test GetById with non-existent ID
			_, err = service.GetById(99999)
			assert.Error(t, err)
		})

		t.Run("update operations", func(t *testing.T) {
			// Create test category
			category := &models.Category{
				Name: "Test Name",
				Description: "Test Description",
				Translation: nil,
				IsActive: true,
				
			}
			err := helper.DB.Create(category).Error
			assert.NoError(t, err)

			// Test update
			updateReq := &models.UpdateCategoryRequest{
				Id: category.Id,
				Name: "Updated Name",
				Description: "Updated Description",
				Translation: nil,
				IsActive: false,
				
			}

			updated, err := service.Update(category.Id, updateReq)
			assert.NoError(t, err)
			assert.NotNil(t, updated)
			assert.Equal(t, "Updated Name", updated.Name)
			assert.Equal(t, "Updated Description", updated.Description)
			assert.Equal(t, nil, updated.Translation)
			assert.Equal(t, false, updated.IsActive)
			

			// Test update with invalid ID
			_, err = service.Update(0, updateReq)
			assert.Error(t, err)

			// Test update with non-existent ID
			_, err = service.Update(99999, updateReq)
			assert.Error(t, err)
		})

		t.Run("delete operations", func(t *testing.T) {
			// Create test category
			category := &models.Category{
				Name: "Test Name",
				Description: "Test Description",
				Translation: nil,
				IsActive: true,
				
			}
			err := helper.DB.Create(category).Error
			assert.NoError(t, err)

			// Test delete
			err = service.Delete(category.Id)
			assert.NoError(t, err)

			// Verify deletion
			_, err = service.GetById(category.Id)
			assert.Error(t, err)

			// Test delete with invalid ID
			err = service.Delete(0)
			assert.Error(t, err)

			// Test delete with non-existent ID
			err = service.Delete(99999)
			assert.Error(t, err)
		})

		t.Run("get all operations", func(t *testing.T) {
			// Create multiple test categories
			for i := 0; i < 5; i++ {
				category := &models.Category{
					Name: fmt.Sprintf("Test Name %d", i),
					Description: fmt.Sprintf("Test Description %d", i),
					Translation: nil,
					IsActive: (i%2 == 0),
					
				}
				err := helper.DB.Create(category).Error
				assert.NoError(t, err)
			}

			// Test GetAll without pagination
			result, err := service.GetAll(nil, nil)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotEmpty(t, result.Data)

			// Test GetAll with pagination
			page := 1
			limit := 2
			result, err = service.GetAll(&page, &limit)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.LessOrEqual(t, len(result.Data), limit)
			assert.Equal(t, page, result.Pagination.Page)
			assert.Equal(t, limit, result.Pagination.PageSize)
		})

		t.Run("error cases and edge conditions", func(t *testing.T) {
			// Test create with duplicate unique field (if applicable)
			
		})
	})
}
