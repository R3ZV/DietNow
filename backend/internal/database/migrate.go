package database

import (
	"fmt"

	"github.com/dietnow/backend/internal/models"
	"gorm.io/gorm"
)

// AllModels is the ordered list of entities AutoMigrate manages. Order matters
// for foreign keys: referenced tables (User, Food) come before the tables that
// point at them.
var AllModels = []any{
	&models.User{},
	&models.UserProfile{},
	&models.Food{},
	&models.WeightLog{},
	&models.MealEntry{},
	&models.Session{},
}

// Migrate creates/updates all tables, columns, indexes, foreign keys, and CHECK
// constraints directly from the Go struct definitions. This is GORM's built-in
// migrator — no SQL scripts, no hand-written DDL. All invariants (uniqueness,
// checks) are declared via struct tags on the models.
//
// AutoMigrate is additive and idempotent: it adds missing tables/columns/indexes
// but never drops columns, so it is safe to run on every boot.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(AllModels...); err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}
	return nil
}
