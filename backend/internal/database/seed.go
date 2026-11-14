package database

import (
	"fmt"
	"time"

	"github.com/dietnow/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SeedDatabase inserts a deterministic test dataset: one user + biometric
// profile, a small catalog of global foods, and a few meal entries for "today".
//
// It is fully IDEMPOTENT — safe to run on every boot and across restarts:
//   - users/foods are matched on their natural keys (email / name+owner) with
//     FirstOrCreate, so re-running never duplicates them;
//   - the profile is upserted on its unique user_id;
//   - meal entries are cleared and re-created for the seed day so the sample
//     dashboard is stable regardless of prior runs.
//
// Everything runs in a single transaction: a failure leaves the DB untouched.
func SeedDatabase(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		user, err := seedUser(tx)
		if err != nil {
			return err
		}
		if err := seedProfile(tx, user.ID); err != nil {
			return err
		}
		foods, err := seedGlobalFoods(tx)
		if err != nil {
			return err
		}
		if err := seedPrivateFood(tx, user.ID); err != nil {
			return err
		}
		return seedMealEntries(tx, user, foods)
	})
}

func seedUser(tx *gorm.DB) (models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, fmt.Errorf("hash seed password: %w", err)
	}
	user := models.User{
		Email:        "test@dietnow.app",
		PasswordHash: string(hash),
		DisplayName:  "Test User",
	}
	// Match on the unique email; only create if absent. On an existing row the
	// stored PasswordHash is left as-is.
	if err := tx.Where(models.User{Email: user.Email}).
		Attrs(models.User{PasswordHash: user.PasswordHash, DisplayName: user.DisplayName}).
		FirstOrCreate(&user).Error; err != nil {
		return models.User{}, fmt.Errorf("seed user: %w", err)
	}
	return user, nil
}

func seedProfile(tx *gorm.DB, userID uint) error {
	profile := models.UserProfile{
		UserID:          userID,
		Gender:          models.GenderFemale,
		BirthDate:       models.NewLocalDate(1998, time.March, 14),
		HeightCM:        168,
		CurrentWeightKG: 62,
		TargetWeightKG:  58,
		ActivityLevel:   models.ActivityModerate,
		Goal:            models.GoalCut,
		Timezone:        "Europe/Bucharest",
		// Persisted targets (as if the calorie engine already ran).
		BMR:            1430,
		TDEE:           2216,
		TargetCalories: 1800,
		TargetProteinG: 140,
		TargetCarbsG:   160,
		TargetFatG:     60,
	}
	// Upsert on the unique user_id so re-seeding refreshes the targets in place.
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		UpdateAll: true,
	}).Create(&profile).Error; err != nil {
		return fmt.Errorf("seed profile: %w", err)
	}
	return nil
}

