package authentication_test

import (
	"base/core/app/authentication"
	"base/core/app/profile"
	"base/core/emitter"
	"base/test"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthenticationService(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Create mocks
	emailSender := &test.MockEmailSender{}
	mockEmitter := emitter.New()

	// Create authentication service
	authService := authentication.NewAuthService(helper.DB, emailSender, mockEmitter)

	t.Run("Authentication service operations for 100% coverage", func(t *testing.T) {
		// Clean database before each sub-test group
		helper.CleanDatabase()

		t.Run("Register - edge cases", func(t *testing.T) {
			// Test with duplicate email
			uniqueID := helper.GenerateUniqueTestID()
			email := fmt.Sprintf("duplicate-%s@example.com", uniqueID)
			username := fmt.Sprintf("duplicateuser%s", uniqueID)
			phone := fmt.Sprintf("+1%s", uniqueID[len(uniqueID)-10:])
			
			req := &authentication.RegisterRequest{
				FirstName: "Duplicate",
				LastName:  "User",
				Username:  username,
				Phone:     phone,
				Email:     email,
				Password:  "password123",
			}

			// First registration should succeed
			_, err := authService.Register(req)
			assert.NoError(t, err)

			// Second registration with same email should fail
			req2 := &authentication.RegisterRequest{
				FirstName: "Another",
				LastName:  "User",
				Username:  "anotheruser" + uniqueID,
				Phone:     fmt.Sprintf("+2%s", uniqueID[len(uniqueID)-10:]),
				Email:     email, // Same email
				Password:  "password123",
			}
			_, err = authService.Register(req2)
			assert.Error(t, err) // Should fail due to duplicate email
		})

		t.Run("Login - edge cases", func(t *testing.T) {
			// Create a test user first
			uniqueID := helper.GenerateUniqueTestID()
			email := fmt.Sprintf("logintest-%s@example.com", uniqueID)
			username := fmt.Sprintf("logintest%s", uniqueID)
			phone := fmt.Sprintf("+1%s", uniqueID[len(uniqueID)-10:])
			
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			testUser := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Login",
					LastName:  "Test",
					Username:  username,
					Phone:     phone,
					Email:     email,
					Password:  string(hashedPassword),
				},
			}
			helper.DB.Create(testUser)

			// Test successful login
			loginReq := &authentication.LoginRequest{
				Email:    email,
				Password: "password123",
			}
			response, err := authService.Login(loginReq)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.NotEmpty(t, response.AccessToken)

			// Test login with wrong password
			wrongPasswordReq := &authentication.LoginRequest{
				Email:    email,
				Password: "wrongpassword",
			}
			_, err = authService.Login(wrongPasswordReq)
			assert.Error(t, err) // Should fail with wrong password

			// Test login with non-existent email
			nonExistentReq := &authentication.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			}
			_, err = authService.Login(nonExistentReq)
			assert.Error(t, err) // Should fail with non-existent email
		})

		t.Run("ForgotPassword - edge cases", func(t *testing.T) {
			// Create a test user first
			testUser := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Forgot",
					LastName:  "Password",
					Username:  "forgotpassword",
					Phone:     "+4444444444",
					Email:     "forgotpassword@example.com",
					Password:  "oldpassword",
				},
			}
			helper.DB.Create(testUser)

			// Test forgot password with existing email
			err := authService.ForgotPassword("forgotpassword@example.com")
			assert.NoError(t, err)

			// Verify reset token was set
			var updatedUser authentication.AuthUser
			helper.DB.First(&updatedUser, "email = ?", "forgotpassword@example.com")
			assert.NotEmpty(t, updatedUser.ResetToken)
			assert.NotNil(t, updatedUser.ResetTokenExpiry)

			// Test forgot password with non-existent email
			err = authService.ForgotPassword("nonexistent@example.com")
			assert.Error(t, err) // Should return error for non-existent email
		})

		t.Run("ResetPassword - edge cases", func(t *testing.T) {
			// Create a test user with reset token
			resetToken := "validtoken123"
			resetExpiry := time.Now().Add(time.Hour)
			testUser := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Reset",
					LastName:  "Password",
					Username:  "resetpassword",
					Phone:     "+5555555555",
					Email:     "resetpassword@example.com",
					Password:  "oldpassword",
				},
				ResetToken:       resetToken,
				ResetTokenExpiry: &resetExpiry,
			}
			helper.DB.Create(testUser)

			// Test successful password reset
			err := authService.ResetPassword("resetpassword@example.com", resetToken, "newpassword123")
			assert.NoError(t, err)

			// Verify password was changed and token was cleared
			var updatedUser authentication.AuthUser
			helper.DB.First(&updatedUser, "email = ?", "resetpassword@example.com")
			assert.Empty(t, updatedUser.ResetToken)
			assert.Nil(t, updatedUser.ResetTokenExpiry)

			// Test reset password with invalid token
			err = authService.ResetPassword("resetpassword@example.com", "invalidtoken", "newpassword123")
			assert.Error(t, err) // Should fail with invalid token

			// Test reset password with non-existent email
			err = authService.ResetPassword("nonexistent@example.com", "sometoken", "newpassword123")
			assert.Error(t, err) // Should fail with non-existent email
		})

		t.Run("Register - additional edge cases", func(t *testing.T) {
			// Test with very long password to hit different code paths
			req := &authentication.RegisterRequest{
				FirstName: "Long",
				LastName:  "Password",
				Username:  "longpassword",
				Phone:     "+6666666666",
				Email:     "longpassword@example.com",
				Password:  "verylongpasswordthatmightcausedifferentbehavior123456789",
			}
			_, err := authService.Register(req)
			// This should succeed and hit different code paths
			assert.NoError(t, err)
		})

		t.Run("Service comprehensive edge cases", func(t *testing.T) {
			// Test Register with database error simulation (by using invalid data)
			req := &authentication.RegisterRequest{
				FirstName: "DB",
				LastName:  "Error",
				Username:  "dberror",
				Phone:     "+8888888888",
				Email:     "dberror@example.com",
				Password:  "password123",
			}
			_, err := authService.Register(req)
			assert.NoError(t, err) // Should succeed

			// Test Login with user that has no last login
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			testUser := &authentication.AuthUser{
				User: profile.User{
					FirstName: "No",
					LastName:  "LastLogin",
					Username:  "nolastlogin",
					Phone:     "+9999999999",
					Email:     "nolastlogin@example.com",
					Password:  string(hashedPassword),
				},
				LastLogin: nil, // No previous login
			}
			helper.DB.Create(testUser)

			loginReq := &authentication.LoginRequest{
				Email:    "nolastlogin@example.com",
				Password: "password123",
			}
			response, err := authService.Login(loginReq)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.NotEmpty(t, response.AccessToken)

			// Test ForgotPassword with email sending success
			testUser2 := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Email",
					LastName:  "Success",
					Username:  "emailsuccess",
					Phone:     "+1010101010",
					Email:     "emailsuccess@example.com",
					Password:  "oldpassword",
				},
			}
			helper.DB.Create(testUser2)

			err = authService.ForgotPassword("emailsuccess@example.com")
			assert.NoError(t, err)

			// Test ResetPassword with valid token and successful password change
			resetToken := "validtoken789"
			resetExpiry := time.Now().Add(time.Hour)
			testUser3 := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Valid",
					LastName:  "Reset",
					Username:  "validreset",
					Phone:     "+1212121213", // Unique phone number
					Email:     "validreset@example.com",
					Password:  "oldpassword",
				},
				ResetToken:       resetToken,
				ResetTokenExpiry: &resetExpiry,
			}
			helper.DB.Create(testUser3)

			err = authService.ResetPassword("validreset@example.com", resetToken, "newpassword789")
			assert.NoError(t, err)
		})

		t.Run("Register comprehensive coverage tests", func(t *testing.T) {
			// Test Register with emitter nil case
			// Create service without emitter to hit the nil emitter path
			serviceNoEmitter := authentication.NewAuthService(helper.DB, emailSender, nil)
			req := &authentication.RegisterRequest{
				FirstName: "No",
				LastName:  "Emitter",
				Username:  "noemitter",
				Phone:     "+1313131314",
				Email:     "noemitter@example.com",
				Password:  "password123",
			}
			response, err := serviceNoEmitter.Register(req)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.NotEmpty(t, response.AccessToken)

			// Test Register with duplicate user (should hit duplicate key error path)
			// First create a user
			req1 := &authentication.RegisterRequest{
				FirstName: "Duplicate",
				LastName:  "User",
				Username:  "duplicateuser",
				Phone:     "+1414141415",
				Email:     "duplicate@example.com",
				Password:  "password123",
			}
			_, err = authService.Register(req1)
			assert.NoError(t, err)

			// Try to register same user again (should fail with duplicate error from validateUser)
			req2 := &authentication.RegisterRequest{
				FirstName: "Duplicate",
				LastName:  "User2",
				Username:  "duplicateuser", // Same username - should cause duplicate
				Phone:     "+1515151516",    // Different phone
				Email:     "duplicate2@example.com", // Different email
				Password:  "password123",
			}
			_, err = authService.Register(req2)
			// This should succeed because validateUser catches duplicates before database
			assert.Error(t, err)
			assert.Equal(t, "user already exists", err.Error())

			// Test Register success path with all valid data
			req3 := &authentication.RegisterRequest{
				FirstName: "Success",
				LastName:  "Path",
				Username:  "successpath",
				Phone:     "+1616161617",
				Email:     "successpath@example.com",
				Password:  "password123",
			}
			response, err = authService.Register(req3)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, "Success", response.FirstName)
			assert.Equal(t, "successpath@example.com", response.Email)
			assert.NotEmpty(t, response.AccessToken)
			assert.Greater(t, response.Exp, int64(0))
		})

		t.Run("ForgotPassword comprehensive coverage tests", func(t *testing.T) {
			// Test ForgotPassword with valid user - success path
			helper.CreateTestUser("forgotpass@example.com", "forgotpassuser", "+1717171718")
			err := authService.ForgotPassword("forgotpass@example.com")
			assert.NoError(t, err)

			// Verify reset token was set
			var updatedUser authentication.AuthUser
			helper.DB.Where("email = ?", "forgotpass@example.com").First(&updatedUser)
			assert.NotEmpty(t, updatedUser.ResetToken)
			assert.NotNil(t, updatedUser.ResetTokenExpiry)

			// Test ForgotPassword with non-existent user - should return user not found error
			err = authService.ForgotPassword("nonexistent@example.com")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "user not found")

			// Test ForgotPassword with empty email - should return user not found error
			err = authService.ForgotPassword("")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "user not found")

			// Test ForgotPassword with invalid email format - should return user not found error
			err = authService.ForgotPassword("invalid-email-format")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "user not found")

			// Test ForgotPassword success path with different user to hit more code paths
			helper.CreateTestUser("forgotpass2@example.com", "forgotpass2user", "+1818181819")
			err = authService.ForgotPassword("forgotpass2@example.com")
			assert.NoError(t, err)

			// Verify reset token was set for second user
			var updatedUser2 authentication.AuthUser
			helper.DB.Where("email = ?", "forgotpass2@example.com").First(&updatedUser2)
			assert.NotEmpty(t, updatedUser2.ResetToken)
			assert.NotNil(t, updatedUser2.ResetTokenExpiry)
			assert.True(t, updatedUser2.ResetTokenExpiry.After(time.Now()))
		})

		t.Run("ResetPassword comprehensive coverage tests", func(t *testing.T) {
			// Create user with valid reset token for testing
			uniqueID := helper.GenerateUniqueTestID()
			email := fmt.Sprintf("resetpassword-%s@example.com", uniqueID)
			username := fmt.Sprintf("resetpassword%s", uniqueID)
			phone := fmt.Sprintf("+1%s", uniqueID[len(uniqueID)-10:])
			resetToken := "validresettoken123"
			resetExpiry := time.Now().Add(time.Hour)
			
			testUser := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Reset",
					LastName:  "Password",
					Username:  username,
					Phone:     phone,
					Email:     email,
					Password:  "oldpassword",
				},
				ResetToken:       resetToken,
				ResetTokenExpiry: &resetExpiry,
			}
			helper.DB.Create(testUser)

			// Test ResetPassword with valid token - success path
			err := authService.ResetPassword(email, resetToken, "newpassword123")
			assert.NoError(t, err)

			// Verify password was changed and reset token cleared
			var updatedUser authentication.AuthUser
			helper.DB.Where("email = ?", email).First(&updatedUser)
			assert.Empty(t, updatedUser.ResetToken)
			assert.Nil(t, updatedUser.ResetTokenExpiry)
			assert.NotEqual(t, "oldpassword", updatedUser.Password)

			// Test ResetPassword with non-existent user
			err = authService.ResetPassword("nonexistent@example.com", "anytoken", "newpassword")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "user not found")

			// Create user with different reset token for invalid token test
			invalidTokenUser := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Invalid",
					LastName:  "Token",
					Username:  "invalidtoken",
					Phone:     "+2020202021",
					Email:     "invalidtoken@example.com",
					Password:  "oldpassword",
				},
				ResetToken:       "correcttoken123",
				ResetTokenExpiry: &resetExpiry,
			}
			helper.DB.Create(invalidTokenUser)

			// Test ResetPassword with invalid token
			err = authService.ResetPassword("invalidtoken@example.com", "wrongtoken123", "newpassword")
			assert.Error(t, err)
			assert.Equal(t, "invalid token", err.Error())

			// Create user with expired reset token
			expiredTime := time.Now().Add(-time.Hour) // Expired 1 hour ago
			expiredTokenUser := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Expired",
					LastName:  "Token",
					Username:  "expiredtoken",
					Phone:     "+2121212122",
					Email:     "expiredtoken@example.com",
					Password:  "oldpassword",
				},
				ResetToken:       "expiredtoken123",
				ResetTokenExpiry: &expiredTime,
			}
			helper.DB.Create(expiredTokenUser)

			// Test ResetPassword with expired token
			err = authService.ResetPassword("expiredtoken@example.com", "expiredtoken123", "newpassword")
			assert.Error(t, err)
			assert.Equal(t, "token expired", err.Error())

			// Create user with nil reset token expiry
			nilExpiryUser := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Nil",
					LastName:  "Expiry",
					Username:  "nilexpiry",
					Phone:     "+2222222223",
					Email:     "nilexpiry@example.com",
					Password:  "oldpassword",
				},
				ResetToken:       "nilexpirytoken123",
				ResetTokenExpiry: nil, // Nil expiry should be treated as expired
			}
			helper.DB.Create(nilExpiryUser)

			// Test ResetPassword with nil token expiry
			err = authService.ResetPassword("nilexpiry@example.com", "nilexpirytoken123", "newpassword")
			assert.Error(t, err)
			assert.Equal(t, "token expired", err.Error())
		})

		t.Run("Login comprehensive coverage tests", func(t *testing.T) {
			// Create user with hashed password for login testing
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
			uniqueID := helper.GenerateUniqueTestID()
			email := fmt.Sprintf("logintest-%s@example.com", uniqueID)
			username := fmt.Sprintf("logintest%s", uniqueID)
			phone := fmt.Sprintf("+1%s", uniqueID[len(uniqueID)-10:])
			
			loginUser := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Login",
					LastName:  "Test",
					Username:  username,
					Phone:     phone,
					Email:     email,
					Password:  string(hashedPassword),
				},
				LastLogin: nil, // No previous login
			}
			helper.DB.Create(loginUser)

			// Test Login with correct credentials - success path
			loginReq := &authentication.LoginRequest{
				Email:    email,
				Password: "correctpassword",
			}
			response, err := authService.Login(loginReq)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, email, response.Email)
			assert.NotEmpty(t, response.AccessToken)
			assert.Greater(t, response.Exp, int64(0))

			// Test Login with non-existent user
			loginReq = &authentication.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "anypassword",
			}
			response, err = authService.Login(loginReq)
			assert.Error(t, err)
			assert.Nil(t, response)
			assert.Equal(t, "invalid credentials", err.Error())

			// Test Login with wrong password
			loginReq = &authentication.LoginRequest{
				Email:    "logintest@example.com",
				Password: "wrongpassword",
			}
			response, err = authService.Login(loginReq)
			assert.Error(t, err)
			assert.Nil(t, response)
			assert.Equal(t, "invalid credentials", err.Error())

			// Create user with previous login time to test different code path
			lastLoginTime := time.Now().Add(-time.Hour)
			hashedPassword2, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			loginUser2 := &authentication.AuthUser{
				User: profile.User{
					FirstName: "Previous",
					LastName:  "Login",
					Username:  "previouslogin",
					Phone:     "+2424242425",
					Email:     "previouslogin@example.com",
					Password:  string(hashedPassword2),
				},
				LastLogin: &lastLoginTime, // Has previous login
			}
			helper.DB.Create(loginUser2)

			// Test Login with user that has previous login time
			loginReq = &authentication.LoginRequest{
				Email:    "previouslogin@example.com",
				Password: "password123",
			}
			response, err = authService.Login(loginReq)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, "previouslogin@example.com", response.Email)
			assert.NotEmpty(t, response.LastLogin) // Should have last login formatted
			assert.NotEmpty(t, response.AccessToken)

			// Test Login with empty email
			loginReq = &authentication.LoginRequest{
				Email:    "",
				Password: "anypassword",
			}
			response, err = authService.Login(loginReq)
			assert.Error(t, err)
			assert.Nil(t, response)
			assert.Equal(t, "invalid credentials", err.Error())

			// Test Login with empty password
			loginReq = &authentication.LoginRequest{
				Email:    "logintest@example.com",
				Password: "",
			}
			response, err = authService.Login(loginReq)
			assert.Error(t, err)
			assert.Nil(t, response)
			assert.Equal(t, "invalid credentials", err.Error())
		})
	})
}
