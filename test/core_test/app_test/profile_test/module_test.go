package profile_test

import (
	"base/core/app/profile"
	"base/core/storage"
	"base/test"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// createMockActiveStorageForModule creates a mock ActiveStorage for testing
func createMockActiveStorageForModule(db *gorm.DB) (*storage.ActiveStorage, error) {
	config := storage.Config{
		Provider: "local",
		Path:     "/tmp/test-storage",
		BaseURL:  "http://localhost:8080/storage",
	}
	return storage.NewActiveStorage(db, config)
}

func TestProfileModule(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	t.Run("Profile module functions for 100% coverage", func(t *testing.T) {

		t.Run("NewUserModule - success", func(t *testing.T) {
			// Create mock storage
			mockStorage, err := createMockActiveStorageForModule(helper.DB)
			assert.NoError(t, err)

			// Create router group
			router := gin.New()
			group := router.Group("/api/v1")

			// Test NewUserModule
			module := profile.NewUserModule(helper.DB, group, helper.Logger, mockStorage)
			assert.NotNil(t, module)
		})

		t.Run("Module Migrate", func(t *testing.T) {
			// Create mock storage
			mockStorage, err := createMockActiveStorageForModule(helper.DB)
			assert.NoError(t, err)

			// Create router group
			router := gin.New()
			group := router.Group("/api/v1")

			// Create module and test Migrate
			module := profile.NewUserModule(helper.DB, group, helper.Logger, mockStorage)
			err = module.Migrate()
			assert.NoError(t, err)
		})

		t.Run("Module Migrate Success Case", func(t *testing.T) {
			// Test successful migration to ensure we hit all code paths
			// Create mock storage
			mockStorage, err := createMockActiveStorageForModule(helper.DB)
			assert.NoError(t, err)

			// Create router group
			router := gin.New()
			group := router.Group("/api/v1")

			// Create module
			module := profile.NewUserModule(helper.DB, group, helper.Logger, mockStorage)
			
			// Test Migrate - should succeed and hit success path
			err = module.Migrate()
			assert.NoError(t, err)
		})

		t.Run("Module Migrate comprehensive coverage", func(t *testing.T) {
			// Test multiple successful migrations to hit all code paths
			// Create mock storage
			mockStorage, err := createMockActiveStorageForModule(helper.DB)
			assert.NoError(t, err)

			// Create router group
			router := gin.New()
			group := router.Group("/api/v1")

			// Create module and test multiple migrations
			module := profile.NewUserModule(helper.DB, group, helper.Logger, mockStorage)
			
			// Test first migration - should succeed
			err = module.Migrate()
			assert.NoError(t, err)

			// Test second migration - should also succeed (idempotent)
			err = module.Migrate()
			assert.NoError(t, err)

			// Test migration with different module instance
			module2 := profile.NewUserModule(helper.DB, group, helper.Logger, mockStorage)
			err = module2.Migrate()
			assert.NoError(t, err)
		})

		t.Run("Module GetModels", func(t *testing.T) {
			// Create mock storage
			mockStorage, err := createMockActiveStorageForModule(helper.DB)
			assert.NoError(t, err)

			// Create router group
			router := gin.New()
			group := router.Group("/api/v1")

			// Create module and test GetModels
			module := profile.NewUserModule(helper.DB, group, helper.Logger, mockStorage)
			models := module.GetModels()
			assert.NotNil(t, models)
			assert.Greater(t, len(models), 0)
		})

		t.Run("Module Routes", func(t *testing.T) {
			// Create mock storage
			mockStorage, err := createMockActiveStorageForModule(helper.DB)
			assert.NoError(t, err)

			// Create router group
			router := gin.New()
			group := router.Group("/api/v1")

			// Create module and test Routes
			module := profile.NewUserModule(helper.DB, group, helper.Logger, mockStorage)
			
			// Cast to concrete type to access Routes method
			userModule, ok := module.(*profile.UserModule)
			assert.True(t, ok)
			
			// Test Routes method - this should not panic and should set up routes
			assert.NotPanics(t, func() {
				userModule.Routes(group)
			})
		})

		t.Run("Module GetModelNames", func(t *testing.T) {
			// Create mock storage
			mockStorage, err := createMockActiveStorageForModule(helper.DB)
			assert.NoError(t, err)

			// Create router group
			router := gin.New()
			group := router.Group("/api/v1")

			// Create module and test GetModelNames
			module := profile.NewUserModule(helper.DB, group, helper.Logger, mockStorage)
			
			// Cast to concrete type to access GetModelNames method
			userModule, ok := module.(*profile.UserModule)
			assert.True(t, ok)
			
			modelNames := userModule.GetModelNames()
			assert.NotNil(t, modelNames)
			assert.Greater(t, len(modelNames), 0)
		})
	})
}
