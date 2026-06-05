// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"
	"time"

	"github.com/alfariesh/backend-memora/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=../usecase/mocks_repo_test.go -package=usecase_test

type (
	// UserRepo -.
	UserRepo interface {
		Store(ctx context.Context, user *entity.User) error
		GetByID(ctx context.Context, id string) (entity.User, error)
		GetByEmail(ctx context.Context, email string) (entity.User, error)
		UpdatePasswordAndReplaceSessions(ctx context.Context, userID, passwordHash string, at time.Time, session *entity.UserSession) error
	}

	// UserSettingsRepo -.
	UserSettingsRepo interface {
		Get(ctx context.Context, userID string) (entity.UserSettings, error)
		Upsert(ctx context.Context, settings *entity.UserSettings) error
	}

	// UserSessionRepo -.
	UserSessionRepo interface {
		Store(ctx context.Context, session *entity.UserSession) error
		Rotate(ctx context.Context, refreshTokenHash string, now time.Time, nextSession entity.UserSession) (entity.UserSession, error)
		RevokeByRefreshTokenHash(ctx context.Context, refreshTokenHash string, at time.Time, reason string) error
		ListActiveByUserID(ctx context.Context, userID string, now time.Time) ([]entity.UserSession, error)
		RevokeByID(ctx context.Context, userID, id string, at time.Time, reason string) error
		RevokeAllByUserID(ctx context.Context, userID string, at time.Time, reason string) error
	}

	// ImportantDayRepo -.
	ImportantDayRepo interface {
		Store(ctx context.Context, day *entity.ImportantDay) error
		GetByID(ctx context.Context, userID, id string) (entity.ImportantDay, error)
		List(ctx context.Context, userID string, filter ImportantDayFilter) ([]entity.ImportantDay, int, error)
		Update(ctx context.Context, day *entity.ImportantDay) error
		Delete(ctx context.Context, userID, id string) error
	}

	// ReminderRuleRepo -.
	ReminderRuleRepo interface {
		ReplaceForImportantDay(ctx context.Context, userID, importantDayID string, rules []entity.ReminderRule) error
		GetForImportantDay(ctx context.Context, userID, importantDayID string) ([]entity.ReminderRule, error)
	}

	// ReminderJobRepo -.
	ReminderJobRepo interface {
		Store(ctx context.Context, job *entity.ReminderJob) error
		ReplacePendingForImportantDay(ctx context.Context, userID, importantDayID string, jobs []entity.ReminderJob) error
		ClaimDue(ctx context.Context, now time.Time, limit int) ([]entity.ReminderJob, error)
		FinishWithNext(ctx context.Context, id string, status entity.ReminderJobStatus, finishedAt time.Time, lastError string, nextJob entity.ReminderJob) error
		MarkFailed(ctx context.Context, id, reason string, retry bool) error
	}

	// NotificationRepo -.
	NotificationRepo interface {
		Store(ctx context.Context, notification *entity.Notification) error
		List(ctx context.Context, userID string, filter NotificationFilter) ([]entity.Notification, int, error)
		CountUnread(ctx context.Context, userID string) (int, error)
		GetByID(ctx context.Context, userID, id string) (entity.Notification, error)
		MarkRead(ctx context.Context, userID, id string, readAt time.Time) error
		MarkAllRead(ctx context.Context, userID string, readAt time.Time) error
	}

	// DeviceTokenRepo -.
	DeviceTokenRepo interface {
		Store(ctx context.Context, token *entity.DeviceToken) error
		Delete(ctx context.Context, userID, id string) error
		ListActiveByUser(ctx context.Context, userID string) ([]entity.DeviceToken, error)
		Deactivate(ctx context.Context, userID, id string, at time.Time) error
	}

	// EmailSender -.
	EmailSender interface {
		Send(ctx context.Context, to, subject, html string) (string, error)
	}

	// PushSender -.
	PushSender interface {
		Send(ctx context.Context, token, title, body string, data map[string]string) (string, error)
	}

	// ImportantDayFilter -.
	ImportantDayFilter struct {
		Type   *entity.ImportantDayType
		Limit  uint64
		Offset uint64
	}

	// NotificationFilter -.
	NotificationFilter struct {
		UnreadOnly bool
		Limit      uint64
		Offset     uint64
	}
)
