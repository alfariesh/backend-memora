package v1

import (
	"context"
	"errors"
	"time"

	v1 "github.com/evrone/go-clean-template/docs/proto/v1"
	grpcmw "github.com/evrone/go-clean-template/internal/controller/grpc/middleware"
	"github.com/evrone/go-clean-template/internal/controller/grpc/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateImportantDay -.
func (c *ImportantDayController) CreateImportantDay(ctx context.Context, req *v1.CreateImportantDayRequest) (*v1.ImportantDayResponse, error) {
	userID, ok := grpcmw.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if err := validateCreateImportantDayRequest(req); err != nil {
		return nil, err
	}

	day, err := c.id.Create(ctx, userID, createImportantDayParams(req))
	if err != nil {
		return nil, importantDayError(c.l, err, "grpc - v1 - CreateImportantDay")
	}

	return response.NewImportantDayResponse(&day), nil
}

// GetImportantDay -.
func (c *ImportantDayController) GetImportantDay(ctx context.Context, req *v1.GetImportantDayRequest) (*v1.ImportantDayResponse, error) {
	userID, ok := grpcmw.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if err := validateRequiredID(req.GetId()); err != nil {
		return nil, err
	}

	day, err := c.id.Get(ctx, userID, req.GetId())
	if err != nil {
		return nil, importantDayError(c.l, err, "grpc - v1 - GetImportantDay")
	}

	return response.NewImportantDayResponse(&day), nil
}

// ListImportantDays -.
func (c *ImportantDayController) ListImportantDays(ctx context.Context, req *v1.ListImportantDaysRequest) (*v1.ListImportantDaysResponse, error) {
	userID, ok := grpcmw.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	var dayType *entity.ImportantDayType
	if req.GetType() != "" {
		t := entity.ImportantDayType(req.GetType())
		if !t.Valid() {
			return nil, status.Error(codes.InvalidArgument, "invalid important day type")
		}

		dayType = &t
	}

	days, total, err := c.id.List(ctx, userID, dayType, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		c.l.Error(err, "grpc - v1 - ListImportantDays")

		return nil, status.Error(codes.Internal, "internal server error")
	}

	return response.NewListImportantDaysResponse(days, total), nil
}

// UpcomingImportantDays -.
func (c *ImportantDayController) UpcomingImportantDays(ctx context.Context, req *v1.UpcomingImportantDaysRequest) (*v1.UpcomingImportantDaysResponse, error) {
	userID, ok := grpcmw.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	days, total, err := c.id.Upcoming(ctx, userID, time.Now().UTC(), int(req.GetDays()), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		c.l.Error(err, "grpc - v1 - UpcomingImportantDays")

		return nil, status.Error(codes.Internal, "internal server error")
	}

	return response.NewUpcomingImportantDaysResponse(days, total), nil
}

// UpdateImportantDay -.
func (c *ImportantDayController) UpdateImportantDay(ctx context.Context, req *v1.UpdateImportantDayRequest) (*v1.ImportantDayResponse, error) {
	userID, ok := grpcmw.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if err := validateUpdateImportantDayRequest(req); err != nil {
		return nil, err
	}

	day, err := c.id.Update(ctx, userID, req.GetId(), updateImportantDayParams(req))
	if err != nil {
		return nil, importantDayError(c.l, err, "grpc - v1 - UpdateImportantDay")
	}

	return response.NewImportantDayResponse(&day), nil
}

// ReplaceImportantDayReminders -.
func (c *ImportantDayController) ReplaceImportantDayReminders(ctx context.Context, req *v1.ReplaceReminderRulesRequest) (*v1.ReminderRulesResponse, error) {
	userID, ok := grpcmw.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if err := validateReplaceReminderRulesRequest(req); err != nil {
		return nil, err
	}

	rules, err := c.id.ReplaceReminderRules(ctx, userID, req.GetId(), reminderRuleParams(req.GetRules()))
	if err != nil {
		return nil, importantDayError(c.l, err, "grpc - v1 - ReplaceImportantDayReminders")
	}

	return response.NewReminderRulesResponse(rules), nil
}

// DeleteImportantDay -.
func (c *ImportantDayController) DeleteImportantDay(ctx context.Context, req *v1.DeleteImportantDayRequest) (*v1.DeleteImportantDayResponse, error) {
	userID, ok := grpcmw.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if err := validateRequiredID(req.GetId()); err != nil {
		return nil, err
	}

	if err := c.id.Delete(ctx, userID, req.GetId()); err != nil {
		return nil, importantDayError(c.l, err, "grpc - v1 - DeleteImportantDay")
	}

	return &v1.DeleteImportantDayResponse{}, nil
}

// ListNotifications -.
func (c *NotificationController) ListNotifications(ctx context.Context, req *v1.ListNotificationsRequest) (*v1.ListNotificationsResponse, error) {
	userID, ok := grpcmw.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	notifications, total, err := c.n.List(ctx, userID, req.GetUnreadOnly(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		c.l.Error(err, "grpc - v1 - ListNotifications")

		return nil, status.Error(codes.Internal, "internal server error")
	}

	return response.NewListNotificationsResponse(notifications, total), nil
}

// MarkNotificationRead -.
func (c *NotificationController) MarkNotificationRead(ctx context.Context, req *v1.MarkNotificationReadRequest) (*v1.NotificationResponse, error) {
	userID, ok := grpcmw.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if err := validateRequiredID(req.GetId()); err != nil {
		return nil, err
	}

	notification, err := c.n.MarkRead(ctx, userID, req.GetId())
	if err != nil {
		c.l.Error(err, "grpc - v1 - MarkNotificationRead")

		if errors.Is(err, entity.ErrNotificationNotFound) {
			return nil, status.Error(codes.NotFound, "notification not found")
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	return response.NewNotificationResponse(&notification), nil
}

// MarkAllNotificationsRead -.
func (c *NotificationController) MarkAllNotificationsRead(ctx context.Context, _ *v1.MarkAllNotificationsReadRequest) (*v1.MarkAllNotificationsReadResponse, error) {
	userID, ok := grpcmw.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if err := c.n.MarkAllRead(ctx, userID); err != nil {
		c.l.Error(err, "grpc - v1 - MarkAllNotificationsRead")

		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &v1.MarkAllNotificationsReadResponse{}, nil
}

// RegisterDevice -.
func (c *DeviceController) RegisterDevice(ctx context.Context, req *v1.RegisterDeviceRequest) (*v1.DeviceTokenResponse, error) {
	userID, ok := grpcmw.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if err := validateRegisterDeviceRequest(req); err != nil {
		return nil, err
	}

	token, err := c.d.Register(ctx, userID, req.GetToken(), req.GetPlatform(), req.GetName())
	if err != nil {
		c.l.Error(err, "grpc - v1 - RegisterDevice")

		if errors.Is(err, entity.ErrInvalidDeviceToken) {
			return nil, status.Error(codes.InvalidArgument, "invalid device token")
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	return response.NewDeviceTokenResponse(&token), nil
}

// DeleteDevice -.
func (c *DeviceController) DeleteDevice(ctx context.Context, req *v1.DeleteDeviceRequest) (*v1.DeleteDeviceResponse, error) {
	userID, ok := grpcmw.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if err := validateRequiredID(req.GetId()); err != nil {
		return nil, err
	}

	if err := c.d.Delete(ctx, userID, req.GetId()); err != nil {
		c.l.Error(err, "grpc - v1 - DeleteDevice")

		if errors.Is(err, entity.ErrDeviceTokenNotFound) {
			return nil, status.Error(codes.NotFound, "device not found")
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &v1.DeleteDeviceResponse{}, nil
}

func createImportantDayParams(req *v1.CreateImportantDayRequest) entity.ImportantDayParams {
	return entity.ImportantDayParams{
		Title:         req.GetTitle(),
		Type:          entity.ImportantDayType(req.GetType()),
		PersonName:    req.GetPersonName(),
		Relationship:  req.GetRelationship(),
		Description:   req.GetDescription(),
		EventYear:     intPointerFromInt32(req.EventYear),
		EventMonth:    int(req.GetEventMonth()),
		EventDay:      int(req.GetEventDay()),
		Timezone:      req.GetTimezone(),
		ReminderTime:  req.GetReminderTime(),
		ReminderRules: reminderRuleParams(req.GetReminderRules()),
	}
}

func updateImportantDayParams(req *v1.UpdateImportantDayRequest) entity.ImportantDayParams {
	return entity.ImportantDayParams{
		Title:        req.GetTitle(),
		Type:         entity.ImportantDayType(req.GetType()),
		PersonName:   req.GetPersonName(),
		Relationship: req.GetRelationship(),
		Description:  req.GetDescription(),
		EventYear:    intPointerFromInt32(req.EventYear),
		EventMonth:   int(req.GetEventMonth()),
		EventDay:     int(req.GetEventDay()),
		Timezone:     req.GetTimezone(),
		ReminderTime: req.GetReminderTime(),
	}
}

func reminderRuleParams(rules []*v1.ReminderRuleRequest) []entity.ReminderRuleParams {
	params := make([]entity.ReminderRuleParams, 0, len(rules))
	for _, rule := range rules {
		channels := make([]entity.ReminderChannel, len(rule.GetChannels()))
		for i, channel := range rule.GetChannels() {
			channels[i] = entity.ReminderChannel(channel)
		}

		params = append(params, entity.ReminderRuleParams{
			OffsetDays: int(rule.GetOffsetDays()),
			Channels:   channels,
		})
	}

	return params
}

func intPointerFromInt32(value *int32) *int {
	if value == nil {
		return nil
	}

	result := int(*value)

	return &result
}

const (
	maxImportantDayTitleLen        = 255
	maxImportantDayPersonNameLen   = 255
	maxImportantDayRelationshipLen = 100
	maxImportantDayDescriptionLen  = 1000
	maxImportantDayTimezoneLen     = 64
	maxDevicePlatformLen           = 40
	maxDeviceNameLen               = 255
)

func validateCreateImportantDayRequest(req *v1.CreateImportantDayRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is required")
	}

	if err := validateImportantDayFields(
		req.GetTitle(),
		req.GetType(),
		req.GetPersonName(),
		req.GetRelationship(),
		req.GetDescription(),
		req.EventYear,
		req.GetEventMonth(),
		req.GetEventDay(),
		req.GetTimezone(),
		req.GetReminderTime(),
	); err != nil {
		return err
	}

	return validateReminderRuleRequests(req.GetReminderRules())
}

func validateUpdateImportantDayRequest(req *v1.UpdateImportantDayRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is required")
	}

	if err := validateRequiredID(req.GetId()); err != nil {
		return err
	}

	return validateImportantDayFields(
		req.GetTitle(),
		req.GetType(),
		req.GetPersonName(),
		req.GetRelationship(),
		req.GetDescription(),
		req.EventYear,
		req.GetEventMonth(),
		req.GetEventDay(),
		req.GetTimezone(),
		req.GetReminderTime(),
	)
}

func validateImportantDayFields(
	title string,
	dayType string,
	personName string,
	relationship string,
	description string,
	eventYear *int32,
	eventMonth int32,
	eventDay int32,
	timezone string,
	reminderTime string,
) error {
	if title == "" {
		return status.Error(codes.InvalidArgument, "title is required")
	}

	if len(title) > maxImportantDayTitleLen ||
		len(personName) > maxImportantDayPersonNameLen ||
		len(relationship) > maxImportantDayRelationshipLen ||
		len(description) > maxImportantDayDescriptionLen ||
		len(timezone) > maxImportantDayTimezoneLen {
		return status.Error(codes.InvalidArgument, "request field exceeds maximum length")
	}

	if dayType != "" && !entity.ImportantDayType(dayType).Valid() {
		return status.Error(codes.InvalidArgument, "invalid important day type")
	}

	if eventYear != nil && *eventYear < 1 {
		return status.Error(codes.InvalidArgument, "invalid event year")
	}

	if eventMonth < 1 || eventMonth > 12 {
		return status.Error(codes.InvalidArgument, "invalid event month")
	}

	if eventDay < 1 || eventDay > 31 {
		return status.Error(codes.InvalidArgument, "invalid event day")
	}

	if reminderTime != "" {
		if _, err := time.Parse("15:04", reminderTime); err != nil {
			return status.Error(codes.InvalidArgument, "invalid reminder time")
		}
	}

	return nil
}

func validateReplaceReminderRulesRequest(req *v1.ReplaceReminderRulesRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is required")
	}

	if err := validateRequiredID(req.GetId()); err != nil {
		return err
	}

	return validateReminderRuleRequests(req.GetRules())
}

func validateReminderRuleRequests(rules []*v1.ReminderRuleRequest) error {
	for _, rule := range rules {
		if rule.GetOffsetDays() < 0 {
			return status.Error(codes.InvalidArgument, "invalid reminder offset")
		}

		for _, channel := range rule.GetChannels() {
			if !entity.ReminderChannel(channel).Valid() {
				return status.Error(codes.InvalidArgument, "invalid reminder channel")
			}
		}
	}

	return nil
}

func validateRegisterDeviceRequest(req *v1.RegisterDeviceRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is required")
	}

	if req.GetToken() == "" {
		return status.Error(codes.InvalidArgument, "token is required")
	}

	if req.GetPlatform() == "" {
		return status.Error(codes.InvalidArgument, "platform is required")
	}

	if len(req.GetPlatform()) > maxDevicePlatformLen || len(req.GetName()) > maxDeviceNameLen {
		return status.Error(codes.InvalidArgument, "request field exceeds maximum length")
	}

	return nil
}

func validateRequiredID(id string) error {
	if id == "" {
		return status.Error(codes.InvalidArgument, "id is required")
	}

	return nil
}

func importantDayError(l interface {
	Error(message any, args ...any)
}, err error, message string) error {
	l.Error(err, message)

	if errors.Is(err, entity.ErrImportantDayNotFound) {
		return status.Error(codes.NotFound, "important day not found")
	}

	if errors.Is(err, entity.ErrImportantDayForbidden) {
		return status.Error(codes.PermissionDenied, "forbidden")
	}

	if errors.Is(err, entity.ErrInvalidImportantDayDate) {
		return status.Error(codes.InvalidArgument, "invalid important day date")
	}

	return status.Error(codes.Internal, "internal server error")
}
