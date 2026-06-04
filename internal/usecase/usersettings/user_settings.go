package usersettings

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/alfariesh/backend-memora/internal/repo"
)

// UseCase -.
type UseCase struct {
	repo repo.UserSettingsRepo
}

// New -.
func New(r repo.UserSettingsRepo) *UseCase {
	return &UseCase{repo: r}
}

// Get -.
func (uc *UseCase) Get(ctx context.Context, userID string) (entity.UserSettings, error) {
	settings, err := uc.repo.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrUserSettingsNotFound) {
			return entity.DefaultUserSettings(userID, time.Now().UTC()), nil
		}

		return entity.UserSettings{}, fmt.Errorf("UserSettingsUseCase - Get - uc.repo.Get: %w", err)
	}

	return settings, nil
}

// Update -.
func (uc *UseCase) Update(ctx context.Context, userID string, params entity.UserSettingsParams) (entity.UserSettings, error) {
	if err := entity.NormalizeUserSettings(&params); err != nil {
		return entity.UserSettings{}, err
	}

	now := time.Now().UTC()
	settings := entity.UserSettings{
		UserID:               userID,
		Timezone:             params.Timezone,
		ReminderTime:         params.ReminderTime,
		NotificationChannels: params.NotificationChannels,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := uc.repo.Upsert(ctx, &settings); err != nil {
		return entity.UserSettings{}, fmt.Errorf("UserSettingsUseCase - Update - uc.repo.Upsert: %w", err)
	}

	settings, err := uc.repo.Get(ctx, userID)
	if err != nil {
		return entity.UserSettings{}, fmt.Errorf("UserSettingsUseCase - Update - uc.repo.Get: %w", err)
	}

	return settings, nil
}
