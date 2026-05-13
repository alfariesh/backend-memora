package entity

import "time"

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
