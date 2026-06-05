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

	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/alfariesh/backend-memora/internal/repo"
	"github.com/alfariesh/backend-memora/pkg/jwt"
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
	username, email, err := entity.NormalizeUserRegistration(username, email, password)
	if err != nil {
		return entity.User{}, err
	}

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
func (uc *UseCase) Login(ctx context.Context, email, password string, metadata entity.SessionMetadata) (entity.AuthTokens, error) {
	user, err := uc.authenticate(ctx, email, password)
	if err != nil {
		return entity.AuthTokens{}, err
	}

	tokens, err := uc.issueTokens(ctx, user.ID, metadata)
	if err != nil {
		return entity.AuthTokens{}, fmt.Errorf("UserUseCase - Login - uc.issueTokens: %w", err)
	}

	return tokens, nil
}

// LoginAccessOnly authenticates the user and returns only an access token.
func (uc *UseCase) LoginAccessOnly(ctx context.Context, email, password string) (entity.AuthTokens, error) {
	user, err := uc.authenticate(ctx, email, password)
	if err != nil {
		return entity.AuthTokens{}, err
	}

	tokens, err := uc.issueAccessToken(user.ID)
	if err != nil {
		return entity.AuthTokens{}, fmt.Errorf("UserUseCase - LoginAccessOnly - uc.issueAccessToken: %w", err)
	}

	return tokens, nil
}

// Refresh -.
func (uc *UseCase) Refresh(ctx context.Context, refreshToken string, metadata entity.SessionMetadata) (entity.AuthTokens, error) {
	if uc.sessionRepo == nil || !entity.ValidRefreshToken(refreshToken) {
		return entity.AuthTokens{}, entity.ErrInvalidRefreshToken
	}

	now := time.Now().UTC()
	newRefreshToken, nextSession, err := uc.newRefreshSession("", now, metadata)
	if err != nil {
		return entity.AuthTokens{}, fmt.Errorf("UserUseCase - Refresh - uc.newRefreshSession: %w", err)
	}

	session, err := uc.sessionRepo.Rotate(ctx, hashRefreshToken(refreshToken), now, nextSession)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidRefreshToken) || errors.Is(err, entity.ErrRefreshTokenReuse) {
			return entity.AuthTokens{}, err
		}

		return entity.AuthTokens{}, fmt.Errorf("UserUseCase - Refresh - uc.sessionRepo.Rotate: %w", err)
	}

	tokens, err := uc.issueAccessToken(session.UserID)
	if err != nil {
		return entity.AuthTokens{}, fmt.Errorf("UserUseCase - Refresh - uc.issueAccessToken: %w", err)
	}

	tokens.RefreshToken = newRefreshToken

	return tokens, nil
}

// Logout -.
func (uc *UseCase) Logout(ctx context.Context, refreshToken string) error {
	if uc.sessionRepo == nil || !entity.ValidRefreshToken(refreshToken) {
		return nil
	}

	err := uc.sessionRepo.RevokeByRefreshTokenHash(ctx, hashRefreshToken(refreshToken), time.Now().UTC(), "logout")
	if err != nil {
		if errors.Is(err, entity.ErrInvalidRefreshToken) {
			return nil
		}

		return fmt.Errorf("UserUseCase - Logout - uc.sessionRepo.RevokeByRefreshTokenHash: %w", err)
	}

	return nil
}

// ListSessions returns active sessions for the user.
func (uc *UseCase) ListSessions(ctx context.Context, userID string) ([]entity.UserSessionView, error) {
	if uc.sessionRepo == nil {
		return nil, nil
	}

	sessions, err := uc.sessionRepo.ListActiveByUserID(ctx, userID, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - ListSessions - uc.sessionRepo.ListActiveByUserID: %w", err)
	}

	views := make([]entity.UserSessionView, len(sessions))
	for i := range sessions {
		views[i] = sessions[i].ToView()
	}

	return views, nil
}

// RevokeSession revokes one active session owned by the user.
func (uc *UseCase) RevokeSession(ctx context.Context, userID, sessionID string) error {
	if uc.sessionRepo == nil || userID == "" || sessionID == "" {
		return entity.ErrSessionNotFound
	}

	err := uc.sessionRepo.RevokeByID(ctx, userID, sessionID, time.Now().UTC(), "user_revoked")
	if err != nil {
		if errors.Is(err, entity.ErrSessionNotFound) {
			return entity.ErrSessionNotFound
		}

		return fmt.Errorf("UserUseCase - RevokeSession - uc.sessionRepo.RevokeByID: %w", err)
	}

	return nil
}

// LogoutAll revokes all active sessions owned by the user.
func (uc *UseCase) LogoutAll(ctx context.Context, userID string) error {
	if uc.sessionRepo == nil {
		return nil
	}

	if userID == "" {
		return entity.ErrUserNotFound
	}

	if err := uc.sessionRepo.RevokeAllByUserID(ctx, userID, time.Now().UTC(), "logout_all"); err != nil {
		return fmt.Errorf("UserUseCase - LogoutAll - uc.sessionRepo.RevokeAllByUserID: %w", err)
	}

	return nil
}

