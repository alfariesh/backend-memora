package reminder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/google/uuid"
)

const maxAttempts = 3

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

	user, err := uc.userRepo.GetByID(ctx, job.UserID)
	if err != nil {
		return fmt.Errorf("ReminderUseCase - deliverJob - uc.userRepo.GetByID: %w", err)
	}

	title, body := reminderCopy(day, job)
	channels, err := uc.enabledChannels(ctx, job)
	if err != nil {
		return fmt.Errorf("ReminderUseCase - deliverJob - uc.enabledChannels: %w", err)
	}

	failures := make([]string, 0)

	if hasChannel(channels, entity.ReminderChannelEmail) {
		if _, err = uc.emailSender.Send(ctx, user.Email, title, reminderHTML(user, day, job, body)); err != nil {
			failures = append(failures, "email: "+err.Error())
		}
	}

	if hasChannel(channels, entity.ReminderChannelPush) {
		if err = uc.sendPush(ctx, job, title, body); err != nil {
			failures = append(failures, "push: "+err.Error())
		}
	}

	if len(failures) > 0 {
		return errors.New(strings.Join(failures, "; "))
	}

	if hasChannel(channels, entity.ReminderChannelInApp) {
		if err = uc.storeNotification(ctx, job, title, body, now); err != nil {
			return fmt.Errorf("in_app: %w", err)
		}
	}

	if err = uc.jobRepo.MarkSent(ctx, job.ID, now); err != nil {
		return fmt.Errorf("ReminderUseCase - deliverJob - uc.jobRepo.MarkSent: %w", err)
	}

	if err = uc.scheduleNext(ctx, day, job, now); err != nil {
		return fmt.Errorf("ReminderUseCase - deliverJob - uc.scheduleNext: %w", err)
	}

	return nil
}

func (uc *UseCase) enabledChannels(ctx context.Context, job entity.ReminderJob) ([]entity.ReminderChannel, error) {
	if uc.settingsRepo == nil {
		return job.Channels, nil
	}

	settings, err := uc.settingsRepo.Get(ctx, job.UserID)
	if err != nil {
		if errors.Is(err, entity.ErrUserSettingsNotFound) {
			return job.Channels, nil
		}

		return nil, err
	}

	return entity.FilterReminderChannels(job.Channels, settings.NotificationChannels), nil
}

func (uc *UseCase) sendPush(ctx context.Context, job entity.ReminderJob, title, body string) error {
	tokens, err := uc.deviceRepo.ListActiveByUser(ctx, job.UserID)
	if err != nil {
		return fmt.Errorf("ReminderUseCase - sendPush - uc.deviceRepo.ListActiveByUser: %w", err)
	}

	if len(tokens) == 0 {
		return nil
	}

	data := map[string]string{
		"type":             "important_day_reminder",
		"important_day_id": job.ImportantDayID,
		"reminder_job_id":  job.ID,
		"occurrence_date":  job.OccurrenceDate.Format("2006-01-02"),
	}

	failures := make([]string, 0)
	for _, token := range tokens {
		if _, err = uc.pushSender.Send(ctx, token.Token, title, body, data); err != nil {
			if errors.Is(err, entity.ErrPushDeviceNotRegistered) {
				if deactivateErr := uc.deviceRepo.Deactivate(ctx, job.UserID, token.ID, time.Now().UTC()); deactivateErr != nil {
					failures = append(failures, deactivateErr.Error())
				}

				continue
			}

			failures = append(failures, err.Error())
		}
	}

	if len(failures) > 0 {
		return errors.New(strings.Join(failures, "; "))
	}

	return nil
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
		CreatedAt:      now,
	}

	return uc.notificationRepo.Store(ctx, &notification)
}

func (uc *UseCase) scheduleNext(ctx context.Context, day entity.ImportantDay, job entity.ReminderJob, now time.Time) error {
	nextFrom := job.OccurrenceDate.AddDate(0, 0, 1)
	nextOccurrence, err := day.NextOccurrence(nextFrom)
	if err != nil {
		return err
	}

	scheduledAt, err := day.ReminderScheduledAt(nextOccurrence, job.OffsetDays)
	if err != nil {
		return err
	}

	next := entity.ReminderJob{
		ID:             uuid.New().String(),
		UserID:         job.UserID,
		ImportantDayID: job.ImportantDayID,
		ReminderRuleID: job.ReminderRuleID,
		OccurrenceDate: nextOccurrence,
		OffsetDays:     job.OffsetDays,
		Channels:       job.Channels,
		ScheduledAt:    scheduledAt,
		Status:         entity.ReminderJobStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return uc.jobRepo.Store(ctx, &next)
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
