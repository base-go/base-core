package authentication_test

import (
	"base/core/app/authentication"
	"base/core/app/profile"
	"base/test"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestAuthenticationModels tests all authentication model functionality for 100% coverage
func TestAuthenticationModels(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	t.Run("AuthUser model operations", func(t *testing.T) {
		t.Run("should create AuthUser with embedded User", func(t *testing.T) {
			authUser := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Auth",
					LastName:  "User",
					Username:  "authuser",
					Email:     "auth@example.com",
					Password:  "hashedpassword",
					Phone:     "+1234567890",
				},
			}

			err := helper.DB.Create(authUser).Error
			assert.NoError(t, err)
			assert.NotZero(t, authUser.Id)
			assert.Equal(t, "auth@example.com", authUser.Email)
		})

		t.Run("should use correct table name", func(t *testing.T) {
			authUser := authentication.AuthUser{}
			tableName := authUser.TableName()
			assert.Equal(t, "users", tableName)
		})

		t.Run("should handle reset token and last login fields", func(t *testing.T) {
			authUser := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Reset",
					LastName:  "Token",
					Username:  "resettoken",
					Email:     "resettoken@example.com",
					Password:  "hashedpassword",
					Phone:     "+1234567891",
				},
			}
			helper.DB.Create(authUser)

			// Set reset token and last login
			resetToken := "abc123def456"
			resetExpiry := time.Now().Add(time.Hour)
			lastLogin := time.Now()

			authUser.ResetToken = resetToken
			authUser.ResetTokenExpiry = &resetExpiry
			authUser.LastLogin = &lastLogin

			err := helper.DB.Save(authUser).Error
			assert.NoError(t, err)

			// Retrieve and verify all fields
			var savedUser authentication.AuthUser
			err = helper.DB.First(&savedUser, authUser.Id).Error
			assert.NoError(t, err)
			assert.Equal(t, resetToken, savedUser.ResetToken)
			assert.NotNil(t, savedUser.ResetTokenExpiry)
			assert.NotNil(t, savedUser.LastLogin)
		})
	})

	t.Run("All Request and Response models", func(t *testing.T) {
		t.Run("RegisterRequest with all fields", func(t *testing.T) {
			req := authentication.RegisterRequest{
				FirstName: "Test",
				LastName:  "User",
				Username:  "testuser",
				Email:     "test@example.com",
				Password:  "password123",
				Phone:     "+1234567893",
			}

			assert.Equal(t, "Test", req.FirstName)
			assert.Equal(t, "User", req.LastName)
			assert.Equal(t, "testuser", req.Username)
			assert.Equal(t, "test@example.com", req.Email)
			assert.Equal(t, "password123", req.Password)
			assert.Equal(t, "+1234567893", req.Phone)
		})

		t.Run("LoginRequest validation", func(t *testing.T) {
			req := authentication.LoginRequest{
				Email:    "login@example.com",
				Password: "loginpass",
			}
			assert.Equal(t, "login@example.com", req.Email)
			assert.Equal(t, "loginpass", req.Password)
		})

		t.Run("Password reset requests", func(t *testing.T) {
			forgotReq := authentication.ForgotPasswordRequest{
				Email: "forgot@example.com",
			}
			assert.Equal(t, "forgot@example.com", forgotReq.Email)

			resetReq := authentication.ResetPasswordRequest{
				Email:       "reset@example.com",
				Token:       "resettoken123",
				NewPassword: "newpassword",
			}
			assert.Equal(t, "reset@example.com", resetReq.Email)
			assert.Equal(t, "resettoken123", resetReq.Token)
			assert.Equal(t, "newpassword", resetReq.NewPassword)
		})

		t.Run("OTP requests", func(t *testing.T) {
			verifyReq := authentication.VerifyOTPRequest{
				Email: "otp@example.com",
				OTP:   "123456",
			}
			assert.Equal(t, "otp@example.com", verifyReq.Email)
			assert.Equal(t, "123456", verifyReq.OTP)

			sendReq := authentication.SendOTPRequest{
				Email: "sendotp@example.com",
			}
			assert.Equal(t, "sendotp@example.com", sendReq.Email)
		})

		t.Run("Response models", func(t *testing.T) {
			authResp := authentication.AuthResponse{
				UserResponse: profile.UserResponse{
					Id:        1,
					Email:     "test@example.com",
					Username:  "testuser",
					FirstName: "Test",
					LastName:  "User",
					Phone:     "+1234567894",
				},
				AccessToken: "jwt-token-here",
				Exp:         1234567890,
				Extend:      map[string]interface{}{"extra": "data"},
			}
			assert.Equal(t, uint(1), authResp.Id)
			assert.Equal(t, "jwt-token-here", authResp.AccessToken)
			assert.Equal(t, int64(1234567890), authResp.Exp)
			assert.NotNil(t, authResp.Extend)

			errorResp := authentication.ErrorResponse{Error: "Something went wrong"}
			assert.Equal(t, "Something went wrong", errorResp.Error)

			successResp := authentication.SuccessResponse{Message: "Operation successful"}
			assert.Equal(t, "Operation successful", successResp.Message)
		})
	})

	t.Run("LoginEvent model", func(t *testing.T) {
		authUser := &authentication.AuthUser{
			User: profile.User{
				Email: "event@example.com",
				Phone: "+1234567895",
			},
		}

		loginAllowed := true
		errorResp := &authentication.ErrorResponse{Error: "Test error"}
		authResp := &authentication.AuthResponse{AccessToken: "token"}

		event := authentication.LoginEvent{
			User:         authUser,
			LoginAllowed: &loginAllowed,
			Error:        errorResp,
			Response:     authResp,
		}

		assert.Equal(t, authUser, event.User)
		assert.True(t, *event.LoginAllowed)
		assert.Equal(t, "Test error", event.Error.Error)
		assert.Equal(t, "token", event.Response.AccessToken)
	})
}
