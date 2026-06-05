package v1

import (
	"context"
	"errors"

	v1 "github.com/alfariesh/backend-memora/docs/proto/v1"
	grpcmw "github.com/alfariesh/backend-memora/internal/controller/grpc/middleware"
	"github.com/alfariesh/backend-memora/internal/controller/grpc/v1/response"
	"github.com/alfariesh/backend-memora/internal/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Register -.
func (c *AuthController) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	username, email, err := entity.NormalizeUserRegistration(req.GetUsername(), req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user input")
	}

	user, err := c.u.Register(ctx, username, email, req.GetPassword())
	if err != nil {
		c.l.Error(err, "grpc - v1 - Register")

		if errors.Is(err, entity.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		if errors.Is(err, entity.ErrInvalidUserInput) {
			return nil, status.Error(codes.InvalidArgument, "invalid user input")
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	return response.NewRegisterResponse(&user), nil
}

// Login -.
func (c *AuthController) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	email, err := entity.NormalizeUserLogin(req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user input")
	}

	tokens, err := c.u.Login(ctx, email, req.GetPassword())
	if err != nil {
		c.l.Error(err, "grpc - v1 - Login")

		if errors.Is(err, entity.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}

		if errors.Is(err, entity.ErrInvalidUserInput) {
			return nil, status.Error(codes.InvalidArgument, "invalid user input")
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &v1.LoginResponse{Token: tokens.AccessToken}, nil
}

// GetProfile -.
func (c *AuthController) GetProfile(ctx context.Context, _ *v1.GetProfileRequest) (*v1.GetProfileResponse, error) {
	userID, ok := grpcmw.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	user, err := c.u.GetUser(ctx, userID)
	if err != nil {
		c.l.Error(err, "grpc - v1 - GetProfile")

		if errors.Is(err, entity.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	return response.NewGetProfileResponse(&user), nil
}
