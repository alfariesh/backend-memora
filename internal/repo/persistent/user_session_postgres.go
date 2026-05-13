package persistent

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

const userSessionColumns = "id, user_id, refresh_token_hash, expires_at, revoked_at, created_at, updated_at"

// UserSessionRepo -.
type UserSessionRepo struct {
	*postgres.Postgres
}

// NewUserSessionRepo -.
func NewUserSessionRepo(pg *postgres.Postgres) *UserSessionRepo {
	return &UserSessionRepo{pg}
}

// Store -.
func (r *UserSessionRepo) Store(ctx context.Context, session *entity.UserSession) error {
	sql, args, err := r.Builder.
		Insert("user_sessions").
		Columns(userSessionColumns).
		Values(
			session.ID,
			session.UserID,
			session.RefreshTokenHash,
			session.ExpiresAt,
			session.RevokedAt,
			session.CreatedAt,
			session.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserSessionRepo - Store - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserSessionRepo - Store - r.Pool.Exec: %w", err)
	}

	return nil
}

// GetActiveByRefreshTokenHash -.
func (r *UserSessionRepo) GetActiveByRefreshTokenHash(ctx context.Context, refreshTokenHash string, now time.Time) (entity.UserSession, error) {
	sql, args, err := r.Builder.
		Select(userSessionColumns).
		From("user_sessions").
		Where(sq.Eq{"refresh_token_hash": refreshTokenHash, "revoked_at": nil}).
		Where(sq.Gt{"expires_at": now}).
		ToSql()
	if err != nil {
		return entity.UserSession{}, fmt.Errorf("UserSessionRepo - GetActiveByRefreshTokenHash - r.Builder: %w", err)
	}

	session, err := scanUserSession(r.Pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.UserSession{}, entity.ErrInvalidRefreshToken
		}

		return entity.UserSession{}, fmt.Errorf("UserSessionRepo - GetActiveByRefreshTokenHash - scan: %w", err)
	}

	return session, nil
}

// RevokeByRefreshTokenHash -.
func (r *UserSessionRepo) RevokeByRefreshTokenHash(ctx context.Context, refreshTokenHash string, at time.Time) error {
	sql, args, err := r.Builder.
		Update("user_sessions").
		Set("revoked_at", at).
		Set("updated_at", at).
		Where(sq.Eq{"refresh_token_hash": refreshTokenHash, "revoked_at": nil}).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserSessionRepo - RevokeByRefreshTokenHash - r.Builder: %w", err)
	}

	result, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserSessionRepo - RevokeByRefreshTokenHash - r.Pool.Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrInvalidRefreshToken
	}

	return nil
}

func scanUserSession(row scanner) (entity.UserSession, error) {
	var session entity.UserSession

	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	return session, err
}
