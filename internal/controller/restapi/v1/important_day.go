package v1

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/request"
	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

func (r *V1) createImportantDay(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var body request.CreateImportantDay
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - createImportantDay")

		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - createImportantDay")

		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	day, err := r.id.Create(ctx.UserContext(), userID, body.ToParams())
	if err != nil {
		r.l.Error(err, "restapi - v1 - createImportantDay")

		if errors.Is(err, entity.ErrInvalidImportantDayDate) {
			return errorResponse(ctx, http.StatusBadRequest, "invalid important day date")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusCreated).JSON(day)
}

func (r *V1) listImportantDays(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var dayType *entity.ImportantDayType
	if t := ctx.Query("type"); t != "" {
		parsed := entity.ImportantDayType(t)
		if !parsed.Valid() {
			return errorResponse(ctx, http.StatusBadRequest, "invalid important day type")
		}

		dayType = &parsed
	}

	limit, offset := pagination(ctx, 10)

	days, total, err := r.id.List(ctx.UserContext(), userID, dayType, limit, offset)
	if err != nil {
		r.l.Error(err, "restapi - v1 - listImportantDays")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(response.ImportantDayList{
		ImportantDays: days,
		Total:         total,
	})
}

func (r *V1) upcomingImportantDays(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	days, err := strconv.Atoi(ctx.Query("days", "365"))
	if err != nil {
		days = 365
	}

	limit, offset := pagination(ctx, 10)
	upcoming, total, err := r.id.Upcoming(ctx.UserContext(), userID, time.Now().UTC(), days, limit, offset)
	if err != nil {
		r.l.Error(err, "restapi - v1 - upcomingImportantDays")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(response.UpcomingImportantDayList{
		ImportantDays: upcoming,
		Total:         total,
	})
}

func (r *V1) getImportantDay(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	day, err := r.id.Get(ctx.UserContext(), userID, ctx.Params("id"))
	if err != nil {
		r.l.Error(err, "restapi - v1 - getImportantDay")

		if errors.Is(err, entity.ErrImportantDayNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "important day not found")
		}

		if errors.Is(err, entity.ErrImportantDayForbidden) {
			return errorResponse(ctx, http.StatusForbidden, "forbidden")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(day)
}

func (r *V1) updateImportantDay(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var body request.UpdateImportantDay
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - updateImportantDay")

		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - updateImportantDay")

		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	day, err := r.id.Update(ctx.UserContext(), userID, ctx.Params("id"), body.ToParams())
	if err != nil {
		r.l.Error(err, "restapi - v1 - updateImportantDay")

		if errors.Is(err, entity.ErrImportantDayNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "important day not found")
		}

		if errors.Is(err, entity.ErrInvalidImportantDayDate) {
			return errorResponse(ctx, http.StatusBadRequest, "invalid important day date")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(day)
}

func (r *V1) replaceImportantDayReminders(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var body request.ReplaceReminderRules
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - replaceImportantDayReminders")

		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - replaceImportantDayReminders")

		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	rules, err := r.id.ReplaceReminderRules(ctx.UserContext(), userID, ctx.Params("id"), body.ToParams())
	if err != nil {
		r.l.Error(err, "restapi - v1 - replaceImportantDayReminders")

		if errors.Is(err, entity.ErrImportantDayNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "important day not found")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(response.ReminderRuleList{Rules: rules})
}

func (r *V1) deleteImportantDay(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	if err := r.id.Delete(ctx.UserContext(), userID, ctx.Params("id")); err != nil {
		r.l.Error(err, "restapi - v1 - deleteImportantDay")

		if errors.Is(err, entity.ErrImportantDayNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "important day not found")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.SendStatus(http.StatusNoContent)
}

func pagination(ctx *fiber.Ctx, defaultLimit int) (int, int) {
	limit, err := strconv.Atoi(ctx.Query("limit", strconv.Itoa(defaultLimit)))
	if err != nil {
		limit = defaultLimit
	}

	offset, err := strconv.Atoi(ctx.Query("offset", "0"))
	if err != nil {
		offset = 0
	}

	return limit, offset
}
