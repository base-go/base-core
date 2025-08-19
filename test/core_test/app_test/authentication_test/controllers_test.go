package authentication_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"base/core/app/authentication"
	"base/core/app/profile"
	"base/core/emitter"
	"base/core/router"
	"base/test"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// TestAuthenticationControllers tests all authentication controller endpoints for 100% coverage
func TestAuthenticationControllers(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Setup router for testing

	// Create mock email sender that can simulate both success and failure
	mockEmailSender := &test.MockEmailSender{ShouldFail: false}
	mockEmitter := emitter.New()

	// Create authentication service and controller
	authService := authentication.NewAuthService(helper.DB, mockEmailSender, mockEmitter)
	authController := authentication.NewAuthController(authService, mockEmailSender, helper.Logger)

	// Setup router
	r := router.New()
	authGroup := r.Group("/auth")
	authController.Routes(authGroup)

	t.Run("Authentication controller operations comprehensive coverage", func(t *testing.T) {
		// Clean database before each sub-test group
		helper.CleanDatabase()

		t.Run("POST /auth/register - successful registration", func(t *testing.T) {
			req := &authentication.RegisterRequest{
				FirstName: "John",
				LastName:  "Doe",
				Username:  "johndoe",
				Phone:     "+1234567890",
				Email:     "john@example.com",
				Password:  "password123",
			}

			jsonData, err := json.Marshal(req)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response authentication.AuthResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "john@example.com", response.Email)
			assert.Equal(t, "John", response.FirstName)
		})

		t.Run("POST /auth/register - with email sending failure", func(t *testing.T) {
			// Set mock to fail email sending
			mockEmailSender.ShouldFail = true

			req := &authentication.RegisterRequest{
				FirstName: "Jane",
				LastName:  "Doe",
				Username:  "janedoe",
				Phone:     "+1234567891",
				Email:     "jane@example.com",
				Password:  "password123",
			}

			jsonData, err := json.Marshal(req)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)

			// Should still succeed even if email fails
			assert.Equal(t, http.StatusCreated, w.Code)

			// Reset mock
			mockEmailSender.ShouldFail = false
		})

		t.Run("POST /auth/register - invalid JSON", func(t *testing.T) {
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer([]byte("invalid json")))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response authentication.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response.Error, "invalid character")
		})

		t.Run("POST /auth/register - service error", func(t *testing.T) {
			// Try to register with duplicate email
			req1 := &authentication.RegisterRequest{
				FirstName: "First",
				LastName:  "User",
				Username:  "firstuser",
				Phone:     "+1234567892",
				Email:     "duplicate@example.com",
				Password:  "password123",
			}

			jsonData, err := json.Marshal(req1)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)
			assert.Equal(t, http.StatusCreated, w.Code)

			// Try to register again with same email
			req2 := &authentication.RegisterRequest{
				FirstName: "Second",
				LastName:  "User",
				Username:  "seconduser",
				Phone:     "+1234567893",
				Email:     "duplicate@example.com",
				Password:  "password123",
			}

			jsonData, err = json.Marshal(req2)
			assert.NoError(t, err)

			w = httptest.NewRecorder()
			httpReq, _ = http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusConflict, w.Code)

			var response authentication.ErrorResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "user already exists", response.Error)
		})

		t.Run("POST /auth/login - successful login", func(t *testing.T) {
			// Create a user first
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("loginpass123"), bcrypt.DefaultCost)
			user := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Login",
					LastName:  "User",
					Username:  "loginuser",
					Phone:     "+1234567894",
					Email:     "login@example.com",
					Password:  string(hashedPassword),
				},
			}
			helper.DB.Create(user)

			req := &authentication.LoginRequest{
				Email:    "login@example.com",
				Password: "loginpass123",
			}

			jsonData, err := json.Marshal(req)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusOK, w.Code)

			var response authentication.AuthResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "login@example.com", response.Email)
		})

		t.Run("POST /auth/login - invalid JSON", func(t *testing.T) {
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer([]byte("invalid json")))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("POST /auth/login - invalid credentials", func(t *testing.T) {
			req := &authentication.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "wrongpassword",
			}

			jsonData, err := json.Marshal(req)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})

		t.Run("POST /auth/logout - successful logout", func(t *testing.T) {
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/logout", nil)

			r.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusOK, w.Code)

			var response authentication.SuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Logout successful", response.Message)
		})

		t.Run("POST /auth/forgot-password - successful request", func(t *testing.T) {
			// Create a user first
			user := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Forgot",
					LastName:  "Password",
					Username:  "forgotpass",
					Phone:     "+1234567896",
					Email:     "forgot@example.com",
					Password:  "hashedpass",
				},
			}
			helper.DB.Create(user)

			req := &authentication.ForgotPasswordRequest{
				Email: "forgot@example.com",
			}

			jsonData, err := json.Marshal(req)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/forgot-password", bytes.NewBuffer(jsonData))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusOK, w.Code)

			var response authentication.SuccessResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Password reset email sent", response.Message)
		})

		t.Run("POST /auth/forgot-password - invalid JSON", func(t *testing.T) {
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/forgot-password", bytes.NewBuffer([]byte("invalid json")))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("POST /auth/forgot-password - user not found", func(t *testing.T) {
			req := &authentication.ForgotPasswordRequest{
				Email: "notfound@example.com",
			}

			jsonData, err := json.Marshal(req)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/forgot-password", bytes.NewBuffer(jsonData))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusNotFound, w.Code)

			var response authentication.ErrorResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "User not found", response.Error)
		})

		t.Run("POST /auth/reset-password - successful reset", func(t *testing.T) {
			// Create a user with reset token
			resetToken := "validtoken123"
			resetExpiry := time.Now().Add(time.Hour)
			user := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Reset",
					LastName:  "Password",
					Username:  "resetpass",
					Phone:     "+1234567897",
					Email:     "reset@example.com",
					Password:  "oldpassword",
				},
				ResetToken:       resetToken,
				ResetTokenExpiry: &resetExpiry,
			}
			helper.DB.Create(user)

			req := &authentication.ResetPasswordRequest{
				Email:       "reset@example.com",
				Token:       resetToken,
				NewPassword: "newpassword123",
			}

			jsonData, err := json.Marshal(req)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/reset-password", bytes.NewBuffer(jsonData))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusOK, w.Code)

			var response authentication.SuccessResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Password reset successful", response.Message)
		})

		t.Run("POST /auth/reset-password - invalid JSON", func(t *testing.T) {
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/reset-password", bytes.NewBuffer([]byte("invalid json")))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response authentication.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Invalid request format", response.Error)
		})

		t.Run("POST /auth/reset-password - invalid token", func(t *testing.T) {
			// Create a user with reset token
			resetToken := "validtoken456"
			resetExpiry := time.Now().Add(time.Hour)
			user := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Invalid",
					LastName:  "Token",
					Username:  "invalidtoken",
					Phone:     "+1234567898",
					Email:     "invalidtoken@example.com",
					Password:  "oldpassword",
				},
				ResetToken:       resetToken,
				ResetTokenExpiry: &resetExpiry,
			}
			helper.DB.Create(user)

			req := &authentication.ResetPasswordRequest{
				Email:       "test@example.com",
				Token:       "invalid-token",
				NewPassword: "newpassword123",
			}

			reqBody, _ := json.Marshal(req)
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/reset-password", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusInternalServerError, w.Code)

			var response authentication.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Failed to reset password", response.Error)
		})

		t.Run("POST /auth/reset-password - user not found", func(t *testing.T) {
			req := &authentication.ResetPasswordRequest{
				Email:       "notfound@example.com",
				Token:       "some-token",
				NewPassword: "newpassword123",
			}

			reqBody, _ := json.Marshal(req)
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/reset-password", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusInternalServerError, w.Code)

			var response authentication.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Failed to reset password", response.Error)
		})

		t.Run("POST /auth/login - edge cases for full coverage", func(t *testing.T) {
			// Test with non-existent user to hit error paths
			req := &authentication.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			}

			reqBody, _ := json.Marshal(req)
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)

			// Should return unauthorized
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})

		t.Run("POST /auth/forgot-password - edge cases for full coverage", func(t *testing.T) {
			// Test with non-existent email to hit error paths
			req := &authentication.ForgotPasswordRequest{
				Email: "nonexistent@example.com",
			}

			reqBody, _ := json.Marshal(req)
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/forgot-password", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)

			// Should return 404 for non-existent email
			assert.Equal(t, http.StatusNotFound, w.Code)
		})

		t.Run("POST /auth/reset-password - edge cases for full coverage", func(t *testing.T) {
			// Test with expired token to hit error paths
			resetToken := "expiredtoken123"
			resetExpiry := time.Now().Add(-time.Hour) // Expired token
			user := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Expired",
					LastName:  "Token",
					Username:  "expiredtoken",
					Phone:     "+1234567897",
					Email:     "expired@example.com",
					Password:  "oldpassword",
				},
				ResetToken:       resetToken,
				ResetTokenExpiry: &resetExpiry,
			}
			helper.DB.Create(user)

			req := &authentication.ResetPasswordRequest{
				Email:       "expired@example.com",
				Token:       resetToken,
				NewPassword: "newpassword123",
			}

			reqBody, _ := json.Marshal(req)
			w := httptest.NewRecorder()
			httpReq, _ := http.NewRequest("POST", "/auth/reset-password", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, httpReq)

			// Should fail due to expired token
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})
}
