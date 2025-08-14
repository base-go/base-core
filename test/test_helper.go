package test

import (
	"base/core/app/authentication"
	"base/core/app/profile"
	"base/core/email"
	"base/core/logger"
	"fmt"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
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

	// Create in-memory SQLite database for testing (with silent logging)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
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

// CleanDatabase truncates all tables for a clean test state
func (h *TestHelper) CleanDatabase() {
	if h.DB != nil {
		// Delete all records from tables (in correct order to handle foreign keys)
		h.DB.Exec("DELETE FROM users")
		h.DB.Exec("DELETE FROM sqlite_sequence WHERE name='users'") // Reset auto-increment
	}
}

// GenerateUniqueTestID generates a unique identifier for test data to prevent conflicts
func (h *TestHelper) GenerateUniqueTestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
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

	if err := h.DB.Create(user).Error; err != nil {
		// Log error but continue - let the test handle the failure
		fmt.Printf("Warning: Failed to create test user: %v\n", err)
	}
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

	if err := h.DB.Create(authUser).Error; err != nil {
		// Log error but continue - let the test handle the failure
		fmt.Printf("Warning: Failed to create test auth user: %v\n", err)
	}
	return authUser
}

// CreateUniqueTestUser creates a test user with auto-generated unique identifiers
func (h *TestHelper) CreateUniqueTestUser(prefix string) *profile.User {
	unique := h.GenerateUniqueTestID()
	email := fmt.Sprintf("%s-%s@example.com", prefix, unique)
	username := fmt.Sprintf("%s%s", prefix, unique)
	phone := fmt.Sprintf("+1%s", unique[len(unique)-10:]) // Use last 10 digits
	return h.CreateTestUser(email, username, phone)
}

// CreateUniqueTestAuthUser creates a test auth user with auto-generated unique identifiers
func (h *TestHelper) CreateUniqueTestAuthUser(prefix string) *authentication.AuthUser {
	unique := h.GenerateUniqueTestID()
	email := fmt.Sprintf("%s-%s@example.com", prefix, unique)
	username := fmt.Sprintf("%s%s", prefix, unique)
	phone := fmt.Sprintf("+1%s", unique[len(unique)-10:]) // Use last 10 digits
	return h.CreateTestAuthUser(email, username, phone)
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
