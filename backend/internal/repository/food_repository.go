package repository

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"github.com/dietnow/backend/internal/models"
	"gorm.io/gorm"
)

type FoodRepository struct {
	db *gorm.DB
}

func NewFoodRepository(db *gorm.DB) *FoodRepository {
	return &FoodRepository{db: db}
}

// Search returns foods the user may see — global foods plus their own private
// ones — ranked by minimum word edit distance, so typos still return the
// closest matches. Each query word is compared with every name word and the
// minimum distances are summed. Literal fragments score zero. Visibility and
// soft deletion are applied before returning any results.
func (r *FoodRepository) Search(ctx context.Context, userID uint, nameLike string, limit int) ([]models.Food, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	var foods []models.Food
	query := strings.TrimSpace(nameLike)
	db := r.db.WithContext(ctx).Model(&models.Food{}).Scopes(VisibleFoodsFor(userID))
	if query != "" && strings.IndexFunc(query, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsNumber(r) }) < 0 {
		// Punctuation-only input has no words to compare; match it literally.
		db = db.Where("strpos(lower(name), lower(?)) > 0", query)
	} else if query != "" {
		// fuzzystrmatch accepts at most 255 characters per operand. Parameterized
		// input and literal strpos avoid interpreting SQL or LIKE wildcards.
		db = db.Select("foods.*").Joins(`CROSS JOIN LATERAL (
			SELECT COALESCE(SUM(best.distance), 0) AS distance
			FROM regexp_split_to_table(lower(?), '[^[:alnum:]]+') AS q(word)
			CROSS JOIN LATERAL (
				SELECT MIN(CASE WHEN strpos(w.word, q.word) > 0 THEN 0
					ELSE levenshtein(left(q.word, 255), left(w.word, 255)) END) AS distance
				FROM regexp_split_to_table(lower(foods.name), '[^[:alnum:]]+') AS w(word)
				WHERE w.word <> ''
			) AS best
			WHERE q.word <> ''
		) AS match_score`, query).Order("match_score.distance ASC")
	}
	err := db.Order("is_verified DESC, name ASC, foods.id ASC").Limit(limit).Find(&foods).Error
	return foods, err
}

// GetVisible fetches one food by id, but only if it is visible to the user.
// Used before logging a meal so a user cannot reference someone else's private
// food by guessing its id.
func (r *FoodRepository) GetVisible(ctx context.Context, userID, foodID uint) (models.Food, error) {
	var food models.Food
	err := r.db.WithContext(ctx).
		Scopes(VisibleFoodsFor(userID)).
		Where("id = ?", foodID).
		First(&food).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Food{}, ErrNotFound
	}
	return food, err
}

// CreatePrivate creates a food owned by the user. We force OwnerUserID on the
// server so a client can never smuggle in a NULL owner and forge a global food.
func (r *FoodRepository) CreatePrivate(ctx context.Context, userID uint, food *models.Food) error {
	food.ID = 0
	food.OwnerUserID = &userID
	food.IsVerified = false // only the seed/admin path creates verified globals
	food.FDCID = nil
	food.USDANutrients = nil
	return r.db.WithContext(ctx).Create(food).Error
}

// UpdatePrivate edits a food only if the user OWNS it (owner_user_id = userID).
// The guard explicitly excludes global foods (NULL owner), so no user can edit
// the shared catalog through this path.
func (r *FoodRepository) UpdatePrivate(ctx context.Context, userID, foodID uint, fields map[string]any) error {
	// Never let a caller flip ownership or verification via the update map.
	delete(fields, "owner_user_id")
	delete(fields, "is_verified")
	delete(fields, "fdc_id")
	delete(fields, "usda_nutrients")
	delete(fields, "id")

	res := r.db.WithContext(ctx).
		Model(&models.Food{}).
		Where("id = ? AND owner_user_id = ?", foodID, userID).
		Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeletePrivate soft-deletes a food the user owns. Global foods are untouchable.
func (r *FoodRepository) DeletePrivate(ctx context.Context, userID, foodID uint) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND owner_user_id = ?", foodID, userID).
		Delete(&models.Food{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
