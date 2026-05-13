package request

import "github.com/evrone/go-clean-template/internal/entity"

// ReminderRule -.
type ReminderRule struct {
	OffsetDays int                      `json:"offset_days" validate:"min=0"`
	Channels   []entity.ReminderChannel `json:"channels"    validate:"omitempty,dive,oneof=email in_app push"`
}

// CreateImportantDay -.
type CreateImportantDay struct {
	Title         string         `json:"title"          validate:"required,max=255"`
	Type          string         `json:"type"           validate:"omitempty,oneof=birthday wedding memorial graduation first_day document subscription medical custom"`
	PersonName    string         `json:"person_name"    validate:"max=255"`
	Relationship  string         `json:"relationship"   validate:"max=100"`
	Description   string         `json:"description"    validate:"max=1000"`
	EventYear     *int           `json:"event_year"     validate:"omitempty,min=1"`
	EventMonth    int            `json:"event_month"    validate:"required,min=1,max=12"`
	EventDay      int            `json:"event_day"      validate:"required,min=1,max=31"`
	Timezone      string         `json:"timezone"       validate:"omitempty,max=64"`
	ReminderTime  string         `json:"reminder_time"  validate:"omitempty,datetime=15:04"`
	ReminderRules []ReminderRule `json:"reminder_rules" validate:"omitempty,dive"`
}

// UpdateImportantDay -.
type UpdateImportantDay struct {
	ID           string `json:"id"            validate:"required"`
	Title        string `json:"title"         validate:"required,max=255"`
	Type         string `json:"type"          validate:"omitempty,oneof=birthday wedding memorial graduation first_day document subscription medical custom"`
	PersonName   string `json:"person_name"   validate:"max=255"`
	Relationship string `json:"relationship"  validate:"max=100"`
	Description  string `json:"description"   validate:"max=1000"`
	EventYear    *int   `json:"event_year"    validate:"omitempty,min=1"`
	EventMonth   int    `json:"event_month"   validate:"required,min=1,max=12"`
	EventDay     int    `json:"event_day"     validate:"required,min=1,max=31"`
	Timezone     string `json:"timezone"      validate:"omitempty,max=64"`
	ReminderTime string `json:"reminder_time" validate:"omitempty,datetime=15:04"`
}

// GetImportantDay -.
type GetImportantDay struct {
	ID string `json:"id" validate:"required"`
}

// ListImportantDays -.
type ListImportantDays struct {
	Type   string `json:"type" validate:"omitempty,oneof=birthday wedding memorial graduation first_day document subscription medical custom"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// UpcomingImportantDays -.
type UpcomingImportantDays struct {
	Days   int `json:"days"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// ReplaceReminderRules -.
type ReplaceReminderRules struct {
	ID    string         `json:"id"    validate:"required"`
	Rules []ReminderRule `json:"rules" validate:"omitempty,dive"`
}

// DeleteImportantDay -.
type DeleteImportantDay struct {
	ID string `json:"id" validate:"required"`
}

// ListNotifications -.
type ListNotifications struct {
	UnreadOnly bool `json:"unread_only"`
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
}

// MarkNotificationRead -.
type MarkNotificationRead struct {
	ID string `json:"id" validate:"required"`
}

// RegisterDevice -.
type RegisterDevice struct {
	Token    string `json:"token"    validate:"required"`
	Platform string `json:"platform" validate:"required,max=40"`
	Name     string `json:"name"     validate:"omitempty,max=255"`
}

// DeleteDevice -.
type DeleteDevice struct {
	ID string `json:"id" validate:"required"`
}

func (r CreateImportantDay) ToParams() entity.ImportantDayParams {
	return entity.ImportantDayParams{
		Title:         r.Title,
		Type:          entity.ImportantDayType(r.Type),
		PersonName:    r.PersonName,
		Relationship:  r.Relationship,
		Description:   r.Description,
		EventYear:     r.EventYear,
		EventMonth:    r.EventMonth,
		EventDay:      r.EventDay,
		Timezone:      r.Timezone,
		ReminderTime:  r.ReminderTime,
		ReminderRules: reminderRulesToParams(r.ReminderRules),
	}
}

func (r UpdateImportantDay) ToParams() entity.ImportantDayParams {
	return entity.ImportantDayParams{
		Title:        r.Title,
		Type:         entity.ImportantDayType(r.Type),
		PersonName:   r.PersonName,
		Relationship: r.Relationship,
		Description:  r.Description,
		EventYear:    r.EventYear,
		EventMonth:   r.EventMonth,
		EventDay:     r.EventDay,
		Timezone:     r.Timezone,
		ReminderTime: r.ReminderTime,
	}
}

func (r ReplaceReminderRules) ToParams() []entity.ReminderRuleParams {
	return reminderRulesToParams(r.Rules)
}

func reminderRulesToParams(rules []ReminderRule) []entity.ReminderRuleParams {
	params := make([]entity.ReminderRuleParams, 0, len(rules))
	for _, rule := range rules {
		params = append(params, entity.ReminderRuleParams{
			OffsetDays: rule.OffsetDays,
			Channels:   rule.Channels,
		})
	}

	return params
}
