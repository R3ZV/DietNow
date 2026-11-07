// Package repository holds the data-access layer. Every method that touches
// user-owned data takes an explicit ownerID and filters on it, so no query can
// ever cross the tenant boundary — data isolation is guaranteed at this layer,
// not left to the caller to remember.
package repository

import "gorm.io/gorm"

// OwnedBy is a reusable GORM scope that constrains a query to a single user's
// rows. Attach it with .Scopes(OwnedBy(uid)) on any model that has a user_id
// column (WeightLog, MealEntry, and — for private rows — Food).
//
//	db.Scopes(repository.OwnedBy(userID)).Find(&logs)
//
// Centralizing the WHERE clause here means the "user_id = ?" guard is written
// once and reused, instead of being re-typed (and occasionally forgotten) at
// every call site.
func OwnedBy(userID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	}
}

// VisibleFoodsFor scopes a Food query to what a user is allowed to see:
// all global foods (NULL owner) OR their own private foods.
func VisibleFoodsFor(userID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("owner_user_id IS NULL OR owner_user_id = ?", userID)
	}
}
