package request

import "github.com/evrone/go-clean-template/internal/entity"

// ReminderRule -.
type ReminderRule struct {
	OffsetDays int                      `json:"offset_days" validate:"min=0" example:"7"`
	Channels   []entity.ReminderChannel `json:"channels"    validate:"omitempty,dive,oneof=email in_app push"`
} // @name v1.ReminderRuleRequest

// CreateImportantDay -.
type CreateImportantDay struct {
	Title         string         `json:"title"          validate:"required,max=255"                                example:"Mom birthday"`
	Type          string         `json:"type"           validate:"omitempty,oneof=birthday wedding memorial graduation first_day document subscription medical custom" example:"birthday"`
	PersonName    string         `json:"person_name"    validate:"max=255"                                         example:"Mom"`
	Relationship  string         `json:"relationship"   validate:"max=100"                                         example:"mother"`
	Description   string         `json:"description"    validate:"max=1000"                                        example:"Buy flowers"`
	EventYear     *int           `json:"event_year"     validate:"omitempty,min=1"                                 example:"1970"`
	EventMonth    int            `json:"event_month"    validate:"required,min=1,max=12"                           example:"5"`
	EventDay      int            `json:"event_day"      validate:"required,min=1,max=31"                           example:"13"`
	Timezone      string         `json:"timezone"       validate:"omitempty,max=64"                                example:"Asia/Jakarta"`
	ReminderTime  string         `json:"reminder_time"  validate:"omitempty,datetime=15:04"                        example:"09:00"`
	ReminderRules []ReminderRule `json:"reminder_rules" validate:"omitempty,dive"`
} // @name v1.CreateImportantDay

// UpdateImportantDay -.
type UpdateImportantDay struct {
	Title        string `json:"title"         validate:"required,max=255"                                example:"Mom birthday"`
	Type         string `json:"type"          validate:"omitempty,oneof=birthday wedding memorial graduation first_day document subscription medical custom" example:"birthday"`
	PersonName   string `json:"person_name"   validate:"max=255"                                         example:"Mom"`
	Relationship string `json:"relationship"  validate:"max=100"                                         example:"mother"`
	Description  string `json:"description"   validate:"max=1000"                                        example:"Buy flowers"`
	EventYear    *int   `json:"event_year"    validate:"omitempty,min=1"                                 example:"1970"`
	EventMonth   int    `json:"event_month"   validate:"required,min=1,max=12"                           example:"5"`
	EventDay     int    `json:"event_day"     validate:"required,min=1,max=31"                           example:"13"`
	Timezone     string `json:"timezone"      validate:"omitempty,max=64"                                example:"Asia/Jakarta"`
	ReminderTime string `json:"reminder_time" validate:"omitempty,datetime=15:04"                        example:"09:00"`
} // @name v1.UpdateImportantDay

// ReplaceReminderRules -.
type ReplaceReminderRules struct {
	Rules []ReminderRule `json:"rules" validate:"omitempty,dive"`
} // @name v1.ReplaceReminderRules

// RegisterDevice -.
type RegisterDevice struct {
	Token    string `json:"token"    validate:"required"              example:"ExpoPushToken[xxxxxxxxxxxxxxxxxxxxxx]"`
	Platform string `json:"platform" validate:"required,max=40"       example:"android"`
	Name     string `json:"name"     validate:"omitempty,max=255"     example:"Pixel 8"`
} // @name v1.RegisterDevice

// TestPush -.
type TestPush struct {
	Title string `json:"title" validate:"omitempty,max=100" example:"Memora test"`
	Body  string `json:"body"  validate:"omitempty,max=255" example:"Push notifications are working."`
} // @name v1.TestPush

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
