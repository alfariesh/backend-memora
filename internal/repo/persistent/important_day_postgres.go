package persistent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/alfariesh/backend-memora/internal/repo"
	"github.com/alfariesh/backend-memora/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

// ImportantDayRepo -.
type ImportantDayRepo struct {
	*postgres.Postgres
}

// NewImportantDayRepo -.
func NewImportantDayRepo(pg *postgres.Postgres) *ImportantDayRepo {
	return &ImportantDayRepo{pg}
}

// Store -.
func (r *ImportantDayRepo) Store(ctx context.Context, day *entity.ImportantDay) error {
	sql, args, err := r.Builder.
		Insert("important_days").
		Columns(
			"id, user_id, title, type, person_name, relationship, description, event_year, event_month, event_day, recurrence, timezone, reminder_time, created_at, updated_at",
		).
		Values(
			day.ID,
			day.UserID,
			day.Title,
			day.Type,
			day.PersonName,
			day.Relationship,
			day.Description,
			day.EventYear,
			day.EventMonth,
			day.EventDay,
			day.Recurrence,
			day.Timezone,
			day.ReminderTime,
			day.CreatedAt,
			day.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("ImportantDayRepo - Store - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("ImportantDayRepo - Store - r.Pool.Exec: %w", err)
	}

	return nil
}

// GetByID -.
func (r *ImportantDayRepo) GetByID(ctx context.Context, userID, id string) (entity.ImportantDay, error) {
	sql, args, err := r.Builder.
		Select(importantDayColumns()).
		From("important_days").
		Where(sq.Eq{"id": id, "user_id": userID}).
		ToSql()
	if err != nil {
		return entity.ImportantDay{}, fmt.Errorf("ImportantDayRepo - GetByID - r.Builder: %w", err)
	}

	day, err := scanImportantDay(r.Pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ImportantDay{}, entity.ErrImportantDayNotFound
		}

		return entity.ImportantDay{}, fmt.Errorf("ImportantDayRepo - GetByID - scan: %w", err)
	}

	return day, nil
}

