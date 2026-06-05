package reminder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/alfariesh/backend-memora/internal/repo"
	"github.com/google/uuid"
)

const maxAttempts = 3

type deliveryResult struct {
	status entity.ReminderJobStatus
	reason string
}

// UseCase -.
type UseCase struct {
	jobRepo          repo.ReminderJobRepo
	dayRepo          repo.ImportantDayRepo
	userRepo         repo.UserRepo
	settingsRepo     repo.UserSettingsRepo
	notificationRepo repo.NotificationRepo
	deviceRepo       repo.DeviceTokenRepo
	emailSender      repo.EmailSender
	pushSender       repo.PushSender
}

// New -.
func New(
	jobRepo repo.ReminderJobRepo,
	dayRepo repo.ImportantDayRepo,
	userRepo repo.UserRepo,
	settingsRepo repo.UserSettingsRepo,
	notificationRepo repo.NotificationRepo,
	deviceRepo repo.DeviceTokenRepo,
	emailSender repo.EmailSender,
	pushSender repo.PushSender,
) *UseCase {
	return &UseCase{
		jobRepo:          jobRepo,
		dayRepo:          dayRepo,
		userRepo:         userRepo,
		settingsRepo:     settingsRepo,
		notificationRepo: notificationRepo,
		deviceRepo:       deviceRepo,
		emailSender:      emailSender,
		pushSender:       pushSender,
	}
}

// RunOnce claims and processes due reminder jobs.
func (uc *UseCase) RunOnce(ctx context.Context, now time.Time, limit int) (int, error) {
	jobs, err := uc.jobRepo.ClaimDue(ctx, now, limit)
	if err != nil {
		return 0, fmt.Errorf("ReminderUseCase - RunOnce - uc.jobRepo.ClaimDue: %w", err)
	}

	processed := 0
	for _, job := range jobs {
		processed++
		if err = uc.deliverJob(ctx, job, now); err != nil {
			retry := job.Attempts < maxAttempts
			if markErr := uc.jobRepo.MarkFailed(ctx, job.ID, err.Error(), retry); markErr != nil {
				return processed, fmt.Errorf("ReminderUseCase - RunOnce - uc.jobRepo.MarkFailed: %w", markErr)
			}
		}
	}

	return processed, nil
}

func (uc *UseCase) deliverJob(ctx context.Context, job entity.ReminderJob, now time.Time) error {
	day, err := uc.dayRepo.GetByID(ctx, job.UserID, job.ImportantDayID)
	if err != nil {
		return fmt.Errorf("ReminderUseCase - deliverJob - uc.dayRepo.GetByID: %w", err)
	}

	title, body := reminderCopy(day, job)
	enabled, err := uc.enabledChannel(ctx, job)
	if err != nil {
		return fmt.Errorf("ReminderUseCase - deliverJob - uc.enabledChannel: %w", err)
	}

	result := deliveryResult{status: entity.ReminderJobStatusSkipped, reason: "channel disabled"}
	if enabled {
		result, err = uc.deliverChannel(ctx, day, job, title, body, now)
		if err != nil {
			return err
		}
	}

	next, err := nextReminderJob(day, job, now)
	if err != nil {
		return fmt.Errorf("ReminderUseCase - deliverJob - nextReminderJob: %w", err)
	}

	if err = uc.jobRepo.FinishWithNext(ctx, job.ID, result.status, now, result.reason, next); err != nil {
		return fmt.Errorf("ReminderUseCase - deliverJob - uc.jobRepo.FinishWithNext: %w", err)
	}

	return nil
}

func (uc *UseCase) deliverChannel(ctx context.Context, day entity.ImportantDay, job entity.ReminderJob, title, body string, now time.Time) (deliveryResult, error) {
	switch job.Channel {
	case entity.ReminderChannelEmail:
		return uc.deliverEmail(ctx, day, job, title, body)
	case entity.ReminderChannelInApp:
		return uc.deliverInApp(ctx, job, title, body, now)
	case entity.ReminderChannelPush:
		return uc.deliverPush(ctx, job, title, body, now)
	default:
		return deliveryResult{}, fmt.Errorf("invalid reminder channel: %s", job.Channel)
	}
}

func (uc *UseCase) deliverEmail(ctx context.Context, day entity.ImportantDay, job entity.ReminderJob, title, body string) (deliveryResult, error) {
	if uc.emailSender == nil {
		return deliveryResult{status: entity.ReminderJobStatusSkipped, reason: entity.ErrEmailSenderNotConfigured.Error()}, nil
	}

	user, err := uc.userRepo.GetByID(ctx, job.UserID)
	if err != nil {
		return deliveryResult{}, fmt.Errorf("ReminderUseCase - deliverEmail - uc.userRepo.GetByID: %w", err)
	}

	if _, err := uc.emailSender.Send(ctx, user.Email, title, reminderHTML(user, day, job, body)); err != nil {
		if errors.Is(err, entity.ErrEmailSenderNotConfigured) {
			return deliveryResult{status: entity.ReminderJobStatusSkipped, reason: entity.ErrEmailSenderNotConfigured.Error()}, nil
		}

		return deliveryResult{}, fmt.Errorf("email: %w", err)
	}

	return deliveryResult{status: entity.ReminderJobStatusSent}, nil
}

func (uc *UseCase) deliverInApp(ctx context.Context, job entity.ReminderJob, title, body string, now time.Time) (deliveryResult, error) {
	if err := uc.storeNotification(ctx, job, title, body, now); err != nil {
		return deliveryResult{}, fmt.Errorf("in_app: %w", err)
	}

	return deliveryResult{status: entity.ReminderJobStatusSent}, nil
}

