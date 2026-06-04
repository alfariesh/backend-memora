package restapi

import (
	"net/http"

	"github.com/alfariesh/backend-memora/config"
	_ "github.com/alfariesh/backend-memora/docs" // Swagger docs.
	"github.com/alfariesh/backend-memora/internal/controller/restapi/middleware"
	v1 "github.com/alfariesh/backend-memora/internal/controller/restapi/v1"
	"github.com/alfariesh/backend-memora/internal/usecase"
	"github.com/alfariesh/backend-memora/pkg/jwt"
	"github.com/alfariesh/backend-memora/pkg/logger"
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// NewRouter -.
// Swagger spec:
//
//	@title       Memora API
//	@description Backend API for Memora important days, reminders, notifications, devices, auth, and tasks
//	@version     1.0
//	@host        localhost:8080
//	@BasePath    /v1
//	@securityDefinitions.apikey BearerAuth
//	@in header
//	@name Authorization
func NewRouter(
	app *fiber.App,
	cfg *config.Config,
	u usecase.User,
	us usecase.UserSettings,
	tk usecase.Task,
	id usecase.ImportantDay,
	n usecase.Notification,
	d usecase.DeviceToken,
	jwtManager *jwt.Manager,
	l logger.Interface,
) {
	// Options
	app.Use(middleware.Logger(l))
	app.Use(middleware.Recovery(l))

	// Prometheus metrics
	if cfg.Metrics.Enabled {
		prometheus := fiberprometheus.New("my-service-name")
		prometheus.RegisterAt(app, "/metrics")
		app.Use(prometheus.Middleware)
	}

	// Swagger
	if cfg.Swagger.Enabled {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// K8s probe
	app.Get("/healthz", func(ctx *fiber.Ctx) error { return ctx.SendStatus(http.StatusOK) })

	// Routers
	apiV1Group := app.Group("/v1")
	{
		v1.NewRoutes(apiV1Group, u, us, tk, id, n, d, jwtManager, l)
	}
}
