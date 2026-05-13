package v1

import (
	v1 "github.com/evrone/go-clean-template/docs/proto/v1"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/go-playground/validator/v10"
)

// ImportantDayController -.
type ImportantDayController struct {
	v1.UnimplementedImportantDayServiceServer

	id usecase.ImportantDay
	l  logger.Interface
	v  *validator.Validate
}

// NotificationController -.
type NotificationController struct {
	v1.UnimplementedNotificationServiceServer

	n usecase.Notification
	l logger.Interface
	v *validator.Validate
}

// DeviceController -.
type DeviceController struct {
	v1.UnimplementedDeviceServiceServer

	d usecase.DeviceToken
	l logger.Interface
	v *validator.Validate
}
