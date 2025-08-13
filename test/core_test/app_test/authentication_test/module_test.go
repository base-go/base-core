package authentication_test

import (
	"base/core/app/authentication"
	"base/core/emitter"
	"base/core/router"
	"base/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticationModule(t *testing.T) {
	helper := test.SetupTest(t)
	defer helper.TeardownTest()

	t.Run("Authentication module functions for 100% coverage", func(t *testing.T) {

		t.Run("NewAuthenticationModule - success", func(t *testing.T) {
			// Create router group
			router := router.New()
			group := router.Group("/api/v1")

			// Create mocks
			emailSender := &test.MockEmailSender{}
			mockEmitter := emitter.New()

			// Test NewAuthenticationModule
			module := authentication.NewAuthenticationModule(helper.DB, group, emailSender, helper.Logger, mockEmitter)
			assert.NotNil(t, module)
		})

		t.Run("Module Migrate", func(t *testing.T) {
			// Create router group
			router := router.New()
			group := router.Group("/api/v1")

			// Create mocks
			emailSender := &test.MockEmailSender{}
			mockEmitter := emitter.New()

			// Create module and test Migrate
			module := authentication.NewAuthenticationModule(helper.DB, group, emailSender, helper.Logger, mockEmitter)
			err := module.Migrate()
			assert.NoError(t, err)
		})

		t.Run("Module GetModels", func(t *testing.T) {
			// Create router group
			router := router.New()
			group := router.Group("/api/v1")

			// Create mocks
			emailSender := &test.MockEmailSender{}
			mockEmitter := emitter.New()

			// Create module and test GetModels
			module := authentication.NewAuthenticationModule(helper.DB, group, emailSender, helper.Logger, mockEmitter)
			models := module.GetModels()
			assert.NotNil(t, models)
			assert.Greater(t, len(models), 0)
		})

		t.Run("Module Routes", func(t *testing.T) {
			// Create router group
			router := router.New()
			group := router.Group("/api/v1")

			// Create mocks
			emailSender := &test.MockEmailSender{}
			mockEmitter := emitter.New()

			// Create module and test Routes
			module := authentication.NewAuthenticationModule(helper.DB, group, emailSender, helper.Logger, mockEmitter)

			// Cast to concrete type to access Routes method
			authModule, ok := module.(*authentication.AuthenticationModule)
			assert.True(t, ok)

			// Test Routes method - this should not panic and should set up routes
			assert.NotPanics(t, func() {
				authModule.Routes(group)
			})
		})
	})
}
