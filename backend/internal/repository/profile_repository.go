package repository

import (
	"context"
	"errors"

	"github.com/dietnow/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProfileRepository reads and writes the one-to-one biometric profile.
type ProfileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

// Get returns the user's profile, or ErrNotFound if onboarding is not done.
func (r *ProfileRepository) Get(ctx context.Context, userID uint) (models.UserProfile, error) {
	var p models.UserProfile
	err := r.db.WithContext(ctx).Scopes(OwnedBy(userID)).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.UserProfile{}, ErrNotFound
	}
	return p, err
}

// Upsert creates or replaces the user's profile. The owner is forced from the
// trusted userID argument; the unique user_id index turns a second save into
// an update of the same row.
func (r *ProfileRepository) Upsert(ctx context.Context, userID uint, p *models.UserProfile) error {
	p.ID = 0
	p.UserID = userID
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"gender", "birth_date", "height_cm", "current_weight_kg", "target_weight_kg",
				"activity_level", "goal", "timeframe_weeks", "timezone",
				"bmr", "tdee", "weekly_rate_kg", "daily_adjustment_kcal",
				"target_calories", "target_protein_g", "target_carbs_g", "target_fat_g", "updated_at",
			}),
		}).
		Create(p).Error
}
