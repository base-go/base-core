package helper

import (
	"base/core/config"
	"base/core/database"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FieldMapping struct {
	Source []string
	Target string
}

func FeedData(args []string) {
	var mysqlTable, jsonPath string
	var fieldMappings []FieldMapping

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
			for _, mapping := range fieldMappings {
				if len(mapping.Source) == 1 {
					// Single field mapping
					if value, ok := item[mapping.Source[0]]; ok {
						insertData[mapping.Target] = value
					}
				} else {
					// Multiple field concatenation
					var values []string
					for _, source := range mapping.Source {
						if value, ok := item[source].(string); ok {
							values = append(values, value)
						}
					}
					insertData[mapping.Target] = strings.Join(values, " ")
				}
			}

			// For fields not in the mapping, use the JSON field name as is
			for jsonField, value := range item {
				if !isMapped(jsonField, fieldMappings) {
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

func parseFieldMappings(args []string) []FieldMapping {
	var fieldMappings []FieldMapping
	for _, mapping := range args {
		parts := strings.Split(mapping, ":")
		if len(parts) != 2 {
			fmt.Printf("Warning: Ignoring invalid mapping '%s'\n", mapping)
			continue
		}
		sources := strings.Split(strings.Trim(parts[0], "\""), " ")
		fieldMappings = append(fieldMappings, FieldMapping{Source: sources, Target: parts[1]})
	}
	return fieldMappings
}

func isMapped(field string, mappings []FieldMapping) bool {
	for _, mapping := range mappings {
		for _, source := range mapping.Source {
			if source == field {
				return true
			}
		}
	}
	return false
}
