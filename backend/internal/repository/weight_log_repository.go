package repository

import (
	"context"

	"github.com/dietnow/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WeightLogRepository struct {
	db *gorm.DB
}

func NewWeightLogRepository(db *gorm.DB) *WeightLogRepository {
	return &WeightLogRepository{db: db}
}

// Upsert records (or overwrites) the weigh-in for a given day. The unique
// (user_id, logged_on) index makes "one weigh-in per day" an upsert instead of
// a duplicate; conflict handling updates weight/note in place.
func (r *WeightLogRepository) Upsert(ctx context.Context, log *models.WeightLog) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "logged_on"}},
			DoUpdates: clause.AssignmentColumns([]string{"weight_kg", "note", "updated_at"}),
		}).
		Create(log).Error
}

// History returns a user's weigh-ins within an inclusive date range, oldest
// first — ready to plot the trajectory toward target weight. OwnedBy guarantees
// isolation; the LocalDate bounds are plain, zone-less DATE comparisons.
func (r *WeightLogRepository) History(ctx context.Context, userID uint, from, to models.LocalDate) ([]models.WeightLog, error) {
	var logs []models.WeightLog
	err := r.db.WithContext(ctx).
		Scopes(OwnedBy(userID)).
		Where("logged_on BETWEEN ? AND ?", from, to).
		Order("logged_on ASC").
		Find(&logs).Error
	return logs, err
}

// Latest returns the most recent weigh-in for a user, or ErrNotFound if none.
func (r *WeightLogRepository) Latest(ctx context.Context, userID uint) (models.WeightLog, error) {
	var log models.WeightLog
	err := r.db.WithContext(ctx).
		Scopes(OwnedBy(userID)).
		Order("logged_on DESC").
		First(&log).Error
	if err == gorm.ErrRecordNotFound {
		return models.WeightLog{}, ErrNotFound
	}
	return log, err
}

// Delete removes a weigh-in only if it belongs to the user.
func (r *WeightLogRepository) Delete(ctx context.Context, userID, logID uint) error {
	res := r.db.WithContext(ctx).
		Scopes(OwnedBy(userID)).
		Where("id = ?", logID).
		Delete(&models.WeightLog{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
