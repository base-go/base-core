package disbursements_test

import (
	"fmt"
	"testing"

	"base/app/disbursements"
	"base/app/models"
	"base/core/emitter"
	"base/test"

	"github.com/stretchr/testify/assert"
)

func TestDisbursementServices(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Auto-migrate Disbursement table for testing
	err := helper.DB.AutoMigrate(&models.Disbursement{})
	assert.NoError(t, err)

	// Create service with proper emitter and nil storage
	service := disbursements.NewDisbursementService(
		helper.DB,
		emitter.New(), // proper emitter
		nil,           // storage (can be nil)
		helper.Logger,
	)

	t.Run("Disbursement service operations comprehensive coverage", func(t *testing.T) {
		t.Run("Service creation", func(t *testing.T) {
			// Test that service is created successfully
			assert.NotNil(t, service)
		})

		t.Run("Basic CRUD operations", func(t *testing.T) {
			// Test Create operation
			createReq := &models.CreateDisbursementRequest{
				Amount:      123.45,
				Description: "Test Description",
			}

			disbursement, err := service.Create(createReq)
			assert.NoError(t, err)
			assert.NotNil(t, disbursement)
			assert.NotZero(t, disbursement.Id)
			assert.Equal(t, 123.45, disbursement.Amount)
			assert.Equal(t, "Test Description", disbursement.Description)

			// Test GetById operation
			found, err := service.GetById(disbursement.Id)
			assert.NoError(t, err)
			assert.NotNil(t, found)
			assert.Equal(t, disbursement.Id, found.Id)

			// Test Update operation
			updateReq := &models.UpdateDisbursementRequest{
				Amount:      678.90,
				Description: "Updated Description",
			}

			updated, err := service.Update(disbursement.Id, updateReq)
			assert.NoError(t, err)
			assert.NotNil(t, updated)
			assert.Equal(t, 678.90, updated.Amount)
			assert.Equal(t, "Updated Description", updated.Description)

			// Test GetAll operation
			result, err := service.GetAll(nil, nil, nil, nil)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotNil(t, result.Data)

			// Test Delete operation
			err = service.Delete(disbursement.Id)
			assert.NoError(t, err)

			// Verify deletion
			_, err = service.GetById(disbursement.Id)
			assert.Error(t, err) // Should not find deleted record
		})

		t.Run("get operations", func(t *testing.T) {
			// Create test disbursement
			disbursement := &models.Disbursement{
				Amount:      123.45,
				Description: "Test Description",
			}
			err := helper.DB.Create(disbursement).Error
			assert.NoError(t, err)

			// Test GetById
			found, err := service.GetById(disbursement.Id)
			assert.NoError(t, err)
			assert.NotNil(t, found)
			assert.Equal(t, disbursement.Id, found.Id)

			// Test GetById with invalid ID
			_, err = service.GetById(0)
			assert.Error(t, err)

			// Test GetById with non-existent ID
			_, err = service.GetById(99999)
			assert.Error(t, err)
		})

		t.Run("update operations", func(t *testing.T) {
			// Create test disbursement
			disbursement := &models.Disbursement{
				Amount:      123.45,
				Description: "Test Description",
			}
			err := helper.DB.Create(disbursement).Error
			assert.NoError(t, err)

			// Test update
			updateReq := &models.UpdateDisbursementRequest{
				Amount:      678.90,
				Description: "Updated Description",
			}

			updated, err := service.Update(disbursement.Id, updateReq)
			assert.NoError(t, err)
			assert.NotNil(t, updated)
			assert.Equal(t, 678.90, updated.Amount)
			assert.Equal(t, "Updated Description", updated.Description)

			// Test update with invalid ID
			_, err = service.Update(0, updateReq)
			assert.Error(t, err)

			// Test update with non-existent ID
			_, err = service.Update(99999, updateReq)
			assert.Error(t, err)
		})

		t.Run("delete operations", func(t *testing.T) {
			// Create test disbursement
			disbursement := &models.Disbursement{
				Amount:      123.45,
				Description: "Test Description",
			}
			err := helper.DB.Create(disbursement).Error
			assert.NoError(t, err)

			// Test delete
			err = service.Delete(disbursement.Id)
			assert.NoError(t, err)

			// Verify deletion
			_, err = service.GetById(disbursement.Id)
			assert.Error(t, err)

			// Test delete with invalid ID
			err = service.Delete(0)
			assert.Error(t, err)

			// Test delete with non-existent ID
			err = service.Delete(99999)
			assert.Error(t, err)
		})

		t.Run("get all operations", func(t *testing.T) {
			// Create multiple test disbursements
			for i := 0; i < 5; i++ {
				disbursement := &models.Disbursement{
					Amount:      float64(100.5 + float64(i)),
					Description: fmt.Sprintf("Test Description %d", i),
				}
				err := helper.DB.Create(disbursement).Error
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

		})
	})
}
