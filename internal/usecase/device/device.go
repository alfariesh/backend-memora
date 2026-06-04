package device

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/alfariesh/backend-memora/internal/repo"
	"github.com/google/uuid"
)

// UseCase -.
type UseCase struct {
	repo       repo.DeviceTokenRepo
	pushSender repo.PushSender
}

// Option -.
type Option func(*UseCase)

// PushSender -.
func PushSender(sender repo.PushSender) Option {
	return func(uc *UseCase) {
		uc.pushSender = sender
	}
}

// New -.
func New(r repo.DeviceTokenRepo, opts ...Option) *UseCase {
	uc := &UseCase{repo: r}

	for _, opt := range opts {
		opt(uc)
	}

	return uc
}

// List -.
func (uc *UseCase) List(ctx context.Context, userID string) ([]entity.DeviceToken, error) {
	tokens, err := uc.repo.ListActiveByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("DeviceUseCase - List - uc.repo.ListActiveByUser: %w", err)
	}

	return tokens, nil
}

// Register -.
func (uc *UseCase) Register(ctx context.Context, userID, token, platform, name string) (entity.DeviceToken, error) {
	if !entity.IsExpoPushToken(token) {
		return entity.DeviceToken{}, entity.ErrInvalidDeviceToken
	}

	now := time.Now().UTC()
	deviceToken := entity.DeviceToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		Platform:  platform,
		Name:      name,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.repo.Store(ctx, &deviceToken); err != nil {
		return entity.DeviceToken{}, fmt.Errorf("DeviceUseCase - Register - uc.repo.Store: %w", err)
	}

	return deviceToken, nil
}

// Delete -.
func (uc *UseCase) Delete(ctx context.Context, userID, id string) error {
	if err := uc.repo.Delete(ctx, userID, id); err != nil {
		return fmt.Errorf("DeviceUseCase - Delete - uc.repo.Delete: %w", err)
	}

	return nil
}

// TestPush -.
func (uc *UseCase) TestPush(ctx context.Context, userID, id, title, body string) (entity.PushTestResult, error) {
	if uc.pushSender == nil {
		return entity.PushTestResult{}, entity.ErrPushSenderNotConfigured
	}

	deviceToken, err := uc.getActiveDevice(ctx, userID, id)
	if err != nil {
		return entity.PushTestResult{}, err
	}

	if title == "" {
		title = "Memora test"
	}

	if body == "" {
		body = "Push notifications are working."
	}

	now := time.Now().UTC()
	ticketID, err := uc.pushSender.Send(ctx, deviceToken.Token, title, body, map[string]string{
		"type":      "test_push",
		"device_id": deviceToken.ID,
		"sent_at":   now.Format(time.RFC3339),
	})
	if err != nil {
		if errors.Is(err, entity.ErrPushDeviceNotRegistered) {
			if deactivateErr := uc.repo.Deactivate(ctx, userID, deviceToken.ID, now); deactivateErr != nil {
				return entity.PushTestResult{}, fmt.Errorf("DeviceUseCase - TestPush - uc.repo.Deactivate: %w", deactivateErr)
			}

			return entity.PushTestResult{}, entity.ErrPushDeviceNotRegistered
		}

		return entity.PushTestResult{}, fmt.Errorf("DeviceUseCase - TestPush - uc.pushSender.Send: %w: %v", entity.ErrPushSendFailed, err)
	}

	return entity.PushTestResult{
		DeviceID: deviceToken.ID,
		TicketID: ticketID,
		SentAt:   now,
	}, nil
}

func (uc *UseCase) getActiveDevice(ctx context.Context, userID, id string) (entity.DeviceToken, error) {
	tokens, err := uc.repo.ListActiveByUser(ctx, userID)
	if err != nil {
		return entity.DeviceToken{}, fmt.Errorf("DeviceUseCase - getActiveDevice - uc.repo.ListActiveByUser: %w", err)
	}

	for _, token := range tokens {
		if token.ID == id {
			return token, nil
		}
	}

	return entity.DeviceToken{}, entity.ErrDeviceTokenNotFound
}
