package entity

import (
	"net/mail"
	"strings"
	"time"
)

const (
	MinUsernameLength = 3
	MaxUsernameLength = 255
	MinPasswordLength = 8
	MaxPasswordLength = 72
	MaxEmailLength    = 254
)

// User -.
type User struct {
	ID           string    `json:"id"         example:"550e8400-e29b-41d4-a716-446655440000"`
	Username     string    `json:"username"    example:"johndoe"`
	Email        string    `json:"email"       example:"john@example.com"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"  example:"2026-01-01T00:00:00Z"`
	UpdatedAt    time.Time `json:"updated_at"  example:"2026-01-01T00:00:00Z"`
} // @name entity.User

// AuthTokens -.
type AuthTokens struct {
	Token        string    `json:"token"         example:"eyJhbGciOiJIUzI1NiIs..."`
	AccessToken  string    `json:"access_token"  example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string    `json:"refresh_token" example:"C-Kt0pA3..."`
	ExpiresAt    time.Time `json:"expires_at"    example:"2026-01-01T00:00:00Z"`
} // @name entity.AuthTokens

// UserSession -.
type UserSession struct {
	ID               string     `json:"id"                 example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID           string     `json:"user_id"            example:"550e8400-e29b-41d4-a716-446655440000"`
	RefreshTokenHash string     `json:"-"`
	ExpiresAt        time.Time  `json:"expires_at"         example:"2026-01-31T00:00:00Z"`
	RevokedAt        *time.Time `json:"revoked_at"`
	CreatedAt        time.Time  `json:"created_at"         example:"2026-01-01T00:00:00Z"`
	UpdatedAt        time.Time  `json:"updated_at"         example:"2026-01-01T00:00:00Z"`
}

// NormalizeUserRegistration validates and canonicalizes registration credentials.
func NormalizeUserRegistration(username, email, password string) (string, string, error) {
	username = strings.TrimSpace(username)
	email = normalizeEmail(email)

	if !validUsername(username) || !validEmail(email) || !validPassword(password) {
		return "", "", ErrInvalidUserInput
	}

	return username, email, nil
}

// NormalizeUserLogin validates and canonicalizes login credentials.
func NormalizeUserLogin(email, password string) (string, error) {
	email = normalizeEmail(email)

	if !validEmail(email) || password == "" || len(password) > MaxPasswordLength {
		return "", ErrInvalidUserInput
	}

	return email, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validUsername(username string) bool {
	length := len(username)

	return length >= MinUsernameLength && length <= MaxUsernameLength
}

func validPassword(password string) bool {
	length := len(password)

	return length >= MinPasswordLength && length <= MaxPasswordLength
}

func validEmail(email string) bool {
	if email == "" || len(email) > MaxEmailLength {
		return false
	}

	address, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	return address.Address == email
}
