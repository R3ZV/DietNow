package database

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/dietnow/backend/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type foundationFood struct {
	FDCID         int64  `json:"fdcId"`
	Description   string `json:"description"`
	FoodNutrients []struct {
		Nutrient struct {
			ID   int    `json:"id"`
			Unit string `json:"unitName"`
		} `json:"nutrient"`
		Amount *float64 `json:"amount"`
	} `json:"foodNutrients"`
}

// ImportFoundationFoods imports USDA's per-100g values atomically. Only rows
// with the same FDC ID are updated; private foods and curated seeds are untouched.
// The original nutrient entries are retained, including missing measurements.
func ImportFoundationFoods(db *gorm.DB, reader io.Reader) (int, error) {
	var document struct {
		Foods []json.RawMessage `json:"FoundationFoods"`
	}
	dec := json.NewDecoder(reader)
	if err := dec.Decode(&document); err != nil {
		return 0, fmt.Errorf("decode USDA dataset: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return 0, fmt.Errorf("unexpected trailing dataset content")
	}
	if len(document.Foods) == 0 {
		return 0, fmt.Errorf("dataset contains no FoundationFoods")
	}
	count := 0
	err := db.Transaction(func(tx *gorm.DB) error {
		batch := make([]models.Food, 0, 100)
		flush := func() error {
			if len(batch) == 0 {
				return nil
			}
			err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "fdc_id"}},
				// Explicit list protects ownership, IDs and creation timestamps.
				DoUpdates: clause.AssignmentColumns([]string{
					"name", "is_verified", "usda_nutrients", "updated_at",
					"per100g_calories_kcal", "per100g_protein_g", "per100g_carbs_g", "per100g_fat_g",
					"per100g_fiber_g", "per100g_sugar_g", "per100g_sodium_mg", "per100g_potassium_mg",
					"per100g_calcium_mg", "per100g_iron_mg", "per100g_vitamin_aug", "per100g_vitamin_cmg",
					"per100g_vitamin_dug", "per100g_vitamin_emg", "per100g_vitamin_kug", "per100g_omega3_g",
				}),
			}).Create(&batch).Error
			count += len(batch)
			batch = batch[:0]
			return err
		}
		for _, raw := range document.Foods {
			food, err := decodeFoundationFood(raw)
			if err != nil {
				return err
			}
			batch = append(batch, food)
			if len(batch) == cap(batch) {
				if err := flush(); err != nil {
					return err
				}
			}
		}
		return flush()
	})
	if err != nil {
		return 0, fmt.Errorf("import USDA foods: %w", err)
	}
	return count, nil
}

func decodeFoundationFood(raw json.RawMessage) (models.Food, error) {
	var source foundationFood
	if err := json.Unmarshal(raw, &source); err != nil {
		return models.Food{}, err
	}
	if source.FDCID <= 0 || strings.TrimSpace(source.Description) == "" || len([]rune(source.Description)) > 200 {
		return models.Food{}, fmt.Errorf("invalid USDA food identity/name: %d", source.FDCID)
	}
	var original struct {
		Nutrients json.RawMessage `json:"foodNutrients"`
	}
	if err := json.Unmarshal(raw, &original); err != nil {
		return models.Food{}, err
	}
	food := models.Food{Name: source.Description, FDCID: &source.FDCID, IsVerified: true, USDANutrients: datatypes.JSON(original.Nutrients)}
	// Standardize mass units before mapping. IU values are not interchangeable
	// with micrograms and are intentionally not converted here.
	values := map[int]float64{}
	for _, n := range source.FoodNutrients {
		if n.Amount == nil {
			continue
		}
		v := *n.Amount
		// USDA carbohydrate-by-difference can be slightly negative due to
		// analytical variation. Clamp for tracking, retaining the original JSON.
		if n.Nutrient.ID == 1005 && v < 0 {
			v = 0
		}
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
			return models.Food{}, fmt.Errorf("invalid nutrient in food %d", source.FDCID)
		}
		switch strings.ToLower(n.Nutrient.Unit) {
		case "g":
		case "mg":
			v /= 1000
		case "µg", "μg", "ug":
			v /= 1000000
		case "kcal", "kj":
		default:
			continue
		}
		values[n.Nutrient.ID] = v
	}
	first := func(ids ...int) float64 {
		for _, id := range ids {
			if v, ok := values[id]; ok {
				return v
			}
		}
		return 0
	}
	// Prefer specific Atwater energy, then general Atwater, then legacy kcal.
	food.Nutrition = models.Macros{CaloriesKcal: first(2048, 2047, 1008), ProteinG: first(1003), CarbsG: first(1005, 1050), FatG: first(1004, 1085)}
	if _, a := values[2048]; !a {
		if _, b := values[2047]; !b {
			if _, c := values[1008]; !c {
				food.Nutrition.CaloriesKcal = values[1062] / 4.184
			}
		}
	}
	food.Micros = models.Micros{
		FiberG: first(1079), SugarG: first(2000, 1063),
		SodiumMG: first(1093) * 1000, PotassiumMG: first(1092) * 1000,
		CalciumMG: first(1087) * 1000, IronMG: first(1089) * 1000,
		VitaminAUG: first(1106) * 1000000, VitaminCMG: first(1162) * 1000,
		VitaminDUG: first(1114) * 1000000, VitaminEMG: first(1109) * 1000, VitaminKUG: first(1185) * 1000000,
		Omega3G: first(1404) + first(1278, 2023) + first(1272, 2025) + first(1280, 2024) + first(1405),
	}
	return food, nil
}
