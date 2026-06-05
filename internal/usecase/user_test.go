package usecase_test

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/alfariesh/backend-memora/internal/usecase/user"
	"github.com/alfariesh/backend-memora/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func newUserUseCase(t *testing.T) (*user.UseCase, *MockUserRepo) {
	t.Helper()

	ctrl := gomock.NewController(t)

	repo := NewMockUserRepo(ctrl)
	jwtManager := jwt.New("test-secret", time.Hour)
	useCase := user.New(repo, jwtManager)

	return useCase, repo
}

func newUserUseCaseWithSession(t *testing.T) (*user.UseCase, *MockUserRepo, *MockUserSessionRepo) {
	t.Helper()

	ctrl := gomock.NewController(t)

	repo := NewMockUserRepo(ctrl)
	sessionRepo := NewMockUserSessionRepo(ctrl)
	jwtManager := jwt.New("test-secret", time.Hour)
	useCase := user.New(repo, jwtManager, user.SessionRepo(sessionRepo), user.RefreshTokenTTL(time.Hour))

	return useCase, repo, sessionRepo
}

func TestRegister(t *testing.T) {
	t.Parallel()

	t.Run("register success", func(t *testing.T) {
		t.Parallel()

		uc, repo := newUserUseCase(t)
		repo.EXPECT().Store(context.Background(), gomock.Any()).Return(nil)

		u, err := uc.Register(context.Background(), "testuser", "test@example.com", "password123")

		require.NoError(t, err)
		assert.NotEmpty(t, u.ID)
		assert.Equal(t, "testuser", u.Username)
		assert.Equal(t, "test@example.com", u.Email)
	})

	t.Run("register duplicate", func(t *testing.T) {
		t.Parallel()

		uc, repo := newUserUseCase(t)
		repo.EXPECT().Store(context.Background(), gomock.Any()).Return(entity.ErrUserAlreadyExists)

		_, err := uc.Register(context.Background(), "testuser", "test@example.com", "password123")

		require.ErrorIs(t, err, entity.ErrUserAlreadyExists)
	})
}

func TestRegister_NormalizesInput(t *testing.T) {
	t.Parallel()

	uc, repo := newUserUseCase(t)
	repo.EXPECT().Store(context.Background(), gomock.Any()).DoAndReturn(func(_ context.Context, stored *entity.User) error {
		assert.Equal(t, "testuser", stored.Username)
		assert.Equal(t, "test@example.com", stored.Email)

		return nil
	})

	user, err := uc.Register(context.Background(), "  testuser  ", " Test@Example.COM ", "password123")

	require.NoError(t, err)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestRegister_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		username string
		email    string
		password string
	}{
		{name: "short username", username: "ab", email: "test@example.com", password: "password123"},
		{name: "invalid email", username: "testuser", email: "not-email", password: "password123"},
		{name: "short password", username: "testuser", email: "test@example.com", password: "short"},
		{name: "long password", username: "testuser", email: "test@example.com", password: strings.Repeat("a", entity.MaxPasswordLength+1)},
	}

	for _, tc := range tests {
		localTc := tc

		t.Run(localTc.name, func(t *testing.T) {
			t.Parallel()

			uc, _ := newUserUseCase(t)

			_, err := uc.Register(context.Background(), localTc.username, localTc.email, localTc.password)

			require.ErrorIs(t, err, entity.ErrInvalidUserInput)
		})
	}
}

