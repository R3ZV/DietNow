package models

import (
	"time"

	"gorm.io/gorm"
)

// User is the identity/auth root. Everything a user owns (profile, weigh-ins,
// meal entries, private foods) hangs off this record and is deleted with it
// (OnDelete:CASCADE for owned rows; private foods cascade, global foods are
// untouched because they have a NULL owner).
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash string         `gorm:"not null;size:255" json:"-"` // never serialized
	DisplayName  string         `gorm:"size:120" json:"displayName"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"` // soft delete

	// Associations.
	Profile     *UserProfile `gorm:"constraint:OnDelete:CASCADE;" json:"profile,omitempty"`
	WeightLogs  []WeightLog  `gorm:"constraint:OnDelete:CASCADE;" json:"weightLogs,omitempty"`
	MealEntries []MealEntry  `gorm:"constraint:OnDelete:CASCADE;" json:"mealEntries,omitempty"`
	// Private foods created by this user. Global foods have OwnerUserID = NULL
	// and are intentionally NOT part of this cascade.
	Foods []Food `gorm:"foreignKey:OwnerUserID;constraint:OnDelete:CASCADE;" json:"foods,omitempty"`
}
