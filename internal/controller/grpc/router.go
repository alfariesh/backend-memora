package grpc

import (
	v1 "github.com/alfariesh/backend-memora/internal/controller/grpc/v1"
	"github.com/alfariesh/backend-memora/internal/usecase"
	"github.com/alfariesh/backend-memora/pkg/logger"
	pbgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewRouter -.
func NewRouter(
	app *pbgrpc.Server,
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
		v1.NewImportantDayRoutes(app, id, l)
		v1.NewNotificationRoutes(app, n, l)
		v1.NewDeviceRoutes(app, d, l)
	}

	reflection.Register(app)
}
