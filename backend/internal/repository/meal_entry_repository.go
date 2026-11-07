package repository

import (
	"context"
	"errors"

	"github.com/dietnow/backend/internal/models"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a record does not exist OR is not owned by the
// requesting user. We deliberately do NOT distinguish the two cases, so an
// attacker probing with another user's IDs cannot tell "doesn't exist" from
// "exists but isn't yours".
var ErrNotFound = errors.New("record not found")

type MealEntryRepository struct {
	db *gorm.DB
}

func NewMealEntryRepository(db *gorm.DB) *MealEntryRepository {
	return &MealEntryRepository{db: db}
}

// ListForDay returns a user's entries for a single calendar day, ordered for the
// 4-meal tracker. The OwnedBy scope + the explicit consumed_on filter guarantee
// isolation and hit the (user_id, consumed_on) composite index.
func (r *MealEntryRepository) ListForDay(ctx context.Context, userID uint, day models.LocalDate) ([]models.MealEntry, error) {
	var entries []models.MealEntry
	err := r.db.WithContext(ctx).
		Scopes(OwnedBy(userID)).
		Where("consumed_on = ?", day).
		Order("meal_type, id").
		Find(&entries).Error
	return entries, err
}

// DailyTotals sums the frozen snapshot columns for a day — the dashboard's
// "consumed so far" figure. Summing snapshots (not re-joining Food) is what
// keeps historical totals stable after a food definition changes.
func (r *MealEntryRepository) DailyTotals(ctx context.Context, userID uint, day models.LocalDate) (models.NutritionSnapshot, error) {
	var totals models.NutritionSnapshot
	err := r.db.WithContext(ctx).
		Model(&models.MealEntry{}).
		Scopes(OwnedBy(userID)).
		Where("consumed_on = ?", day).
		Select(`
			COALESCE(SUM(snap_calories_kcal), 0) AS calories_kcal,
			COALESCE(SUM(snap_protein_g),     0) AS protein_g,
			COALESCE(SUM(snap_carbs_g),       0) AS carbs_g,
			COALESCE(SUM(snap_fat_g),         0) AS fat_g`).
		Scan(&totals).Error
	return totals, err
}

// Create inserts an entry. The UserID on the struct is the owner; callers must
// set it from the authenticated session, never from client input.
func (r *MealEntryRepository) Create(ctx context.Context, e *models.MealEntry) error {
	return r.db.WithContext(ctx).Create(e).Error
}

// UpdatePortion changes a portion and re-freezes the macro snapshot. The WHERE
// carries BOTH id AND user_id, so a user can never mutate another user's row —
// a mismatched owner updates zero rows and returns ErrNotFound.
func (r *MealEntryRepository) UpdatePortion(ctx context.Context, userID, entryID uint, grams float64, snap models.NutritionSnapshot) error {
	res := r.db.WithContext(ctx).
		Model(&models.MealEntry{}).
		Scopes(OwnedBy(userID)).
		Where("id = ?", entryID).
		Updates(map[string]any{
			"portion_grams":       grams,
			"snap_calories_kcal":  snap.CaloriesKcal,
			"snap_protein_g":      snap.ProteinG,
			"snap_carbs_g":        snap.CarbsG,
			"snap_fat_g":          snap.FatG,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes an entry only if it belongs to the user. Same owner-guard as
// UpdatePortion: RowsAffected == 0 means "not yours (or gone)".
func (r *MealEntryRepository) Delete(ctx context.Context, userID, entryID uint) error {
	res := r.db.WithContext(ctx).
		Scopes(OwnedBy(userID)).
		Where("id = ?", entryID).
		Delete(&models.MealEntry{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
