package authentication

import (
	"base/core/app/profile"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Setup test database
func setupModelTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto migrate tables - use AuthUser to get all fields including reset_token
	err = db.AutoMigrate(&AuthUser{})
	assert.NoError(t, err)

	return db
}

func TestAuthUser_TableName(t *testing.T) {
	authUser := AuthUser{}
	tableName := authUser.TableName()
	assert.Equal(t, "users", tableName)
}

func TestAuthUser_CRUD_Operations(t *testing.T) {
	db := setupModelTestDB(t)

	t.Run("create AuthUser", func(t *testing.T) {
		authUser := &AuthUser{
			User: profile.User{
				FirstName: "Auth",
				LastName:  "User",
				Username:  "authuser",
				Email:     "auth@example.com",
				Password:  "hashedpassword",
				Phone:     "+1234567890",
			},
		}

		err := db.Create(authUser).Error
		assert.NoError(t, err)
		assert.NotZero(t, authUser.Id)
		assert.Equal(t, "auth@example.com", authUser.Email)
	})

	t.Run("update AuthUser with reset token", func(t *testing.T) {
		authUser := &AuthUser{
			User: profile.User{
				FirstName: "Reset",
				LastName:  "Token",
				Username:  "resettoken",
				Email:     "reset@example.com",
				Password:  "hashedpassword",
				Phone:     "+1234567891",
			},
		}
		db.Create(authUser)

		// Set reset token and last login
		resetToken := "abc123def456"
		resetExpiry := time.Now().Add(time.Hour)
		lastLogin := time.Now()

		authUser.ResetToken = resetToken
		authUser.ResetTokenExpiry = &resetExpiry
		authUser.LastLogin = &lastLogin

		err := db.Save(authUser).Error
		assert.NoError(t, err)

		// Retrieve and verify
		var savedUser AuthUser
		err = db.First(&savedUser, authUser.Id).Error
		assert.NoError(t, err)
		assert.Equal(t, resetToken, savedUser.ResetToken)
		assert.NotNil(t, savedUser.ResetTokenExpiry)
		assert.NotNil(t, savedUser.LastLogin)
	})
}

func TestRegisterRequest_Validation(t *testing.T) {
	t.Run("valid RegisterRequest", func(t *testing.T) {
		req := RegisterRequest{
			FirstName: "Test",
			LastName:  "User",
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "password123",
			Phone:     "+1234567892",
		}

		assert.Equal(t, "Test", req.FirstName)
		assert.Equal(t, "User", req.LastName)
		assert.Equal(t, "testuser", req.Username)
		assert.Equal(t, "test@example.com", req.Email)
		assert.Equal(t, "password123", req.Password)
		assert.Equal(t, "+1234567892", req.Phone)
	})

	t.Run("empty RegisterRequest", func(t *testing.T) {
		req := RegisterRequest{}
		assert.Equal(t, "", req.FirstName)
		assert.Equal(t, "", req.Email)
		assert.Equal(t, "", req.Password)
	})
}

func TestLoginRequest_Validation(t *testing.T) {
	t.Run("valid LoginRequest", func(t *testing.T) {
		req := LoginRequest{
			Email:    "login@example.com",
			Password: "loginpass",
		}
		assert.Equal(t, "login@example.com", req.Email)
		assert.Equal(t, "loginpass", req.Password)
	})

	t.Run("empty LoginRequest", func(t *testing.T) {
		req := LoginRequest{}
		assert.Equal(t, "", req.Email)
		assert.Equal(t, "", req.Password)
	})
}

func TestPasswordResetRequests(t *testing.T) {
	t.Run("ForgotPasswordRequest", func(t *testing.T) {
		req := ForgotPasswordRequest{
			Email: "forgot@example.com",
		}
		assert.Equal(t, "forgot@example.com", req.Email)
	})

	t.Run("ResetPasswordRequest", func(t *testing.T) {
		req := ResetPasswordRequest{
			Email:       "reset@example.com",
			Token:       "resettoken123",
			NewPassword: "newpassword",
		}
		assert.Equal(t, "reset@example.com", req.Email)
		assert.Equal(t, "resettoken123", req.Token)
		assert.Equal(t, "newpassword", req.NewPassword)
	})
}

func TestOTPRequests(t *testing.T) {
	t.Run("VerifyOTPRequest", func(t *testing.T) {
		req := VerifyOTPRequest{
			Email: "otp@example.com",
			OTP:   "123456",
		}
		assert.Equal(t, "otp@example.com", req.Email)
		assert.Equal(t, "123456", req.OTP)
	})

	t.Run("SendOTPRequest", func(t *testing.T) {
		req := SendOTPRequest{
			Email: "sendotp@example.com",
		}
		assert.Equal(t, "sendotp@example.com", req.Email)
	})
}

func TestResponseModels(t *testing.T) {
	t.Run("AuthResponse", func(t *testing.T) {
		authResp := AuthResponse{
			UserResponse: profile.UserResponse{
				Id:        1,
				Email:     "test@example.com",
				Username:  "testuser",
				FirstName: "Test",
				LastName:  "User",
				Phone:     "+1234567893",
			},
			AccessToken: "jwt-token-here",
			Exp:         1234567890,
			Extend:      map[string]any{"extra": "data"},
		}
		assert.Equal(t, uint(1), authResp.Id)
		assert.Equal(t, "jwt-token-here", authResp.AccessToken)
		assert.Equal(t, int64(1234567890), authResp.Exp)
		assert.NotNil(t, authResp.Extend)
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		errorResp := ErrorResponse{Error: "Something went wrong"}
		assert.Equal(t, "Something went wrong", errorResp.Error)
	})

	t.Run("SuccessResponse", func(t *testing.T) {
		successResp := SuccessResponse{Message: "Operation successful"}
		assert.Equal(t, "Operation successful", successResp.Message)
	})
}

func TestLoginEvent(t *testing.T) {
	authUser := &AuthUser{
		User: profile.User{
			Email: "event@example.com",
			Phone: "+1234567894",
		},
	}

	loginAllowed := true
	errorResp := &ErrorResponse{Error: "Test error"}
	authResp := &AuthResponse{AccessToken: "token"}

	event := LoginEvent{
		User:         authUser,
		LoginAllowed: &loginAllowed,
		Error:        errorResp,
		Response:     authResp,
	}

	assert.Equal(t, authUser, event.User)
	assert.True(t, *event.LoginAllowed)
	assert.Equal(t, "Test error", event.Error.Error)
	assert.Equal(t, "token", event.Response.AccessToken)
}
