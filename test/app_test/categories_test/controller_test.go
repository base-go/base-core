package categories_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"base/app/categories"
	"base/app/models"
	"base/core/emitter"
	"base/test"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCategoryControllers(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Auto-migrate Category table for testing
	err := helper.DB.AutoMigrate(&models.Category{})
	assert.NoError(t, err)

	// Create service and controller
	service := categories.NewCategoryService(
		helper.DB,
		emitter.New(),
		nil, // storage can be nil for basic tests
		helper.Logger,
	)
	controller := categories.NewCategoryController(service, nil)

	// Setup router
 
	router := router.New()
	api := router.Group("/api")
	controller.Routes(api)

	t.Run("Category controller operations comprehensive coverage", func(t *testing.T) {
		t.Run("Create Category", func(t *testing.T) {
			createReq := models.CreateCategoryRequest{
				Name: "Test Name",
				Description: "Test Description",
				Translation: nil,
				IsActive: true,
				
			}

			jsonData, _ := json.Marshal(createReq)
			req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response models.CategoryResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotZero(t, response.Id)
			assert.Equal(t, "Test Name", response.Name)
			assert.Equal(t, "Test Description", response.Description)
			assert.Equal(t, nil, response.Translation)
			assert.Equal(t, true, response.IsActive)
			
		})

		t.Run("Get Category by ID", func(t *testing.T) {
			// Create test category
			category := &models.Category{
				Name: "Test Name",
				Description: "Test Description",
				Translation: nil,
				IsActive: true,
				
			}
			err := helper.DB.Create(category).Error
			assert.NoError(t, err)

			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/categories/%d", category.Id), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response models.CategoryResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, category.Id, response.Id)
			assert.Equal(t, "Test Name", response.Name)
			assert.Equal(t, "Test Description", response.Description)
			assert.Equal(t, nil, response.Translation)
			assert.Equal(t, true, response.IsActive)
			
		})

		t.Run("Update Category", func(t *testing.T) {
			// Create test category
			category := &models.Category{
				Name: "Test Name",
				Description: "Test Description",
				Translation: nil,
				IsActive: true,
				
			}
			err := helper.DB.Create(category).Error
			assert.NoError(t, err)

			updateReq := models.UpdateCategoryRequest{
				Name: "Updated Name",
				Description: "Updated Description",
				Translation: nil,
				IsActive: false,
				
			}

			jsonData, _ := json.Marshal(updateReq)
			req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/categories/%d", category.Id), bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response models.CategoryResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, category.Id, response.Id)
			assert.Equal(t, "Updated Name", response.Name)
			assert.Equal(t, "Updated Description", response.Description)
			assert.Equal(t, nil, response.Translation)
			assert.Equal(t, false, response.IsActive)
			
		})

		t.Run("Delete Category", func(t *testing.T) {
			// Create test category
			category := &models.Category{
				Name: "Test Name",
				Description: "Test Description",
				Translation: nil,
				IsActive: true,
				
			}
			err := helper.DB.Create(category).Error
			assert.NoError(t, err)

			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/categories/%d", category.Id), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNoContent, w.Code)

			// Verify deletion
			var found models.Category
			err = helper.DB.First(&found, category.Id).Error
			assert.Error(t, err) // Should not find deleted record
		})

		t.Run("List Categories (paginated)", func(t *testing.T) {
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

			req, _ := http.NewRequest("GET", "/api/categories?page=1&limit=3", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response struct {
				Data       []models.CategoryListResponse `json:"data"`
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

		t.Run("List All Categories (unpaginated)", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/categories/all", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response []models.CategoryListResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotNil(t, response)
		})

		t.Run("Error cases", func(t *testing.T) {
			// Test Get with invalid ID
			req, _ := http.NewRequest("GET", "/api/categories/99999", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code)

			// Test Update with invalid ID
			updateReq := models.UpdateCategoryRequest{
				Name: "Updated Name",
				Description: "Updated Description",
				Translation: nil,
				IsActive: false,
				
			}
			jsonData, _ := json.Marshal(updateReq)
			req, _ = http.NewRequest("PUT", "/api/categories/99999", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code)

			// Test Delete with invalid ID
			req, _ = http.NewRequest("DELETE", "/api/categories/99999", nil)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})
}
