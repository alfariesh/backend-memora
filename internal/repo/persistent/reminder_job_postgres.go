package persistent

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/pkg/postgres"
)

const reminderJobColumns = "id, user_id, important_day_id, reminder_rule_id, occurrence_date, offset_days, channels, scheduled_at, status, attempts, last_error, locked_until, sent_at, created_at, updated_at"

// ReminderJobRepo -.
type ReminderJobRepo struct {
	*postgres.Postgres
}

// NewReminderJobRepo -.
func NewReminderJobRepo(pg *postgres.Postgres) *ReminderJobRepo {
	return &ReminderJobRepo{pg}
}

// Store -.
func (r *ReminderJobRepo) Store(ctx context.Context, job *entity.ReminderJob) error {
	channels, err := marshalChannels(job.Channels)
	if err != nil {
		return fmt.Errorf("ReminderJobRepo - Store - marshal: %w", err)
	}

	sql, args, err := r.Builder.
		Insert("reminder_jobs").
		Columns("id, user_id, important_day_id, reminder_rule_id, occurrence_date, offset_days, channels, scheduled_at, status, attempts, last_error, locked_until, sent_at, created_at, updated_at").
		Values(
			job.ID,
			job.UserID,
			job.ImportantDayID,
			job.ReminderRuleID,
			job.OccurrenceDate,
			job.OffsetDays,
			channels,
			job.ScheduledAt,
			job.Status,
			job.Attempts,
			job.LastError,
			job.LockedUntil,
			job.SentAt,
			job.CreatedAt,
			job.UpdatedAt,
		).
		Suffix("ON CONFLICT (important_day_id, occurrence_date, offset_days) DO NOTHING").
		ToSql()
	if err != nil {
		return fmt.Errorf("ReminderJobRepo - Store - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("ReminderJobRepo - Store - r.Pool.Exec: %w", err)
	}

	return nil
}

// ReplacePendingForImportantDay -.
func (r *ReminderJobRepo) ReplacePendingForImportantDay(ctx context.Context, userID, importantDayID string, jobs []entity.ReminderJob) (err error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ReminderJobRepo - ReplacePendingForImportantDay - Begin: %w", err)
	}
	defer rollbackTx(ctx, tx, &err, "ReminderJobRepo - ReplacePendingForImportantDay - Rollback")

	if err = replacePendingReminderJobsTx(ctx, r.Builder, tx, userID, importantDayID, jobs); err != nil {
		return fmt.Errorf("ReminderJobRepo - ReplacePendingForImportantDay - replacePendingReminderJobsTx: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("ReminderJobRepo - ReplacePendingForImportantDay - Commit: %w", err)
	}

	return nil
}

// ClaimDue -.
func (r *ReminderJobRepo) ClaimDue(ctx context.Context, now time.Time, limit int) ([]entity.ReminderJob, error) {
	if limit <= 0 {
		limit = 50
	}

	sql := `
WITH candidates AS (
    SELECT id
    FROM reminder_jobs
    WHERE status = 'pending'
      AND scheduled_at <= $1
      AND (locked_until IS NULL OR locked_until < $1)
    ORDER BY scheduled_at ASC
    LIMIT $2
    FOR UPDATE SKIP LOCKED
)
UPDATE reminder_jobs
SET locked_until = $3,
    attempts = attempts + 1,
    updated_at = $1
WHERE id IN (SELECT id FROM candidates)
RETURNING ` + reminderJobColumns

	rows, err := r.Pool.Query(ctx, sql, now, limit, now.Add(5*time.Minute))
	if err != nil {
		return nil, fmt.Errorf("ReminderJobRepo - ClaimDue - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	jobs := make([]entity.ReminderJob, 0, limit)
	for rows.Next() {
		job, scanErr := scanReminderJob(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("ReminderJobRepo - ClaimDue - rows.Scan: %w", scanErr)
		}

		jobs = append(jobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ReminderJobRepo - ClaimDue - rows.Err: %w", err)
	}

	return jobs, nil
}

// MarkSent -.
func (r *ReminderJobRepo) MarkSent(ctx context.Context, id string, sentAt time.Time) error {
	sql, args, err := r.Builder.
		Update("reminder_jobs").
		Set("status", entity.ReminderJobStatusSent).
		Set("sent_at", sentAt).
		Set("locked_until", nil).
		Set("updated_at", sentAt).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("ReminderJobRepo - MarkSent - r.Builder: %w", err)
	}

	result, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("ReminderJobRepo - MarkSent - r.Pool.Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrReminderJobNotFound
	}

	return nil
}

// MarkFailed -.
func (r *ReminderJobRepo) MarkFailed(ctx context.Context, id, reason string, retry bool) error {
	status := entity.ReminderJobStatusFailed
	var lockedUntil any
	if retry {
		status = entity.ReminderJobStatusPending
		lockedUntil = time.Now().UTC().Add(15 * time.Minute)
	}

	sql, args, err := r.Builder.
		Update("reminder_jobs").
		Set("status", status).
		Set("last_error", reason).
		Set("locked_until", lockedUntil).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("ReminderJobRepo - MarkFailed - r.Builder: %w", err)
	}

	result, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("ReminderJobRepo - MarkFailed - r.Pool.Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrReminderJobNotFound
	}

	return nil
}

func scanReminderJob(row scanner) (entity.ReminderJob, error) {
	var (
		job         entity.ReminderJob
		channelsRaw []byte
	)

	err := row.Scan(
		&job.ID,
		&job.UserID,
		&job.ImportantDayID,
		&job.ReminderRuleID,
		&job.OccurrenceDate,
		&job.OffsetDays,
		&channelsRaw,
		&job.ScheduledAt,
		&job.Status,
		&job.Attempts,
		&job.LastError,
		&job.LockedUntil,
		&job.SentAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return entity.ReminderJob{}, err
	}

	job.Channels, err = unmarshalChannels(channelsRaw)
	if err != nil {
		return entity.ReminderJob{}, err
	}

	return job, nil
}
