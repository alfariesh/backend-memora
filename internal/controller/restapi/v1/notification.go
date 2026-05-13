package v1

import (
	"errors"
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary     List notifications
// @Description List in-app notifications for the current user
// @ID          list-notifications
// @Tags        notifications
// @Produce     json
// @Param       unread_only query    bool false "Only unread notifications" default(false)
// @Param       limit       query    int  false "Limit"                     default(20)
// @Param       offset      query    int  false "Offset"                    default(0)
// @Success     200         {object} response.NotificationList
// @Failure     401         {object} response.Error
// @Failure     500         {object} response.Error
// @Security    BearerAuth
// @Router      /notifications [get]
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

// @Summary     Count unread notifications
// @Description Get unread in-app notification count for the current user
// @ID          count-unread-notifications
// @Tags        notifications
// @Produce     json
// @Success     200 {object} response.UnreadNotificationCount
// @Failure     401 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /notifications/unread-count [get]
func (r *V1) countUnreadNotifications(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	count, err := r.n.CountUnread(ctx.UserContext(), userID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - countUnreadNotifications")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(response.UnreadNotificationCount{UnreadCount: count})
}

// @Summary     Mark notification read
// @Description Mark one notification as read
// @ID          mark-notification-read
// @Tags        notifications
// @Produce     json
// @Param       id  path     string true "Notification ID"
// @Success     200 {object} entity.Notification
// @Failure     401 {object} response.Error
// @Failure     404 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /notifications/{id}/read [patch]
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

// @Summary     Mark all notifications read
// @Description Mark all notifications for the current user as read
// @ID          mark-all-notifications-read
// @Tags        notifications
// @Success     204 "No Content"
// @Failure     401 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /notifications/read-all [patch]
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
