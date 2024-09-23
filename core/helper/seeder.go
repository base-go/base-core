package helper

import (
	"base/app"
	"base/core/config"
	"base/core/database"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func FeedData(args []string) {
	var mysqlTable, jsonPath string
	var fieldMappings map[string]string

	// Parse table and file path
	if !strings.Contains(args[0], ":") {
		mysqlTable = args[0]
		jsonPath = filepath.Join("data", mysqlTable+".json")
		fieldMappings = parseFieldMappings(args[1:])
	} else {
		tablePath := strings.Split(args[0], ":")
		if len(tablePath) != 2 {
			fmt.Println("Error: Invalid table:path format")
			return
		}
		mysqlTable, jsonPath = tablePath[0], tablePath[1]
		fieldMappings = parseFieldMappings(args[1:])
	}

	// Read JSON file
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Printf("Error reading JSON file: %v\n", err)
		return
	}

	// Parse JSON data
	var data []map[string]interface{}
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		fmt.Printf("Error parsing JSON data: %v\n", err)
		return
	}

	// Initialize config
	cfg := config.NewConfig()

	// Initialize the database
	db, err := database.InitDB(cfg)
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		return
	}

	// Insert data into MySQL table
	for _, item := range data {
		insertData := make(map[string]interface{})

		if len(fieldMappings) == 0 {
			// If no mappings provided, use all JSON fields
			insertData = item
		} else {
			// Apply specific mappings
			for jsonField, mysqlField := range fieldMappings {
				if value, ok := item[jsonField]; ok {
					insertData[mysqlField] = value
				}
			}

			// For fields not in the mapping, use the JSON field name as is
			for jsonField, value := range item {
				if _, mapped := fieldMappings[jsonField]; !mapped {
					insertData[jsonField] = value
				}
			}
		}

		if len(insertData) > 0 {
			result := db.Table(mysqlTable).Create(insertData)
			if result.Error != nil {
				fmt.Printf("Error inserting data: %v\n", result.Error)
			} else {
				fmt.Printf("Inserted data: %v\n", insertData)
			}
		}
	}

	fmt.Println("Data feed complete.")
}

func parseFieldMappings(args []string) map[string]string {
	fieldMappings := make(map[string]string)
	for _, mapping := range args {
		parts := strings.Split(mapping, ":")
		if len(parts) != 2 {
			fmt.Printf("Warning: Ignoring invalid mapping '%s'\n", mapping)
			continue
		}
		fieldMappings[parts[0]] = parts[1]
	}
	return fieldMappings
}
