// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=./mocks_usecase_test.go -package=usecase_test

type (
	// Translation -.
	Translation interface {
		Translate(ctx context.Context, userID string, t entity.Translation) (entity.Translation, error)
		History(ctx context.Context, userID string) (entity.TranslationHistory, error)
	}

	// User -.
	User interface {
		Register(ctx context.Context, username, email, password string) (entity.User, error)
		Login(ctx context.Context, email, password string) (entity.AuthTokens, error)
		Refresh(ctx context.Context, refreshToken string) (entity.AuthTokens, error)
		Logout(ctx context.Context, refreshToken string) error
		GetUser(ctx context.Context, userID string) (entity.User, error)
	}

	// UserSettings -.
	UserSettings interface {
		Get(ctx context.Context, userID string) (entity.UserSettings, error)
		Update(ctx context.Context, userID string, params entity.UserSettingsParams) (entity.UserSettings, error)
	}

	// Task -.
	Task interface {
		Create(ctx context.Context, userID, title, description string) (entity.Task, error)
		Get(ctx context.Context, userID, taskID string) (entity.Task, error)
		List(ctx context.Context, userID string, status *entity.TaskStatus, limit, offset int) ([]entity.Task, int, error)
		Update(ctx context.Context, userID, taskID, title, description string) (entity.Task, error)
		Transition(ctx context.Context, userID, taskID string, newStatus entity.TaskStatus) (entity.Task, error)
		Delete(ctx context.Context, userID, taskID string) error
	}

	// ImportantDay -.
	ImportantDay interface {
		Create(ctx context.Context, userID string, params entity.ImportantDayParams) (entity.ImportantDay, error)
		Get(ctx context.Context, userID, id string) (entity.ImportantDay, error)
		List(ctx context.Context, userID string, dayType *entity.ImportantDayType, limit, offset int) ([]entity.ImportantDay, int, error)
		Upcoming(ctx context.Context, userID string, from time.Time, days, limit, offset int) ([]entity.ImportantDayUpcoming, int, error)
		Update(ctx context.Context, userID, id string, params entity.ImportantDayParams) (entity.ImportantDay, error)
		Delete(ctx context.Context, userID, id string) error
		ReplaceReminderRules(ctx context.Context, userID, id string, rules []entity.ReminderRuleParams) ([]entity.ReminderRule, error)
	}

	// Notification -.
	Notification interface {
		List(ctx context.Context, userID string, unreadOnly bool, limit, offset int) ([]entity.Notification, int, error)
		MarkRead(ctx context.Context, userID, id string) (entity.Notification, error)
		MarkAllRead(ctx context.Context, userID string) error
	}

	// DeviceToken -.
	DeviceToken interface {
		Register(ctx context.Context, userID, token, platform, name string) (entity.DeviceToken, error)
		TestPush(ctx context.Context, userID, id, title, body string) (entity.PushTestResult, error)
		Delete(ctx context.Context, userID, id string) error
	}

	// ReminderWorker -.
	ReminderWorker interface {
		RunOnce(ctx context.Context, now time.Time, limit int) (int, error)
	}
)
