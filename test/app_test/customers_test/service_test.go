package customers_test

import (
	"fmt"
	"testing"

	"base/app/customers"
	"base/app/models"
	"base/core/emitter"
	"base/test"

	"github.com/stretchr/testify/assert"
)

func TestCustomerServices(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Auto-migrate Customer table for testing
	err := helper.DB.AutoMigrate(&models.Customer{})
	assert.NoError(t, err)

	// Create service with proper emitter and nil storage
	service := customers.NewCustomerService(
		helper.DB,
		emitter.New(), // proper emitter
		nil,           // storage (can be nil)
		helper.Logger,
	)

	t.Run("Customer service operations comprehensive coverage", func(t *testing.T) {
		t.Run("Service creation", func(t *testing.T) {
			// Test that service is created successfully
			assert.NotNil(t, service)
		})

		t.Run("Basic CRUD operations", func(t *testing.T) {
			// Test Create operation
			createReq := &models.CreateCustomerRequest{
				Name:  "Test Name",
				Email: "Test Email",
			}

			customer, err := service.Create(createReq)
			assert.NoError(t, err)
			assert.NotNil(t, customer)
			assert.NotZero(t, customer.Id)
			assert.Equal(t, "Test Name", customer.Name)
			assert.Equal(t, "Test Email", customer.Email)

			// Test GetById operation
			found, err := service.GetById(customer.Id)
			assert.NoError(t, err)
			assert.NotNil(t, found)
			assert.Equal(t, customer.Id, found.Id)

			// Test Update operation
			updateReq := &models.UpdateCustomerRequest{
				Name:  "Updated Name",
				Email: "Updated Email",
			}

			updated, err := service.Update(customer.Id, updateReq)
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
			err = service.Delete(customer.Id)
			assert.NoError(t, err)

			// Verify deletion
			_, err = service.GetById(customer.Id)
			assert.Error(t, err) // Should not find deleted record
		})

		t.Run("get operations", func(t *testing.T) {
			// Create test customer
			customer := &models.Customer{
				Name:    "Test Name",
				Email:   "Test Email",
				Country: nil,
			}
			err := helper.DB.Create(customer).Error
			assert.NoError(t, err)

			// Test GetById
			found, err := service.GetById(customer.Id)
			assert.NoError(t, err)
			assert.NotNil(t, found)
			assert.Equal(t, customer.Id, found.Id)

			// Test GetById with invalid ID
			_, err = service.GetById(0)
			assert.Error(t, err)

			// Test GetById with non-existent ID
			_, err = service.GetById(99999)
			assert.Error(t, err)
		})

		t.Run("update operations", func(t *testing.T) {
			// Create test customer
			customer := &models.Customer{
				Name:    "Test Name",
				Email:   "Test Email",
				Country: nil,
			}
			err := helper.DB.Create(customer).Error
			assert.NoError(t, err)

			// Test update
			updateReq := &models.UpdateCustomerRequest{
				Name:  "Updated Name",
				Email: "Updated Email",
			}

			updated, err := service.Update(customer.Id, updateReq)
			assert.NoError(t, err)
			assert.NotNil(t, updated)
			assert.Equal(t, "Updated Name", updated.Name)
			assert.Equal(t, "Updated Email", updated.Email)
			assert.Equal(t, nil, updated.Country)

			// Test update with invalid ID
			_, err = service.Update(0, updateReq)
			assert.Error(t, err)

			// Test update with non-existent ID
			_, err = service.Update(99999, updateReq)
			assert.Error(t, err)
		})

		t.Run("delete operations", func(t *testing.T) {
			// Create test customer
			customer := &models.Customer{
				Name:    "Test Name",
				Email:   "Test Email",
				Country: nil,
			}
			err := helper.DB.Create(customer).Error
			assert.NoError(t, err)

			// Test delete
			err = service.Delete(customer.Id)
			assert.NoError(t, err)

			// Verify deletion
			_, err = service.GetById(customer.Id)
			assert.Error(t, err)

			// Test delete with invalid ID
			err = service.Delete(0)
			assert.Error(t, err)

			// Test delete with non-existent ID
			err = service.Delete(99999)
			assert.Error(t, err)
		})

		t.Run("get all operations", func(t *testing.T) {
			// Create multiple test customers
			for i := 0; i < 5; i++ {
				customer := &models.Customer{
					Name:    fmt.Sprintf("Test Name %d", i),
					Email:   fmt.Sprintf("Test Email %d", i),
					Country: nil,
				}
				err := helper.DB.Create(customer).Error
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

			customer1 := &models.Customer{
				Name:      "Test Name",
				Email:     "Test Email", // Duplicate unique value
				CountryId: 1,
			}
			err := helper.DB.Create(customer1).Error
			assert.NoError(t, err)

			createReq := &models.CreateCustomerRequest{
				Name:      "Test Name",
				Email:     "Test Email", // Duplicate unique value
				CountryId: 1,
			}
			_, err = service.Create(createReq)
			assert.Error(t, err) // Should fail due to duplicate unique constraint

		})
	})
}
