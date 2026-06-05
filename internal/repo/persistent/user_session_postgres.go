package persistent

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/alfariesh/backend-memora/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const userSessionColumns = "id, user_id, refresh_token_hash, expires_at, revoked_at, revoked_reason, created_ip, created_user_agent, last_used_at, last_used_ip, last_used_user_agent, created_at, updated_at"

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
	if err := insertUserSession(ctx, r.Builder, r.Pool, session); err != nil {
		return fmt.Errorf("UserSessionRepo - Store - r.Pool.Exec: %w", err)
	}

	return nil
}

// Rotate revokes the old refresh token and creates the next session atomically.
func (r *UserSessionRepo) Rotate(
	ctx context.Context,
	refreshTokenHash string,
	now time.Time,
	nextSession entity.UserSession,
) (session entity.UserSession, err error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return entity.UserSession{}, fmt.Errorf("UserSessionRepo - Rotate - Begin: %w", err)
	}
	defer rollbackTx(ctx, tx, &err, "UserSessionRepo - Rotate - Rollback")

	current, err := getUserSessionByRefreshHashTx(ctx, r.Builder, tx, refreshTokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.UserSession{}, entity.ErrInvalidRefreshToken
		}

		return entity.UserSession{}, fmt.Errorf("UserSessionRepo - Rotate - getUserSessionByRefreshHashTx: %w", err)
	}

	if current.RevokedAt != nil {
		if err = revokeAllUserSessionsTx(ctx, r.Builder, tx, current.UserID, now, "reuse_detected"); err != nil {
			return entity.UserSession{}, fmt.Errorf("UserSessionRepo - Rotate - revokeAllUserSessionsTx: %w", err)
		}

		if err = tx.Commit(ctx); err != nil {
			return entity.UserSession{}, fmt.Errorf("UserSessionRepo - Rotate - Commit reuse: %w", err)
		}

		return entity.UserSession{}, entity.ErrRefreshTokenReuse
	}

	if !current.ExpiresAt.After(now) {
		return entity.UserSession{}, entity.ErrInvalidRefreshToken
	}

	sql, args, err := r.Builder.
		Update("user_sessions").
		Set("revoked_at", now).
		Set("revoked_reason", "rotated").
		Set("last_used_at", now).
		Set("last_used_ip", nextSession.CreatedIP).
		Set("last_used_user_agent", nextSession.CreatedUserAgent).
		Set("updated_at", now).
		Where(sq.Eq{"id": current.ID, "revoked_at": nil}).
		ToSql()
	if err != nil {
		return entity.UserSession{}, fmt.Errorf("UserSessionRepo - Rotate - update builder: %w", err)
	}

	result, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return entity.UserSession{}, fmt.Errorf("UserSessionRepo - Rotate - update: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.UserSession{}, entity.ErrInvalidRefreshToken
	}

	nextSession.UserID = current.UserID
	if err = insertUserSession(ctx, r.Builder, tx, &nextSession); err != nil {
		return entity.UserSession{}, fmt.Errorf("UserSessionRepo - Rotate - insertUserSession: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return entity.UserSession{}, fmt.Errorf("UserSessionRepo - Rotate - Commit: %w", err)
	}

	return nextSession, nil
}

