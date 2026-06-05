package v1

import (
	"context"
	"testing"
	"time"

	protov1 "github.com/alfariesh/backend-memora/docs/proto/v1"
	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/alfariesh/backend-memora/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authUserStub struct {
	registerCalled bool
	loginCalled    bool
	registerEmail  string
	loginEmail     string
	registerErr    error
	loginErr       error
}

func (s *authUserStub) Register(_ context.Context, username, email, _ string) (entity.User, error) {
	s.registerCalled = true
	s.registerEmail = email

	if s.registerErr != nil {
		return entity.User{}, s.registerErr
	}

	return entity.User{
		ID:        "user-id-123",
		Username:  username,
		Email:     email,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}

func (s *authUserStub) Login(_ context.Context, email, _ string) (entity.AuthTokens, error) {
	s.loginCalled = true
	s.loginEmail = email

	if s.loginErr != nil {
		return entity.AuthTokens{}, s.loginErr
	}

	return entity.AuthTokens{AccessToken: "access-token"}, nil
}

func (s *authUserStub) Refresh(_ context.Context, _ string) (entity.AuthTokens, error) {
	return entity.AuthTokens{}, nil
}

func (s *authUserStub) Logout(_ context.Context, _ string) error {
	return nil
}

func (s *authUserStub) GetUser(_ context.Context, _ string) (entity.User, error) {
	return entity.User{}, nil
}

func TestAuthControllerRegisterValidation(t *testing.T) {
	t.Parallel()

	stub := &authUserStub{}
	controller := &AuthController{u: stub, l: logger.New("error")}

	resp, err := controller.Register(t.Context(), &protov1.RegisterRequest{
		Username: "ab",
		Email:    "not-email",
		Password: "short",
	})

	assert.Nil(t, resp)
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.False(t, stub.registerCalled)
}

func TestAuthControllerRegisterNormalizesInput(t *testing.T) {
	t.Parallel()

	stub := &authUserStub{}
	controller := &AuthController{u: stub, l: logger.New("error")}

	resp, err := controller.Register(t.Context(), &protov1.RegisterRequest{
		Username: " testuser ",
		Email:    " Test@Example.COM ",
		Password: "password123",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, stub.registerCalled)
	assert.Equal(t, "test@example.com", stub.registerEmail)
	assert.Equal(t, "testuser", resp.GetUsername())
	assert.Equal(t, "test@example.com", resp.GetEmail())
}

func TestAuthControllerLoginValidation(t *testing.T) {
	t.Parallel()

	stub := &authUserStub{}
	controller := &AuthController{u: stub, l: logger.New("error")}

	resp, err := controller.Login(t.Context(), &protov1.LoginRequest{
		Email:    "not-email",
		Password: "password123",
	})

	assert.Nil(t, resp)
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.False(t, stub.loginCalled)
}

func TestAuthControllerLoginInvalidCredentials(t *testing.T) {
	t.Parallel()

	stub := &authUserStub{loginErr: entity.ErrInvalidCredentials}
	controller := &AuthController{u: stub, l: logger.New("error")}

	resp, err := controller.Login(t.Context(), &protov1.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	assert.Nil(t, resp)
	require.Error(t, err)
	assert.True(t, stub.loginCalled)
	assert.Equal(t, "test@example.com", stub.loginEmail)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}
