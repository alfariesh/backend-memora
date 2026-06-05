package usecase_test

import (
	"context"
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

		tokens, err := uc.Login(context.Background(), "test@example.com", "password123")

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

		tokens, err := uc.Login(context.Background(), "test@example.com", "wrongpassword")

		require.ErrorIs(t, err, entity.ErrInvalidCredentials)
		assert.Empty(t, tokens.Token)
		assert.Empty(t, tokens.AccessToken)
		assert.Empty(t, tokens.RefreshToken)
	})

	t.Run("login user not found", func(t *testing.T) {
		t.Parallel()

		uc, repo := newUserUseCase(t)
		repo.EXPECT().GetByEmail(context.Background(), "notfound@example.com").Return(entity.User{}, entity.ErrUserNotFound)

		tokens, err := uc.Login(context.Background(), "notfound@example.com", "password123")

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

	tokens, err := uc.Login(context.Background(), " Test@Example.COM ", "password123")

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

			tokens, err := uc.Login(context.Background(), localTc.email, localTc.password)

			require.ErrorIs(t, err, entity.ErrInvalidUserInput)
			assert.Empty(t, tokens.AccessToken)
		})
	}
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

func TestGetUser_GenericError(t *testing.T) {
	t.Parallel()

	uc, repo := newUserUseCase(t)

	repo.EXPECT().GetByID(context.Background(), "user-id-123").Return(entity.User{}, errInternalServErr)

	_, err := uc.GetUser(context.Background(), "user-id-123")

	require.Error(t, err)
	require.ErrorIs(t, err, errInternalServErr)
}
