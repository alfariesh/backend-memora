package entity

import (
	"time"
)

const (
	DefaultTimezone     = "Asia/Jakarta"
	DefaultReminderTime = "09:00"
)

// ImportantDayType -.
type ImportantDayType string // @name entity.ImportantDayType

const (
	ImportantDayTypeBirthday     ImportantDayType = "birthday"
	ImportantDayTypeWedding      ImportantDayType = "wedding"
	ImportantDayTypeMemorial     ImportantDayType = "memorial"
	ImportantDayTypeGraduation   ImportantDayType = "graduation"
	ImportantDayTypeFirstDay     ImportantDayType = "first_day"
	ImportantDayTypeDocument     ImportantDayType = "document"
	ImportantDayTypeSubscription ImportantDayType = "subscription"
	ImportantDayTypeMedical      ImportantDayType = "medical"
	ImportantDayTypeCustom       ImportantDayType = "custom"
)

// Valid reports whether t is a known important day type.
func (t ImportantDayType) Valid() bool {
	switch t {
	case ImportantDayTypeBirthday,
		ImportantDayTypeWedding,
		ImportantDayTypeMemorial,
		ImportantDayTypeGraduation,
		ImportantDayTypeFirstDay,
		ImportantDayTypeDocument,
		ImportantDayTypeSubscription,
		ImportantDayTypeMedical,
		ImportantDayTypeCustom:
		return true
	default:
		return false
	}
}

// Recurrence -.
type Recurrence string // @name entity.Recurrence

const (
	RecurrenceYearly Recurrence = "yearly"
)

// ReminderChannel -.
type ReminderChannel string // @name entity.ReminderChannel

const (
	ReminderChannelEmail ReminderChannel = "email"
	ReminderChannelInApp ReminderChannel = "in_app"
	ReminderChannelPush  ReminderChannel = "push"
)

var DefaultReminderOffsets = []int{7, 1, 0}

var DefaultReminderChannels = []ReminderChannel{
	ReminderChannelEmail,
	ReminderChannelInApp,
	ReminderChannelPush,
}

