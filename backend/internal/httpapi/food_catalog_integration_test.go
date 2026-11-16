package httpapi_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dietnow/backend/internal/database"
	"github.com/dietnow/backend/internal/models"
	"github.com/dietnow/backend/internal/repository"
)

func TestFoundationImportIdempotent(t *testing.T) {
	e := setup(t)
	tx := e.db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	data := `{"FoundationFoods":[{"fdcId":9000000001,"description":"Import fixture","foodNutrients":[{"nutrient":{"id":1003,"unitName":"g"},"amount":12}]}]}`
	for i := 0; i < 2; i++ {
		count, err := database.ImportFoundationFoods(tx, strings.NewReader(data))
		if err != nil || count != 1 {
			t.Fatalf("import: %d, %v", count, err)
		}
	}
	var foods []models.Food
	if err := tx.Where("fdc_id = ?", 9000000001).Find(&foods).Error; err != nil {
		t.Fatal(err)
	}
	if len(foods) != 1 || foods[0].Nutrition.ProteinG != 12 {
		t.Fatalf("foods: %+v", foods)
	}
	id := foods[0].ID
	updated := strings.ReplaceAll(data, `"amount":12`, `"amount":15`)
	if _, err := database.ImportFoundationFoods(tx, strings.NewReader(updated)); err != nil {
		t.Fatal(err)
	}
	if err := tx.First(&foods[0], id).Error; err != nil {
		t.Fatal(err)
	}
	if foods[0].Nutrition.ProteinG != 15 || foods[0].ID != id {
		t.Fatal("import failed to update in place")
	}
	// An error after a flushed batch must roll back the entire import.
	valid := `{"fdcId":9000000002,"description":"Rollback fixture","foodNutrients":[]}`
	items := make([]string, 100)
	for i := range items {
		items[i] = strings.Replace(valid, "9000000002", fmt.Sprint(9000000002+int64(i)), 1)
	}
	broken := `{"FoundationFoods":[` + strings.Join(items, ",") + `,{}]}`
	if _, err := database.ImportFoundationFoods(tx, strings.NewReader(broken)); err == nil {
		t.Fatal("accepted invalid record")
	}
	var count int64
	if err := tx.Model(&models.Food{}).Where("fdc_id >= ? AND fdc_id < ?", 9000000002, 9000000102).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("partial import persisted")
	}
	for _, invalid := range []string{`{}`, `{"FoundationFoods":[]}`, data + `{}`} {
		if _, err := database.ImportFoundationFoods(tx, strings.NewReader(invalid)); err == nil {
			t.Fatalf("accepted %s", invalid)
		}
	}
}

func TestFoodSearchEditDistanceAndVisibility(t *testing.T) {
	e := setup(t)
	tx := e.db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	user := models.User{Email: fmt.Sprintf("fuzzy-%d@example.com", time.Now().UnixNano()), PasswordHash: "unused"}
	if err := tx.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	other := models.User{Email: "other-" + user.Email, PasswordHash: "unused"}
	if err := tx.Create(&other).Error; err != nil {
		t.Fatal(err)
	}
	foods := []models.Food{
		{Name: "Zucchinix, steamed", IsVerified: true},
		{Name: "Zucchinixx, steamed", IsVerified: true},
		{Name: "Zuchinix", OwnerUserID: &other.ID},
		{Name: "Zuchinix deleted", OwnerUserID: &user.ID},
		{Name: "Zuchinix private", OwnerUserID: &user.ID},
	}
	if err := tx.Create(&foods).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Delete(&foods[3]).Error; err != nil {
		t.Fatal(err)
	}
	repo := repository.NewFoodRepository(tx)
	for _, query := range []string{"zuchinix", "ZUCHINIX steamd", "zuchinix, steamd"} {
		found, err := repo.Search(context.Background(), user.ID, query, 100)
		if err != nil {
			t.Fatal(err)
		}
		if len(found) == 0 {
			t.Fatalf("no matches for %q", query)
		}
		for _, f := range found {
			if f.ID == foods[2].ID || f.ID == foods[3].ID {
				t.Fatal("private/deleted food leaked")
			}
		}
		positions := map[uint]int{}
		for i, f := range found {
			positions[f.ID] = i
		}
		a, okA := positions[foods[0].ID]
		b, okB := positions[foods[1].ID]
		if !okA || !okB || a >= b {
			t.Fatalf("incorrect distance ranking for %q: %+v", query, found)
		}
	}
	found, err := repo.Search(context.Background(), user.ID, "zuchinix", 1)
	if err != nil || len(found) != 1 || found[0].ID != foods[4].ID {
		t.Fatalf("own exact match should rank first: %+v, %v", found, err)
	}
	for _, query := range []string{"", "%_", "' OR 1=1 --", strings.Repeat("a", 300), "蒸西葫芦"} {
		if _, err := repo.Search(context.Background(), user.ID, query, 2); err != nil {
			t.Fatalf("query %q: %v", query, err)
		}
	}
	// Client-supplied USDA identities cannot be attached to private foods.
	fdcID := int64(9000000200)
	private := models.Food{Name: "Private fixture", FDCID: &fdcID, IsVerified: true}
	if err := repo.CreatePrivate(context.Background(), user.ID, &private); err != nil {
		t.Fatal(err)
	}
	if private.FDCID != nil || private.IsVerified {
		t.Fatal("private food retained trusted-source identity")
	}
}
