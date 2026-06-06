package persistent

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/alfariesh/backend-memora/pkg/postgres"
)

const reminderDeliveryColumns = "id, reminder_job_id, channel, target_id, provider, idempotency_key, status, attempts, provider_message_id, last_error, sent_at, created_at, updated_at"

// ReminderDeliveryRepo -.
type ReminderDeliveryRepo struct {
	*postgres.Postgres
}

// NewReminderDeliveryRepo -.
func NewReminderDeliveryRepo(pg *postgres.Postgres) *ReminderDeliveryRepo {
	return &ReminderDeliveryRepo{pg}
}

// Begin creates or claims a reminder delivery attempt.
func (r *ReminderDeliveryRepo) Begin(ctx context.Context, delivery entity.ReminderDelivery, now time.Time) (entity.ReminderDelivery, error) {
	sql, args, err := r.Builder.
		Insert("reminder_deliveries").
		Columns("id, reminder_job_id, channel, target_id, provider, idempotency_key, status, attempts, provider_message_id, last_error, sent_at, created_at, updated_at").
		Values(
			delivery.ID,
			delivery.ReminderJobID,
			delivery.Channel,
			delivery.TargetID,
			delivery.Provider,
			delivery.IdempotencyKey,
			entity.ReminderDeliveryStatusSending,
			1,
			"",
			"",
			nil,
			now,
			now,
		).
		Suffix(`
ON CONFLICT (reminder_job_id, channel, target_id) DO UPDATE SET
    status = CASE
        WHEN reminder_deliveries.status IN ('sent', 'skipped') THEN reminder_deliveries.status
        ELSE EXCLUDED.status
    END,
    attempts = CASE
        WHEN reminder_deliveries.status IN ('sent', 'skipped') THEN reminder_deliveries.attempts
        ELSE reminder_deliveries.attempts + 1
    END,
    last_error = CASE
        WHEN reminder_deliveries.status IN ('sent', 'skipped') THEN reminder_deliveries.last_error
        ELSE ''
    END,
    updated_at = EXCLUDED.updated_at
RETURNING ` + reminderDeliveryColumns).
		ToSql()
	if err != nil {
		return entity.ReminderDelivery{}, fmt.Errorf("ReminderDeliveryRepo - Begin - r.Builder: %w", err)
	}

	return scanReminderDelivery(r.Pool.QueryRow(ctx, sql, args...))
}

// MarkSent -.
func (r *ReminderDeliveryRepo) MarkSent(ctx context.Context, id, providerMessageID string, sentAt time.Time) error {
	return r.updateStatus(ctx, id, entity.ReminderDeliveryStatusSent, providerMessageID, "", sentAt, sentAt)
}

// MarkSkipped -.
func (r *ReminderDeliveryRepo) MarkSkipped(ctx context.Context, id, reason string, skippedAt time.Time) error {
	return r.updateStatus(ctx, id, entity.ReminderDeliveryStatusSkipped, "", reason, nil, skippedAt)
}

// MarkFailed -.
func (r *ReminderDeliveryRepo) MarkFailed(ctx context.Context, id, reason string, failedAt time.Time) error {
	return r.updateStatus(ctx, id, entity.ReminderDeliveryStatusFailed, "", reason, nil, failedAt)
}

func (r *ReminderDeliveryRepo) updateStatus(
	ctx context.Context,
	id string,
	status entity.ReminderDeliveryStatus,
	providerMessageID string,
	lastError string,
	sentAt any,
	updatedAt time.Time,
) error {
	builder := r.Builder.
		Update("reminder_deliveries").
		Set("status", status).
		Set("last_error", lastError).
		Set("updated_at", updatedAt).
		Where(sq.Eq{"id": id})

	if providerMessageID != "" {
		builder = builder.Set("provider_message_id", providerMessageID)
	}
	if sentAt != nil {
		builder = builder.Set("sent_at", sentAt)
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("ReminderDeliveryRepo - updateStatus - r.Builder: %w", err)
	}

	result, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("ReminderDeliveryRepo - updateStatus - r.Pool.Exec: %w", err)
	}
	if result.RowsAffected() == 0 {
		return entity.ErrReminderDeliveryNotFound
	}

	return nil
}

func scanReminderDelivery(row scanner) (entity.ReminderDelivery, error) {
	var delivery entity.ReminderDelivery

	err := row.Scan(
		&delivery.ID,
		&delivery.ReminderJobID,
		&delivery.Channel,
		&delivery.TargetID,
		&delivery.Provider,
		&delivery.IdempotencyKey,
		&delivery.Status,
		&delivery.Attempts,
		&delivery.ProviderMessageID,
		&delivery.LastError,
		&delivery.SentAt,
		&delivery.CreatedAt,
		&delivery.UpdatedAt,
	)

	return delivery, err
}
