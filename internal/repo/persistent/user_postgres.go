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

// UserRepo -.
type UserRepo struct {
	*postgres.Postgres
}

// NewUserRepo -.
func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

// Store -.
func (r *UserRepo) Store(ctx context.Context, user *entity.User) error {
	sql, args, err := r.Builder.
		Insert("users").
		Columns("id, username, email, password_hash, created_at, updated_at").
		Values(user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserRepo - Store - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entity.ErrUserAlreadyExists
		}

		return fmt.Errorf("UserRepo - Store - r.Pool.Exec: %w", err)
	}

	return nil
}

// GetByID -.
func (r *UserRepo) GetByID(ctx context.Context, id string) (entity.User, error) {
	return r.getUser(ctx, "id", id)
}

// GetByEmail -.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	return r.getUser(ctx, "email", email)
}

// UpdatePasswordAndReplaceSessions updates the user password and replaces active sessions atomically.
func (r *UserRepo) UpdatePasswordAndReplaceSessions(
	ctx context.Context,
	userID,
	passwordHash string,
	at time.Time,
	session *entity.UserSession,
) (err error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("UserRepo - UpdatePasswordAndReplaceSessions - Begin: %w", err)
	}
	defer rollbackTx(ctx, tx, &err, "UserRepo - UpdatePasswordAndReplaceSessions - Rollback")

	sql, args, err := r.Builder.
		Update("users").
		Set("password_hash", passwordHash).
		Set("updated_at", at).
		Where(sq.Eq{"id": userID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserRepo - UpdatePasswordAndReplaceSessions - builder: %w", err)
	}

	result, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserRepo - UpdatePasswordAndReplaceSessions - update: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrUserNotFound
	}

	if err = revokeAllUserSessionsTx(ctx, r.Builder, tx, userID, at, "password_changed"); err != nil {
		return fmt.Errorf("UserRepo - UpdatePasswordAndReplaceSessions - revokeAllUserSessionsTx: %w", err)
	}

	if session != nil {
		session.UserID = userID
		if err = insertUserSession(ctx, r.Builder, tx, session); err != nil {
			return fmt.Errorf("UserRepo - UpdatePasswordAndReplaceSessions - insertUserSession: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("UserRepo - UpdatePasswordAndReplaceSessions - Commit: %w", err)
	}

	return nil
}

func (r *UserRepo) getUser(ctx context.Context, column, value string) (entity.User, error) {
	sql, args, err := r.Builder.
		Select("id, username, email, password_hash, created_at, updated_at").
		From("users").
		Where(sq.Eq{column: value}).
		ToSql()
	if err != nil {
		return entity.User{}, fmt.Errorf("UserRepo - getUser - r.Builder: %w", err)
	}

	var user entity.User

	err = r.Pool.QueryRow(ctx, sql, args...).
		Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, entity.ErrUserNotFound
		}

		return entity.User{}, fmt.Errorf("UserRepo - getUser - r.Pool.QueryRow: %w", err)
	}

	return user, nil
}
