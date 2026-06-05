package v1

import (
	"net/http"
	"strings"
	"time"

	"github.com/alfariesh/backend-memora/internal/controller/restapi/middleware"
	"github.com/alfariesh/backend-memora/internal/usecase"
	"github.com/alfariesh/backend-memora/pkg/jwt"
	"github.com/alfariesh/backend-memora/pkg/logger"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

const (
	registerRateLimit = 3
	loginRateLimit    = 5
	refreshRateLimit  = 30
	authRateWindow    = time.Minute
)

// NewRoutes -.
func NewRoutes(
	apiV1Group fiber.Router,
	u usecase.User,
	us usecase.UserSettings,
	id usecase.ImportantDay,
	n usecase.Notification,
	d usecase.DeviceToken,
	jwtManager *jwt.Manager,
	l logger.Interface,
) {
	r := &V1{u: u, us: us, id: id, n: n, d: d, l: l, v: newValidator()}

	// Public routes
	authGroup := apiV1Group.Group("/auth")
	{
		authGroup.Post("/register", authRateLimiter("register", registerRateLimit, ipRateLimitKey), r.register)
		authGroup.Post("/login", authRateLimiter("login", loginRateLimit, loginRateLimitKey), r.login)
		authGroup.Post("/refresh", authRateLimiter("refresh", refreshRateLimit, ipRateLimitKey), r.refreshToken)
		authGroup.Post("/logout", r.logout)
	}

	// Protected routes
	protected := apiV1Group.Group("", middleware.Auth(jwtManager))

	authProtectedGroup := protected.Group("/auth")
	{
		authProtectedGroup.Get("/sessions", r.listSessions)
		authProtectedGroup.Delete("/sessions/:id", r.revokeSession)
		authProtectedGroup.Post("/logout-all", r.logoutAll)
	}

	userGroup := protected.Group("/user")
	{
		userGroup.Get("/profile", r.profile)
		userGroup.Post("/password", r.changePassword)
		userGroup.Get("/settings", r.getUserSettings)
		userGroup.Put("/settings", r.updateUserSettings)
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

func authRateLimiter(name string, max int, keyGenerator func(*fiber.Ctx) string) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: authRateWindow,
		KeyGenerator: func(ctx *fiber.Ctx) string {
			return "auth:" + name + ":" + keyGenerator(ctx)
		},
		LimitReached: func(ctx *fiber.Ctx) error {
			return errorResponse(ctx, http.StatusTooManyRequests, "too many requests")
		},
	})
}

func ipRateLimitKey(ctx *fiber.Ctx) string {
	return ctx.IP()
}

func loginRateLimitKey(ctx *fiber.Ctx) string {
	var body struct {
		Email string `json:"email"`
	}

	if err := json.Unmarshal(ctx.Body(), &body); err != nil {
		return ctx.IP()
	}

	email := strings.ToLower(strings.TrimSpace(body.Email))
	if email == "" {
		return ctx.IP()
	}

	return ctx.IP() + ":" + email
}