func (uc *UseCase) deliverPush(ctx context.Context, job entity.ReminderJob, title, body string, now time.Time) (deliveryResult, error) {
	if uc.pushSender == nil {
		return deliveryResult{status: entity.ReminderJobStatusSkipped, reason: entity.ErrPushSenderNotConfigured.Error()}, nil
	}

	result, err := uc.sendPush(ctx, job, title, body, now)
	if err != nil {
		return deliveryResult{}, fmt.Errorf("push: %w", err)
	}

	return result, nil
}

func (uc *UseCase) enabledChannel(ctx context.Context, job entity.ReminderJob) (bool, error) {
	if uc.settingsRepo == nil {
		return true, nil
	}

	settings, err := uc.settingsRepo.Get(ctx, job.UserID)
	if err != nil {
		if errors.Is(err, entity.ErrUserSettingsNotFound) {
			return true, nil
		}

		return false, err
	}

	return hasChannel(settings.NotificationChannels, job.Channel), nil
}

func (uc *UseCase) sendPush(ctx context.Context, job entity.ReminderJob, title, body string, now time.Time) (deliveryResult, error) {
	tokens, err := uc.deviceRepo.ListActiveByUser(ctx, job.UserID)
	if err != nil {
		return deliveryResult{}, fmt.Errorf("ReminderUseCase - sendPush - uc.deviceRepo.ListActiveByUser: %w", err)
	}

	if len(tokens) == 0 {
		return deliveryResult{status: entity.ReminderJobStatusSkipped, reason: "no active push tokens"}, nil
	}

	data := map[string]string{
		"type":             "important_day_reminder",
		"important_day_id": job.ImportantDayID,
		"reminder_job_id":  job.ID,
		"occurrence_date":  job.OccurrenceDate.Format("2006-01-02"),
	}

	failures := make([]string, 0)
	sent := false
	for _, token := range tokens {
		if _, err = uc.pushSender.Send(ctx, token.Token, title, body, data); err != nil {
			if errors.Is(err, entity.ErrPushDeviceNotRegistered) {
				if deactivateErr := uc.deviceRepo.Deactivate(ctx, job.UserID, token.ID, now); deactivateErr != nil {
					failures = append(failures, deactivateErr.Error())
				}

				continue
			}

			failures = append(failures, err.Error())
			continue
		}

		sent = true
	}

	if len(failures) > 0 {
		return deliveryResult{}, errors.New(strings.Join(failures, "; "))
	}

	if !sent {
		return deliveryResult{status: entity.ReminderJobStatusSkipped, reason: "no registered push devices"}, nil
	}

	return deliveryResult{status: entity.ReminderJobStatusSent}, nil
}

func (uc *UseCase) storeNotification(ctx context.Context, job entity.ReminderJob, title, body string, now time.Time) error {
	importantDayID := job.ImportantDayID
	data, err := json.Marshal(map[string]string{
		"important_day_id": job.ImportantDayID,
		"reminder_job_id":  job.ID,
		"occurrence_date":  job.OccurrenceDate.Format("2006-01-02"),
	})
	if err != nil {
		return err
	}

	notification := entity.Notification{
		ID:             uuid.New().String(),
		UserID:         job.UserID,
		ImportantDayID: &importantDayID,
		Type:           "important_day_reminder",
		Title:          title,
		Body:           body,
		Data:           string(data),
		DedupeKey:      reminderNotificationDedupeKey(job),
		CreatedAt:      now,
	}

	return uc.notificationRepo.Store(ctx, &notification)
}

func nextReminderJob(day entity.ImportantDay, job entity.ReminderJob, now time.Time) (entity.ReminderJob, error) {
	nextFrom := job.OccurrenceDate.AddDate(0, 0, 1)
	nextOccurrence, err := day.NextOccurrence(nextFrom)
	if err != nil {
		return entity.ReminderJob{}, err
	}

	scheduledAt, err := day.ReminderScheduledAt(nextOccurrence, job.OffsetDays)
	if err != nil {
		return entity.ReminderJob{}, err
	}

	return entity.ReminderJob{
		ID:             uuid.New().String(),
		UserID:         job.UserID,
		ImportantDayID: job.ImportantDayID,
		ReminderRuleID: job.ReminderRuleID,
		OccurrenceDate: nextOccurrence,
		OffsetDays:     job.OffsetDays,
		Channel:        job.Channel,
		ScheduledAt:    scheduledAt,
		Status:         entity.ReminderJobStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func reminderNotificationDedupeKey(job entity.ReminderJob) string {
	return "reminder_job:" + job.ID + ":in_app"
}

func reminderCopy(day entity.ImportantDay, job entity.ReminderJob) (string, string) {
	if job.OffsetDays == 0 {
		title := fmt.Sprintf("%s is today", day.Title)

		return title, fmt.Sprintf("%s is today.", day.Title)
	}

	title := fmt.Sprintf("%s is in %d days", day.Title, job.OffsetDays)

	return title, fmt.Sprintf("%s is coming in %d days.", day.Title, job.OffsetDays)
}

func reminderHTML(user entity.User, day entity.ImportantDay, job entity.ReminderJob, body string) string {
	return fmt.Sprintf(
		"<p>Hi %s,</p><p>%s</p><p>Date: %s</p>",
		html.EscapeString(user.Username),
		html.EscapeString(body),
		html.EscapeString(job.OccurrenceDate.Format("2006-01-02")),
	) + fmt.Sprintf("<p>Event: %s</p>", html.EscapeString(day.Title))
}

func hasChannel(channels []entity.ReminderChannel, expected entity.ReminderChannel) bool {
	for _, channel := range channels {
		if channel == expected {
			return true
		}
	}

	return false
}
