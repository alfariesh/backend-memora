package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
)

const (
	defaultNotificationListLimit = 20
	maxNotificationListLimit     = 100
)

// UseCase -.
type UseCase struct {
	repo repo.NotificationRepo
}

// New -.
func New(r repo.NotificationRepo) *UseCase {
	return &UseCase{repo: r}
}

// List -.
func (uc *UseCase) List(ctx context.Context, userID string, unreadOnly bool, limit, offset int) ([]entity.Notification, int, error) {
	if limit <= 0 {
		limit = defaultNotificationListLimit
	}

	if limit > maxNotificationListLimit {
		limit = maxNotificationListLimit
	}

	if offset < 0 {
		offset = 0
	}

	notifications, total, err := uc.repo.List(ctx, userID, repo.NotificationFilter{
		UnreadOnly: unreadOnly,
		Limit:      uint64(limit),
		Offset:     uint64(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("NotificationUseCase - List - uc.repo.List: %w", err)
	}

	return notifications, total, nil
}

// CountUnread -.
func (uc *UseCase) CountUnread(ctx context.Context, userID string) (int, error) {
	count, err := uc.repo.CountUnread(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("NotificationUseCase - CountUnread - uc.repo.CountUnread: %w", err)
	}

	return count, nil
}

// MarkRead -.
func (uc *UseCase) MarkRead(ctx context.Context, userID, id string) (entity.Notification, error) {
	readAt := time.Now().UTC()
	if err := uc.repo.MarkRead(ctx, userID, id, readAt); err != nil {
		return entity.Notification{}, fmt.Errorf("NotificationUseCase - MarkRead - uc.repo.MarkRead: %w", err)
	}

	notification, err := uc.repo.GetByID(ctx, userID, id)
	if err != nil {
		return entity.Notification{}, fmt.Errorf("NotificationUseCase - MarkRead - uc.repo.GetByID: %w", err)
	}

	return notification, nil
}

// MarkAllRead -.
func (uc *UseCase) MarkAllRead(ctx context.Context, userID string) error {
	if err := uc.repo.MarkAllRead(ctx, userID, time.Now().UTC()); err != nil {
		return fmt.Errorf("NotificationUseCase - MarkAllRead - uc.repo.MarkAllRead: %w", err)
	}

	return nil
}
