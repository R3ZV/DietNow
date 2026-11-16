package database

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestDecodeFoundationFood(t *testing.T) {
	food, err := decodeFoundationFood(json.RawMessage(`{"fdcId":321358,"description":"Hummus","foodNutrients":[
 {"nutrient":{"id":1003,"unitName":"g"},"amount":7.35},
 {"nutrient":{"id":1004,"unitName":"g"},"amount":17.1},
 {"nutrient":{"id":1005,"unitName":"g"},"amount":14.9},
 {"nutrient":{"id":1008,"unitName":"kcal"},"amount":229},
 {"nutrient":{"id":2048,"unitName":"kcal"},"amount":230},
 {"nutrient":{"id":1093,"unitName":"mg"},"amount":438},
 {"nutrient":{"id":1106,"unitName":"µg"},"amount":1},
 {"nutrient":{"id":1114,"unitName":"µg"}},
 {"nutrient":{"id":1278,"unitName":"g"},"amount":0.2},
 {"nutrient":{"id":2023,"unitName":"g"},"amount":0.2}
 ]}`))
	if err != nil {
		t.Fatal(err)
	}
	if *food.FDCID != 321358 || food.Name != "Hummus" || !food.IsGlobal() || !food.IsVerified {
		t.Fatalf("identity: %+v", food)
	}
	if food.Nutrition.CaloriesKcal != 230 || food.Nutrition.ProteinG != 7.35 || food.Nutrition.FatG != 17.1 || food.Nutrition.CarbsG != 14.9 {
		t.Fatalf("macros: %+v", food.Nutrition)
	}
	if math.Abs(food.Micros.SodiumMG-438) > 1e-9 || food.Micros.VitaminAUG != 1 || food.Micros.Omega3G != 0.2 {
		t.Fatalf("micros: %+v", food.Micros)
	}
	if !strings.Contains(string(food.USDANutrients), `"id":1114`) {
		t.Fatal("original missing nutrient not retained")
	}
}

func TestDecodeFoundationEnergyFallback(t *testing.T) {
	food, err := decodeFoundationFood(json.RawMessage(`{"fdcId":1,"description":"Test","foodNutrients":[{"nutrient":{"id":1062,"unitName":"kJ"},"amount":418.4}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(food.Nutrition.CaloriesKcal-100) > 1e-9 {
		t.Fatal(food.Nutrition)
	}
}

func TestDecodeFoundationNegativeCarbs(t *testing.T) {
	food, err := decodeFoundationFood(json.RawMessage(`{"fdcId":2727566,"description":"Chicken","foodNutrients":[{"nutrient":{"id":1005,"unitName":"g"},"amount":-0.475}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if food.Nutrition.CarbsG != 0 || !strings.Contains(string(food.USDANutrients), "-0.475") {
		t.Fatal("negative calculated carbs must be clamped without losing the source value")
	}
}

func TestDecodeFoundationInvalid(t *testing.T) {
	for _, raw := range []string{`{}`, `{"fdcId":1}`, `{"fdcId":1,"description":"Test","foodNutrients":[{"nutrient":{"id":1003,"unitName":"g"},"amount":-1}]}`} {
		if _, err := decodeFoundationFood(json.RawMessage(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}
