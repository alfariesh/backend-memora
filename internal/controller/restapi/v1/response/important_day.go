package response

import "github.com/evrone/go-clean-template/internal/entity"

// ImportantDayList -.
type ImportantDayList struct {
	ImportantDays []entity.ImportantDay `json:"important_days"`
	Total         int                   `json:"total" example:"42"`
} // @name v1.ImportantDayList

// UpcomingImportantDayList -.
type UpcomingImportantDayList struct {
	ImportantDays []entity.ImportantDayUpcoming `json:"important_days"`
	Total         int                           `json:"total" example:"42"`
} // @name v1.UpcomingImportantDayList

// ReminderRuleList -.
type ReminderRuleList struct {
	Rules []entity.ReminderRule `json:"rules"`
} // @name v1.ReminderRuleList

// NotificationList -.
type NotificationList struct {
	Notifications []entity.Notification `json:"notifications"`
	Total         int                   `json:"total" example:"42"`
} // @name v1.NotificationList

// UnreadNotificationCount -.
type UnreadNotificationCount struct {
	UnreadCount int `json:"unread_count" example:"3"`
} // @name v1.UnreadNotificationCount
