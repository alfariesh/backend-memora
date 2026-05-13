package persistent

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/pkg/postgres"
)

// DeviceTokenRepo -.
type DeviceTokenRepo struct {
	*postgres.Postgres
}

// NewDeviceTokenRepo -.
func NewDeviceTokenRepo(pg *postgres.Postgres) *DeviceTokenRepo {
	return &DeviceTokenRepo{pg}
}

// Store -.
func (r *DeviceTokenRepo) Store(ctx context.Context, token *entity.DeviceToken) error {
	sql, args, err := r.Builder.
		Insert("device_tokens").
		Columns("id, user_id, token, platform, name, active, created_at, updated_at").
		Values(token.ID, token.UserID, token.Token, token.Platform, token.Name, token.Active, token.CreatedAt, token.UpdatedAt).
		Suffix("ON CONFLICT (user_id, token) DO UPDATE SET platform = EXCLUDED.platform, name = EXCLUDED.name, active = true, updated_at = EXCLUDED.updated_at RETURNING id, created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("DeviceTokenRepo - Store - r.Builder: %w", err)
	}

	err = r.Pool.QueryRow(ctx, sql, args...).Scan(&token.ID, &token.CreatedAt)
	if err != nil {
		return fmt.Errorf("DeviceTokenRepo - Store - r.Pool.QueryRow: %w", err)
	}

	return nil
}

// Delete -.
func (r *DeviceTokenRepo) Delete(ctx context.Context, userID, id string) error {
	sql, args, err := r.Builder.
		Update("device_tokens").
		Set("active", false).
		Where(sq.Eq{"id": id, "user_id": userID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("DeviceTokenRepo - Delete - r.Builder: %w", err)
	}

	result, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("DeviceTokenRepo - Delete - r.Pool.Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrDeviceTokenNotFound
	}

	return nil
}

// ListActiveByUser -.
func (r *DeviceTokenRepo) ListActiveByUser(ctx context.Context, userID string) ([]entity.DeviceToken, error) {
	sql, args, err := r.Builder.
		Select("id, user_id, token, platform, name, active, created_at, updated_at").
		From("device_tokens").
		Where(sq.Eq{"user_id": userID, "active": true}).
		OrderBy("updated_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("DeviceTokenRepo - ListActiveByUser - r.Builder: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("DeviceTokenRepo - ListActiveByUser - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	tokens := make([]entity.DeviceToken, 0)
	for rows.Next() {
		var token entity.DeviceToken

		err = rows.Scan(
			&token.ID,
			&token.UserID,
			&token.Token,
			&token.Platform,
			&token.Name,
			&token.Active,
			&token.CreatedAt,
			&token.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("DeviceTokenRepo - ListActiveByUser - rows.Scan: %w", err)
		}

		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("DeviceTokenRepo - ListActiveByUser - rows.Err: %w", err)
	}

	return tokens, nil
}