// seedGlobalFoods inserts verified, longevity-relevant foods (OwnerUserID = NULL).
// Returns them keyed by name for use when building meal entries.
func seedGlobalFoods(tx *gorm.DB) (map[string]models.Food, error) {
	catalog := []models.Food{
		{
			Name:       "Rolled Oats",
			IsVerified: true,
			Nutrition:  models.Macros{CaloriesKcal: 389, ProteinG: 16.9, CarbsG: 66.3, FatG: 6.9},
			Micros:     models.Micros{FiberG: 10.6, IronMG: 4.7, PotassiumMG: 429, CalciumMG: 54},
		},
		{
			Name:       "Blueberries",
			IsVerified: true,
			Nutrition:  models.Macros{CaloriesKcal: 57, ProteinG: 0.7, CarbsG: 14.5, FatG: 0.3},
			Micros:     models.Micros{FiberG: 2.4, VitaminCMG: 9.7, VitaminKUG: 19.3},
			// Flavonoid antioxidants are a key longevity marker for berries.
			Antioxidants: datatypes.JSONMap{"anthocyanins_mg": 163.0, "polyphenols_mg": 560.0},
		},
		{
			Name:       "Chicken Breast, cooked",
			IsVerified: true,
			Nutrition:  models.Macros{CaloriesKcal: 165, ProteinG: 31, CarbsG: 0, FatG: 3.6},
			Micros:     models.Micros{PotassiumMG: 256, SodiumMG: 74},
		},
		{
			Name:       "Extra Virgin Olive Oil",
			IsVerified: true,
			Nutrition:  models.Macros{CaloriesKcal: 884, ProteinG: 0, CarbsG: 0, FatG: 100},
			Micros:     models.Micros{VitaminEMG: 14.4, VitaminKUG: 60.2},
			Antioxidants: datatypes.JSONMap{"polyphenols_mg": 62.0, "oleocanthal_mg": 9.0},
		},
		{
			Name:       "Broccoli, steamed",
			IsVerified: true,
			Nutrition:  models.Macros{CaloriesKcal: 35, ProteinG: 2.4, CarbsG: 7.2, FatG: 0.4},
			Micros:     models.Micros{FiberG: 3.3, VitaminCMG: 65, VitaminKUG: 141, CalciumMG: 40},
			Antioxidants: datatypes.JSONMap{"sulforaphane_mg": 1.5},
		},
	}

	out := make(map[string]models.Food, len(catalog))
	for i := range catalog {
		f := catalog[i]
		// Natural key for a GLOBAL food = name with a NULL owner. FirstOrCreate
		// with that filter makes re-seeding a no-op.
		if err := tx.Where("name = ? AND owner_user_id IS NULL", f.Name).
			Attrs(f).
			FirstOrCreate(&f).Error; err != nil {
			return nil, fmt.Errorf("seed food %q: %w", f.Name, err)
		}
		out[f.Name] = f
	}
	return out, nil
}

// seedPrivateFood demonstrates a user-owned (private) food.
func seedPrivateFood(tx *gorm.DB, userID uint) error {
	food := models.Food{
		Name:        "Grandma's Protein Pancakes",
		OwnerUserID: &userID,
		Nutrition:   models.Macros{CaloriesKcal: 180, ProteinG: 15, CarbsG: 18, FatG: 5},
		Micros:      models.Micros{FiberG: 2, CalciumMG: 120},
	}
	if err := tx.Where("name = ? AND owner_user_id = ?", food.Name, userID).
		Attrs(food).
		FirstOrCreate(&food).Error; err != nil {
		return fmt.Errorf("seed private food: %w", err)
	}
	return nil
}

func seedMealEntries(tx *gorm.DB, user models.User, foods map[string]models.Food) error {
	// Build "today" in the seed user's own timezone — the correct, bug-free way
	// to decide which calendar day the sample data belongs to.
	loc, err := time.LoadLocation("Europe/Bucharest")
	if err != nil {
		loc = time.UTC
	}
	today := models.TodayIn(loc)

	// Clear any existing entries for this user on the seed day, then re-create,
	// so the sample dashboard is deterministic across restarts.
	if err := tx.Where("user_id = ? AND consumed_on = ?", user.ID, today).
		Delete(&models.MealEntry{}).Error; err != nil {
		return fmt.Errorf("reset seed meal entries: %w", err)
	}

	entries := []models.MealEntry{
		models.NewMealEntry(user.ID, today, models.MealBreakfast, foods["Rolled Oats"], 80),
		models.NewMealEntry(user.ID, today, models.MealBreakfast, foods["Blueberries"], 100),
		models.NewMealEntry(user.ID, today, models.MealLunch, foods["Chicken Breast, cooked"], 200),
		models.NewMealEntry(user.ID, today, models.MealLunch, foods["Broccoli, steamed"], 150),
		models.NewMealEntry(user.ID, today, models.MealLunch, foods["Extra Virgin Olive Oil"], 10),
	}
	if err := tx.Create(&entries).Error; err != nil {
		return fmt.Errorf("seed meal entries: %w", err)
	}
	return nil
}
