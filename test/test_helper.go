package test

import (
	"base/core/app/authentication"
	"base/core/app/profile"
	"base/core/email"
	"base/core/logger"
	"fmt"
	"os"
	"testing"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestHelper provides Rails-style test utilities
type TestHelper struct {
	DB     *gorm.DB
	Logger *MockLogger
}

// SetupTest initializes test environment (Rails-style setup)
func SetupTest(t *testing.T) *TestHelper {
	// Set test environment
	os.Setenv("GIN_MODE", "test")

	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate tables
	err = db.AutoMigrate(&profile.User{}, &authentication.AuthUser{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return &TestHelper{
		DB:     db,
		Logger: &MockLogger{},
	}
}

// TeardownTest cleans up test environment (Rails-style teardown)
func (h *TestHelper) TeardownTest() {
	// Clean up database
	if h.DB != nil {
		sqlDB, _ := h.DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

// CreateTestUser creates a test user with unique email and phone (Rails-style factory)
func (h *TestHelper) CreateTestUser(email, username, phone string) *profile.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpassword123"), bcrypt.DefaultCost)

	user := &profile.User{
		FirstName: "Test",
		LastName:  "User",
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		Phone:     phone,
	}

	h.DB.Create(user)
	return user
}

// CreateTestAuthUser creates a test auth user (Rails-style factory)
func (h *TestHelper) CreateTestAuthUser(email, username, phone string) *authentication.AuthUser {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpassword123"), bcrypt.DefaultCost)

	authUser := &authentication.AuthUser{
		User: profile.User{
			FirstName: "Auth",
			LastName:  "User",
			Username:  username,
			Email:     email,
			Password:  string(hashedPassword),
			Phone:     phone,
		},
	}

	h.DB.Create(authUser)
	return authUser
}

// MockEmailSender provides a simple mock for email sending
type MockEmailSender struct {
	ShouldFail bool
}

func (m *MockEmailSender) Send(msg email.Message) error {
	if m.ShouldFail {
		return fmt.Errorf("mock email send failure")
	}
	return nil
}

// MockLogger provides a simple mock for logging
type MockLogger struct{}

func (m *MockLogger) Info(msg string, fields ...logger.Field) {
	// Mock implementation - do nothing
}

func (m *MockLogger) Error(msg string, fields ...logger.Field) {
	// Mock implementation - do nothing
}

func (m *MockLogger) Debug(msg string, fields ...logger.Field) {
	// Mock implementation - do nothing
}

func (m *MockLogger) Warn(msg string, fields ...logger.Field) {
	// Mock implementation - do nothing
}

func (m *MockLogger) Fatal(msg string, fields ...logger.Field) {
	// Mock implementation - do nothing
}

func (m *MockLogger) With(fields ...logger.Field) logger.Logger {
	// Mock implementation - return self
	return m
}

func (m *MockLogger) GetZapLogger() *zap.Logger {
	// Mock implementation - return nil
	return nil
}

// MockActiveStorage provides a simple mock for storage operations
type MockActiveStorage struct{}

func (m *MockActiveStorage) Store(filename string, data []byte) (string, error) {
	// Mock implementation - return fake URL
	return "http://example.com/fake-file.jpg", nil
}

func (m *MockActiveStorage) Delete(url string) error {
	// Mock implementation - just return nil
	return nil
}
