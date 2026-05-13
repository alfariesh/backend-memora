package integration_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo/persistent"
	"github.com/evrone/go-clean-template/internal/usecase/reminder"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/google/uuid"
)

const defaultIntegrationPostgresURL = "postgres://user:myAwEsOm3pa55%40w0rd@db:5432/db?sslmode=disable"

type unexpectedIntegrationEmailSender struct {
	t *testing.T
}

func (s unexpectedIntegrationEmailSender) Send(_ context.Context, _, _, _ string) (string, error) {
	s.t.Helper()
	s.t.Fatal("email sender should not be called for in-app only reminder job")

	return "", nil
}

type unexpectedIntegrationPushSender struct {
	t *testing.T
}

func (s unexpectedIntegrationPushSender) Send(_ context.Context, _, _, _ string, _ map[string]string) (string, error) {
	s.t.Helper()
	s.t.Fatal("push sender should not be called for in-app only reminder job")

	return "", nil
}

func TestReminderWorkerRunOnceProcessesDueInAppJob(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	pg := openIntegrationPostgres(t)
	userRepo := persistent.NewUserRepo(pg)
	settingsRepo := persistent.NewUserSettingsRepo(pg)
	dayRepo := persistent.NewImportantDayRepo(pg)
	ruleRepo := persistent.NewReminderRuleRepo(pg)
	jobRepo := persistent.NewReminderJobRepo(pg)
	notificationRepo := persistent.NewNotificationRepo(pg)

	now := time.Date(2026, 5, 13, 9, 0, 0, 0, time.UTC)
	userID := uuid.NewString()
	dayID := uuid.NewString()
	ruleID := uuid.NewString()
	jobID := uuid.NewString()

	t.Cleanup(func() {
		_, _ = pg.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID)
	})

	user := entity.User{
		ID:           userID,
		Username:     "worker_" + strings.ReplaceAll(userID[:8], "-", ""),
		Email:        userID + "@test.com",
		PasswordHash: "hash",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := userRepo.Store(ctx, &user); err != nil {
		t.Fatalf("store user: %v", err)
	}

	settings := entity.UserSettings{
		UserID:               userID,
		Timezone:             "UTC",
		ReminderTime:         "09:00",
		NotificationChannels: []entity.ReminderChannel{entity.ReminderChannelInApp},
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := settingsRepo.Upsert(ctx, &settings); err != nil {
		t.Fatalf("store user settings: %v", err)
	}

	day := entity.ImportantDay{
		ID:           dayID,
		UserID:       userID,
		Title:        "Mom birthday",
		Type:         entity.ImportantDayTypeBirthday,
		EventMonth:   5,
		EventDay:     20,
		Recurrence:   entity.RecurrenceYearly,
		Timezone:     "UTC",
		ReminderTime: "09:00",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := dayRepo.Store(ctx, &day); err != nil {
		t.Fatalf("store important day: %v", err)
	}

	rule := entity.ReminderRule{
		ID:             ruleID,
		UserID:         userID,
		ImportantDayID: dayID,
		OffsetDays:     7,
		Channels:       []entity.ReminderChannel{entity.ReminderChannelInApp},
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := ruleRepo.ReplaceForImportantDay(ctx, userID, dayID, []entity.ReminderRule{rule}); err != nil {
		t.Fatalf("store reminder rule: %v", err)
	}

	job := entity.ReminderJob{
		ID:             jobID,
		UserID:         userID,
		ImportantDayID: dayID,
		ReminderRuleID: &ruleID,
		OccurrenceDate: time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC),
		OffsetDays:     7,
		Channels:       []entity.ReminderChannel{entity.ReminderChannelInApp},
		ScheduledAt:    now.Add(-time.Hour),
		Status:         entity.ReminderJobStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := jobRepo.Store(ctx, &job); err != nil {
		t.Fatalf("store reminder job: %v", err)
	}

	uc := reminder.New(
		jobRepo,
		dayRepo,
		userRepo,
		settingsRepo,
		notificationRepo,
		persistent.NewDeviceTokenRepo(pg),
		unexpectedIntegrationEmailSender{t: t},
		unexpectedIntegrationPushSender{t: t},
	)

	processed, err := uc.RunOnce(ctx, now, 1)
	if err != nil {
		t.Fatalf("run reminder worker once: %v", err)
	}
	if processed != 1 {
		t.Fatalf("expected 1 processed job, got %d", processed)
	}

	assertOriginalReminderJobSent(t, ctx, pg, jobID, now)
	assertReminderNotificationStored(t, ctx, pg, userID, dayID, jobID)
	assertNextReminderJobScheduled(t, ctx, pg, userID, dayID, jobID, ruleID)
}

func openIntegrationPostgres(t *testing.T) *postgres.Postgres {
	t.Helper()

	pg, err := postgres.New(integrationPostgresURL(), postgres.MaxPoolSize(2))
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}

	t.Cleanup(pg.Close)

	return pg
}

func integrationPostgresURL() string {
	url := os.Getenv("PG_URL")
	if url == "" {
		return defaultIntegrationPostgresURL
	}
	if strings.Contains(url, "sslmode=") {
		return url
	}

	separator := "?"
	if strings.Contains(url, "?") {
		separator = "&"
	}

	return url + separator + "sslmode=disable"
}

func assertOriginalReminderJobSent(t *testing.T, ctx context.Context, pg *postgres.Postgres, jobID string, sentAt time.Time) {
	t.Helper()

	var (
		status      string
		attempts    int
		lastError   string
		lockedUntil *time.Time
		actualSent  *time.Time
	)

	err := pg.Pool.QueryRow(
		ctx,
		"SELECT status, attempts, last_error, locked_until, sent_at FROM reminder_jobs WHERE id = $1",
		jobID,
	).Scan(&status, &attempts, &lastError, &lockedUntil, &actualSent)
	if err != nil {
		t.Fatalf("query sent reminder job: %v", err)
	}

	if status != string(entity.ReminderJobStatusSent) {
		t.Fatalf("expected original job status sent, got %s", status)
	}
	if attempts != 1 {
		t.Fatalf("expected original job attempts 1, got %d", attempts)
	}
	if lastError != "" {
		t.Fatalf("expected empty last error, got %q", lastError)
	}
	if lockedUntil != nil {
		t.Fatalf("expected original job lock cleared, got %v", *lockedUntil)
	}
	if actualSent == nil || actualSent.Format("2006-01-02T15:04:05") != sentAt.Format("2006-01-02T15:04:05") {
		t.Fatalf("expected sent_at %s, got %v", sentAt.Format(time.RFC3339), actualSent)
	}
}

func assertReminderNotificationStored(t *testing.T, ctx context.Context, pg *postgres.Postgres, userID, dayID, jobID string) {
	t.Helper()

	var count int
	if err := pg.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM notifications WHERE user_id = $1", userID).Scan(&count); err != nil {
		t.Fatalf("count notifications: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 notification, got %d", count)
	}

	var (
		importantDayID   string
		notificationType string
		title            string
		body             string
		dataRaw          string
	)

	err := pg.Pool.QueryRow(
		ctx,
		"SELECT important_day_id::text, type, title, body, data::text FROM notifications WHERE user_id = $1",
		userID,
	).Scan(&importantDayID, &notificationType, &title, &body, &dataRaw)
	if err != nil {
		t.Fatalf("query notification: %v", err)
	}

	if importantDayID != dayID {
		t.Fatalf("expected notification important_day_id %s, got %s", dayID, importantDayID)
	}
	if notificationType != "important_day_reminder" {
		t.Fatalf("expected notification type important_day_reminder, got %s", notificationType)
	}
	if title != "Mom birthday is in 7 days" {
		t.Fatalf("unexpected notification title: %s", title)
	}
	if body != "Mom birthday is coming in 7 days." {
		t.Fatalf("unexpected notification body: %s", body)
	}

	var data map[string]string
	if err := json.Unmarshal([]byte(dataRaw), &data); err != nil {
		t.Fatalf("decode notification data: %v", err)
	}
	if data["important_day_id"] != dayID || data["reminder_job_id"] != jobID || data["occurrence_date"] != "2026-05-20" {
		t.Fatalf("unexpected notification data: %+v", data)
	}
}

func assertNextReminderJobScheduled(t *testing.T, ctx context.Context, pg *postgres.Postgres, userID, dayID, originalJobID, ruleID string) {
	t.Helper()

	var count int
	err := pg.Pool.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM reminder_jobs WHERE user_id = $1 AND important_day_id = $2 AND id <> $3",
		userID,
		dayID,
		originalJobID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("count next reminder jobs: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 next reminder job, got %d", count)
	}

	var (
		nextID         string
		nextRuleID     string
		occurrenceDate time.Time
		scheduledAt    time.Time
		status         string
		channelsRaw    string
	)

	err = pg.Pool.QueryRow(
		ctx,
		`SELECT id, reminder_rule_id::text, occurrence_date, scheduled_at, status, channels::text
		FROM reminder_jobs
		WHERE user_id = $1 AND important_day_id = $2 AND id <> $3`,
		userID,
		dayID,
		originalJobID,
	).Scan(&nextID, &nextRuleID, &occurrenceDate, &scheduledAt, &status, &channelsRaw)
	if err != nil {
		t.Fatalf("query next reminder job: %v", err)
	}

	if nextID == "" {
		t.Fatal("expected next reminder job id")
	}
	if nextRuleID != ruleID {
		t.Fatalf("expected next reminder rule id %s, got %s", ruleID, nextRuleID)
	}
	if occurrenceDate.Format("2006-01-02") != "2027-05-20" {
		t.Fatalf("expected next occurrence 2027-05-20, got %s", occurrenceDate.Format("2006-01-02"))
	}
	if scheduledAt.Format("2006-01-02T15:04:05") != "2027-05-13T09:00:00" {
		t.Fatalf("expected next scheduled_at 2027-05-13T09:00:00, got %s", scheduledAt.Format("2006-01-02T15:04:05"))
	}
	if status != string(entity.ReminderJobStatusPending) {
		t.Fatalf("expected next job status pending, got %s", status)
	}

	var channels []entity.ReminderChannel
	if err := json.Unmarshal([]byte(channelsRaw), &channels); err != nil {
		t.Fatalf("decode next job channels: %v", err)
	}
	if len(channels) != 1 || channels[0] != entity.ReminderChannelInApp {
		t.Fatalf("expected next job channels [in_app], got %+v", channels)
	}
}
