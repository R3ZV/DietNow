package models

import "time"

// WeightLog is a single historical weigh-in used for the progress trajectory.
//
// The composite UNIQUE index (user_id, logged_on) enforces at most one weigh-in
// per calendar day per user — re-logging the same day is an upsert, not a
// duplicate. LoggedOn is a zone-less DATE (see LocalDate), so a weigh-in never
// drifts to the previous/next day due to server timezone.
type WeightLog struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"not null;index:idx_weightlog_user_date,unique,priority:1" json:"userId"`

	LoggedOn LocalDate `gorm:"type:date;not null;index:idx_weightlog_user_date,unique,priority:2" json:"loggedOn"`
	WeightKG float64   `gorm:"not null;check:chk_weightlog_weight_positive,weight_kg > 0" json:"weightKg"`
	Note     string    `gorm:"size:255" json:"note,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
