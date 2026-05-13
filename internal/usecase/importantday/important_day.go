package importantday

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/google/uuid"
)

// UseCase -.
type UseCase struct {
	dayRepo  repo.ImportantDayRepo
	ruleRepo repo.ReminderRuleRepo
	jobRepo  repo.ReminderJobRepo
}

// New -.
func New(dayRepo repo.ImportantDayRepo, ruleRepo repo.ReminderRuleRepo, jobRepo repo.ReminderJobRepo) *UseCase {
	return &UseCase{
		dayRepo:  dayRepo,
		ruleRepo: ruleRepo,
		jobRepo:  jobRepo,
	}
}

// Create -.
func (uc *UseCase) Create(ctx context.Context, userID string, params entity.ImportantDayParams) (entity.ImportantDay, error) {
	if err := entity.NormalizeImportantDay(&params); err != nil {
		return entity.ImportantDay{}, err
	}

	now := time.Now().UTC()
	day := entity.ImportantDay{
		ID:           uuid.New().String(),
		UserID:       userID,
		Title:        params.Title,
		Type:         params.Type,
		PersonName:   params.PersonName,
		Relationship: params.Relationship,
		Description:  params.Description,
		EventYear:    params.EventYear,
		EventMonth:   params.EventMonth,
		EventDay:     params.EventDay,
		Recurrence:   entity.RecurrenceYearly,
		Timezone:     params.Timezone,
		ReminderTime: params.ReminderTime,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.dayRepo.Store(ctx, &day); err != nil {
		return entity.ImportantDay{}, fmt.Errorf("ImportantDayUseCase - Create - uc.dayRepo.Store: %w", err)
	}

	rules := buildReminderRules(userID, day.ID, entity.NormalizeReminderRules(params.ReminderRules), now)
	if err := uc.ruleRepo.ReplaceForImportantDay(ctx, userID, day.ID, rules); err != nil {
		_ = uc.dayRepo.Delete(ctx, userID, day.ID)

		return entity.ImportantDay{}, fmt.Errorf("ImportantDayUseCase - Create - uc.ruleRepo.ReplaceForImportantDay: %w", err)
	}

	if err := uc.scheduleJobs(ctx, day, rules, now); err != nil {
		return entity.ImportantDay{}, fmt.Errorf("ImportantDayUseCase - Create - uc.scheduleJobs: %w", err)
	}

	return day, nil
}

// Get -.
func (uc *UseCase) Get(ctx context.Context, userID, id string) (entity.ImportantDay, error) {
	day, err := uc.dayRepo.GetByID(ctx, userID, id)
	if err != nil {
		return entity.ImportantDay{}, fmt.Errorf("ImportantDayUseCase - Get - uc.dayRepo.GetByID: %w", err)
	}

	return day, nil
}

// List -.
func (uc *UseCase) List(ctx context.Context, userID string, dayType *entity.ImportantDayType, limit, offset int) ([]entity.ImportantDay, int, error) {
	if limit <= 0 {
		limit = 10
	}

	if offset < 0 {
		offset = 0
	}

	days, total, err := uc.dayRepo.List(ctx, userID, repo.ImportantDayFilter{
		Type:   dayType,
		Limit:  uint64(limit),
		Offset: uint64(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("ImportantDayUseCase - List - uc.dayRepo.List: %w", err)
	}

	return days, total, nil
}

// Upcoming -.
func (uc *UseCase) Upcoming(ctx context.Context, userID string, from time.Time, days, limit, offset int) ([]entity.ImportantDayUpcoming, int, error) {
	if days <= 0 {
		days = 365
	}

	if limit <= 0 {
		limit = 10
	}

	if offset < 0 {
		offset = 0
	}

	allDays, _, err := uc.dayRepo.List(ctx, userID, repo.ImportantDayFilter{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("ImportantDayUseCase - Upcoming - uc.dayRepo.List: %w", err)
	}

	until := from.AddDate(0, 0, days)
	upcoming := make([]entity.ImportantDayUpcoming, 0, len(allDays))

	for _, day := range allDays {
		occurrence, occErr := day.NextOccurrence(from)
		if occErr != nil {
			return nil, 0, fmt.Errorf("ImportantDayUseCase - Upcoming - day.NextOccurrence: %w", occErr)
		}

		if occurrence.After(until) {
			continue
		}

		daysUntil := int(occurrence.Sub(dateOnlyInLocation(from, occurrence.Location())).Hours() / 24)
		upcoming = append(upcoming, entity.ImportantDayUpcoming{
			ImportantDay:   day,
			OccurrenceDate: occurrence.Format("2006-01-02"),
			DaysUntil:      daysUntil,
			Anniversary:    day.AnniversaryFor(occurrence),
		})
	}

	sort.SliceStable(upcoming, func(i, j int) bool {
		return upcoming[i].OccurrenceDate < upcoming[j].OccurrenceDate
	})

	total := len(upcoming)
	if offset >= total {
		return []entity.ImportantDayUpcoming{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return upcoming[offset:end], total, nil
}

// Update -.
func (uc *UseCase) Update(ctx context.Context, userID, id string, params entity.ImportantDayParams) (entity.ImportantDay, error) {
	if err := entity.NormalizeImportantDay(&params); err != nil {
		return entity.ImportantDay{}, err
	}

	day, err := uc.dayRepo.GetByID(ctx, userID, id)
	if err != nil {
		return entity.ImportantDay{}, fmt.Errorf("ImportantDayUseCase - Update - uc.dayRepo.GetByID: %w", err)
	}

	day.Title = params.Title
	day.Type = params.Type
	day.PersonName = params.PersonName
	day.Relationship = params.Relationship
	day.Description = params.Description
	day.EventYear = params.EventYear
	day.EventMonth = params.EventMonth
	day.EventDay = params.EventDay
	day.Timezone = params.Timezone
	day.ReminderTime = params.ReminderTime
	day.UpdatedAt = time.Now().UTC()

	if err = uc.dayRepo.Update(ctx, &day); err != nil {
		return entity.ImportantDay{}, fmt.Errorf("ImportantDayUseCase - Update - uc.dayRepo.Update: %w", err)
	}

	rules, err := uc.ruleRepo.GetForImportantDay(ctx, userID, id)
	if err != nil {
		return entity.ImportantDay{}, fmt.Errorf("ImportantDayUseCase - Update - uc.ruleRepo.GetForImportantDay: %w", err)
	}

	if len(rules) > 0 {
		if err = uc.scheduleJobs(ctx, day, rules, time.Now().UTC()); err != nil {
			return entity.ImportantDay{}, fmt.Errorf("ImportantDayUseCase - Update - uc.scheduleJobs: %w", err)
		}
	}

	return day, nil
}

// Delete -.
func (uc *UseCase) Delete(ctx context.Context, userID, id string) error {
	err := uc.dayRepo.Delete(ctx, userID, id)
	if err != nil {
		return fmt.Errorf("ImportantDayUseCase - Delete - uc.dayRepo.Delete: %w", err)
	}

	return nil
}

// ReplaceReminderRules -.
func (uc *UseCase) ReplaceReminderRules(ctx context.Context, userID, id string, params []entity.ReminderRuleParams) ([]entity.ReminderRule, error) {
	day, err := uc.dayRepo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, fmt.Errorf("ImportantDayUseCase - ReplaceReminderRules - uc.dayRepo.GetByID: %w", err)
	}

	now := time.Now().UTC()
	rules := buildReminderRules(userID, id, entity.NormalizeReminderRules(params), now)

	if err = uc.ruleRepo.ReplaceForImportantDay(ctx, userID, id, rules); err != nil {
		return nil, fmt.Errorf("ImportantDayUseCase - ReplaceReminderRules - uc.ruleRepo.ReplaceForImportantDay: %w", err)
	}

	if err = uc.scheduleJobs(ctx, day, rules, now); err != nil {
		return nil, fmt.Errorf("ImportantDayUseCase - ReplaceReminderRules - uc.scheduleJobs: %w", err)
	}

	return rules, nil
}

func (uc *UseCase) scheduleJobs(ctx context.Context, day entity.ImportantDay, rules []entity.ReminderRule, now time.Time) error {
	occurrence, err := day.NextOccurrence(now)
	if err != nil {
		return err
	}

	jobs := make([]entity.ReminderJob, 0, len(rules))
	for _, rule := range rules {
		scheduledAt, scheduleErr := day.ReminderScheduledAt(occurrence, rule.OffsetDays)
		if scheduleErr != nil {
			return scheduleErr
		}

		if scheduledAt.Before(now) {
			scheduledAt = now
		}

		ruleID := rule.ID
		jobs = append(jobs, entity.ReminderJob{
			ID:             uuid.New().String(),
			UserID:         day.UserID,
			ImportantDayID: day.ID,
			ReminderRuleID: &ruleID,
			OccurrenceDate: occurrence,
			OffsetDays:     rule.OffsetDays,
			Channels:       rule.Channels,
			ScheduledAt:    scheduledAt,
			Status:         entity.ReminderJobStatusPending,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}

	return uc.jobRepo.ReplacePendingForImportantDay(ctx, day.UserID, day.ID, jobs)
}

func buildReminderRules(userID, importantDayID string, params []entity.ReminderRuleParams, now time.Time) []entity.ReminderRule {
	rules := make([]entity.ReminderRule, 0, len(params))
	for _, param := range params {
		rules = append(rules, entity.ReminderRule{
			ID:             uuid.New().String(),
			UserID:         userID,
			ImportantDayID: importantDayID,
			OffsetDays:     param.OffsetDays,
			Channels:       param.Channels,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}

	return rules
}

func dateOnlyInLocation(t time.Time, loc *time.Location) time.Time {
	local := t.In(loc)

	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
}
