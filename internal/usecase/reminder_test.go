package usecase_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase/reminder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type reminderUseCaseDeps struct {
	jobRepo          *MockReminderJobRepo
	dayRepo          *MockImportantDayRepo
	userRepo         *MockUserRepo
	settingsRepo     *MockUserSettingsRepo
	notificationRepo *MockNotificationRepo
	deviceRepo       *MockDeviceTokenRepo
	emailSender      *MockEmailSender
	pushSender       *MockPushSender
}

func newReminderUseCase(t *testing.T) (*reminder.UseCase, reminderUseCaseDeps) {
	t.Helper()

	ctrl := gomock.NewController(t)
	deps := reminderUseCaseDeps{
		jobRepo:          NewMockReminderJobRepo(ctrl),
		dayRepo:          NewMockImportantDayRepo(ctrl),
		userRepo:         NewMockUserRepo(ctrl),
		settingsRepo:     NewMockUserSettingsRepo(ctrl),
		notificationRepo: NewMockNotificationRepo(ctrl),
		deviceRepo:       NewMockDeviceTokenRepo(ctrl),
		emailSender:      NewMockEmailSender(ctrl),
		pushSender:       NewMockPushSender(ctrl),
	}

	useCase := reminder.New(
		deps.jobRepo,
		deps.dayRepo,
		deps.userRepo,
		deps.settingsRepo,
		deps.notificationRepo,
		deps.deviceRepo,
		deps.emailSender,
		deps.pushSender,
	)

	return useCase, deps
}

func reminderFixtures() (time.Time, entity.ReminderJob, entity.ImportantDay, entity.User) {
	now := time.Date(2026, 5, 13, 9, 0, 0, 0, time.UTC)
	reminderRuleID := "rule-id-123"
	job := entity.ReminderJob{
		ID:             "job-id-123",
		UserID:         "user-id-123",
		ImportantDayID: "day-id-123",
		ReminderRuleID: &reminderRuleID,
		OccurrenceDate: time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC),
		OffsetDays:     7,
		Channels: []entity.ReminderChannel{
			entity.ReminderChannelInApp,
			entity.ReminderChannelPush,
		},
		ScheduledAt: now.Add(-time.Minute),
		Status:      entity.ReminderJobStatusPending,
		Attempts:    1,
	}
	day := entity.ImportantDay{
		ID:           "day-id-123",
		UserID:       "user-id-123",
		Title:        "Mom birthday",
		Type:         entity.ImportantDayTypeBirthday,
		EventMonth:   5,
		EventDay:     20,
		Recurrence:   entity.RecurrenceYearly,
		Timezone:     "UTC",
		ReminderTime: "09:00",
	}
	user := entity.User{
		ID:       "user-id-123",
		Username: "Ayu",
		Email:    "ayu@example.com",
	}

	return now, job, day, user
}

func expectReminderJobLoaded(
	deps reminderUseCaseDeps,
	now time.Time,
	job entity.ReminderJob,
	day entity.ImportantDay,
	user entity.User,
) {
	deps.jobRepo.EXPECT().
		ClaimDue(context.Background(), now, 10).
		Return([]entity.ReminderJob{job}, nil)
	deps.dayRepo.EXPECT().
		GetByID(context.Background(), job.UserID, job.ImportantDayID).
		Return(day, nil)
	deps.userRepo.EXPECT().
		GetByID(context.Background(), job.UserID).
		Return(user, nil)
	deps.settingsRepo.EXPECT().
		Get(context.Background(), job.UserID).
		Return(entity.UserSettings{
			UserID: job.UserID,
			NotificationChannels: []entity.ReminderChannel{
				entity.ReminderChannelEmail,
				entity.ReminderChannelInApp,
				entity.ReminderChannelPush,
			},
		}, nil)
}

func expectNextReminderScheduled(t *testing.T, deps reminderUseCaseDeps, now time.Time, job entity.ReminderJob) {
	t.Helper()

	deps.jobRepo.EXPECT().
		MarkSent(context.Background(), job.ID, now).
		Return(nil)
	deps.jobRepo.EXPECT().
		Store(context.Background(), gomock.AssignableToTypeOf(&entity.ReminderJob{})).
		DoAndReturn(func(_ context.Context, next *entity.ReminderJob) error {
			require.NotEmpty(t, next.ID)
			assert.Equal(t, job.UserID, next.UserID)
			assert.Equal(t, job.ImportantDayID, next.ImportantDayID)
			require.NotNil(t, next.ReminderRuleID)
			assert.Equal(t, *job.ReminderRuleID, *next.ReminderRuleID)
			assert.Equal(t, time.Date(2027, 5, 20, 0, 0, 0, 0, time.UTC), next.OccurrenceDate)
			assert.Equal(t, job.OffsetDays, next.OffsetDays)
			assert.Equal(t, job.Channels, next.Channels)
			assert.Equal(t, time.Date(2027, 5, 13, 9, 0, 0, 0, time.UTC), next.ScheduledAt)
			assert.Equal(t, entity.ReminderJobStatusPending, next.Status)
			assert.Equal(t, now, next.CreatedAt)
			assert.Equal(t, now, next.UpdatedAt)

			return nil
		})
}

