// Command migrate runs GORM AutoMigrate against the configured database.
//
//	go run ./cmd/migrate
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
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migration complete")
}
