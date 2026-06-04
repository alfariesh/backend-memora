package v1

import (
	"github.com/alfariesh/backend-memora/internal/controller/restapi/middleware"
	"github.com/alfariesh/backend-memora/internal/usecase"
	"github.com/alfariesh/backend-memora/pkg/jwt"
	"github.com/alfariesh/backend-memora/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// NewRoutes -.
func NewRoutes(
	apiV1Group fiber.Router,
	u usecase.User,
	us usecase.UserSettings,
	tk usecase.Task,
	id usecase.ImportantDay,
	n usecase.Notification,
	d usecase.DeviceToken,
	jwtManager *jwt.Manager,
	l logger.Interface,
) {
	r := &V1{u: u, us: us, tk: tk, id: id, n: n, d: d, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	// Public routes
	authGroup := apiV1Group.Group("/auth")
	{
		authGroup.Post("/register", r.register)
		authGroup.Post("/login", r.login)
		authGroup.Post("/refresh", r.refreshToken)
		authGroup.Post("/logout", r.logout)
	}

	// Protected routes
	protected := apiV1Group.Group("", middleware.Auth(jwtManager))

	userGroup := protected.Group("/user")
	{
		userGroup.Get("/profile", r.profile)
		userGroup.Get("/settings", r.getUserSettings)
		userGroup.Put("/settings", r.updateUserSettings)
	}

	taskGroup := protected.Group("/tasks")
	{
		taskGroup.Post("/", r.createTask)
		taskGroup.Get("/", r.listTasks)
		taskGroup.Get("/:id", r.getTask)
		taskGroup.Put("/:id", r.updateTask)
		taskGroup.Patch("/:id/status", r.transitionTask)
		taskGroup.Delete("/:id", r.deleteTask)
	}

	mobileGroup := protected.Group("/mobile")
	{
		mobileGroup.Get("/bootstrap", r.mobileBootstrap)
	}

	importantDayGroup := protected.Group("/important-days")
	{
		importantDayGroup.Post("/", r.createImportantDay)
		importantDayGroup.Get("/", r.listImportantDays)
		importantDayGroup.Get("/upcoming", r.upcomingImportantDays)
		importantDayGroup.Get("/:id/reminders", r.getImportantDayReminders)
		importantDayGroup.Get("/:id", r.getImportantDay)
		importantDayGroup.Put("/:id", r.updateImportantDay)
		importantDayGroup.Put("/:id/reminders", r.replaceImportantDayReminders)
		importantDayGroup.Delete("/:id", r.deleteImportantDay)
	}

	notificationGroup := protected.Group("/notifications")
	{
		notificationGroup.Get("/", r.listNotifications)
		notificationGroup.Get("/unread-count", r.countUnreadNotifications)
		notificationGroup.Patch("/:id/read", r.markNotificationRead)
		notificationGroup.Patch("/read-all", r.markAllNotificationsRead)
	}

	deviceGroup := protected.Group("/devices")
	{
		deviceGroup.Get("/", r.listDevices)
		deviceGroup.Post("/", r.registerDevice)
		deviceGroup.Post("/:id/test-push", r.testPush)
		deviceGroup.Delete("/:id", r.deleteDevice)
	}
}