func TestReminderRunOnceSuccessStoresNotificationPushAndSchedulesNext(t *testing.T) {
	t.Parallel()

	now, job, day, user := reminderFixtures()
	uc, deps := newReminderUseCase(t)
	expectReminderJobLoaded(deps, now, job, day, user)
	deps.deviceRepo.EXPECT().
		ListActiveByUser(context.Background(), job.UserID).
		Return([]entity.DeviceToken{
			{ID: "device-id-123", UserID: job.UserID, Token: "ExpoPushToken[test]", Active: true},
		}, nil)
	deps.pushSender.EXPECT().
		Send(
			context.Background(),
			"ExpoPushToken[test]",
			"Mom birthday is in 7 days",
			"Mom birthday is coming in 7 days.",
			gomock.Any(),
		).
		DoAndReturn(func(_ context.Context, _ string, _ string, _ string, data map[string]string) (string, error) {
			assert.Equal(t, map[string]string{
				"type":             "important_day_reminder",
				"important_day_id": job.ImportantDayID,
				"reminder_job_id":  job.ID,
				"occurrence_date":  "2026-05-20",
			}, data)

			return "ticket-id-123", nil
		})
	deps.notificationRepo.EXPECT().
		Store(context.Background(), gomock.AssignableToTypeOf(&entity.Notification{})).
		DoAndReturn(func(_ context.Context, notification *entity.Notification) error {
			require.NotEmpty(t, notification.ID)
			assert.Equal(t, job.UserID, notification.UserID)
			require.NotNil(t, notification.ImportantDayID)
			assert.Equal(t, job.ImportantDayID, *notification.ImportantDayID)
			assert.Equal(t, "important_day_reminder", notification.Type)
			assert.Equal(t, "Mom birthday is in 7 days", notification.Title)
			assert.Equal(t, "Mom birthday is coming in 7 days.", notification.Body)
			assert.Equal(t, now, notification.CreatedAt)

			var data map[string]string
			require.NoError(t, json.Unmarshal([]byte(notification.Data), &data))
			assert.Equal(t, map[string]string{
				"important_day_id": job.ImportantDayID,
				"reminder_job_id":  job.ID,
				"occurrence_date":  "2026-05-20",
			}, data)

			return nil
		})
	expectNextReminderScheduled(t, deps, now, job)

	processed, err := uc.RunOnce(context.Background(), now, 10)

	require.NoError(t, err)
	assert.Equal(t, 1, processed)
}

func TestReminderRunOnceSkipsUnconfiguredEmailAndStoresInApp(t *testing.T) {
	t.Parallel()

	now, job, day, user := reminderFixtures()
	job.Channels = []entity.ReminderChannel{
		entity.ReminderChannelEmail,
		entity.ReminderChannelInApp,
	}
	uc, deps := newReminderUseCase(t)
	expectReminderJobLoaded(deps, now, job, day, user)
	deps.emailSender.EXPECT().
		Send(
			context.Background(),
			user.Email,
			"Mom birthday is in 7 days",
			gomock.Any(),
		).
		Return("", entity.ErrEmailSenderNotConfigured)
	deps.notificationRepo.EXPECT().
		Store(context.Background(), gomock.AssignableToTypeOf(&entity.Notification{})).
		DoAndReturn(func(_ context.Context, notification *entity.Notification) error {
			require.NotEmpty(t, notification.ID)
			assert.Equal(t, job.UserID, notification.UserID)
			assert.Equal(t, "important_day_reminder", notification.Type)
			assert.Equal(t, "Mom birthday is in 7 days", notification.Title)
			assert.Equal(t, "Mom birthday is coming in 7 days.", notification.Body)

			return nil
		})
	expectNextReminderScheduled(t, deps, now, job)

	processed, err := uc.RunOnce(context.Background(), now, 10)

	require.NoError(t, err)
	assert.Equal(t, 1, processed)
}

