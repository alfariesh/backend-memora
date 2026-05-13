package grpc

import (
	v1 "github.com/evrone/go-clean-template/internal/controller/grpc/v1"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	pbgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewRouter -.
func NewRouter(
	app *pbgrpc.Server,
	t usecase.Translation,
	u usecase.User,
	tk usecase.Task,
	id usecase.ImportantDay,
	n usecase.Notification,
	d usecase.DeviceToken,
	l logger.Interface,
) {
	{
		v1.NewAuthRoutes(app, u, l)
		v1.NewTaskRoutes(app, tk, l)
		v1.NewTranslationRoutes(app, t, l)
		v1.NewImportantDayRoutes(app, id, l)
		v1.NewNotificationRoutes(app, n, l)
		v1.NewDeviceRoutes(app, d, l)
	}

	reflection.Register(app)
}
