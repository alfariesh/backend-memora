package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/gofiber/fiber/v2"
)

const (
	defaultBootstrapUpcomingDays   = 30
	defaultBootstrapUpcomingLimit  = 5
	defaultBootstrapUpcomingOffset = 0
)

// @Summary     Get mobile bootstrap
// @Description Get initial data needed by the mobile app after login
// @ID          mobile-bootstrap
// @Tags        mobile
// @Produce     json
// @Param       upcoming_days   query    int false "Upcoming lookahead window in days" default(30)
// @Param       upcoming_limit  query    int false "Upcoming important days limit"     default(5)
// @Param       upcoming_offset query    int false "Upcoming important days offset"    default(0)
// @Success     200             {object} response.MobileBootstrap
// @Failure     401             {object} response.Error
// @Failure     500             {object} response.Error
// @Security    BearerAuth
// @Router      /mobile/bootstrap [get]
func (r *V1) mobileBootstrap(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	settings, err := r.us.Get(ctx.UserContext(), userID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - mobileBootstrap - settings")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	upcomingDays := positiveQueryInt(ctx, "upcoming_days", defaultBootstrapUpcomingDays)
	upcomingLimit := positiveQueryInt(ctx, "upcoming_limit", defaultBootstrapUpcomingLimit)
	upcomingOffset := nonNegativeQueryInt(ctx, "upcoming_offset", defaultBootstrapUpcomingOffset)

	upcoming, upcomingTotal, err := r.id.Upcoming(ctx.UserContext(), userID, time.Now().UTC(), upcomingDays, upcomingLimit, upcomingOffset)
	if err != nil {
		r.l.Error(err, "restapi - v1 - mobileBootstrap - upcoming")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	unreadCount, err := r.n.CountUnread(ctx.UserContext(), userID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - mobileBootstrap - unreadCount")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	devices, err := r.d.List(ctx.UserContext(), userID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - mobileBootstrap - devices")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(response.MobileBootstrap{
		Settings:                settings,
		UpcomingImportantDays:   upcoming,
		UpcomingTotal:           upcomingTotal,
		UnreadNotificationCount: unreadCount,
		Devices:                 devices,
		DevicesTotal:            len(devices),
	})
}

func positiveQueryInt(ctx *fiber.Ctx, name string, defaultValue int) int {
	value := nonNegativeQueryInt(ctx, name, defaultValue)
	if value <= 0 {
		return defaultValue
	}

	return value
}

func nonNegativeQueryInt(ctx *fiber.Ctx, name string, defaultValue int) int {
	value, err := strconv.Atoi(ctx.Query(name, strconv.Itoa(defaultValue)))
	if err != nil || value < 0 {
		return defaultValue
	}

	return value
}
