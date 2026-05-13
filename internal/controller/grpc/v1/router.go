package v1

import (
	v1 "github.com/evrone/go-clean-template/docs/proto/v1"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/go-playground/validator/v10"
	pbgrpc "google.golang.org/grpc"
)

// NewTranslationRoutes -.
func NewTranslationRoutes(app *pbgrpc.Server, t usecase.Translation, l logger.Interface) {
	r := &TranslationController{t: t, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	v1.RegisterTranslationServer(app, r)
}

// NewAuthRoutes -.
func NewAuthRoutes(app *pbgrpc.Server, u usecase.User, l logger.Interface) {
	r := &AuthController{u: u, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	v1.RegisterAuthServiceServer(app, r)
}

// NewTaskRoutes -.
func NewTaskRoutes(app *pbgrpc.Server, tk usecase.Task, l logger.Interface) {
	r := &TaskController{tk: tk, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	v1.RegisterTaskServiceServer(app, r)
}

// NewImportantDayRoutes -.
func NewImportantDayRoutes(app *pbgrpc.Server, id usecase.ImportantDay, l logger.Interface) {
	r := &ImportantDayController{id: id, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	v1.RegisterImportantDayServiceServer(app, r)
}

// NewNotificationRoutes -.
func NewNotificationRoutes(app *pbgrpc.Server, n usecase.Notification, l logger.Interface) {
	r := &NotificationController{n: n, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	v1.RegisterNotificationServiceServer(app, r)
}

// NewDeviceRoutes -.
func NewDeviceRoutes(app *pbgrpc.Server, d usecase.DeviceToken, l logger.Interface) {
	r := &DeviceController{d: d, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	v1.RegisterDeviceServiceServer(app, r)
}
