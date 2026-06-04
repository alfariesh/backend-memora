package response

import "github.com/alfariesh/backend-memora/internal/entity"

// ImportantDayList -.
type ImportantDayList struct {
	ImportantDays []entity.ImportantDay `json:"important_days"`
	Total         int                   `json:"total"`
}

// UpcomingImportantDayList -.
type UpcomingImportantDayList struct {
	ImportantDays []entity.ImportantDayUpcoming `json:"important_days"`
	Total         int                           `json:"total"`
}

// ReminderRuleList -.
type ReminderRuleList struct {
	Rules []entity.ReminderRule `json:"rules"`
}

// NotificationList -.
type NotificationList struct {
	Notifications []entity.Notification `json:"notifications"`
	Total         int                   `json:"total"`
}
