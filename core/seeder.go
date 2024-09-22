package core

import (
	"base/app"
	"base/core/config"
	"base/core/database"
	"fmt"
	"os"

	"gorm.io/gorm"
)

func SeedDatabase(clean bool) {
	cfg := config.NewConfig()
	db, err := database.InitDB(cfg)
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}

	if clean {
		fmt.Println("Cleaning database...")
		if err := cleanDatabase(db.DB); err != nil {
			fmt.Printf("Error cleaning database: %v\n", err)
			os.Exit(1)
		}
	}

	if err := runSeeders(db.DB); err != nil {
		fmt.Printf("Error seeding database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Database seeded successfully.")
}

func runSeeders(db *gorm.DB) error {
	seeders := app.InitializeSeeders()
	for _, seeder := range seeders {
		fmt.Printf("Seeding %T...\n", seeder)
		if err := seeder.Seed(db); err != nil {
			return fmt.Errorf("error seeding %T: %v", seeder, err)
		}
	}
	return nil
}

func cleanDatabase(db *gorm.DB) error {
	// Get all table names
	var tableNames []string
	if err := db.Table("information_schema.tables").Where("table_schema = ?", "public").Pluck("table_name", &tableNames).Error; err != nil {
		return err
	}

	// Disable foreign key checks
	db.Exec("SET CONSTRAINTS ALL DEFERRED")

	// Truncate all tables
	for _, tableName := range tableNames {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tableName)).Error; err != nil {
			return err
		}
	}

	// Re-enable foreign key checks
	db.Exec("SET CONSTRAINTS ALL IMMEDIATE")

	return nil
}