func TestLogin(t *testing.T) {
	t.Parallel()

	t.Run("login success", func(t *testing.T) {
		t.Parallel()

		uc, repo := newUserUseCase(t)
		hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		require.NoError(t, err)

		storedUser := entity.User{
			ID: "user-id-123", Username: "testuser",
			Email: "test@example.com", PasswordHash: string(hash),
		}
		repo.EXPECT().GetByEmail(context.Background(), "test@example.com").Return(storedUser, nil)

		tokens, err := uc.Login(context.Background(), "test@example.com", "password123", entity.SessionMetadata{})

		require.NoError(t, err)
		assert.NotEmpty(t, tokens.Token)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.Equal(t, tokens.Token, tokens.AccessToken)
		assert.Empty(t, tokens.RefreshToken)
		assert.NotZero(t, tokens.ExpiresAt)
	})

	t.Run("login wrong password", func(t *testing.T) {
		t.Parallel()

		uc, repo := newUserUseCase(t)
		hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		require.NoError(t, err)

		storedUser := entity.User{
			ID: "user-id-123", Username: "testuser",
			Email: "test@example.com", PasswordHash: string(hash),
		}
		repo.EXPECT().GetByEmail(context.Background(), "test@example.com").Return(storedUser, nil)

		tokens, err := uc.Login(context.Background(), "test@example.com", "wrongpassword", entity.SessionMetadata{})

		require.ErrorIs(t, err, entity.ErrInvalidCredentials)
		assert.Empty(t, tokens.Token)
		assert.Empty(t, tokens.AccessToken)
		assert.Empty(t, tokens.RefreshToken)
	})

	t.Run("login user not found", func(t *testing.T) {
		t.Parallel()

		uc, repo := newUserUseCase(t)
		repo.EXPECT().GetByEmail(context.Background(), "notfound@example.com").Return(entity.User{}, entity.ErrUserNotFound)

		tokens, err := uc.Login(context.Background(), "notfound@example.com", "password123", entity.SessionMetadata{})

		require.ErrorIs(t, err, entity.ErrInvalidCredentials)
		assert.Empty(t, tokens.Token)
		assert.Empty(t, tokens.AccessToken)
		assert.Empty(t, tokens.RefreshToken)
	})
}

func TestLogin_NormalizesEmail(t *testing.T) {
	t.Parallel()

	uc, repo := newUserUseCase(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	storedUser := entity.User{
		ID: "user-id-123", Username: "testuser",
		Email: "test@example.com", PasswordHash: string(hash),
	}
	repo.EXPECT().GetByEmail(context.Background(), "test@example.com").Return(storedUser, nil)

	tokens, err := uc.Login(context.Background(), " Test@Example.COM ", "password123", entity.SessionMetadata{})

	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
}

func TestLogin_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		email    string
		password string
	}{
		{name: "invalid email", email: "not-email", password: "password123"},
		{name: "empty password", email: "test@example.com", password: ""},
		{name: "long password", email: "test@example.com", password: strings.Repeat("a", entity.MaxPasswordLength+1)},
	}

	for _, tc := range tests {
		localTc := tc

		t.Run(localTc.name, func(t *testing.T) {
			t.Parallel()

			uc, _ := newUserUseCase(t)

			tokens, err := uc.Login(context.Background(), localTc.email, localTc.password, entity.SessionMetadata{})

			require.ErrorIs(t, err, entity.ErrInvalidUserInput)
			assert.Empty(t, tokens.AccessToken)
		})
	}
}

