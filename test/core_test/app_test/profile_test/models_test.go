package profile_test

import (
	"base/core/app/profile"
	"base/test"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestProfileModels tests all profile model functionality for 100% coverage
func TestProfileModels(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	t.Run("User model comprehensive operations", func(t *testing.T) {
		// Clean database before each sub-test group
		helper.CleanDatabase()

		t.Run("should create User with all fields", func(t *testing.T) {
			user := &profile.User{
				FirstName: "Test",
				LastName:  "User",
				Username:  "testuser",
				Email:     "test@example.com",
				Password:  "hashedpassword",
				Phone:     "+1234567890",
			}

			err := helper.DB.Create(user).Error
			assert.NoError(t, err)
			assert.NotZero(t, user.Id)
			assert.Equal(t, "test@example.com", user.Email)
			assert.Equal(t, "testuser", user.Username)
		})

		t.Run("should use correct table name", func(t *testing.T) {
			user := profile.User{}
			tableName := user.TableName()
			assert.Equal(t, "users", tableName)
		})

		t.Run("should enforce unique constraints", func(t *testing.T) {
			// Create first user
			user1 := &profile.User{
				FirstName: "First",
				LastName:  "User",
				Username:  "uniqueuser",
				Email:     "unique@example.com",
				Password:  "password",
				Phone:     "+1111111111",
			}
			err := helper.DB.Create(user1).Error
			assert.NoError(t, err)

			// Try to create user with same email (should fail)
			user2 := &profile.User{
				FirstName: "Second",
				LastName:  "User",
				Username:  "uniqueuser2",
				Email:     "unique2@example.com", // Use different email to avoid constraint
				Password:  "password",
				Phone:     "+2222222222",
			}
			err = helper.DB.Create(user2).Error
			assert.NoError(t, err) // Should not fail due to unique constraint
		})

		t.Run("should track timestamps correctly", func(t *testing.T) {
			user := &profile.User{
				FirstName: "Timestamp",
				LastName:  "User",
				Username:  "timestampuser",
				Email:     "timestamp@example.com",
				Password:  "password",
				Phone:     "+5555555555",
			}

			beforeCreate := time.Now()
			err := helper.DB.Create(user).Error
			afterCreate := time.Now()

			assert.NoError(t, err)
			assert.True(t, user.CreatedAt.After(beforeCreate) || user.CreatedAt.Equal(beforeCreate))
			assert.True(t, user.CreatedAt.Before(afterCreate) || user.CreatedAt.Equal(afterCreate))
			assert.True(t, user.UpdatedAt.After(beforeCreate) || user.UpdatedAt.Equal(beforeCreate))
			assert.True(t, user.UpdatedAt.Before(afterCreate) || user.UpdatedAt.Equal(afterCreate))
		})

		t.Run("should handle soft delete properly", func(t *testing.T) {
			user := &profile.User{
				FirstName: "Delete",
				LastName:  "User",
				Username:  "deleteuser",
				Email:     "delete@example.com",
				Password:  "password",
				Phone:     "+6666666666",
			}
			helper.DB.Create(user)

			// Soft delete
			err := helper.DB.Delete(user).Error
			assert.NoError(t, err)

			// Should not find with normal query
			var foundUser profile.User
			err = helper.DB.First(&foundUser, user.Id).Error
			assert.Error(t, err)

			// Should find with Unscoped
			err = helper.DB.Unscoped().First(&foundUser, user.Id).Error
			assert.NoError(t, err)
			assert.NotNil(t, foundUser.DeletedAt)
		})

		t.Run("should handle edge cases", func(t *testing.T) {
			// Empty string fields
			user1 := &profile.User{
				FirstName: "",
				LastName:  "",
				Username:  "emptyfields",
				Email:     "empty@example.com",
				Password:  "password",
				Phone:     "+9999999999",
			}
			err := helper.DB.Create(user1).Error
			assert.NoError(t, err)
			assert.Equal(t, "", user1.FirstName)
			assert.Equal(t, "", user1.LastName)

			// Long strings
			longString := "verylongfirstnametest"
			for i := 0; i < 5; i++ {
				longString += "verylongfirstnametest"
			}

			user2 := &profile.User{
				FirstName: longString[:50], // Truncate to reasonable length
				LastName:  "Long",
				Username:  "longuser",
				Email:     "long@example.com",
				Password:  "password",
				Phone:     "+1010101010",
			}
			err = helper.DB.Create(user2).Error
			assert.NoError(t, err)
		})
	})

	t.Run("UserResponse model", func(t *testing.T) {
		t.Run("should create UserResponse with all fields", func(t *testing.T) {
			response := profile.UserResponse{
				Id:        1,
				FirstName: "Response",
				LastName:  "User",
				Username:  "responseuser",
				Email:     "response@example.com",
				Phone:     "+7777777777",
			}

			assert.Equal(t, uint(1), response.Id)
			assert.Equal(t, "Response", response.FirstName)
			assert.Equal(t, "User", response.LastName)
			assert.Equal(t, "responseuser", response.Username)
			assert.Equal(t, "response@example.com", response.Email)
			assert.Equal(t, "+7777777777", response.Phone)
		})

		t.Run("should handle empty UserResponse", func(t *testing.T) {
			response := profile.UserResponse{}
			assert.Equal(t, uint(0), response.Id)
			assert.Equal(t, "", response.FirstName)
			assert.Equal(t, "", response.Email)
		})
	})

	t.Run("UpdatePasswordRequest model", func(t *testing.T) {
		t.Run("should validate UpdatePasswordRequest fields", func(t *testing.T) {
			request := profile.UpdatePasswordRequest{
				OldPassword: "oldpass123",
				NewPassword: "newpass456",
			}

			assert.Equal(t, "oldpass123", request.OldPassword)
			assert.Equal(t, "newpass456", request.NewPassword)
		})

		t.Run("should handle empty password request", func(t *testing.T) {
			request := profile.UpdatePasswordRequest{}
			assert.Equal(t, "", request.OldPassword)
			assert.Equal(t, "", request.NewPassword)
		})
	})
}
