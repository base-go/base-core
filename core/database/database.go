package database

import (
	"base/core/config"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB initializes the database connection based on the provided configuration.
func InitDB(cfg *config.Config) error {
	var err error
	switch cfg.DBDriver {
	case "sqlite":
		DB, err = gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	case "mysql":
		if cfg.DBURL == "" {
			cfg.DBURL = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
		}
		DB, err = gorm.Open(mysql.Open(cfg.DBURL), &gorm.Config{})
	case "postgres":
		if cfg.DBURL == "" {
			cfg.DBURL = fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
				cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBName, cfg.DBPassword)
		}
		DB, err = gorm.Open(postgres.Open(cfg.DBURL), &gorm.Config{})
	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.DBDriver)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to the database: %v", err)
	}

	log.Println("Database connection established successfully")
	return nil
}