func TestReminderRunOnceDeactivatesUnregisteredPushToken(t *testing.T) {
	t.Parallel()

	now, job, day, user := reminderFixtures()
	job.Channels = []entity.ReminderChannel{entity.ReminderChannelPush}
	uc, deps := newReminderUseCase(t)
	expectReminderJobLoaded(deps, now, job, day, user)
	deps.deviceRepo.EXPECT().
		ListActiveByUser(context.Background(), job.UserID).
		Return([]entity.DeviceToken{
			{ID: "device-id-123", UserID: job.UserID, Token: "ExpoPushToken[test]", Active: true},
		}, nil)
	deps.pushSender.EXPECT().
		Send(
			context.Background(),
			"ExpoPushToken[test]",
			"Mom birthday is in 7 days",
			"Mom birthday is coming in 7 days.",
			gomock.Any(),
		).
		Return("", fmt.Errorf("%w: inactive token", entity.ErrPushDeviceNotRegistered))
	deps.deviceRepo.EXPECT().
		Deactivate(context.Background(), job.UserID, "device-id-123", gomock.AssignableToTypeOf(time.Time{})).
		Return(nil)
	expectNextReminderScheduled(t, deps, now, job)

	processed, err := uc.RunOnce(context.Background(), now, 10)

	require.NoError(t, err)
	assert.Equal(t, 1, processed)
}

func TestReminderRunOnceMarksFailedOnPushFailure(t *testing.T) {
	t.Parallel()

	now, job, day, user := reminderFixtures()
	job.Channels = []entity.ReminderChannel{entity.ReminderChannelPush}
	uc, deps := newReminderUseCase(t)
	expectReminderJobLoaded(deps, now, job, day, user)
	deps.deviceRepo.EXPECT().
		ListActiveByUser(context.Background(), job.UserID).
		Return([]entity.DeviceToken{
			{ID: "device-id-123", UserID: job.UserID, Token: "ExpoPushToken[test]", Active: true},
		}, nil)
	deps.pushSender.EXPECT().
		Send(
			context.Background(),
			"ExpoPushToken[test]",
			"Mom birthday is in 7 days",
			"Mom birthday is coming in 7 days.",
			gomock.Any(),
		).
		Return("", errInternalServErr)
	deps.jobRepo.EXPECT().
		MarkFailed(context.Background(), job.ID, gomock.Any(), true).
		DoAndReturn(func(_ context.Context, _ string, reason string, retry bool) error {
			assert.Contains(t, reason, "push:")
			assert.Contains(t, reason, errInternalServErr.Error())
			assert.True(t, retry)

			return nil
		})

	processed, err := uc.RunOnce(context.Background(), now, 10)

	require.NoError(t, err)
	assert.Equal(t, 1, processed)
}

func TestReminderRunOnceFinalAttemptMarksFailedWithoutRetry(t *testing.T) {
	t.Parallel()

	now, job, day, user := reminderFixtures()
	job.Channels = []entity.ReminderChannel{entity.ReminderChannelPush}
	job.Attempts = 3
	uc, deps := newReminderUseCase(t)
	expectReminderJobLoaded(deps, now, job, day, user)
	deps.deviceRepo.EXPECT().
		ListActiveByUser(context.Background(), job.UserID).
		Return([]entity.DeviceToken{
			{ID: "device-id-123", UserID: job.UserID, Token: "ExpoPushToken[test]", Active: true},
		}, nil)
	deps.pushSender.EXPECT().
		Send(
			context.Background(),
			"ExpoPushToken[test]",
			"Mom birthday is in 7 days",
			"Mom birthday is coming in 7 days.",
			gomock.Any(),
		).
		Return("", errInternalServErr)
	deps.jobRepo.EXPECT().
		MarkFailed(context.Background(), job.ID, gomock.Any(), false).
		DoAndReturn(func(_ context.Context, _ string, reason string, retry bool) error {
			assert.Contains(t, reason, "push:")
			assert.Contains(t, reason, errInternalServErr.Error())
			assert.False(t, retry)

			return nil
		})

	processed, err := uc.RunOnce(context.Background(), now, 10)

	require.NoError(t, err)
	assert.Equal(t, 1, processed)
}

func TestReminderRunOnceClaimDueError(t *testing.T) {
	t.Parallel()

	now, _, _, _ := reminderFixtures()
	uc, deps := newReminderUseCase(t)
	deps.jobRepo.EXPECT().
		ClaimDue(context.Background(), now, 10).
		Return(nil, errInternalServErr)

	processed, err := uc.RunOnce(context.Background(), now, 10)

	require.ErrorIs(t, err, errInternalServErr)
	assert.Equal(t, 0, processed)
}
