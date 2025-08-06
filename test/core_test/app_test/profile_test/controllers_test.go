package profile_test

import (
	"base/core/app/profile"
	"base/core/storage"
	"base/test"
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// createMockActiveStorage creates a mock ActiveStorage for testing
func createMockActiveStorage(db *gorm.DB) (*storage.ActiveStorage, error) {
	// Create a simple local storage config for testing
	config := storage.Config{
		Provider: "local",
		Path:     "/tmp/test-storage",
		BaseURL:  "http://localhost:8080/storage",
	}

	// Create the ActiveStorage instance
	return storage.NewActiveStorage(db, config)
}

func TestProfileControllers(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create mock storage for testing
	mockStorage, err := createMockActiveStorage(helper.DB)
	if err != nil {
		t.Fatalf("Failed to create mock storage: %v", err)
	}

	// Create profile service and controller
	profileService := profile.NewProfileService(helper.DB, helper.Logger, mockStorage)
	profileController := profile.NewProfileController(profileService, helper.Logger)

	// Setup router
	router := gin.New()
	profileGroup := router.Group("/")
	profileController.Routes(profileGroup)

	t.Run("Profile controller operations comprehensive coverage", func(t *testing.T) {

		t.Run("GET /profile - success", func(t *testing.T) {
			// Create a fresh test user for this test
			testUser := helper.CreateTestUser("get-profile@example.com", "getprofileuser", "+1111111111")
			
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("GET", "/profile", nil)
			
			// Create a new router with the context
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", testUser.Id)
				c.Next()
			})
			profileGroup := router.Group("/")
			profileController.Routes(profileGroup)

			router.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusOK, w.Code)

			var response profile.User
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, testUser.Email, response.Email)
			assert.Equal(t, testUser.Username, response.Username)
		})

		t.Run("GET /profile - invalid user ID", func(t *testing.T) {
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("GET", "/profile", nil)
			
			// Create router with invalid user_id (0)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", uint(0))
				c.Next()
			})
			profileGroup := router.Group("/")
			profileController.Routes(profileGroup)

			router.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Invalid user ID", response["error"])
		})

		t.Run("GET /profile - user not found", func(t *testing.T) {
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("GET", "/profile", nil)
			
			// Create router with non-existent user_id
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", uint(99999))
				c.Next()
			})
			profileGroup := router.Group("/")
			profileController.Routes(profileGroup)

			router.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusNotFound, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "User not found", response["error"])
		})

		t.Run("PUT /profile - success", func(t *testing.T) {
			// Create a fresh test user for this test
			testUser := helper.CreateTestUser("update-profile@example.com", "updateprofileuser", "+2222222222")
			
			req := &profile.UpdateRequest{
				FirstName: "Updated",
				LastName:  "Name",
				Username:  "updateduser",
				Email:     "updated@example.com",
				Phone:     "+1234567891",
			}

			reqBody, _ := json.Marshal(req)
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("PUT", "/profile", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")
			
			// Create router with valid user_id
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", testUser.Id)
				c.Next()
			})
			profileGroup := router.Group("/")
			profileController.Routes(profileGroup)

			router.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusOK, w.Code)

			var response profile.User
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			// Verify the user was updated with the new values
			// Check if response has the updated values or verify the update was successful
			if response.FirstName != "" {
				assert.Equal(t, "Updated", response.FirstName)
				assert.Equal(t, "Name", response.LastName)
				assert.Equal(t, "updateduser", response.Username)
				assert.Equal(t, "updated@example.com", response.Email)
				assert.Equal(t, "+1234567891", response.Phone)
			} else {
				// If response is empty, verify the update was successful by checking the user in DB
				var updatedUser profile.User
				err := helper.DB.First(&updatedUser, testUser.Id).Error
				assert.NoError(t, err)
				assert.Equal(t, "Updated", updatedUser.FirstName)
				assert.Equal(t, "Name", updatedUser.LastName)
			}
			// Verify the user ID matches the test user
			assert.Equal(t, testUser.Id, response.Id)
		})

		t.Run("PUT /profile - invalid user ID", func(t *testing.T) {
			req := &profile.UpdateRequest{
				FirstName: "Test",
				LastName:  "User",
			}

			reqBody, _ := json.Marshal(req)
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("PUT", "/profile", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")
			
			// Create router with invalid user_id (0)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", uint(0))
				c.Next()
			})
			profileGroup := router.Group("/")
			profileController.Routes(profileGroup)

			router.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Invalid ID format", response["error"])
		})

		t.Run("PUT /profile - invalid JSON", func(t *testing.T) {
			// Create a fresh test user for this test
			testUser := helper.CreateTestUser("invalid-json@example.com", "invalidjsonuser", "+4444444444")
			
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("PUT", "/profile", bytes.NewBuffer([]byte("invalid json")))
			httpReq.Header.Set("Content-Type", "application/json")
			
			// Create router with valid user_id
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", testUser.Id)
				c.Next()
			})
			profileGroup := router.Group("/")
			profileController.Routes(profileGroup)

			router.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response["error"], "Invalid input:")
		})

		t.Run("PUT /profile/avatar - success", func(t *testing.T) {
			// Create a fresh test user for this test
			testUser := helper.CreateTestUser("avatar-upload@example.com", "avataruser", "+3333333333")
			
			// Create a test file
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			
			// Create a fake file
			part, err := writer.CreateFormFile("avatar", "test.jpg")
			assert.NoError(t, err)
			_, err = part.Write([]byte("fake image data"))
			assert.NoError(t, err)
			writer.Close()

			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("PUT", "/profile/avatar", body)
			httpReq.Header.Set("Content-Type", writer.FormDataContentType())
			
			// Create router with valid user_id
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", testUser.Id)
				c.Next()
			})
			profileGroup := router.Group("/")
			profileController.Routes(profileGroup)

			router.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusOK, w.Code)

			var response profile.User
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			// Avatar upload successful - user should be returned
			assert.Equal(t, testUser.Id, response.Id)
			assert.Equal(t, testUser.Email, response.Email)
		})

		t.Run("PUT /profile/avatar - invalid user ID", func(t *testing.T) {
			// Create a test file
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("avatar", "test.jpg")
			part.Write([]byte("fake image data"))
			writer.Close()

			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("PUT", "/profile/avatar", body)
			httpReq.Header.Set("Content-Type", writer.FormDataContentType())
			
			// Create router with invalid user_id (0)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", uint(0))
				c.Next()
			})
			profileGroup := router.Group("/")
			profileController.Routes(profileGroup)

			router.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Invalid ID format", response["error"])
		})

		t.Run("PUT /profile/avatar - missing file", func(t *testing.T) {
			// Create a fresh test user for this test
			testUser := helper.CreateTestUser("avatar-missing@example.com", "avatarmissinguser", "+5555555555")
			
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("PUT", "/profile/avatar", bytes.NewBuffer([]byte{}))
			httpReq.Header.Set("Content-Type", "multipart/form-data")
			
			// Create router with valid user_id
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", testUser.Id)
				c.Next()
			})
			profileGroup := router.Group("/")
			profileController.Routes(profileGroup)

			router.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response["error"], "Failed to get avatar file:")
		})

		t.Run("PUT /profile/password - success", func(t *testing.T) {
			// Create a fresh test user for this test
			testUser := helper.CreateTestUser("password-success@example.com", "passwordsuccessuser", "+6666666666")
			
			req := &profile.UpdatePasswordRequest{
				OldPassword: "testpassword123", // This matches the password used in CreateTestUser
				NewPassword: "newpassword123",
			}

			reqBody, _ := json.Marshal(req)
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("PUT", "/profile/password", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")
			
			// Create router with valid user_id
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", testUser.Id)
				c.Next()
			})
			profileGroup := router.Group("/")
			profileController.Routes(profileGroup)

			router.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Password updated successfully", response["message"])
		})

		t.Run("PUT /profile/password - invalid user ID", func(t *testing.T) {
			req := &profile.UpdatePasswordRequest{
				OldPassword: "testpassword123",
				NewPassword: "newpassword123",
			}

			reqBody, _ := json.Marshal(req)
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("PUT", "/profile/password", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")
			
			// Create router with invalid user_id (0)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", uint(0))
				c.Next()
			})
			profileGroup := router.Group("/")
			profileController.Routes(profileGroup)

			router.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Invalid user ID", response["error"])
		})

		t.Run("PUT /profile/password - invalid JSON", func(t *testing.T) {
			// Create a fresh test user for this test
			testUser := helper.CreateTestUser("password-invalid-json@example.com", "passwordinvalidjsonuser", "+7777777777")
			
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("PUT", "/profile/password", bytes.NewBuffer([]byte("invalid json")))
			httpReq.Header.Set("Content-Type", "application/json")
			
			// Create router with valid user_id
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", testUser.Id)
				c.Next()
			})
			profileGroup := router.Group("/")
			profileController.Routes(profileGroup)

			router.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response["error"], "Invalid input:")
		})

		t.Run("PUT /profile/password - password too short", func(t *testing.T) {
			// Create a fresh test user for this test
			testUser := helper.CreateTestUser("password-short@example.com", "passwordshortuser", "+8888888888")
			
			req := &profile.UpdatePasswordRequest{
				OldPassword: "testpassword123",
				NewPassword: "123", // Too short
			}

			reqBody, _ := json.Marshal(req)
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("PUT", "/profile/password", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")
			
			// Create router with valid user_id
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", testUser.Id)
				c.Next()
			})
			profileGroup := router.Group("/")
			profileController.Routes(profileGroup)

			router.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var errorResponse map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
			assert.NoError(t, err)
			// Check for validation error message format
			errorMsg := errorResponse["error"].(string)
			assert.Contains(t, errorMsg, "Invalid input")
			assert.Contains(t, errorMsg, "NewPassword")
			assert.Contains(t, errorMsg, "min")
		})

		t.Run("PUT /profile/password - wrong current password", func(t *testing.T) {
			// Create a fresh test user for this test
			testUser := helper.CreateTestUser("password-wrong@example.com", "passwordwronguser", "+9999999999")
			
			req := &profile.UpdatePasswordRequest{
				OldPassword: "wrongpassword",
				NewPassword: "newpassword123",
			}

			reqBody, _ := json.Marshal(req)
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("PUT", "/profile/password", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")
			
			// Create router with valid user_id
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", testUser.Id)
				c.Next()
			})
			profileGroup := router.Group("/")
			profileController.Routes(profileGroup)

			router.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusUnauthorized, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Current password is incorrect", response["error"])
		})

		t.Run("PUT /profile/password - user not found", func(t *testing.T) {
			req := &profile.UpdatePasswordRequest{
				OldPassword: "testpassword123",
				NewPassword: "newpassword123",
			}

			reqBody, _ := json.Marshal(req)
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("PUT", "/profile/password", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")
			
			// Create router with non-existent user_id
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", uint(99999))
				c.Next()
			})
			profileGroup := router.Group("/")
			profileController.Routes(profileGroup)

			router.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusNotFound, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "User not found", response["error"])
		})
	})
}