func TestLogin_CreatesRefreshSessionWithMetadata(t *testing.T) {
	t.Parallel()

	uc, repo, sessionRepo := newUserUseCaseWithSession(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	storedUser := entity.User{
		ID: "user-id-123", Username: "testuser",
		Email: "test@example.com", PasswordHash: string(hash),
	}
	repo.EXPECT().GetByEmail(context.Background(), "test@example.com").Return(storedUser, nil)
	sessionRepo.EXPECT().Store(context.Background(), gomock.Any()).DoAndReturn(func(_ context.Context, session *entity.UserSession) error {
		assert.Equal(t, "user-id-123", session.UserID)
		assert.Len(t, session.RefreshTokenHash, 64)
		assert.Equal(t, "127.0.0.1", session.CreatedIP)
		assert.Len(t, session.CreatedUserAgent, entity.MaxSessionUserAgentSize)
		assert.True(t, session.ExpiresAt.After(session.CreatedAt))

		return nil
	})

	tokens, err := uc.Login(context.Background(), "test@example.com", "password123", entity.SessionMetadata{
		IP:        " 127.0.0.1 ",
		UserAgent: strings.Repeat("a", entity.MaxSessionUserAgentSize+10),
	})

	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.True(t, entity.ValidRefreshToken(tokens.RefreshToken))
}

func TestLoginAccessOnly_DoesNotStoreRefreshSession(t *testing.T) {
	t.Parallel()

	uc, repo, _ := newUserUseCaseWithSession(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	storedUser := entity.User{
		ID: "user-id-123", Username: "testuser",
		Email: "test@example.com", PasswordHash: string(hash),
	}
	repo.EXPECT().GetByEmail(context.Background(), "test@example.com").Return(storedUser, nil)

	tokens, err := uc.LoginAccessOnly(context.Background(), "test@example.com", "password123")

	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.Empty(t, tokens.RefreshToken)
}

func TestRefresh_RotatesRefreshToken(t *testing.T) {
	t.Parallel()

	uc, _, sessionRepo := newUserUseCaseWithSession(t)
	refreshToken := testRefreshToken()

	sessionRepo.EXPECT().
		Rotate(context.Background(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, refreshTokenHash string, _ time.Time, nextSession entity.UserSession) (entity.UserSession, error) {
			assert.Len(t, refreshTokenHash, 64)
			assert.Empty(t, nextSession.UserID)
			assert.Len(t, nextSession.RefreshTokenHash, 64)
			assert.Equal(t, "127.0.0.1", nextSession.CreatedIP)
			assert.Equal(t, "Memora/1.0", nextSession.CreatedUserAgent)

			nextSession.UserID = "user-id-123"
			return nextSession, nil
		})

	tokens, err := uc.Refresh(context.Background(), refreshToken, entity.SessionMetadata{
		IP:        "127.0.0.1",
		UserAgent: "Memora/1.0",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.True(t, entity.ValidRefreshToken(tokens.RefreshToken))
	assert.NotEqual(t, refreshToken, tokens.RefreshToken)
}

func TestRefresh_InvalidAndReusedTokens(t *testing.T) {
	t.Parallel()

	t.Run("malformed token", func(t *testing.T) {
		t.Parallel()

		uc, _, _ := newUserUseCaseWithSession(t)

		tokens, err := uc.Refresh(context.Background(), "bad-token", entity.SessionMetadata{})

		require.ErrorIs(t, err, entity.ErrInvalidRefreshToken)
		assert.Empty(t, tokens.AccessToken)
	})

	t.Run("reuse bubbles reuse error", func(t *testing.T) {
		t.Parallel()

		uc, _, sessionRepo := newUserUseCaseWithSession(t)
		sessionRepo.EXPECT().
			Rotate(context.Background(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(entity.UserSession{}, entity.ErrRefreshTokenReuse)

		tokens, err := uc.Refresh(context.Background(), testRefreshToken(), entity.SessionMetadata{})

		require.ErrorIs(t, err, entity.ErrRefreshTokenReuse)
		assert.Empty(t, tokens.AccessToken)
	})
}

func TestSessionManagement(t *testing.T) {
	t.Parallel()

	t.Run("list sessions", func(t *testing.T) {
		t.Parallel()

		uc, _, sessionRepo := newUserUseCaseWithSession(t)
		sessionRepo.EXPECT().
			ListActiveByUserID(context.Background(), "user-id-123", gomock.Any()).
			Return([]entity.UserSession{{ID: "session-id-1"}}, nil)

		sessions, err := uc.ListSessions(context.Background(), "user-id-123")

		require.NoError(t, err)
		require.Len(t, sessions, 1)
		assert.Equal(t, "session-id-1", sessions[0].ID)
	})

	t.Run("revoke session", func(t *testing.T) {
		t.Parallel()

		uc, _, sessionRepo := newUserUseCaseWithSession(t)
		sessionRepo.EXPECT().
			RevokeByID(context.Background(), "user-id-123", "session-id-1", gomock.Any(), "user_revoked").
			Return(nil)

		err := uc.RevokeSession(context.Background(), "user-id-123", "session-id-1")

		require.NoError(t, err)
	})

	t.Run("logout all", func(t *testing.T) {
		t.Parallel()

		uc, _, sessionRepo := newUserUseCaseWithSession(t)
		sessionRepo.EXPECT().
			RevokeAllByUserID(context.Background(), "user-id-123", gomock.Any(), "logout_all").
			Return(nil)

		err := uc.LogoutAll(context.Background(), "user-id-123")

		require.NoError(t, err)
	})
}

func TestChangePassword(t *testing.T) {
	t.Parallel()

	t.Run("success returns new tokens and session", func(t *testing.T) {
		t.Parallel()

		uc, repo, _ := newUserUseCaseWithSession(t)
		hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		require.NoError(t, err)

		repo.EXPECT().GetByID(context.Background(), "user-id-123").Return(entity.User{
			ID:           "user-id-123",
			PasswordHash: string(hash),
		}, nil)
		repo.EXPECT().
			UpdatePasswordAndReplaceSessions(context.Background(), "user-id-123", gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, passwordHash string, _ time.Time, session *entity.UserSession) error {
				require.NoError(t, bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("newpassword123")))
				require.NotNil(t, session)
				assert.Equal(t, "user-id-123", session.UserID)
				assert.Equal(t, "127.0.0.1", session.CreatedIP)

				return nil
			})

		tokens, err := uc.ChangePassword(
			context.Background(),
			"user-id-123",
			"password123",
			"newpassword123",
			entity.SessionMetadata{IP: "127.0.0.1"},
		)

		require.NoError(t, err)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.True(t, entity.ValidRefreshToken(tokens.RefreshToken))
	})

	t.Run("wrong current password is generic", func(t *testing.T) {
		t.Parallel()

		uc, repo, _ := newUserUseCaseWithSession(t)
		hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		require.NoError(t, err)

		repo.EXPECT().GetByID(context.Background(), "user-id-123").Return(entity.User{
			ID:           "user-id-123",
			PasswordHash: string(hash),
		}, nil)

		tokens, err := uc.ChangePassword(
			context.Background(),
			"user-id-123",
			"wrongpassword",
			"newpassword123",
			entity.SessionMetadata{},
		)

		require.ErrorIs(t, err, entity.ErrInvalidCredentials)
		assert.Empty(t, tokens.AccessToken)
	})
}

func TestGetUser(t *testing.T) {
	t.Parallel()

	expectedUser := entity.User{
		ID:       "user-id-123",
		Username: "testuser",
		Email:    "test@example.com",
	}

	t.Run("get user success", func(t *testing.T) {
		t.Parallel()

		uc, repo := newUserUseCase(t)
		repo.EXPECT().GetByID(context.Background(), "user-id-123").Return(expectedUser, nil)

		u, err := uc.GetUser(context.Background(), "user-id-123")

		require.NoError(t, err)
		assert.Equal(t, expectedUser, u)
	})

	t.Run("get user not found", func(t *testing.T) {
		t.Parallel()

		uc, repo := newUserUseCase(t)
		repo.EXPECT().GetByID(context.Background(), "missing-id").Return(entity.User{}, entity.ErrUserNotFound)

		_, err := uc.GetUser(context.Background(), "missing-id")

		require.ErrorIs(t, err, entity.ErrUserNotFound)
	})
}

func testRefreshToken() string {
	b := make([]byte, entity.RefreshTokenBytes)
	for i := range b {
		b[i] = byte(i + 1)
	}

	return base64.RawURLEncoding.EncodeToString(b)
}

func TestGetUser_GenericError(t *testing.T) {
	t.Parallel()

	uc, repo := newUserUseCase(t)

	repo.EXPECT().GetByID(context.Background(), "user-id-123").Return(entity.User{}, errInternalServErr)

	_, err := uc.GetUser(context.Background(), "user-id-123")

	require.Error(t, err)
	require.ErrorIs(t, err, errInternalServErr)
}
