package device

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/google/uuid"
)

// UseCase -.
type UseCase struct {
	repo repo.DeviceTokenRepo
}

// New -.
func New(r repo.DeviceTokenRepo) *UseCase {
	return &UseCase{repo: r}
}

// Register -.
func (uc *UseCase) Register(ctx context.Context, userID, token, platform, name string) (entity.DeviceToken, error) {
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
