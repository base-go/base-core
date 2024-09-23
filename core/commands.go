package core

import (
	"base/core/helper"
	"fmt"
)

func ExecuteCommand(args []string) {
	if len(args) < 2 {
		fmt.Println("Error: No command provided")
		printUsage()
		return
	}

	switch args[1] {
	case "seed":
		helper.SeedDatabase(false)
	case "replant":
		helper.SeedDatabase(true)
	case "feed":
		if len(args) < 3 {
			fmt.Println("Error: Not enough arguments for feed command")
			printFeedUsage()
			return
		}
		helper.FeedData(args[2:])
	default:
		fmt.Printf("Error: Unknown command '%s'\n", args[1])
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  base seed                 - Seed the database")
	fmt.Println("  base replant              - Clean and reseed the database")
	fmt.Println("  base feed [args]          - Feed data from JSON file into MySQL table")
	fmt.Println("For more information on the feed command, use: base feed")
}

func printFeedUsage() {
	fmt.Println("Usage: base feed <mysql_table[:json_path]> [field_mappings...]")
	fmt.Println("Examples:")
	fmt.Println("  base feed products")
	fmt.Println("  base feed products title:name")
	fmt.Println("  base feed products:data/custom_products.json")
	fmt.Println("  base feed products:data/custom_products.json title:name description:description")
	fmt.Println("If only the table name is provided, it will look for a JSON file named '<table>.json' in the 'data' directory.")
	fmt.Println("If no field mappings are provided, all JSON fields will be used as-is.")
	fmt.Println("If field mappings are provided, unmapped fields will still be included using their JSON field names.")
}
