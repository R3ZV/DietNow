package models

import "time"

// NutritionSnapshot is the absolute macro contribution of a logged portion,
// frozen at the moment of logging.
type NutritionSnapshot struct {
	CaloriesKcal float64 `json:"caloriesKcal"`
	ProteinG     float64 `json:"proteinG"`
	CarbsG       float64 `json:"carbsG"`
	FatG         float64 `json:"fatG"`
}

// MealEntry is one logged item in the 4-meal daily tracker.
//
// Snapshotting: besides the FoodID reference we persist FoodName and the scaled
// Nutrition (per PortionGrams) directly on the row. This means a historical log
// never silently changes if the underlying Food is later edited, and it survives
// the Food being deleted (FoodID is nullable with ON DELETE SET NULL). The
// dashboard sums these snapshot columns — it never has to re-join and re-scale.
type MealEntry struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"not null;index:idx_meal_user_date,priority:1" json:"userId"`

	// Zone-less calendar day the food was consumed (see LocalDate). Combined
	// with UserID in a non-unique index: fetching "today's" entries for a user
	// is an index range scan.
	ConsumedOn LocalDate `gorm:"type:date;not null;index:idx_meal_user_date,priority:2" json:"consumedOn"`
	MealType   MealType  `gorm:"size:16;not null" json:"mealType"`

	// Reference to the source food. Nullable + SET NULL so deleting a Food keeps
	// the historical entry intact (snapshot fields below still describe it).
	FoodID *uint `gorm:"index" json:"foodId,omitempty"`
	Food   *Food `gorm:"foreignKey:FoodID;constraint:OnDelete:SET NULL;" json:"food,omitempty"`

	PortionGrams float64 `gorm:"not null;check:chk_meal_portion_positive,portion_grams > 0" json:"portionGrams"`

	// --- Immutable snapshot (columns: snap_food_name, snap_calories_kcal, ...) ---
	FoodName  string            `gorm:"column:snap_food_name;size:200;not null" json:"foodName"`
	Nutrition NutritionSnapshot `gorm:"embedded;embeddedPrefix:snap_" json:"nutrition"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NewMealEntry builds an entry from a food + portion, freezing the snapshot.
// Keeping this in the model guarantees the reference and the snapshot can never
// disagree at creation time.
func NewMealEntry(userID uint, day LocalDate, meal MealType, food Food, grams float64) MealEntry {
	return MealEntry{
		UserID:       userID,
		ConsumedOn:   day,
		MealType:     meal,
		FoodID:       &food.ID,
		PortionGrams: grams,
		FoodName:     food.Name,
		Nutrition:    food.NutritionFor(grams),
	}
}
