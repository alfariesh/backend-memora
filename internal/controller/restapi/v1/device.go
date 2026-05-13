package v1

import (
	"errors"
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/request"
	_ "github.com/evrone/go-clean-template/internal/controller/restapi/v1/response" // for swaggo
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary     Register device
// @Description Register or reactivate an Expo push token for the current user
// @ID          register-device
// @Tags        devices
// @Accept      json
// @Produce     json
// @Param       request body     request.RegisterDevice true "Device token data"
// @Success     201     {object} entity.DeviceToken
// @Failure     400     {object} response.Error
// @Failure     401     {object} response.Error
// @Failure     500     {object} response.Error
// @Security    BearerAuth
// @Router      /devices [post]
func (r *V1) registerDevice(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var body request.RegisterDevice
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - registerDevice")

		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - registerDevice")

		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	token, err := r.d.Register(ctx.UserContext(), userID, body.Token, body.Platform, body.Name)
	if err != nil {
		r.l.Error(err, "restapi - v1 - registerDevice")

		if errors.Is(err, entity.ErrInvalidDeviceToken) {
			return errorResponse(ctx, http.StatusBadRequest, "invalid device token")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusCreated).JSON(token)
}

// @Summary     Send test push
// @Description Send a test Expo push notification to a registered device
// @ID          test-push
// @Tags        devices
// @Accept      json
// @Produce     json
// @Param       id      path string           true  "Device token ID"
// @Param       request body request.TestPush false "Test push copy"
// @Success     200 {object} entity.PushTestResult
// @Failure     400 {object} response.Error
// @Failure     401 {object} response.Error
// @Failure     404 {object} response.Error
// @Failure     410 {object} response.Error
// @Failure     502 {object} response.Error
// @Failure     503 {object} response.Error
// @Security    BearerAuth
// @Router      /devices/{id}/test-push [post]
func (r *V1) testPush(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var body request.TestPush
	if len(ctx.Body()) > 0 {
		if err := ctx.BodyParser(&body); err != nil {
			r.l.Error(err, "restapi - v1 - testPush")

			return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
		}
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - testPush")

		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	result, err := r.d.TestPush(ctx.UserContext(), userID, ctx.Params("id"), body.Title, body.Body)
	if err != nil {
		r.l.Error(err, "restapi - v1 - testPush")

		if errors.Is(err, entity.ErrDeviceTokenNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "device not found")
		}

		if errors.Is(err, entity.ErrPushDeviceNotRegistered) {
			return errorResponse(ctx, http.StatusGone, "push device not registered")
		}

		if errors.Is(err, entity.ErrPushSenderNotConfigured) {
			return errorResponse(ctx, http.StatusServiceUnavailable, "push sender not configured")
		}

		if errors.Is(err, entity.ErrPushSendFailed) {
			return errorResponse(ctx, http.StatusBadGateway, "push send failed")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(result)
}

// @Summary     Delete device
// @Description Deactivate a registered device token by ID
// @ID          delete-device
// @Tags        devices
// @Param       id  path     string true "Device token ID"
// @Success     204 "No Content"
// @Failure     401 {object} response.Error
// @Failure     404 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /devices/{id} [delete]
func (r *V1) deleteDevice(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	if err := r.d.Delete(ctx.UserContext(), userID, ctx.Params("id")); err != nil {
		r.l.Error(err, "restapi - v1 - deleteDevice")

		if errors.Is(err, entity.ErrDeviceTokenNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "device not found")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.SendStatus(http.StatusNoContent)
}