// ListActiveByUserID -.
func (r *UserSessionRepo) ListActiveByUserID(ctx context.Context, userID string, now time.Time) ([]entity.UserSession, error) {
	sql, args, err := r.Builder.
		Select(userSessionColumns).
		From("user_sessions").
		Where(sq.Eq{"user_id": userID, "revoked_at": nil}).
		Where(sq.Gt{"expires_at": now}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("UserSessionRepo - ListActiveByUserID - r.Builder: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("UserSessionRepo - ListActiveByUserID - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	sessions := make([]entity.UserSession, 0)
	for rows.Next() {
		session, scanErr := scanUserSession(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("UserSessionRepo - ListActiveByUserID - scan: %w", scanErr)
		}

		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("UserSessionRepo - ListActiveByUserID - rows.Err: %w", err)
	}

	return sessions, nil
}

// RevokeByRefreshTokenHash -.
func (r *UserSessionRepo) RevokeByRefreshTokenHash(ctx context.Context, refreshTokenHash string, at time.Time, reason string) error {
	sql, args, err := r.Builder.
		Update("user_sessions").
		Set("revoked_at", at).
		Set("revoked_reason", reason).
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

// RevokeByID -.
func (r *UserSessionRepo) RevokeByID(ctx context.Context, userID, id string, at time.Time, reason string) error {
	sql, args, err := r.Builder.
		Update("user_sessions").
		Set("revoked_at", at).
		Set("revoked_reason", reason).
		Set("updated_at", at).
		Where(sq.Eq{"id": id, "user_id": userID, "revoked_at": nil}).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserSessionRepo - RevokeByID - r.Builder: %w", err)
	}

	result, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserSessionRepo - RevokeByID - r.Pool.Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrSessionNotFound
	}

	return nil
}

// RevokeAllByUserID -.
func (r *UserSessionRepo) RevokeAllByUserID(ctx context.Context, userID string, at time.Time, reason string) error {
	sql, args, err := r.Builder.
		Update("user_sessions").
		Set("revoked_at", at).
		Set("revoked_reason", reason).
		Set("updated_at", at).
		Where(sq.Eq{"user_id": userID, "revoked_at": nil}).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserSessionRepo - RevokeAllByUserID - r.Builder: %w", err)
	}

	if _, err = r.Pool.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("UserSessionRepo - RevokeAllByUserID - r.Pool.Exec: %w", err)
	}

	return nil
}

func getUserSessionByRefreshHashTx(
	ctx context.Context,
	builder sq.StatementBuilderType,
	tx pgx.Tx,
	refreshTokenHash string,
) (entity.UserSession, error) {
	sql, args, err := builder.
		Select(userSessionColumns).
		From("user_sessions").
		Where(sq.Eq{"refresh_token_hash": refreshTokenHash}).
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return entity.UserSession{}, fmt.Errorf("getUserSessionByRefreshHashTx - builder: %w", err)
	}

	return scanUserSession(tx.QueryRow(ctx, sql, args...))
}

func insertUserSession(
	ctx context.Context,
	builder sq.StatementBuilderType,
	execer sessionExecer,
	session *entity.UserSession,
) error {
	sql, args, err := builder.
		Insert("user_sessions").
		Columns(userSessionColumns).
		Values(
			session.ID,
			session.UserID,
			session.RefreshTokenHash,
			session.ExpiresAt,
			session.RevokedAt,
			session.RevokedReason,
			session.CreatedIP,
			session.CreatedUserAgent,
			session.LastUsedAt,
			session.LastUsedIP,
			session.LastUsedUserAgent,
			session.CreatedAt,
			session.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("insertUserSession - builder: %w", err)
	}

	if _, err = execer.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("insertUserSession - exec: %w", err)
	}

	return nil
}

type sessionExecer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func revokeAllUserSessionsTx(
	ctx context.Context,
	builder sq.StatementBuilderType,
	tx pgx.Tx,
	userID string,
	at time.Time,
	reason string,
) error {
	sql, args, err := builder.
		Update("user_sessions").
		Set("revoked_at", at).
		Set("revoked_reason", reason).
		Set("updated_at", at).
		Where(sq.Eq{"user_id": userID, "revoked_at": nil}).
		ToSql()
	if err != nil {
		return fmt.Errorf("revokeAllUserSessionsTx - builder: %w", err)
	}

	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("revokeAllUserSessionsTx - exec: %w", err)
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
		&session.RevokedReason,
		&session.CreatedIP,
		&session.CreatedUserAgent,
		&session.LastUsedAt,
		&session.LastUsedIP,
		&session.LastUsedUserAgent,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	return session, err
}
