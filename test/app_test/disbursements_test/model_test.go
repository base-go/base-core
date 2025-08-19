package disbursements_test

import (
	"testing"
	"time"

	"base/app/models"
	"base/test"

	"github.com/stretchr/testify/assert"
)

func TestDisbursementModels(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Auto-migrate Disbursement table for testing
	err := helper.DB.AutoMigrate(&models.Disbursement{})
	assert.NoError(t, err)

	t.Run("Disbursement model operations comprehensive coverage", func(t *testing.T) {
		t.Run("create and basic operations", func(t *testing.T) {
			// Test Disbursement creation
			disbursement := &models.Disbursement{
				Amount:      123.45,
				Description: "Test Description",
			}
			err := helper.DB.Create(disbursement).Error
			assert.NoError(t, err)
			assert.NotZero(t, disbursement.Id)
			assert.NotZero(t, disbursement.CreatedAt)
			assert.NotZero(t, disbursement.UpdatedAt)

			// Test find by ID
			var foundDisbursement models.Disbursement
			err = helper.DB.First(&foundDisbursement, disbursement.Id).Error
			assert.NoError(t, err)
			assert.Equal(t, disbursement.Id, foundDisbursement.Id)
			assert.Equal(t, 123.45, foundDisbursement.Amount)
			assert.Equal(t, "Test Description", foundDisbursement.Description)

		})

		t.Run("Disbursement model methods", func(t *testing.T) {
			disbursement := &models.Disbursement{
				Amount:      123.45,
				Description: "Test Description",
			}
			err := helper.DB.Create(disbursement).Error
			assert.NoError(t, err)

			// Test TableName
			assert.Equal(t, "disbursements", disbursement.TableName())

			// Test GetId
			assert.Equal(t, disbursement.Id, disbursement.GetId())

			// Test GetModelName
			assert.Equal(t, "disbursement", disbursement.GetModelName())

			// Test ToListResponse
			listResponse := disbursement.ToListResponse()
			assert.NotNil(t, listResponse)
			assert.Equal(t, disbursement.Id, listResponse.Id)
			assert.Equal(t, 123.45, listResponse.Amount)
			assert.Equal(t, "Test Description", listResponse.Description)

			// Test ToResponse
			response := disbursement.ToResponse()
			assert.NotNil(t, response)
			assert.Equal(t, disbursement.Id, response.Id)
			assert.Equal(t, 123.45, response.Amount)
			assert.Equal(t, "Test Description", response.Description)

			// Test Preload
			query := disbursement.Preload(helper.DB)
			assert.NotNil(t, query)
		})

		t.Run("Disbursement soft delete", func(t *testing.T) {
			disbursement := &models.Disbursement{
				Amount:      123.45,
				Description: "Test Description",
			}
			err := helper.DB.Create(disbursement).Error
			assert.NoError(t, err)

			// Test timestamps
			assert.True(t, disbursement.CreatedAt.Before(time.Now().Add(time.Second)))
			assert.True(t, disbursement.UpdatedAt.Before(time.Now().Add(time.Second)))

			// Test soft delete
			err = helper.DB.Delete(disbursement).Error
			assert.NoError(t, err)

			// Should not find deleted record
			var foundDisbursement models.Disbursement
			err = helper.DB.First(&foundDisbursement, disbursement.Id).Error
			assert.Error(t, err)

			// Should find with Unscoped
			err = helper.DB.Unscoped().First(&foundDisbursement, disbursement.Id).Error
			assert.NoError(t, err)
			assert.Equal(t, disbursement.Id, foundDisbursement.Id)
			assert.NotZero(t, foundDisbursement.DeletedAt)
		})
	})
}
