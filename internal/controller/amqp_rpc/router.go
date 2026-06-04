package v1

import (
	v1 "github.com/alfariesh/backend-memora/internal/controller/amqp_rpc/v1"
	"github.com/alfariesh/backend-memora/internal/usecase"
	"github.com/alfariesh/backend-memora/pkg/jwt"
	"github.com/alfariesh/backend-memora/pkg/logger"
	"github.com/alfariesh/backend-memora/pkg/rabbitmq/rmq_rpc/server"
)

// NewRouter -.
func NewRouter(
	u usecase.User,
	tk usecase.Task,
	id usecase.ImportantDay,
	n usecase.Notification,
	d usecase.DeviceToken,
	j *jwt.Manager,
	l logger.Interface,
) map[string]server.CallHandler {
	routes := make(map[string]server.CallHandler)

	{
		v1.NewRoutes(routes, u, tk, id, n, d, j, l)
	}

	return routes
}
