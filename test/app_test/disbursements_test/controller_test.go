package disbursements_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"base/app/disbursements"
	"base/app/models"
	"base/core/emitter"
	"base/core/router"
	"base/test"

	"github.com/stretchr/testify/assert"
)

func TestDisbursementControllers(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Auto-migrate Disbursement table for testing
	err := helper.DB.AutoMigrate(&models.Disbursement{})
	assert.NoError(t, err)

	// Create service and controller
	service := disbursements.NewDisbursementService(
		helper.DB,
		emitter.New(),
		nil, // storage can be nil for basic tests
		helper.Logger,
	)
	controller := disbursements.NewDisbursementController(service, nil)

	// Setup router
	testRouter := router.New()
	api := testRouter.Group("/api")
	controller.Routes(api)

	t.Run("Disbursement controller operations comprehensive coverage", func(t *testing.T) {
		t.Run("Create Disbursement", func(t *testing.T) {
			createReq := models.CreateDisbursementRequest{
				Amount:      123.45,
				Description: "Test Description",
			}

			jsonData, _ := json.Marshal(createReq)
			req, _ := http.NewRequest("POST", "/api/disbursements", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response models.DisbursementResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotZero(t, response.Id)
			assert.Equal(t, 123.45, response.Amount)
			assert.Equal(t, "Test Description", response.Description)

		})

		t.Run("Get Disbursement by ID", func(t *testing.T) {
			// Create test disbursement
			disbursement := &models.Disbursement{
				Amount:      123.45,
				Description: "Test Description",
			}
			err := helper.DB.Create(disbursement).Error
			assert.NoError(t, err)

			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/disbursements/%d", disbursement.Id), nil)
			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response models.DisbursementResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, disbursement.Id, response.Id)
			assert.Equal(t, 123.45, response.Amount)
			assert.Equal(t, "Test Description", response.Description)

		})

		t.Run("Update Disbursement", func(t *testing.T) {
			// Create test disbursement
			disbursement := &models.Disbursement{
				Amount:      123.45,
				Description: "Test Description",
			}
			err := helper.DB.Create(disbursement).Error
			assert.NoError(t, err)

			updateReq := models.UpdateDisbursementRequest{
				Amount:      678.90,
				Description: "Updated Description",
			}

			jsonData, _ := json.Marshal(updateReq)
			req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/disbursements/%d", disbursement.Id), bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response models.DisbursementResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, disbursement.Id, response.Id)
			assert.Equal(t, 678.90, response.Amount)
			assert.Equal(t, "Updated Description", response.Description)

		})

		t.Run("Delete Disbursement", func(t *testing.T) {
			// Create test disbursement
			disbursement := &models.Disbursement{
				Amount:      123.45,
				Description: "Test Description",
			}
			err := helper.DB.Create(disbursement).Error
			assert.NoError(t, err)

			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/disbursements/%d", disbursement.Id), nil)
			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNoContent, w.Code)

			// Verify deletion
			var found models.Disbursement
			err = helper.DB.First(&found, disbursement.Id).Error
			assert.Error(t, err) // Should not find deleted record
		})

		t.Run("List Disbursements (paginated)", func(t *testing.T) {
			// Create multiple test disbursements
			for i := 0; i < 5; i++ {
				disbursement := &models.Disbursement{
					Amount:      float64(100.5 + float64(i)),
					Description: fmt.Sprintf("Test Description %d", i),
				}
				err := helper.DB.Create(disbursement).Error
				assert.NoError(t, err)
			}

			req, _ := http.NewRequest("GET", "/api/disbursements?page=1&limit=3", nil)
			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response struct {
				Data       []models.DisbursementListResponse `json:"data"`
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

		t.Run("List All Disbursements (unpaginated)", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/disbursements/all", nil)
			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response []models.DisbursementListResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotNil(t, response)
		})

		t.Run("Error cases", func(t *testing.T) {
			// Test Get with invalid ID
			req, _ := http.NewRequest("GET", "/api/disbursements/99999", nil)
			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code)

			// Test Update with invalid ID
			updateReq := models.UpdateDisbursementRequest{
				Amount:      678.90,
				Description: "Updated Description",
			}
			jsonData, _ := json.Marshal(updateReq)
			req, _ = http.NewRequest("PUT", "/api/disbursements/99999", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w = httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code)

			// Test Delete with invalid ID
			req, _ = http.NewRequest("DELETE", "/api/disbursements/99999", nil)
			w = httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})
}
