package countries_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"base/app/countries"
	"base/app/models"
	"base/core/emitter"
	"base/core/router"
	"base/test"

	"github.com/stretchr/testify/assert"
)

func TestCountryControllers(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Auto-migrate Country table for testing
	err := helper.DB.AutoMigrate(&models.Country{})
	assert.NoError(t, err)

	// Create service and controller
	service := countries.NewCountryService(
		helper.DB,
		emitter.New(),
		nil, // storage can be nil for basic tests
		helper.Logger,
	)
	controller := countries.NewCountryController(service, nil)

	// Setup router
	testRouter := router.New()
	api := testRouter.Group("/api")
	controller.Routes(api)

	t.Run("Country controller operations comprehensive coverage", func(t *testing.T) {
		t.Run("Create Country", func(t *testing.T) {
			createReq := models.CreateCountryRequest{
				Name:  "Test Name",
				Email: "Test Email",
			}

			jsonData, _ := json.Marshal(createReq)
			req, _ := http.NewRequest("POST", "/api/countries", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response models.CountryResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotZero(t, response.Id)
			assert.Equal(t, "Test Name", response.Name)
			assert.Equal(t, "Test Email", response.Email)

		})

		t.Run("Get Country by ID", func(t *testing.T) {
			// Create test country
			country := &models.Country{
				Name:  "Test Name",
				Email: "Test Email",
			}
			err := helper.DB.Create(country).Error
			assert.NoError(t, err)

			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/countries/%d", country.Id), nil)
			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response models.CountryResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, country.Id, response.Id)
			assert.Equal(t, "Test Name", response.Name)
			assert.Equal(t, "Test Email", response.Email)

		})

		t.Run("Update Country", func(t *testing.T) {
			// Create test country
			country := &models.Country{
				Name:  "Test Name",
				Email: "Test Email",
			}
			err := helper.DB.Create(country).Error
			assert.NoError(t, err)

			updateReq := models.UpdateCountryRequest{
				Name:  "Updated Name",
				Email: "Updated Email",
			}

			jsonData, _ := json.Marshal(updateReq)
			req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/countries/%d", country.Id), bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response models.CountryResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, country.Id, response.Id)
			assert.Equal(t, "Updated Name", response.Name)
			assert.Equal(t, "Updated Email", response.Email)

		})

		t.Run("Delete Country", func(t *testing.T) {
			// Create test country
			country := &models.Country{
				Name:  "Test Name",
				Email: "Test Email",
			}
			err := helper.DB.Create(country).Error
			assert.NoError(t, err)

			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/countries/%d", country.Id), nil)
			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNoContent, w.Code)

			// Verify deletion
			var found models.Country
			err = helper.DB.First(&found, country.Id).Error
			assert.Error(t, err) // Should not find deleted record
		})

		t.Run("List Countries (paginated)", func(t *testing.T) {
			// Create multiple test countries
			for i := 0; i < 5; i++ {
				country := &models.Country{
					Name:  fmt.Sprintf("Test Name %d", i),
					Email: fmt.Sprintf("Test Email %d", i),
				}
				err := helper.DB.Create(country).Error
				assert.NoError(t, err)
			}

			req, _ := http.NewRequest("GET", "/api/countries?page=1&limit=3", nil)
			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response struct {
				Data       []models.CountryListResponse `json:"data"`
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

		t.Run("List All Countries (unpaginated)", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/countries/all", nil)
			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response []models.CountryListResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotNil(t, response)
		})

		t.Run("Error cases", func(t *testing.T) {
			// Test Get with invalid ID
			req, _ := http.NewRequest("GET", "/api/countries/99999", nil)
			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code)

			// Test Update with invalid ID
			updateReq := models.UpdateCountryRequest{
				Name:  "Updated Name",
				Email: "Updated Email",
			}
			jsonData, _ := json.Marshal(updateReq)
			req, _ = http.NewRequest("PUT", "/api/countries/99999", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w = httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code)

			// Test Delete with invalid ID
			req, _ = http.NewRequest("DELETE", "/api/countries/99999", nil)
			w = httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})
}