// ImportantDay -.
type ImportantDay struct {
	ID           string           `json:"id"             example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID       string           `json:"user_id"        example:"550e8400-e29b-41d4-a716-446655440000"`
	Title        string           `json:"title"          example:"Mom birthday"`
	Type         ImportantDayType `json:"type"           example:"birthday"`
	PersonName   string           `json:"person_name"    example:"Mom"`
	Relationship string           `json:"relationship"   example:"mother"`
	Description  string           `json:"description"    example:"Buy flowers"`
	EventYear    *int             `json:"event_year"     example:"1970"`
	EventMonth   int              `json:"event_month"    example:"5"`
	EventDay     int              `json:"event_day"      example:"13"`
	Recurrence   Recurrence       `json:"recurrence"     example:"yearly"`
	Timezone     string           `json:"timezone"       example:"Asia/Jakarta"`
	ReminderTime string           `json:"reminder_time"  example:"09:00"`
	CreatedAt    time.Time        `json:"created_at"     example:"2026-01-01T00:00:00Z"`
	UpdatedAt    time.Time        `json:"updated_at"     example:"2026-01-01T00:00:00Z"`
} // @name entity.ImportantDay

// ImportantDayParams carries create/update values across transports.
type ImportantDayParams struct {
	Title         string
	Type          ImportantDayType
	PersonName    string
	Relationship  string
	Description   string
	EventYear     *int
	EventMonth    int
	EventDay      int
	Timezone      string
	ReminderTime  string
	ReminderRules []ReminderRuleParams
}

// ImportantDayUpcoming -.
type ImportantDayUpcoming struct {
	ImportantDay
	OccurrenceDate string `json:"occurrence_date" example:"2026-05-13"`
	DaysUntil      int    `json:"days_until"      example:"7"`
	Anniversary    *int   `json:"anniversary"     example:"56"`
} // @name entity.ImportantDayUpcoming

// ReminderRule -.
type ReminderRule struct {
	ID             string            `json:"id"               example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID         string            `json:"user_id"          example:"550e8400-e29b-41d4-a716-446655440000"`
	ImportantDayID string            `json:"important_day_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OffsetDays     int               `json:"offset_days"      example:"7"`
	Channels       []ReminderChannel `json:"channels"         example:"email"`
	CreatedAt      time.Time         `json:"created_at"       example:"2026-01-01T00:00:00Z"`
	UpdatedAt      time.Time         `json:"updated_at"       example:"2026-01-01T00:00:00Z"`
} // @name entity.ReminderRule

// ReminderRuleParams -.
type ReminderRuleParams struct {
	OffsetDays int
	Channels   []ReminderChannel
}

// ReminderJobStatus -.
type ReminderJobStatus string

const (
	ReminderJobStatusPending ReminderJobStatus = "pending"
	ReminderJobStatusSent    ReminderJobStatus = "sent"
	ReminderJobStatusFailed  ReminderJobStatus = "failed"
)

// ReminderJob -.
type ReminderJob struct {
	ID             string            `json:"id"`
	UserID         string            `json:"user_id"`
	ImportantDayID string            `json:"important_day_id"`
	ReminderRuleID *string           `json:"reminder_rule_id"`
	OccurrenceDate time.Time         `json:"occurrence_date"`
	OffsetDays     int               `json:"offset_days"`
	Channels       []ReminderChannel `json:"channels"`
	ScheduledAt    time.Time         `json:"scheduled_at"`
	Status         ReminderJobStatus `json:"status"`
	Attempts       int               `json:"attempts"`
	LastError      string            `json:"last_error"`
	LockedUntil    *time.Time        `json:"locked_until"`
	SentAt         *time.Time        `json:"sent_at"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// Notification -.
type Notification struct {
	ID             string     `json:"id"               example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID         string     `json:"user_id"          example:"550e8400-e29b-41d4-a716-446655440000"`
	ImportantDayID *string    `json:"important_day_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Type           string     `json:"type"             example:"reminder"`
	Title          string     `json:"title"            example:"Mom birthday is coming"`
	Body           string     `json:"body"             example:"Mom birthday is in 7 days."`
	Data           string     `json:"data"             example:"{}"`
	ReadAt         *time.Time `json:"read_at"`
	CreatedAt      time.Time  `json:"created_at"       example:"2026-01-01T00:00:00Z"`
} // @name entity.Notification

// DeviceToken -.
type DeviceToken struct {
	ID        string    `json:"id"         example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID    string    `json:"user_id"    example:"550e8400-e29b-41d4-a716-446655440000"`
	Token     string    `json:"token"      example:"ExpoPushToken[xxxxxxxxxxxxxxxxxxxxxx]"`
	Platform  string    `json:"platform"   example:"android"`
	Name      string    `json:"name"       example:"Pixel 8"`
	Active    bool      `json:"active"     example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2026-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2026-01-01T00:00:00Z"`
} // @name entity.DeviceToken

// NormalizeImportantDay fills defaults and validates date-only fields.
func NormalizeImportantDay(params *ImportantDayParams) error {
	if params.Type == "" {
		params.Type = ImportantDayTypeCustom
	}

	if !params.Type.Valid() {
		return ErrInvalidImportantDayDate
	}

	if params.Timezone == "" {
		params.Timezone = DefaultTimezone
	}

	if params.ReminderTime == "" {
		params.ReminderTime = DefaultReminderTime
	}

	if _, err := time.LoadLocation(params.Timezone); err != nil {
		return ErrInvalidImportantDayDate
	}

	if _, err := time.Parse("15:04", params.ReminderTime); err != nil {
		return ErrInvalidImportantDayDate
	}

	if !validMonthDay(params.EventMonth, params.EventDay) {
		return ErrInvalidImportantDayDate
	}

	return nil
}

// NormalizeReminderRules fills defaults for empty reminder input.
func NormalizeReminderRules(rules []ReminderRuleParams) []ReminderRuleParams {
	if len(rules) == 0 {
		rules = make([]ReminderRuleParams, 0, len(DefaultReminderOffsets))
		for _, offset := range DefaultReminderOffsets {
			rules = append(rules, ReminderRuleParams{
				OffsetDays: offset,
				Channels:   append([]ReminderChannel(nil), DefaultReminderChannels...),
			})
		}

		return rules
	}

	normalized := make([]ReminderRuleParams, 0, len(rules))
	for _, rule := range rules {
		channels := rule.Channels
		if len(channels) == 0 {
			channels = append([]ReminderChannel(nil), DefaultReminderChannels...)
		}

		normalized = append(normalized, ReminderRuleParams{
			OffsetDays: rule.OffsetDays,
			Channels:   channels,
		})
	}

	return normalized
}

// NextOccurrence returns the next occurrence date in the important day's timezone.
func (d ImportantDay) NextOccurrence(from time.Time) (time.Time, error) {
	loc, err := time.LoadLocation(d.Timezone)
	if err != nil {
		return time.Time{}, err
	}

	localFrom := from.In(loc)
	year := localFrom.Year()
	occurrence := occurrenceDate(year, d.EventMonth, d.EventDay, loc)
	today := dateOnly(localFrom, loc)

	if occurrence.Before(today) {
		occurrence = occurrenceDate(year+1, d.EventMonth, d.EventDay, loc)
	}

	return occurrence, nil
}

// AnniversaryFor returns the anniversary count for an occurrence.
func (d ImportantDay) AnniversaryFor(occurrence time.Time) *int {
	if d.EventYear == nil {
		return nil
	}

	value := occurrence.Year() - *d.EventYear
	if value < 0 {
		value = 0
	}

	return &value
}

// ReminderScheduledAt returns the UTC send time for an occurrence and offset.
func (d ImportantDay) ReminderScheduledAt(occurrence time.Time, offsetDays int) (time.Time, error) {
	loc, err := time.LoadLocation(d.Timezone)
	if err != nil {
		return time.Time{}, err
	}

	parsed, err := time.Parse("15:04", d.ReminderTime)
	if err != nil {
		return time.Time{}, err
	}

	local := time.Date(
		occurrence.In(loc).Year(),
		occurrence.In(loc).Month(),
		occurrence.In(loc).Day(),
		parsed.Hour(),
		parsed.Minute(),
		0,
		0,
		loc,
	)

	return local.AddDate(0, 0, -offsetDays).UTC(), nil
}

func validMonthDay(month int, day int) bool {
	if month < 1 || month > 12 || day < 1 {
		return false
	}

	if month == 2 && day == 29 {
		return true
	}

	return !time.Date(2024, time.Month(month), day, 0, 0, 0, 0, time.UTC).IsZero() &&
		time.Date(2024, time.Month(month), day, 0, 0, 0, 0, time.UTC).Month() == time.Month(month)
}

func occurrenceDate(year int, month int, day int, loc *time.Location) time.Time {
	if month == 2 && day == 29 && !isLeapYear(year) {
		day = 28
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc)
}

func dateOnly(t time.Time, loc *time.Location) time.Time {
	local := t.In(loc)

	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
}

func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
