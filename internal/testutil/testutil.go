// Package testutil holds helpers shared by the framework's tests.
package testutil

import (
	"path/filepath"
	"testing"

	"base/core/logger"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// NewDB opens a SQLite database in a temporary folder that the test removes when it ends.
func NewDB(t testing.TB) *gorm.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: gormLogger.Discard})
	if err != nil {
		t.Fatalf("open sqlite database %s: %v", path, err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})
	return db
}

// Logger returns a logger that discards everything.
func Logger() logger.Logger {
	return logger.NewLoggerFromZap(zap.NewNop())
}
