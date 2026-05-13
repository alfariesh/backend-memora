package persistent

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

// NotificationRepo -.
type NotificationRepo struct {
	*postgres.Postgres
}

// NewNotificationRepo -.
func NewNotificationRepo(pg *postgres.Postgres) *NotificationRepo {
	return &NotificationRepo{pg}
}

// Store -.
func (r *NotificationRepo) Store(ctx context.Context, notification *entity.Notification) error {
	sql, args, err := r.Builder.
		Insert("notifications").
		Columns("id, user_id, important_day_id, type, title, body, data, read_at, created_at").
		Values(
			notification.ID,
			notification.UserID,
			notification.ImportantDayID,
			notification.Type,
			notification.Title,
			notification.Body,
			notification.Data,
			notification.ReadAt,
			notification.CreatedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("NotificationRepo - Store - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("NotificationRepo - Store - r.Pool.Exec: %w", err)
	}

	return nil
}

// List -.
func (r *NotificationRepo) List(ctx context.Context, userID string, filter repo.NotificationFilter) ([]entity.Notification, int, error) {
	countBuilder := r.Builder.
		Select("COUNT(*)").
		From("notifications").
		Where(sq.Eq{"user_id": userID})

	if filter.UnreadOnly {
		countBuilder = countBuilder.Where("read_at IS NULL")
	}

	countSQL, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("NotificationRepo - List - countBuilder: %w", err)
	}

	var total int

	err = r.Pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("NotificationRepo - List - count query: %w", err)
	}

	dataBuilder := r.Builder.
		Select(notificationColumns()).
		From("notifications").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset)

	if filter.UnreadOnly {
		dataBuilder = dataBuilder.Where("read_at IS NULL")
	}

	dataSQL, dataArgs, err := dataBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("NotificationRepo - List - dataBuilder: %w", err)
	}

	rows, err := r.Pool.Query(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("NotificationRepo - List - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	notifications := make([]entity.Notification, 0, filter.Limit)
	for rows.Next() {
		notification, scanErr := scanNotification(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("NotificationRepo - List - rows.Scan: %w", scanErr)
		}

		notifications = append(notifications, notification)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("NotificationRepo - List - rows.Err: %w", err)
	}

	return notifications, total, nil
}

// CountUnread -.
func (r *NotificationRepo) CountUnread(ctx context.Context, userID string) (int, error) {
	sql, args, err := r.Builder.
		Select("COUNT(*)").
		From("notifications").
		Where(sq.Eq{"user_id": userID}).
		Where("read_at IS NULL").
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("NotificationRepo - CountUnread - r.Builder: %w", err)
	}

	var count int

	err = r.Pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("NotificationRepo - CountUnread - r.Pool.QueryRow: %w", err)
	}

	return count, nil
}

// GetByID -.
func (r *NotificationRepo) GetByID(ctx context.Context, userID, id string) (entity.Notification, error) {
	sql, args, err := r.Builder.
		Select(notificationColumns()).
		From("notifications").
		Where(sq.Eq{"id": id, "user_id": userID}).
		ToSql()
	if err != nil {
		return entity.Notification{}, fmt.Errorf("NotificationRepo - GetByID - r.Builder: %w", err)
	}

	notification, err := scanNotification(r.Pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Notification{}, entity.ErrNotificationNotFound
		}

		return entity.Notification{}, fmt.Errorf("NotificationRepo - GetByID - scan: %w", err)
	}

	return notification, nil
}

// MarkRead -.
func (r *NotificationRepo) MarkRead(ctx context.Context, userID, id string, readAt time.Time) error {
	sql, args, err := r.Builder.
		Update("notifications").
		Set("read_at", readAt).
		Where(sq.Eq{"id": id, "user_id": userID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("NotificationRepo - MarkRead - r.Builder: %w", err)
	}

	result, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("NotificationRepo - MarkRead - r.Pool.Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrNotificationNotFound
	}

	return nil
}

// MarkAllRead -.
func (r *NotificationRepo) MarkAllRead(ctx context.Context, userID string, readAt time.Time) error {
	sql, args, err := r.Builder.
		Update("notifications").
		Set("read_at", readAt).
		Where(sq.Eq{"user_id": userID}).
		Where("read_at IS NULL").
		ToSql()
	if err != nil {
		return fmt.Errorf("NotificationRepo - MarkAllRead - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("NotificationRepo - MarkAllRead - r.Pool.Exec: %w", err)
	}

	return nil
}

func notificationColumns() string {
	return "id, user_id, important_day_id, type, title, body, data, read_at, created_at"
}

func scanNotification(row scanner) (entity.Notification, error) {
	var notification entity.Notification

	err := row.Scan(
		&notification.ID,
		&notification.UserID,
		&notification.ImportantDayID,
		&notification.Type,
		&notification.Title,
		&notification.Body,
		&notification.Data,
		&notification.ReadAt,
		&notification.CreatedAt,
	)

	return notification, err
}