// List -.
func (r *ImportantDayRepo) List(ctx context.Context, userID string, filter repo.ImportantDayFilter) ([]entity.ImportantDay, int, error) {
	countBuilder := r.Builder.
		Select("COUNT(*)").
		From("important_days").
		Where(sq.Eq{"user_id": userID})

	if filter.Type != nil {
		countBuilder = countBuilder.Where(sq.Eq{"type": *filter.Type})
	}

	countSQL, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("ImportantDayRepo - List - countBuilder: %w", err)
	}

	var total int

	err = r.Pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("ImportantDayRepo - List - count query: %w", err)
	}

	dataBuilder := r.Builder.
		Select(importantDayColumns()).
		From("important_days").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset)

	if filter.Type != nil {
		dataBuilder = dataBuilder.Where(sq.Eq{"type": *filter.Type})
	}

	dataSQL, dataArgs, err := dataBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("ImportantDayRepo - List - dataBuilder: %w", err)
	}

	rows, err := r.Pool.Query(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("ImportantDayRepo - List - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	days := make([]entity.ImportantDay, 0, filter.Limit)
	for rows.Next() {
		day, scanErr := scanImportantDay(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("ImportantDayRepo - List - rows.Scan: %w", scanErr)
		}

		days = append(days, day)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("ImportantDayRepo - List - rows.Err: %w", err)
	}

	return days, total, nil
}

// Update -.
func (r *ImportantDayRepo) Update(ctx context.Context, day *entity.ImportantDay) error {
	sql, args, err := r.Builder.
		Update("important_days").
		Set("title", day.Title).
		Set("type", day.Type).
		Set("person_name", day.PersonName).
		Set("relationship", day.Relationship).
		Set("description", day.Description).
		Set("event_year", day.EventYear).
		Set("event_month", day.EventMonth).
		Set("event_day", day.EventDay).
		Set("recurrence", day.Recurrence).
		Set("timezone", day.Timezone).
		Set("reminder_time", day.ReminderTime).
		Set("updated_at", day.UpdatedAt).
		Where(sq.Eq{"id": day.ID, "user_id": day.UserID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("ImportantDayRepo - Update - r.Builder: %w", err)
	}

	result, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("ImportantDayRepo - Update - r.Pool.Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrImportantDayNotFound
	}

	return nil
}

// Delete -.
func (r *ImportantDayRepo) Delete(ctx context.Context, userID, id string) error {
	sql, args, err := r.Builder.
		Delete("important_days").
		Where(sq.Eq{"id": id, "user_id": userID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("ImportantDayRepo - Delete - r.Builder: %w", err)
	}

	result, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("ImportantDayRepo - Delete - r.Pool.Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrImportantDayNotFound
	}

	return nil
}

// StoreWithReminderRulesAndJobs stores an important day, reminder rules, and reminder jobs atomically.
func (r *ImportantDayRepo) StoreWithReminderRulesAndJobs(ctx context.Context, day *entity.ImportantDay, rules []entity.ReminderRule, jobs []entity.ReminderJob) (err error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ImportantDayRepo - StoreWithReminderRulesAndJobs - Begin: %w", err)
	}
	defer rollbackTx(ctx, tx, &err, "ImportantDayRepo - StoreWithReminderRulesAndJobs - Rollback")

	if err = r.storeImportantDayTx(ctx, tx, day); err != nil {
		return err
	}

	if err = replaceReminderRulesTx(ctx, r.Builder, tx, day.UserID, day.ID, rules); err != nil {
		return err
	}

	if err = replacePendingReminderJobsTx(ctx, r.Builder, tx, day.UserID, day.ID, jobs); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("ImportantDayRepo - StoreWithReminderRulesAndJobs - Commit: %w", err)
	}

	return nil
}

// UpdateWithReminderJobs updates an important day and replaces its pending reminder jobs atomically.
func (r *ImportantDayRepo) UpdateWithReminderJobs(ctx context.Context, day *entity.ImportantDay, jobs []entity.ReminderJob) (err error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ImportantDayRepo - UpdateWithReminderJobs - Begin: %w", err)
	}
	defer rollbackTx(ctx, tx, &err, "ImportantDayRepo - UpdateWithReminderJobs - Rollback")

	if err = r.updateImportantDayTx(ctx, tx, day); err != nil {
		return err
	}

	if err = replacePendingReminderJobsTx(ctx, r.Builder, tx, day.UserID, day.ID, jobs); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("ImportantDayRepo - UpdateWithReminderJobs - Commit: %w", err)
	}

	return nil
}

// ReplaceReminderRulesAndJobs replaces reminder rules and pending reminder jobs atomically.
func (r *ImportantDayRepo) ReplaceReminderRulesAndJobs(ctx context.Context, userID, importantDayID string, rules []entity.ReminderRule, jobs []entity.ReminderJob) (err error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ImportantDayRepo - ReplaceReminderRulesAndJobs - Begin: %w", err)
	}
	defer rollbackTx(ctx, tx, &err, "ImportantDayRepo - ReplaceReminderRulesAndJobs - Rollback")

	if err = replaceReminderRulesTx(ctx, r.Builder, tx, userID, importantDayID, rules); err != nil {
		return err
	}

	if err = replacePendingReminderJobsTx(ctx, r.Builder, tx, userID, importantDayID, jobs); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("ImportantDayRepo - ReplaceReminderRulesAndJobs - Commit: %w", err)
	}

	return nil
}

func (r *ImportantDayRepo) storeImportantDayTx(ctx context.Context, tx pgx.Tx, day *entity.ImportantDay) error {
	sql, args, err := r.Builder.
		Insert("important_days").
		Columns(
			"id, user_id, title, type, person_name, relationship, description, event_year, event_month, event_day, recurrence, timezone, reminder_time, created_at, updated_at",
		).
		Values(
			day.ID,
			day.UserID,
			day.Title,
			day.Type,
			day.PersonName,
			day.Relationship,
			day.Description,
			day.EventYear,
			day.EventMonth,
			day.EventDay,
			day.Recurrence,
			day.Timezone,
			day.ReminderTime,
			day.CreatedAt,
			day.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("ImportantDayRepo - storeImportantDayTx - r.Builder: %w", err)
	}

	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("ImportantDayRepo - storeImportantDayTx - tx.Exec: %w", err)
	}

	return nil
}

