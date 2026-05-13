package jwt

import (
	"errors"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// ErrUnexpectedSigningMethod is returned when the JWT signing method is not expected.
var ErrUnexpectedSigningMethod = errors.New("unexpected signing method")

// Manager handles JWT token generation and parsing.
type Manager struct {
	secret   string
	duration time.Duration
}

// New -.
func New(secret string, duration time.Duration) *Manager {
	return &Manager{
		secret:   secret,
		duration: duration,
	}
}

// GenerateToken creates a new JWT token for the given user ID.
func (m *Manager) GenerateToken(userID string) (string, error) {
	tokenString, _, err := m.GenerateTokenWithExpiry(userID)

	return tokenString, err
}

// GenerateTokenWithExpiry creates a new JWT token and returns its expiration time.
func (m *Manager) GenerateTokenWithExpiry(userID string) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(m.duration)
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, jwtlib.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwtlib.NewNumericDate(expiresAt),
	})

	tokenString, err := token.SignedString([]byte(m.secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("jwt - GenerateTokenWithExpiry - token.SignedString: %w", err)
	}

	return tokenString, expiresAt, nil
}

// ParseToken validates a JWT token and returns the user ID.
func (m *Manager) ParseToken(tokenString string) (string, error) {
	token, err := jwtlib.Parse(tokenString, func(token *jwtlib.Token) (any, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, token.Header["alg"])
		}

		return []byte(m.secret), nil
	})
	if err != nil {
		return "", fmt.Errorf("jwt - ParseToken - jwtlib.Parse: %w", err)
	}

	sub, err := token.Claims.GetSubject()
	if err != nil {
		return "", fmt.Errorf("jwt - ParseToken - GetSubject: %w", err)
	}

	return sub, nil
}
