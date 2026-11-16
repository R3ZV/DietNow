// Command import-foods migrates PostgreSQL and imports a USDA Foundation JSON export.
package main

import (
	"flag"
	"log"
	"os"

	"github.com/dietnow/backend/internal/database"
)

func main() {
	path := flag.String("file", "../dataset/FoodData_Central_foundation_food_json_2026-10-29.json", "USDA Foundation JSON file")
	flag.Parse()
	file, err := os.Open(*path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	db, err := database.Connect(database.ConfigFromEnv())
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatal(err)
	}
	count, err := database.ImportFoundationFoods(db, file)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Imported/updated %d USDA Foundation foods", count)
}
