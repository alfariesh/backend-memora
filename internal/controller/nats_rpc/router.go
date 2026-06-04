package v1

import (
	v1 "github.com/alfariesh/backend-memora/internal/controller/nats_rpc/v1"
	"github.com/alfariesh/backend-memora/internal/usecase"
	"github.com/alfariesh/backend-memora/pkg/jwt"
	"github.com/alfariesh/backend-memora/pkg/logger"
	"github.com/alfariesh/backend-memora/pkg/nats/nats_rpc/server"
)

// NewRouter -.
func NewRouter(
	u usecase.User,
	id usecase.ImportantDay,
	n usecase.Notification,
	d usecase.DeviceToken,
	j *jwt.Manager,
	l logger.Interface,
) map[string]server.CallHandler {
	routes := make(map[string]server.CallHandler)

	{
		v1.NewRoutes(routes, u, id, n, d, j, l)
	}

	return routes
}