func (r *ImportantDayRepo) updateImportantDayTx(ctx context.Context, tx pgx.Tx, day *entity.ImportantDay) error {
	sql, args, err := r.Builder.
		Update("important_days").
		Set("title", day.Title).
		Set("type", day.Type).
		Set("person_name", day.PersonName).
		Set("relationship", day.Relationship).
		Set("description", day.Description).
		Set("event_year", day.EventYear).
		Set("event_month", day.EventMonth).
		Set("event_day", day.EventDay).
		Set("recurrence", day.Recurrence).
		Set("timezone", day.Timezone).
		Set("reminder_time", day.ReminderTime).
		Set("updated_at", day.UpdatedAt).
		Where(sq.Eq{"id": day.ID, "user_id": day.UserID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("ImportantDayRepo - updateImportantDayTx - r.Builder: %w", err)
	}

	result, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("ImportantDayRepo - updateImportantDayTx - tx.Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrImportantDayNotFound
	}

	return nil
}

func replaceReminderRulesTx(ctx context.Context, builder sq.StatementBuilderType, tx pgx.Tx, userID, importantDayID string, rules []entity.ReminderRule) error {
	sql, args, err := builder.
		Delete("reminder_rules").
		Where(sq.Eq{"user_id": userID, "important_day_id": importantDayID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("ImportantDayRepo - replaceReminderRulesTx - delete builder: %w", err)
	}

	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("ImportantDayRepo - replaceReminderRulesTx - delete: %w", err)
	}

	for _, rule := range rules {
		channels, marshalErr := marshalChannels(rule.Channels)
		if marshalErr != nil {
			return fmt.Errorf("ImportantDayRepo - replaceReminderRulesTx - marshal: %w", marshalErr)
		}

		sql, args, err = builder.
			Insert("reminder_rules").
			Columns("id, user_id, important_day_id, offset_days, channels, created_at, updated_at").
			Values(rule.ID, rule.UserID, rule.ImportantDayID, rule.OffsetDays, channels, rule.CreatedAt, rule.UpdatedAt).
			ToSql()
		if err != nil {
			return fmt.Errorf("ImportantDayRepo - replaceReminderRulesTx - insert builder: %w", err)
		}

		if _, err = tx.Exec(ctx, sql, args...); err != nil {
			return fmt.Errorf("ImportantDayRepo - replaceReminderRulesTx - insert: %w", err)
		}
	}

	return nil
}

func replacePendingReminderJobsTx(ctx context.Context, builder sq.StatementBuilderType, tx pgx.Tx, userID, importantDayID string, jobs []entity.ReminderJob) error {
	sql, args, err := builder.
		Delete("reminder_jobs").
		Where(sq.Eq{"user_id": userID, "important_day_id": importantDayID, "status": entity.ReminderJobStatusPending}).
		ToSql()
	if err != nil {
		return fmt.Errorf("ImportantDayRepo - replacePendingReminderJobsTx - delete builder: %w", err)
	}

	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("ImportantDayRepo - replacePendingReminderJobsTx - delete: %w", err)
	}

	for _, job := range jobs {
		sql, args, err = builder.
			Insert("reminder_jobs").
			Columns("id, user_id, important_day_id, reminder_rule_id, occurrence_date, offset_days, channel, scheduled_at, status, attempts, last_error, locked_until, sent_at, created_at, updated_at").
			Values(
				job.ID,
				job.UserID,
				job.ImportantDayID,
				job.ReminderRuleID,
				job.OccurrenceDate,
				job.OffsetDays,
				job.Channel,
				job.ScheduledAt,
				job.Status,
				job.Attempts,
				job.LastError,
				job.LockedUntil,
				job.SentAt,
				job.CreatedAt,
				job.UpdatedAt,
			).
			Suffix("ON CONFLICT (important_day_id, occurrence_date, offset_days, channel) DO UPDATE SET scheduled_at = EXCLUDED.scheduled_at, status = EXCLUDED.status, attempts = 0, last_error = '', locked_until = NULL, sent_at = NULL, updated_at = EXCLUDED.updated_at").
			ToSql()
		if err != nil {
			return fmt.Errorf("ImportantDayRepo - replacePendingReminderJobsTx - insert builder: %w", err)
		}

		if _, err = tx.Exec(ctx, sql, args...); err != nil {
			return fmt.Errorf("ImportantDayRepo - replacePendingReminderJobsTx - insert: %w", err)
		}
	}

	return nil
}

