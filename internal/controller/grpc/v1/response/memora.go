package response

import (
	"math"
	"time"

	v1 "github.com/alfariesh/backend-memora/docs/proto/v1"
	"github.com/alfariesh/backend-memora/internal/entity"
)

// NewImportantDayResponse -.
func NewImportantDayResponse(day *entity.ImportantDay) *v1.ImportantDayResponse {
	var eventYear *int32
	if day.EventYear != nil {
		value := safeInt32(*day.EventYear)
		eventYear = &value
	}

	return &v1.ImportantDayResponse{
		Id:           day.ID,
		UserId:       day.UserID,
		Title:        day.Title,
		Type:         string(day.Type),
		PersonName:   day.PersonName,
		Relationship: day.Relationship,
		Description:  day.Description,
		EventYear:    eventYear,
		EventMonth:   int32(day.EventMonth),
		EventDay:     int32(day.EventDay),
		Recurrence:   string(day.Recurrence),
		Timezone:     day.Timezone,
		ReminderTime: day.ReminderTime,
		CreatedAt:    formatTime(day.CreatedAt),
		UpdatedAt:    formatTime(day.UpdatedAt),
	}
}

// NewListImportantDaysResponse -.
func NewListImportantDaysResponse(days []entity.ImportantDay, total int) *v1.ListImportantDaysResponse {
	pbDays := make([]*v1.ImportantDayResponse, len(days))
	for i := range days {
		pbDays[i] = NewImportantDayResponse(&days[i])
	}

	return &v1.ListImportantDaysResponse{
		ImportantDays: pbDays,
		Total:         safeInt32(total),
	}
}

// NewUpcomingImportantDaysResponse -.
func NewUpcomingImportantDaysResponse(days []entity.ImportantDayUpcoming, total int) *v1.UpcomingImportantDaysResponse {
	pbDays := make([]*v1.UpcomingImportantDayResponse, len(days))
	for i := range days {
		var anniversary *int32
		if days[i].Anniversary != nil {
			value := safeInt32(*days[i].Anniversary)
			anniversary = &value
		}

		day := days[i].ImportantDay
		pbDays[i] = &v1.UpcomingImportantDayResponse{
			ImportantDay:   NewImportantDayResponse(&day),
			OccurrenceDate: days[i].OccurrenceDate,
			DaysUntil:      int32(days[i].DaysUntil),
			Anniversary:    anniversary,
		}
	}

	return &v1.UpcomingImportantDaysResponse{
		ImportantDays: pbDays,
		Total:         safeInt32(total),
	}
}

// NewReminderRulesResponse -.
func NewReminderRulesResponse(rules []entity.ReminderRule) *v1.ReminderRulesResponse {
	pbRules := make([]*v1.ReminderRuleResponse, len(rules))
	for i := range rules {
		channels := make([]string, len(rules[i].Channels))
		for j := range rules[i].Channels {
			channels[j] = string(rules[i].Channels[j])
		}

		pbRules[i] = &v1.ReminderRuleResponse{
			Id:             rules[i].ID,
			UserId:         rules[i].UserID,
			ImportantDayId: rules[i].ImportantDayID,
			OffsetDays:     int32(rules[i].OffsetDays),
			Channels:       channels,
			CreatedAt:      formatTime(rules[i].CreatedAt),
			UpdatedAt:      formatTime(rules[i].UpdatedAt),
		}
	}

	return &v1.ReminderRulesResponse{Rules: pbRules}
}

// NewNotificationResponse -.
func NewNotificationResponse(notification *entity.Notification) *v1.NotificationResponse {
	var readAt *string
	if notification.ReadAt != nil {
		value := formatTime(*notification.ReadAt)
		readAt = &value
	}

	return &v1.NotificationResponse{
		Id:             notification.ID,
		UserId:         notification.UserID,
		ImportantDayId: notification.ImportantDayID,
		Type:           notification.Type,
		Title:          notification.Title,
		Body:           notification.Body,
		Data:           notification.Data,
		ReadAt:         readAt,
		CreatedAt:      formatTime(notification.CreatedAt),
	}
}

// NewListNotificationsResponse -.
func NewListNotificationsResponse(notifications []entity.Notification, total int) *v1.ListNotificationsResponse {
	pbNotifications := make([]*v1.NotificationResponse, len(notifications))
	for i := range notifications {
		pbNotifications[i] = NewNotificationResponse(&notifications[i])
	}

	return &v1.ListNotificationsResponse{
		Notifications: pbNotifications,
		Total:         safeInt32(total),
	}
}

// NewDeviceTokenResponse -.
func NewDeviceTokenResponse(token *entity.DeviceToken) *v1.DeviceTokenResponse {
	return &v1.DeviceTokenResponse{
		Id:        token.ID,
		UserId:    token.UserID,
		Token:     token.Token,
		Platform:  token.Platform,
		Name:      token.Name,
		Active:    token.Active,
		CreatedAt: formatTime(token.CreatedAt),
		UpdatedAt: formatTime(token.UpdatedAt),
	}
}

func safeInt32(value int) int32 {
	if value > math.MaxInt32 {
		return math.MaxInt32
	}

	if value < math.MinInt32 {
		return math.MinInt32
	}

	return int32(value)
}

func formatTime(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05Z")
}
