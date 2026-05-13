package v1

import (
	v1 "github.com/evrone/go-clean-template/internal/controller/amqp_rpc/v1"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/jwt"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/evrone/go-clean-template/pkg/rabbitmq/rmq_rpc/server"
)

// NewRouter -.
func NewRouter(
	t usecase.Translation,
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
		v1.NewRoutes(routes, t, u, tk, id, n, d, j, l)
	}

	return routes
}
