package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/dietnow/backend/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrEmailTaken         = errors.New("an account with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrNoSession          = errors.New("no valid session")
)

// lastSeenGranularity limits how often LastSeenAt is written.
const lastSeenGranularity = time.Minute

// Service owns identity and session state.
type Service struct {
	db  *gorm.DB
	cfg Config
	now func() time.Time

	// dummyHash equalizes login timing for unknown emails.
	dummyHash string
}

func NewService(db *gorm.DB, cfg Config) (*Service, error) {
	dummy, err := bcrypt.GenerateFromPassword([]byte("dietnow-timing-equalizer"), cfg.BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("init auth service: %w", err)
	}
	return &Service{db: db, cfg: cfg, now: time.Now, dummyHash: string(dummy)}, nil
}

func (s *Service) Config() Config { return s.cfg }

// NormalizeEmail trims, lower-cases and validates a bare email address.
func NormalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email || len(email) > 255 {
		return "", ErrInvalidEmail
	}
	return email, nil
}

func (s *Service) Register(ctx context.Context, email, password, displayName string) (models.User, error) {
	email, err := NormalizeEmail(email)
	if err != nil {
		return models.User{}, err
	}
	hash, err := HashPassword(password, s.cfg.BcryptCost)
	if err != nil {
		return models.User{}, err
	}
	displayName = strings.TrimSpace(displayName)
	if len([]rune(displayName)) > 120 {
		displayName = string([]rune(displayName)[:120])
	}
	user := models.User{Email: email, PasswordHash: hash, DisplayName: displayName}
	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return models.User{}, ErrEmailTaken
		}
		return models.User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

// Authenticate returns ErrInvalidCredentials for both unknown email and wrong password.
func (s *Service) Authenticate(ctx context.Context, email, password string) (models.User, error) {
	email, err := NormalizeEmail(email)
	if err != nil || len(password) > MaxPasswordBytes {
		checkPassword(s.dummyHash, password)
		return models.User{}, ErrInvalidCredentials
	}
	// Find+Limit avoids GORM logging the attempted email as "record not found".
	var user models.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).Limit(1).Find(&user).Error; err != nil {
		return models.User{}, fmt.Errorf("load user: %w", err)
	}
	if user.ID == 0 {
		checkPassword(s.dummyHash, password)
		return models.User{}, ErrInvalidCredentials
	}
	if !checkPassword(user.PasswordHash, password) {
		return models.User{}, ErrInvalidCredentials
	}
	return user, nil
}

// CreateSession returns the raw token for the cookie; only its hash is stored.
func (s *Service) CreateSession(ctx context.Context, userID uint, userAgent, ip string) (string, models.Session, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", models.Session{}, fmt.Errorf("generate session token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(buf)
	now := s.now().UTC()
	sess := models.Session{
		UserID:     userID,
		TokenHash:  hashToken(token),
		ExpiresAt:  now.Add(s.cfg.SessionTTL),
		LastSeenAt: now,
		UserAgent:  truncate(userAgent, 255),
		IP:         truncate(ip, 64),
	}
	if err := s.db.WithContext(ctx).Create(&sess).Error; err != nil {
		return "", models.Session{}, fmt.Errorf("create session: %w", err)
	}
	return token, sess, nil
}

// ValidateSession resolves a cookie token to its user, deleting expired or idle sessions.
func (s *Service) ValidateSession(ctx context.Context, token string) (models.User, models.Session, error) {
	if token == "" || len(token) > 128 {
		return models.User{}, models.Session{}, ErrNoSession
	}
	db := s.db.WithContext(ctx)
	var sess models.Session
	if err := db.Where("token_hash = ?", hashToken(token)).Limit(1).Find(&sess).Error; err != nil {
		return models.User{}, models.Session{}, fmt.Errorf("load session: %w", err)
	}
	if sess.ID == 0 {
		return models.User{}, models.Session{}, ErrNoSession
	}

	now := s.now().UTC()
	if !now.Before(sess.ExpiresAt) || now.Sub(sess.LastSeenAt) >= s.cfg.IdleTimeout {
		db.Delete(&models.Session{}, sess.ID)
		return models.User{}, models.Session{}, ErrNoSession
	}

	var user models.User
	err := db.First(&user, sess.UserID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		db.Delete(&models.Session{}, sess.ID)
		return models.User{}, models.Session{}, ErrNoSession
	}
	if err != nil {
		return models.User{}, models.Session{}, fmt.Errorf("load session user: %w", err)
	}

	if now.Sub(sess.LastSeenAt) >= lastSeenGranularity {
		sess.LastSeenAt = now
		db.Model(&models.Session{}).Where("id = ?", sess.ID).Update("last_seen_at", now)
	}
	return user, sess, nil
}

// RevokeSession deletes the session for a token. It is idempotent.
func (s *Service) RevokeSession(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.db.WithContext(ctx).Where("token_hash = ?", hashToken(token)).Delete(&models.Session{}).Error
}

// DeleteExpired removes expired and idle sessions.
func (s *Service) DeleteExpired(ctx context.Context) (int64, error) {
	now := s.now().UTC()
	res := s.db.WithContext(ctx).
		Where("expires_at <= ? OR last_seen_at <= ?", now, now.Add(-s.cfg.IdleTimeout)).
		Delete(&models.Session{})
	return res.RowsAffected, res.Error
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
