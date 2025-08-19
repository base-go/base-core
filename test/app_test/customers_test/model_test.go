package customers_test

import (
	"testing"
	"time"

	"base/app/models"
	"base/test"

	"github.com/stretchr/testify/assert"
)

func TestCustomerModels(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Auto-migrate Customer table for testing
	err := helper.DB.AutoMigrate(&models.Customer{})
	assert.NoError(t, err)

	t.Run("Customer model operations comprehensive coverage", func(t *testing.T) {
		t.Run("create and basic operations", func(t *testing.T) {
			// Test Customer creation
			customer := &models.Customer{
				Name:  "Test Name",
				Email: "Test Email",
			}
			err := helper.DB.Create(customer).Error
			assert.NoError(t, err)
			assert.NotZero(t, customer.Id)
			assert.NotZero(t, customer.CreatedAt)
			assert.NotZero(t, customer.UpdatedAt)

			// Test find by ID
			var foundCustomer models.Customer
			err = helper.DB.First(&foundCustomer, customer.Id).Error
			assert.NoError(t, err)
			assert.Equal(t, customer.Id, foundCustomer.Id)
			assert.Equal(t, "Test Name", foundCustomer.Name)
			assert.Equal(t, "Test Email", foundCustomer.Email)

		})

		t.Run("Customer model methods", func(t *testing.T) {
			customer := &models.Customer{
				Name:  "Test Name",
				Email: "Test Email",
			}
			err := helper.DB.Create(customer).Error
			assert.NoError(t, err)

			// Test TableName
			assert.Equal(t, "customers", customer.TableName())

			// Test GetId
			assert.Equal(t, customer.Id, customer.GetId())

			// Test GetModelName
			assert.Equal(t, "customer", customer.GetModelName())

			// Test ToListResponse
			listResponse := customer.ToListResponse()
			assert.NotNil(t, listResponse)
			assert.Equal(t, customer.Id, listResponse.Id)
			assert.Equal(t, "Test Name", listResponse.Name)
			assert.Equal(t, "Test Email", listResponse.Email)

			// Test ToResponse
			response := customer.ToResponse()
			assert.NotNil(t, response)
			assert.Equal(t, customer.Id, response.Id)
			assert.Equal(t, "Test Name", response.Name)
			assert.Equal(t, "Test Email", response.Email)

			// Test Preload
			query := customer.Preload(helper.DB)
			assert.NotNil(t, query)
		})

		t.Run("Customer soft delete", func(t *testing.T) {
			customer := &models.Customer{
				Name:  "Test Name",
				Email: "Test Email",
			}
			err := helper.DB.Create(customer).Error
			assert.NoError(t, err)

			// Test timestamps
			assert.True(t, customer.CreatedAt.Before(time.Now().Add(time.Second)))
			assert.True(t, customer.UpdatedAt.Before(time.Now().Add(time.Second)))

			// Test soft delete
			err = helper.DB.Delete(customer).Error
			assert.NoError(t, err)

			// Should not find deleted record
			var foundCustomer models.Customer
			err = helper.DB.First(&foundCustomer, customer.Id).Error
			assert.Error(t, err)

			// Should find with Unscoped
			err = helper.DB.Unscoped().First(&foundCustomer, customer.Id).Error
			assert.NoError(t, err)
			assert.Equal(t, customer.Id, foundCustomer.Id)
			assert.NotZero(t, foundCustomer.DeletedAt)
		})
	})
}