// ChangePassword changes the user password, revokes old sessions, and issues a new session.
func (uc *UseCase) ChangePassword(
	ctx context.Context,
	userID,
	currentPassword,
	newPassword string,
	metadata entity.SessionMetadata,
) (entity.AuthTokens, error) {
	if err := entity.ValidatePasswordChange(currentPassword, newPassword); err != nil {
		return entity.AuthTokens{}, err
	}

	storedUser, err := uc.repo.GetByID(ctx, userID)
	if err != nil {
		return entity.AuthTokens{}, fmt.Errorf("UserUseCase - ChangePassword - uc.repo.GetByID: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(storedUser.PasswordHash), []byte(currentPassword)); err != nil {
		return entity.AuthTokens{}, entity.ErrInvalidCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return entity.AuthTokens{}, fmt.Errorf("UserUseCase - ChangePassword - bcrypt.GenerateFromPassword: %w", err)
	}

	tokens, err := uc.issueAccessToken(userID)
	if err != nil {
		return entity.AuthTokens{}, fmt.Errorf("UserUseCase - ChangePassword - uc.issueAccessToken: %w", err)
	}

	var session *entity.UserSession
	if uc.sessionRepo != nil {
		var refreshToken string
		var newSession entity.UserSession

		refreshToken, newSession, err = uc.newRefreshSession(userID, time.Now().UTC(), metadata)
		if err != nil {
			return entity.AuthTokens{}, fmt.Errorf("UserUseCase - ChangePassword - uc.newRefreshSession: %w", err)
		}

		tokens.RefreshToken = refreshToken
		session = &newSession
	}

	if err = uc.repo.UpdatePasswordAndReplaceSessions(ctx, userID, string(hash), time.Now().UTC(), session); err != nil {
		return entity.AuthTokens{}, fmt.Errorf("UserUseCase - ChangePassword - uc.repo.UpdatePasswordAndReplaceSessions: %w", err)
	}

	return tokens, nil
}

// GetUser -.
func (uc *UseCase) GetUser(ctx context.Context, userID string) (entity.User, error) {
	user, err := uc.repo.GetByID(ctx, userID)
	if err != nil {
		return entity.User{}, fmt.Errorf("UserUseCase - GetUser - uc.repo.GetByID: %w", err)
	}

	return user, nil
}

func (uc *UseCase) authenticate(ctx context.Context, email, password string) (entity.User, error) {
	email, err := entity.NormalizeUserLogin(email, password)
	if err != nil {
		return entity.User{}, err
	}

	user, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return entity.User{}, entity.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return entity.User{}, entity.ErrInvalidCredentials
	}

	return user, nil
}

func (uc *UseCase) issueAccessToken(userID string) (entity.AuthTokens, error) {
	accessToken, expiresAt, err := uc.jwt.GenerateTokenWithExpiry(userID)
	if err != nil {
		return entity.AuthTokens{}, fmt.Errorf("uc.jwt.GenerateTokenWithExpiry: %w", err)
	}

	tokens := entity.AuthTokens{
		Token:       accessToken,
		AccessToken: accessToken,
		ExpiresAt:   expiresAt,
	}

	return tokens, nil
}

func (uc *UseCase) issueTokens(ctx context.Context, userID string, metadata entity.SessionMetadata) (entity.AuthTokens, error) {
	tokens, err := uc.issueAccessToken(userID)
	if err != nil {
		return entity.AuthTokens{}, err
	}

	if uc.sessionRepo == nil {
		return tokens, nil
	}

	now := time.Now().UTC()
	refreshToken, session, err := uc.newRefreshSession(userID, now, metadata)
	if err != nil {
		return entity.AuthTokens{}, err
	}

	if err = uc.sessionRepo.Store(ctx, &session); err != nil {
		return entity.AuthTokens{}, fmt.Errorf("uc.sessionRepo.Store: %w", err)
	}

	tokens.RefreshToken = refreshToken

	return tokens, nil
}

func (uc *UseCase) newRefreshSession(
	userID string,
	now time.Time,
	metadata entity.SessionMetadata,
) (string, entity.UserSession, error) {
	refreshToken, err := randomRefreshToken()
	if err != nil {
		return "", entity.UserSession{}, fmt.Errorf("randomRefreshToken: %w", err)
	}

	metadata = entity.NormalizeSessionMetadata(metadata)

	return refreshToken, entity.UserSession{
		ID:               uuid.New().String(),
		UserID:           userID,
		RefreshTokenHash: hashRefreshToken(refreshToken),
		ExpiresAt:        now.Add(uc.refreshTokenTTL),
		CreatedIP:        metadata.IP,
		CreatedUserAgent: metadata.UserAgent,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func randomRefreshToken() (string, error) {
	b := make([]byte, entity.RefreshTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand.Read: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))

	return hex.EncodeToString(sum[:])
}
