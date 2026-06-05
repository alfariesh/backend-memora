package response

import "github.com/alfariesh/backend-memora/internal/entity"

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

// DeviceTokenList -.
type DeviceTokenList struct {
	Devices []entity.DeviceToken `json:"devices"`
	Total   int                  `json:"total" example:"2"`
} // @name v1.DeviceTokenList

// NotificationList -.
type NotificationList struct {
	Notifications []entity.Notification `json:"notifications"`
	Total         int                   `json:"total" example:"42"`
} // @name v1.NotificationList

// UnreadNotificationCount -.
type UnreadNotificationCount struct {
	UnreadCount int `json:"unread_count" example:"3"`
} // @name v1.UnreadNotificationCount

// MobileBootstrap -.
type MobileBootstrap struct {
	Settings                entity.UserSettings           `json:"settings"`
	UpcomingImportantDays   []entity.ImportantDayUpcoming `json:"upcoming_important_days"`
	UpcomingTotal           int                           `json:"upcoming_total"             example:"3"`
	UnreadNotificationCount int                           `json:"unread_notification_count"  example:"2"`
	Devices                 []entity.DeviceToken          `json:"devices"`
	DevicesTotal            int                           `json:"devices_total"              example:"1"`
} // @name v1.MobileBootstrap

// UserSessionList -.
type UserSessionList struct {
	Sessions []entity.UserSessionView `json:"sessions"`
	Total    int                      `json:"total" example:"2"`
} // @name v1.UserSessionList
