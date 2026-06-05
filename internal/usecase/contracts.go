// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"
	"time"

	"github.com/alfariesh/backend-memora/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=./mocks_usecase_test.go -package=usecase_test

type (
	// User -.
	User interface {
		Register(ctx context.Context, username, email, password string) (entity.User, error)
		Login(ctx context.Context, email, password string, metadata entity.SessionMetadata) (entity.AuthTokens, error)
		LoginAccessOnly(ctx context.Context, email, password string) (entity.AuthTokens, error)
		Refresh(ctx context.Context, refreshToken string, metadata entity.SessionMetadata) (entity.AuthTokens, error)
		Logout(ctx context.Context, refreshToken string) error
		ListSessions(ctx context.Context, userID string) ([]entity.UserSessionView, error)
		RevokeSession(ctx context.Context, userID, sessionID string) error
		LogoutAll(ctx context.Context, userID string) error
		ChangePassword(ctx context.Context, userID, currentPassword, newPassword string, metadata entity.SessionMetadata) (entity.AuthTokens, error)
		GetUser(ctx context.Context, userID string) (entity.User, error)
	}

	// UserSettings -.
	UserSettings interface {
		Get(ctx context.Context, userID string) (entity.UserSettings, error)
		Update(ctx context.Context, userID string, params entity.UserSettingsParams) (entity.UserSettings, error)
	}

	// ImportantDay -.
	ImportantDay interface {
		Create(ctx context.Context, userID string, params entity.ImportantDayParams) (entity.ImportantDay, error)
		Get(ctx context.Context, userID, id string) (entity.ImportantDay, error)
		List(ctx context.Context, userID string, dayType *entity.ImportantDayType, limit, offset int) ([]entity.ImportantDay, int, error)
		Upcoming(ctx context.Context, userID string, from time.Time, days, limit, offset int) ([]entity.ImportantDayUpcoming, int, error)
		Update(ctx context.Context, userID, id string, params entity.ImportantDayParams) (entity.ImportantDay, error)
		Delete(ctx context.Context, userID, id string) error
		GetReminderRules(ctx context.Context, userID, id string) ([]entity.ReminderRule, error)
		ReplaceReminderRules(ctx context.Context, userID, id string, rules []entity.ReminderRuleParams) ([]entity.ReminderRule, error)
	}

	// Notification -.
	Notification interface {
		List(ctx context.Context, userID string, unreadOnly bool, limit, offset int) ([]entity.Notification, int, error)
		CountUnread(ctx context.Context, userID string) (int, error)
		MarkRead(ctx context.Context, userID, id string) (entity.Notification, error)
		MarkAllRead(ctx context.Context, userID string) error
	}

	// DeviceToken -.
	DeviceToken interface {
		List(ctx context.Context, userID string) ([]entity.DeviceToken, error)
		Register(ctx context.Context, userID, token, platform, name string) (entity.DeviceToken, error)
		TestPush(ctx context.Context, userID, id, title, body string) (entity.PushTestResult, error)
		Delete(ctx context.Context, userID, id string) error
	}

	// ReminderWorker -.
	ReminderWorker interface {
		RunOnce(ctx context.Context, now time.Time, limit int) (int, error)
	}
)
