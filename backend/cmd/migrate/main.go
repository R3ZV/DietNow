// Command migrate applies the database schema (GORM AutoMigrate) and the
// reference food catalog. It is idempotent and additive, so it is safe to run
// on every deploy.
//
//	go run ./cmd/migrate            (locally, reads backend/.env)
//	/app/migrate                    (inside the release image)
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
	if err := database.SeedFoodCatalog(db); err != nil {
		log.Fatalf("seed food catalog: %v", err)
	}
	log.Println("migration complete")
}
