package persistent

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

const userSettingsColumns = "user_id, timezone, reminder_time, notification_channels, created_at, updated_at"

// UserSettingsRepo -.
type UserSettingsRepo struct {
	*postgres.Postgres
}

// NewUserSettingsRepo -.
func NewUserSettingsRepo(pg *postgres.Postgres) *UserSettingsRepo {
	return &UserSettingsRepo{pg}
}

// Get -.
func (r *UserSettingsRepo) Get(ctx context.Context, userID string) (entity.UserSettings, error) {
	sql, args, err := r.Builder.
		Select(userSettingsColumns).
		From("user_settings").
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return entity.UserSettings{}, fmt.Errorf("UserSettingsRepo - Get - r.Builder: %w", err)
	}

	settings, err := scanUserSettings(r.Pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.UserSettings{}, entity.ErrUserSettingsNotFound
		}

		return entity.UserSettings{}, fmt.Errorf("UserSettingsRepo - Get - scan: %w", err)
	}

	return settings, nil
}

// Upsert -.
func (r *UserSettingsRepo) Upsert(ctx context.Context, settings *entity.UserSettings) error {
	channels, err := marshalChannels(settings.NotificationChannels)
	if err != nil {
		return fmt.Errorf("UserSettingsRepo - Upsert - marshal: %w", err)
	}

	sql, args, err := r.Builder.
		Insert("user_settings").
		Columns(userSettingsColumns).
		Values(
			settings.UserID,
			settings.Timezone,
			settings.ReminderTime,
			channels,
			settings.CreatedAt,
			settings.UpdatedAt,
		).
		Suffix("ON CONFLICT (user_id) DO UPDATE SET timezone = EXCLUDED.timezone, reminder_time = EXCLUDED.reminder_time, notification_channels = EXCLUDED.notification_channels, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("UserSettingsRepo - Upsert - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserSettingsRepo - Upsert - r.Pool.Exec: %w", err)
	}

	return nil
}

func scanUserSettings(row scanner) (entity.UserSettings, error) {
	var (
		settings    entity.UserSettings
		channelsRaw []byte
	)

	err := row.Scan(
		&settings.UserID,
		&settings.Timezone,
		&settings.ReminderTime,
		&channelsRaw,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		return entity.UserSettings{}, err
	}

	settings.NotificationChannels, err = unmarshalChannels(channelsRaw)
	if err != nil {
		return entity.UserSettings{}, err
	}

	return settings, nil
}
