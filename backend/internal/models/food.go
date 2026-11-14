package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Macros holds the four tracked macronutrients. On a Food they are expressed
// PER 100g (the normalization unit); on a MealEntry snapshot they are absolute
// amounts for the logged portion.
type Macros struct {
	CaloriesKcal float64 `json:"caloriesKcal"`
	ProteinG     float64 `json:"proteinG"`
	CarbsG       float64 `json:"carbsG"`
	FatG         float64 `json:"fatG"`
}

// Micros holds the micronutrients that feed the Longevity Engine, PER 100g.
// Common, frequently-queried nutrients get first-class columns; anything more
// exotic (specific antioxidants, polyphenols, flavonoids) lives in the flexible
// Antioxidants JSONB map so the longevity catalog can grow without migrations.
type Micros struct {
	FiberG      float64 `json:"fiberG"`
	SugarG      float64 `json:"sugarG"`
	SodiumMG    float64 `json:"sodiumMg"`
	PotassiumMG float64 `json:"potassiumMg"`
	CalciumMG   float64 `json:"calciumMg"`
	IronMG      float64 `json:"ironMg"`
	VitaminAUG  float64 `json:"vitaminAUg"`
	VitaminCMG  float64 `json:"vitaminCMg"`
	VitaminDUG  float64 `json:"vitaminDUg"`
	VitaminEMG  float64 `json:"vitaminEMg"`
	VitaminKUG  float64 `json:"vitaminKUg"`
	Omega3G     float64 `json:"omega3G"`
}

// Food is a nutrition reference item, normalized PER 100g.
//
// Global vs. private:
//   - OwnerUserID == NULL  -> global app food, visible to everyone.
//   - OwnerUserID == &id   -> private food, visible only to that user.
//
// This single-table design (rather than two tables) keeps meal logging uniform:
// a MealEntry can reference either kind via one FoodID. Visibility is enforced
// at the repository layer (see FoodRepository.Visible).
type Food struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Name  string `gorm:"not null;size:200;index" json:"name"`
	Brand string `gorm:"size:120" json:"brand,omitempty"`

	// NULL => global/shared food. Pointer so the zero value is a real SQL NULL.
	OwnerUserID *uint `gorm:"index" json:"ownerUserId,omitempty"`
	Owner       *User `gorm:"foreignKey:OwnerUserID;constraint:OnDelete:CASCADE;" json:"-"`

	// IsVerified flags curated/trusted global foods (e.g. from a seeded catalog).
	IsVerified bool `gorm:"not null;default:false" json:"isVerified"`

	// Normalized per-100g nutrition. embeddedPrefix keeps column names explicit:
	// per100g_calories_kcal, per100g_protein_g, ... per100g_vitamin_c_mg, etc.
	Nutrition Macros `gorm:"embedded;embeddedPrefix:per100g_" json:"nutrition"`
	Micros    Micros `gorm:"embedded;embeddedPrefix:per100g_" json:"micros"`

	// Extensible longevity data, e.g. {"polyphenols_mg": 210, "lycopene_ug": 4600}.
	Antioxidants datatypes.JSONMap `gorm:"type:jsonb" json:"antioxidants,omitempty"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// IsGlobal reports whether this is a shared app food (no owner).
func (f Food) IsGlobal() bool { return f.OwnerUserID == nil }

// NutritionFor scales the per-100g macros to an absolute portion in grams.
// This is what we freeze into a MealEntry snapshot at log time.
func (f Food) NutritionFor(grams float64) NutritionSnapshot {
	k := grams / 100.0
	return NutritionSnapshot{
		CaloriesKcal: f.Nutrition.CaloriesKcal * k,
		ProteinG:     f.Nutrition.ProteinG * k,
		CarbsG:       f.Nutrition.CarbsG * k,
		FatG:         f.Nutrition.FatG * k,
	}
}
