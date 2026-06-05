package usecase_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/alfariesh/backend-memora/internal/usecase/reminder"
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
		Channel:        entity.ReminderChannelInApp,
		ScheduledAt:    now.Add(-time.Minute),
		Status:         entity.ReminderJobStatusPending,
		Attempts:       1,
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
) {
	deps.jobRepo.EXPECT().
		ClaimDue(context.Background(), now, 10).
		Return([]entity.ReminderJob{job}, nil)
	deps.dayRepo.EXPECT().
		GetByID(context.Background(), job.UserID, job.ImportantDayID).
		Return(day, nil)
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

func expectEmailUserLoaded(deps reminderUseCaseDeps, job entity.ReminderJob, user entity.User) {
	deps.userRepo.EXPECT().
		GetByID(context.Background(), job.UserID).
		Return(user, nil)
}

func expectJobFinishedWithNext(
	t *testing.T,
	deps reminderUseCaseDeps,
	now time.Time,
	job entity.ReminderJob,
	status entity.ReminderJobStatus,
	reason string,
) {
	t.Helper()

	deps.jobRepo.EXPECT().
		FinishWithNext(context.Background(), job.ID, status, now, reason, gomock.AssignableToTypeOf(entity.ReminderJob{})).
		DoAndReturn(func(_ context.Context, _ string, _ entity.ReminderJobStatus, _ time.Time, _ string, next entity.ReminderJob) error {
			require.NotEmpty(t, next.ID)
			assert.Equal(t, job.UserID, next.UserID)
			assert.Equal(t, job.ImportantDayID, next.ImportantDayID)
			require.NotNil(t, next.ReminderRuleID)
			assert.Equal(t, *job.ReminderRuleID, *next.ReminderRuleID)
			assert.Equal(t, time.Date(2027, 5, 20, 0, 0, 0, 0, time.UTC), next.OccurrenceDate)
			assert.Equal(t, job.OffsetDays, next.OffsetDays)
			assert.Equal(t, job.Channel, next.Channel)
			assert.Equal(t, time.Date(2027, 5, 13, 9, 0, 0, 0, time.UTC), next.ScheduledAt)
			assert.Equal(t, entity.ReminderJobStatusPending, next.Status)
			assert.Equal(t, now, next.CreatedAt)
			assert.Equal(t, now, next.UpdatedAt)

			return nil
		})
}

func TestReminderRunOnceSuccessStoresNotificationAndSchedulesNext(t *testing.T) {
	t.Parallel()

	now, job, day, _ := reminderFixtures()
	uc, deps := newReminderUseCase(t)
	expectReminderJobLoaded(deps, now, job, day)
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
			assert.Equal(t, "reminder_job:"+job.ID+":in_app", notification.DedupeKey)

			var data map[string]string
			require.NoError(t, json.Unmarshal([]byte(notification.Data), &data))
			assert.Equal(t, map[string]string{
				"important_day_id": job.ImportantDayID,
				"reminder_job_id":  job.ID,
				"occurrence_date":  "2026-05-20",
			}, data)

			return nil
		})
	expectJobFinishedWithNext(t, deps, now, job, entity.ReminderJobStatusSent, "")

	processed, err := uc.RunOnce(context.Background(), now, 10)

	require.NoError(t, err)
	assert.Equal(t, 1, processed)
}

func TestReminderRunOnceSkipsUnconfiguredEmail(t *testing.T) {
	t.Parallel()

	now, job, day, user := reminderFixtures()
	job.Channel = entity.ReminderChannelEmail
	uc, deps := newReminderUseCase(t)
	expectReminderJobLoaded(deps, now, job, day)
	expectEmailUserLoaded(deps, job, user)
	deps.emailSender.EXPECT().
		Send(
			context.Background(),
			user.Email,
			"Mom birthday is in 7 days",
			gomock.Any(),
		).
		Return("", entity.ErrEmailSenderNotConfigured)
	expectJobFinishedWithNext(t, deps, now, job, entity.ReminderJobStatusSkipped, entity.ErrEmailSenderNotConfigured.Error())

	processed, err := uc.RunOnce(context.Background(), now, 10)

	require.NoError(t, err)
	assert.Equal(t, 1, processed)
}

