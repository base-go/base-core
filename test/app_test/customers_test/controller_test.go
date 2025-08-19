package customers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"base/app/customers"
	"base/app/models"
	"base/core/emitter"
	"base/core/router"
	"base/test"

	"github.com/stretchr/testify/assert"
)

func TestCustomerControllers(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Auto-migrate Customer table for testing
	err := helper.DB.AutoMigrate(&models.Customer{})
	assert.NoError(t, err)

	// Create service and controller
	service := customers.NewCustomerService(
		helper.DB,
		emitter.New(),
		nil, // storage can be nil for basic tests
		helper.Logger,
	)
	controller := customers.NewCustomerController(service, nil)

	// Setup router
	testRouter := router.New()
	api := testRouter.Group("/api")
	controller.Routes(api)

	t.Run("Customer controller operations comprehensive coverage", func(t *testing.T) {
		t.Run("Create Customer", func(t *testing.T) {
			createReq := models.CreateCustomerRequest{
				Name:      "Test Name",
				Email:     "Test Email",
				CountryId: 1,
			}

			jsonData, _ := json.Marshal(createReq)
			req, _ := http.NewRequest("POST", "/api/customers", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response models.CustomerResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotZero(t, response.Id)
			assert.Equal(t, "Test Name", response.Name)
			assert.Equal(t, "Test Email", response.Email)

		})

		t.Run("Get Customer by ID", func(t *testing.T) {
			// Create test customer
			customer := &models.Customer{
				Name:  "Test Name",
				Email: "Test Email",
			}
			err := helper.DB.Create(customer).Error
			assert.NoError(t, err)

			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/customers/%d", customer.Id), nil)
			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response models.CustomerResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, customer.Id, response.Id)
			assert.Equal(t, "Test Name", response.Name)
			assert.Equal(t, "Test Email", response.Email)

		})

		t.Run("Update Customer", func(t *testing.T) {
			// Create test customer
			customer := &models.Customer{
				Name:  "Test Name",
				Email: "Test Email",
			}
			err := helper.DB.Create(customer).Error
			assert.NoError(t, err)

			updateReq := models.UpdateCustomerRequest{
				Name:      "Updated Name",
				Email:     "Updated Email",
				CountryId: 2,
			}

			jsonData, _ := json.Marshal(updateReq)
			req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/customers/%d", customer.Id), bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response models.CustomerResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, customer.Id, response.Id)
			assert.Equal(t, "Updated Name", response.Name)
			assert.Equal(t, "Updated Email", response.Email)

		})

		t.Run("Delete Customer", func(t *testing.T) {
			// Create test customer
			customer := &models.Customer{
				Name:  "Test Name",
				Email: "Test Email",
			}
			err := helper.DB.Create(customer).Error
			assert.NoError(t, err)

			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/customers/%d", customer.Id), nil)
			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNoContent, w.Code)

			// Verify deletion
			var found models.Customer
			err = helper.DB.First(&found, customer.Id).Error
			assert.Error(t, err) // Should not find deleted record
		})

		t.Run("List Customers (paginated)", func(t *testing.T) {
			// Create multiple test customers
			for i := 0; i < 5; i++ {
				customer := &models.Customer{
					Name:  fmt.Sprintf("Test Name %d", i),
					Email: fmt.Sprintf("Test Email %d", i),
				}
				err := helper.DB.Create(customer).Error
				assert.NoError(t, err)
			}

			req, _ := http.NewRequest("GET", "/api/customers?page=1&limit=3", nil)
			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response struct {
				Data       []models.CustomerListResponse `json:"data"`
				Pagination struct {
					Page     int `json:"page"`
					PageSize int `json:"page_size"`
					Total    int `json:"total"`
				} `json:"pagination"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotEmpty(t, response.Data)
			assert.LessOrEqual(t, len(response.Data), 3)
			assert.Equal(t, 1, response.Pagination.Page)
			assert.Equal(t, 3, response.Pagination.PageSize)
		})

		t.Run("List All Customers (unpaginated)", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/customers/all", nil)
			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response []models.CustomerListResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotNil(t, response)
		})

		t.Run("Error cases", func(t *testing.T) {
			// Test Get with invalid ID
			req, _ := http.NewRequest("GET", "/api/customers/99999", nil)
			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code)

			// Test Update with invalid ID
			updateReq := models.UpdateCustomerRequest{
				Name:  "Updated Name",
				Email: "Updated Email",
			}
			jsonData, _ := json.Marshal(updateReq)
			req, _ = http.NewRequest("PUT", "/api/customers/99999", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w = httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code)

			// Test Delete with invalid ID
			req, _ = http.NewRequest("DELETE", "/api/customers/99999", nil)
			w = httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})
}