func rollbackTx(ctx context.Context, tx pgx.Tx, err *error, message string) {
	if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
		*err = errors.Join(*err, fmt.Errorf("%s: %w", message, rollbackErr))
	}
}

// ReminderRuleRepo -.
type ReminderRuleRepo struct {
	*postgres.Postgres
}

// NewReminderRuleRepo -.
func NewReminderRuleRepo(pg *postgres.Postgres) *ReminderRuleRepo {
	return &ReminderRuleRepo{pg}
}

// ReplaceForImportantDay -.
func (r *ReminderRuleRepo) ReplaceForImportantDay(ctx context.Context, userID, importantDayID string, rules []entity.ReminderRule) (err error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ReminderRuleRepo - ReplaceForImportantDay - Begin: %w", err)
	}
	defer rollbackTx(ctx, tx, &err, "ReminderRuleRepo - ReplaceForImportantDay - Rollback")

	if err = replaceReminderRulesTx(ctx, r.Builder, tx, userID, importantDayID, rules); err != nil {
		return fmt.Errorf("ReminderRuleRepo - ReplaceForImportantDay - replaceReminderRulesTx: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("ReminderRuleRepo - ReplaceForImportantDay - Commit: %w", err)
	}

	return nil
}

// GetForImportantDay -.
func (r *ReminderRuleRepo) GetForImportantDay(ctx context.Context, userID, importantDayID string) ([]entity.ReminderRule, error) {
	sql, args, err := r.Builder.
		Select("id, user_id, important_day_id, offset_days, channels, created_at, updated_at").
		From("reminder_rules").
		Where(sq.Eq{"user_id": userID, "important_day_id": importantDayID}).
		OrderBy("offset_days DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("ReminderRuleRepo - GetForImportantDay - r.Builder: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ReminderRuleRepo - GetForImportantDay - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	rules := make([]entity.ReminderRule, 0)
	for rows.Next() {
		rule, scanErr := scanReminderRule(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("ReminderRuleRepo - GetForImportantDay - rows.Scan: %w", scanErr)
		}

		rules = append(rules, rule)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ReminderRuleRepo - GetForImportantDay - rows.Err: %w", err)
	}

	return rules, nil
}

func importantDayColumns() string {
	return "id, user_id, title, type, person_name, relationship, description, event_year, event_month, event_day, recurrence, timezone, reminder_time, created_at, updated_at"
}

type scanner interface {
	Scan(dest ...any) error
}

func scanImportantDay(row scanner) (entity.ImportantDay, error) {
	var day entity.ImportantDay

	err := row.Scan(
		&day.ID,
		&day.UserID,
		&day.Title,
		&day.Type,
		&day.PersonName,
		&day.Relationship,
		&day.Description,
		&day.EventYear,
		&day.EventMonth,
		&day.EventDay,
		&day.Recurrence,
		&day.Timezone,
		&day.ReminderTime,
		&day.CreatedAt,
		&day.UpdatedAt,
	)

	return day, err
}

func scanReminderRule(row scanner) (entity.ReminderRule, error) {
	var (
		rule        entity.ReminderRule
		channelsRaw []byte
	)

	err := row.Scan(
		&rule.ID,
		&rule.UserID,
		&rule.ImportantDayID,
		&rule.OffsetDays,
		&channelsRaw,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)
	if err != nil {
		return entity.ReminderRule{}, err
	}

	rule.Channels, err = unmarshalChannels(channelsRaw)
	if err != nil {
		return entity.ReminderRule{}, err
	}

	return rule, nil
}

func marshalChannels(channels []entity.ReminderChannel) ([]byte, error) {
	return json.Marshal(channels)
}

func unmarshalChannels(raw []byte) ([]entity.ReminderChannel, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var channels []entity.ReminderChannel
	if err := json.Unmarshal(raw, &channels); err != nil {
		return nil, err
	}

	return channels, nil
}
