package v1

import (
	"errors"
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

func (r *V1) listNotifications(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	limit, offset := pagination(ctx, 20)
	notifications, total, err := r.n.List(ctx.UserContext(), userID, ctx.QueryBool("unread_only"), limit, offset)
	if err != nil {
		r.l.Error(err, "restapi - v1 - listNotifications")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(response.NotificationList{
		Notifications: notifications,
		Total:         total,
	})
}

func (r *V1) markNotificationRead(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	notification, err := r.n.MarkRead(ctx.UserContext(), userID, ctx.Params("id"))
	if err != nil {
		r.l.Error(err, "restapi - v1 - markNotificationRead")

		if errors.Is(err, entity.ErrNotificationNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "notification not found")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(notification)
}

func (r *V1) markAllNotificationsRead(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	if err := r.n.MarkAllRead(ctx.UserContext(), userID); err != nil {
		r.l.Error(err, "restapi - v1 - markAllNotificationsRead")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.SendStatus(http.StatusNoContent)
}