func TestReminderRunOnceEmailFailureDoesNotBlockInAppJob(t *testing.T) {
	t.Parallel()

	now, job, day, user := reminderFixtures()
	emailJob := job
	emailJob.ID = "email-job-id"
	emailJob.Channel = entity.ReminderChannelEmail
	inAppJob := job
	inAppJob.ID = "in-app-job-id"
	inAppJob.Channel = entity.ReminderChannelInApp
	uc, deps := newReminderUseCase(t)
	settings := entity.UserSettings{
		UserID: job.UserID,
		NotificationChannels: []entity.ReminderChannel{
			entity.ReminderChannelEmail,
			entity.ReminderChannelInApp,
			entity.ReminderChannelPush,
		},
	}

	deps.jobRepo.EXPECT().
		ClaimDue(context.Background(), now, 10).
		Return([]entity.ReminderJob{emailJob, inAppJob}, nil)
	deps.dayRepo.EXPECT().
		GetByID(context.Background(), job.UserID, job.ImportantDayID).
		Return(day, nil).
		Times(2)
	deps.settingsRepo.EXPECT().
		Get(context.Background(), job.UserID).
		Return(settings, nil).
		Times(2)
	expectEmailUserLoaded(deps, emailJob, user)
	deps.emailSender.EXPECT().
		Send(context.Background(), user.Email, "Mom birthday is in 7 days", gomock.Any()).
		Return("", errInternalServErr)
	deps.jobRepo.EXPECT().
		MarkFailed(context.Background(), emailJob.ID, gomock.Any(), true).
		DoAndReturn(func(_ context.Context, _ string, reason string, retry bool) error {
			assert.Contains(t, reason, "email:")
			assert.Contains(t, reason, errInternalServErr.Error())
			assert.True(t, retry)

			return nil
		})
	deps.notificationRepo.EXPECT().
		Store(context.Background(), gomock.AssignableToTypeOf(&entity.Notification{})).
		DoAndReturn(func(_ context.Context, notification *entity.Notification) error {
			assert.Equal(t, "reminder_job:"+inAppJob.ID+":in_app", notification.DedupeKey)

			return nil
		})
	expectJobFinishedWithNext(t, deps, now, inAppJob, entity.ReminderJobStatusSent, "")

	processed, err := uc.RunOnce(context.Background(), now, 10)

	require.NoError(t, err)
	assert.Equal(t, 2, processed)
}

func TestReminderRunOnceDeactivatesUnregisteredPushToken(t *testing.T) {
	t.Parallel()

	now, job, day, _ := reminderFixtures()
	job.Channel = entity.ReminderChannelPush
	uc, deps := newReminderUseCase(t)
	expectReminderJobLoaded(deps, now, job, day)
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
		Deactivate(context.Background(), job.UserID, "device-id-123", now).
		Return(nil)
	expectJobFinishedWithNext(t, deps, now, job, entity.ReminderJobStatusSkipped, "no registered push devices")

	processed, err := uc.RunOnce(context.Background(), now, 10)

	require.NoError(t, err)
	assert.Equal(t, 1, processed)
}

func TestReminderRunOnceSkipsPushWithoutActiveTokens(t *testing.T) {
	t.Parallel()

	now, job, day, _ := reminderFixtures()
	job.Channel = entity.ReminderChannelPush
	uc, deps := newReminderUseCase(t)
	expectReminderJobLoaded(deps, now, job, day)
	deps.deviceRepo.EXPECT().
		ListActiveByUser(context.Background(), job.UserID).
		Return([]entity.DeviceToken{}, nil)
	expectJobFinishedWithNext(t, deps, now, job, entity.ReminderJobStatusSkipped, "no active push tokens")

	processed, err := uc.RunOnce(context.Background(), now, 10)

	require.NoError(t, err)
	assert.Equal(t, 1, processed)
}

func TestReminderRunOnceMarksFailedWhenFinishWithNextFails(t *testing.T) {
	t.Parallel()

	now, job, day, _ := reminderFixtures()
	uc, deps := newReminderUseCase(t)
	expectReminderJobLoaded(deps, now, job, day)
	deps.notificationRepo.EXPECT().
		Store(context.Background(), gomock.AssignableToTypeOf(&entity.Notification{})).
		Return(nil)
	deps.jobRepo.EXPECT().
		FinishWithNext(context.Background(), job.ID, entity.ReminderJobStatusSent, now, "", gomock.AssignableToTypeOf(entity.ReminderJob{})).
		Return(errInternalServErr)
	deps.jobRepo.EXPECT().
		MarkFailed(context.Background(), job.ID, gomock.Any(), true).
		DoAndReturn(func(_ context.Context, _ string, reason string, retry bool) error {
			assert.Contains(t, reason, "FinishWithNext")
			assert.True(t, retry)

			return nil
		})

	processed, err := uc.RunOnce(context.Background(), now, 10)

	require.NoError(t, err)
	assert.Equal(t, 1, processed)
}

func TestReminderRunOnceMarksFailedOnPushFailure(t *testing.T) {
	t.Parallel()

	now, job, day, _ := reminderFixtures()
	job.Channel = entity.ReminderChannelPush
	uc, deps := newReminderUseCase(t)
	expectReminderJobLoaded(deps, now, job, day)
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

	now, job, day, _ := reminderFixtures()
	job.Channel = entity.ReminderChannelPush
	job.Attempts = 3
	uc, deps := newReminderUseCase(t)
	expectReminderJobLoaded(deps, now, job, day)
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
