package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/pkg/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UseCase -.
type UseCase struct {
	repo            repo.UserRepo
	sessionRepo     repo.UserSessionRepo
	jwt             *jwt.Manager
	refreshTokenTTL time.Duration
}

// Option -.
type Option func(*UseCase)

// SessionRepo -.
func SessionRepo(r repo.UserSessionRepo) Option {
	return func(uc *UseCase) {
		uc.sessionRepo = r
	}
}

// RefreshTokenTTL -.
func RefreshTokenTTL(ttl time.Duration) Option {
	return func(uc *UseCase) {
		uc.refreshTokenTTL = ttl
	}
}

// New -.
func New(r repo.UserRepo, j *jwt.Manager, opts ...Option) *UseCase {
	uc := &UseCase{
		repo:            r,
		jwt:             j,
		refreshTokenTTL: 720 * time.Hour,
	}

	for _, opt := range opts {
		opt(uc)
	}

	return uc
}

// Register -.
func (uc *UseCase) Register(ctx context.Context, username, email, password string) (entity.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return entity.User{}, fmt.Errorf("UserUseCase - Register - bcrypt.GenerateFromPassword: %w", err)
	}

	now := time.Now().UTC()

	user := entity.User{
		ID:           uuid.New().String(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = uc.repo.Store(ctx, &user)
	if err != nil {
		return entity.User{}, fmt.Errorf("UserUseCase - Register - uc.repo.Store: %w", err)
	}

	return user, nil
}

// Login -.
func (uc *UseCase) Login(ctx context.Context, email, password string) (entity.AuthTokens, error) {
	user, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return entity.AuthTokens{}, entity.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return entity.AuthTokens{}, entity.ErrInvalidCredentials
	}

	tokens, err := uc.issueTokens(ctx, user.ID)
	if err != nil {
		return entity.AuthTokens{}, fmt.Errorf("UserUseCase - Login - uc.issueTokens: %w", err)
	}

	return tokens, nil
}

// Refresh -.
func (uc *UseCase) Refresh(ctx context.Context, refreshToken string) (entity.AuthTokens, error) {
	if uc.sessionRepo == nil || refreshToken == "" {
		return entity.AuthTokens{}, entity.ErrInvalidRefreshToken
	}

	now := time.Now().UTC()
	refreshTokenHash := hashRefreshToken(refreshToken)

	session, err := uc.sessionRepo.GetActiveByRefreshTokenHash(ctx, refreshTokenHash, now)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidRefreshToken) {
			return entity.AuthTokens{}, entity.ErrInvalidRefreshToken
		}

		return entity.AuthTokens{}, fmt.Errorf("UserUseCase - Refresh - uc.sessionRepo.GetActiveByRefreshTokenHash: %w", err)
	}

	if err = uc.sessionRepo.RevokeByRefreshTokenHash(ctx, refreshTokenHash, now); err != nil {
		if errors.Is(err, entity.ErrInvalidRefreshToken) {
			return entity.AuthTokens{}, entity.ErrInvalidRefreshToken
		}

		return entity.AuthTokens{}, fmt.Errorf("UserUseCase - Refresh - uc.sessionRepo.RevokeByRefreshTokenHash: %w", err)
	}

	tokens, err := uc.issueTokens(ctx, session.UserID)
	if err != nil {
		return entity.AuthTokens{}, fmt.Errorf("UserUseCase - Refresh - uc.issueTokens: %w", err)
	}

	return tokens, nil
}

// Logout -.
func (uc *UseCase) Logout(ctx context.Context, refreshToken string) error {
	if uc.sessionRepo == nil || refreshToken == "" {
		return nil
	}

	err := uc.sessionRepo.RevokeByRefreshTokenHash(ctx, hashRefreshToken(refreshToken), time.Now().UTC())
	if err != nil {
		if errors.Is(err, entity.ErrInvalidRefreshToken) {
			return nil
		}

		return fmt.Errorf("UserUseCase - Logout - uc.sessionRepo.RevokeByRefreshTokenHash: %w", err)
	}

	return nil
}

// GetUser -.
func (uc *UseCase) GetUser(ctx context.Context, userID string) (entity.User, error) {
	user, err := uc.repo.GetByID(ctx, userID)
	if err != nil {
		return entity.User{}, fmt.Errorf("UserUseCase - GetUser - uc.repo.GetByID: %w", err)
	}

	return user, nil
}

func (uc *UseCase) issueTokens(ctx context.Context, userID string) (entity.AuthTokens, error) {
	accessToken, expiresAt, err := uc.jwt.GenerateTokenWithExpiry(userID)
	if err != nil {
		return entity.AuthTokens{}, fmt.Errorf("uc.jwt.GenerateTokenWithExpiry: %w", err)
	}

	tokens := entity.AuthTokens{
		Token:       accessToken,
		AccessToken: accessToken,
		ExpiresAt:   expiresAt,
	}

	if uc.sessionRepo == nil {
		return tokens, nil
	}

	refreshToken, err := randomRefreshToken()
	if err != nil {
		return entity.AuthTokens{}, fmt.Errorf("randomRefreshToken: %w", err)
	}

	now := time.Now().UTC()
	session := entity.UserSession{
		ID:               uuid.New().String(),
		UserID:           userID,
		RefreshTokenHash: hashRefreshToken(refreshToken),
		ExpiresAt:        now.Add(uc.refreshTokenTTL),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err = uc.sessionRepo.Store(ctx, &session); err != nil {
		return entity.AuthTokens{}, fmt.Errorf("uc.sessionRepo.Store: %w", err)
	}

	tokens.RefreshToken = refreshToken

	return tokens, nil
}

func randomRefreshToken() (string, error) {
	const refreshTokenBytes = 32

	b := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand.Read: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))

	return hex.EncodeToString(sum[:])
}
