package v1

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/alfariesh/backend-memora/internal/controller/restapi/v1/request"
	"github.com/alfariesh/backend-memora/internal/controller/restapi/v1/response"
	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary     Create important day
// @Description Create an important day and its reminder rules for the current user
// @ID          create-important-day
// @Tags        important-days
// @Accept      json
// @Produce     json
// @Param       request body     request.CreateImportantDay true "Important day data"
// @Success     201     {object} entity.ImportantDay
// @Failure     400     {object} response.Error
// @Failure     401     {object} response.Error
// @Failure     500     {object} response.Error
// @Security    BearerAuth
// @Router      /important-days [post]
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

// @Summary     List important days
// @Description List important days for the current user with optional type filtering
// @ID          list-important-days
// @Tags        important-days
// @Produce     json
// @Param       type   query    string false "Filter by important day type" Enums(birthday, wedding, memorial, graduation, first_day, document, subscription, medical, custom)
// @Param       limit  query    int    false "Limit"  default(10)
// @Param       offset query    int    false "Offset" default(0)
// @Success     200    {object} response.ImportantDayList
// @Failure     400    {object} response.Error
// @Failure     401    {object} response.Error
// @Failure     500    {object} response.Error
// @Security    BearerAuth
// @Router      /important-days [get]
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

// @Summary     List upcoming important days
// @Description List upcoming occurrences for the current user's important days
// @ID          upcoming-important-days
// @Tags        important-days
// @Produce     json
// @Param       days   query    int false "Lookahead window in days" default(365)
// @Param       limit  query    int false "Limit"                    default(10)
// @Param       offset query    int false "Offset"                   default(0)
// @Success     200    {object} response.UpcomingImportantDayList
// @Failure     401    {object} response.Error
// @Failure     500    {object} response.Error
// @Security    BearerAuth
// @Router      /important-days/upcoming [get]
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

// @Summary     Get important day
// @Description Get an important day by ID
// @ID          get-important-day
// @Tags        important-days
// @Produce     json
// @Param       id  path     string true "Important day ID"
// @Success     200 {object} entity.ImportantDay
// @Failure     401 {object} response.Error
// @Failure     403 {object} response.Error
// @Failure     404 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /important-days/{id} [get]
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

// @Summary     Update important day
// @Description Update an important day by ID
// @ID          update-important-day
// @Tags        important-days
// @Accept      json
// @Produce     json
// @Param       id      path     string                     true "Important day ID"
// @Param       request body     request.UpdateImportantDay true "Important day data"
// @Success     200     {object} entity.ImportantDay
// @Failure     400     {object} response.Error
// @Failure     401     {object} response.Error
// @Failure     404     {object} response.Error
// @Failure     500     {object} response.Error
// @Security    BearerAuth
// @Router      /important-days/{id} [put]
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

// @Summary     Replace important day reminders
// @Description Replace reminder rules for an important day
// @ID          replace-important-day-reminders
// @Tags        important-days
// @Accept      json
// @Produce     json
// @Param       id      path     string                       true "Important day ID"
// @Param       request body     request.ReplaceReminderRules true "Reminder rules"
// @Success     200     {object} response.ReminderRuleList
// @Failure     400     {object} response.Error
// @Failure     401     {object} response.Error
// @Failure     404     {object} response.Error
// @Failure     500     {object} response.Error
// @Security    BearerAuth
// @Router      /important-days/{id}/reminders [put]
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

// @Summary     Get important day reminders
// @Description Get reminder rules for an important day
// @ID          get-important-day-reminders
// @Tags        important-days
// @Produce     json
// @Param       id  path     string true "Important day ID"
// @Success     200 {object} response.ReminderRuleList
// @Failure     401 {object} response.Error
// @Failure     403 {object} response.Error
// @Failure     404 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /important-days/{id}/reminders [get]
func (r *V1) getImportantDayReminders(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	rules, err := r.id.GetReminderRules(ctx.UserContext(), userID, ctx.Params("id"))
	if err != nil {
		r.l.Error(err, "restapi - v1 - getImportantDayReminders")

		if errors.Is(err, entity.ErrImportantDayNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "important day not found")
		}

		if errors.Is(err, entity.ErrImportantDayForbidden) {
			return errorResponse(ctx, http.StatusForbidden, "forbidden")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(response.ReminderRuleList{Rules: rules})
}

// @Summary     Delete important day
// @Description Delete an important day by ID
// @ID          delete-important-day
// @Tags        important-days
// @Param       id  path     string true "Important day ID"
// @Success     204 "No Content"
// @Failure     401 {object} response.Error
// @Failure     404 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /important-days/{id} [delete]
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
