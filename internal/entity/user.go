package entity

import (
	"encoding/base64"
	"net/mail"
	"strings"
	"time"
)

const (
	MinUsernameLength       = 3
	MaxUsernameLength       = 255
	MinPasswordLength       = 8
	MaxPasswordLength       = 72
	MaxEmailLength          = 254
	RefreshTokenBytes       = 32
	RefreshTokenLength      = 43
	MaxSessionIPLength      = 64
	MaxSessionUserAgentSize = 512
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
	ID                string     `json:"id"                   example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID            string     `json:"user_id"              example:"550e8400-e29b-41d4-a716-446655440000"`
	RefreshTokenHash  string     `json:"-"`
	ExpiresAt         time.Time  `json:"expires_at"           example:"2026-01-31T00:00:00Z"`
	RevokedAt         *time.Time `json:"revoked_at"`
	RevokedReason     string     `json:"revoked_reason"`
	CreatedIP         string     `json:"created_ip"`
	CreatedUserAgent  string     `json:"created_user_agent"`
	LastUsedAt        *time.Time `json:"last_used_at"`
	LastUsedIP        string     `json:"last_used_ip"`
	LastUsedUserAgent string     `json:"last_used_user_agent"`
	CreatedAt         time.Time  `json:"created_at"           example:"2026-01-01T00:00:00Z"`
	UpdatedAt         time.Time  `json:"updated_at"           example:"2026-01-01T00:00:00Z"`
}

// SessionMetadata describes the request context that created or used a refresh session.
type SessionMetadata struct {
	IP        string
	UserAgent string
}

// UserSessionView is the safe public representation of an active user session.
type UserSessionView struct {
	ID                string     `json:"id"                   example:"550e8400-e29b-41d4-a716-446655440000"`
	ExpiresAt         time.Time  `json:"expires_at"           example:"2026-01-31T00:00:00Z"`
	CreatedIP         string     `json:"created_ip"           example:"127.0.0.1"`
	CreatedUserAgent  string     `json:"created_user_agent"   example:"Memora/1.0"`
	LastUsedAt        *time.Time `json:"last_used_at"`
	LastUsedIP        string     `json:"last_used_ip"         example:"127.0.0.1"`
	LastUsedUserAgent string     `json:"last_used_user_agent" example:"Memora/1.0"`
	CreatedAt         time.Time  `json:"created_at"           example:"2026-01-01T00:00:00Z"`
	UpdatedAt         time.Time  `json:"updated_at"           example:"2026-01-01T00:00:00Z"`
} // @name entity.UserSessionView

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

// ValidatePasswordChange validates password-change credentials.
func ValidatePasswordChange(currentPassword, newPassword string) error {
	if currentPassword == "" || len(currentPassword) > MaxPasswordLength || !validPassword(newPassword) {
		return ErrInvalidUserInput
	}

	return nil
}

// NormalizeSessionMetadata canonicalizes request metadata for storage.
func NormalizeSessionMetadata(metadata SessionMetadata) SessionMetadata {
	return SessionMetadata{
		IP:        truncate(strings.TrimSpace(metadata.IP), MaxSessionIPLength),
		UserAgent: truncate(strings.TrimSpace(metadata.UserAgent), MaxSessionUserAgentSize),
	}
}

// ValidRefreshToken reports whether token matches the generated opaque refresh token format.
func ValidRefreshToken(token string) bool {
	if len(token) != RefreshTokenLength {
		return false
	}

	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return false
	}

	return len(decoded) == RefreshTokenBytes
}

// ToView returns the safe public representation of a user session.
func (s UserSession) ToView() UserSessionView {
	return UserSessionView{
		ID:                s.ID,
		ExpiresAt:         s.ExpiresAt,
		CreatedIP:         s.CreatedIP,
		CreatedUserAgent:  s.CreatedUserAgent,
		LastUsedAt:        s.LastUsedAt,
		LastUsedIP:        s.LastUsedIP,
		LastUsedUserAgent: s.LastUsedUserAgent,
		CreatedAt:         s.CreatedAt,
		UpdatedAt:         s.UpdatedAt,
	}
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

func truncate(value string, maxLength int) string {
	if len(value) <= maxLength {
		return value
	}

	return value[:maxLength]
}
