// Command seed migrates then inserts the idempotent test dataset.
//
//	go run ./cmd/seed
package main

import (
	"log"

	"github.com/dietnow/backend/internal/database"
)

func main() {
	db, err := database.Connect(database.ConfigFromEnv())
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	// Ensure the schema exists before seeding (safe/idempotent).
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := database.SeedDatabase(db); err != nil {
		log.Fatalf("seed: %v", err)
	}
	log.Println("seed complete")
}
