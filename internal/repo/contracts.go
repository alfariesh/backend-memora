// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=../usecase/mocks_repo_test.go -package=usecase_test

type (
	// TranslationRepo -.
	TranslationRepo interface {
		Store(ctx context.Context, userID string, t entity.Translation) error
		GetHistory(ctx context.Context, userID string) ([]entity.Translation, error)
	}

	// TranslationWebAPI -.
	TranslationWebAPI interface {
		Translate(ctx context.Context, t entity.Translation) (entity.Translation, error)
	}

	// UserRepo -.
	UserRepo interface {
		Store(ctx context.Context, user *entity.User) error
		GetByID(ctx context.Context, id string) (entity.User, error)
		GetByEmail(ctx context.Context, email string) (entity.User, error)
	}

	// TaskRepo -.
	TaskRepo interface {
		Store(ctx context.Context, task *entity.Task) error
		GetByID(ctx context.Context, userID, taskID string) (entity.Task, error)
		List(ctx context.Context, userID string, filter TaskFilter) ([]entity.Task, int, error)
		Update(ctx context.Context, task *entity.Task) error
		Delete(ctx context.Context, userID, taskID string) error
	}

	// TaskFilter -.
	TaskFilter struct {
		Status *entity.TaskStatus
		Limit  uint64
		Offset uint64
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
		MarkSent(ctx context.Context, id string, sentAt time.Time) error
		MarkFailed(ctx context.Context, id, reason string, retry bool) error
	}

	// NotificationRepo -.
	NotificationRepo interface {
		Store(ctx context.Context, notification *entity.Notification) error
		List(ctx context.Context, userID string, filter NotificationFilter) ([]entity.Notification, int, error)
		GetByID(ctx context.Context, userID, id string) (entity.Notification, error)
		MarkRead(ctx context.Context, userID, id string, readAt time.Time) error
		MarkAllRead(ctx context.Context, userID string, readAt time.Time) error
	}

	// DeviceTokenRepo -.
	DeviceTokenRepo interface {
		Store(ctx context.Context, token *entity.DeviceToken) error
		Delete(ctx context.Context, userID, id string) error
		ListActiveByUser(ctx context.Context, userID string) ([]entity.DeviceToken, error)
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
