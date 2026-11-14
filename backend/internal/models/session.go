package models

import "time"

// Session is a server-side login session. Only the SHA-256 hash of the cookie
// token is stored.
type Session struct {
	ID         uint      `gorm:"primaryKey" json:"-"`
	UserID     uint      `gorm:"not null;index" json:"-"`
	TokenHash  string    `gorm:"uniqueIndex;not null;size:64" json:"-"`
	ExpiresAt  time.Time `gorm:"not null;index" json:"expiresAt"`
	LastSeenAt time.Time `gorm:"not null" json:"lastSeenAt"`
	UserAgent  string    `gorm:"size:255" json:"-"`
	IP         string    `gorm:"size:64" json:"-"`
	CreatedAt  time.Time `json:"createdAt"`
}
