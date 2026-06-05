package v1

import (
	"errors"
	"net/http"

	"github.com/alfariesh/backend-memora/internal/controller/restapi/v1/request"
	"github.com/alfariesh/backend-memora/internal/controller/restapi/v1/response"
	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary     Register
// @Description Register a new user
// @ID          register
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body     request.Register true "Registration data"
// @Success     201     {object} entity.User
// @Failure     400     {object} response.Error
// @Failure     409     {object} response.Error
// @Failure     429     {object} response.Error
// @Failure     500     {object} response.Error
// @Router      /auth/register [post]
func (r *V1) register(ctx *fiber.Ctx) error {
	var body request.Register

	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - register")

		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - register")

		return validationErrorResponse(ctx, err)
	}

	user, err := r.u.Register(ctx.UserContext(), body.Username, body.Email, body.Password)
	if err != nil {
		r.l.Error(err, "restapi - v1 - register")

		if errors.Is(err, entity.ErrUserAlreadyExists) {
			return errorResponse(ctx, http.StatusConflict, "user already exists")
		}

		if errors.Is(err, entity.ErrInvalidUserInput) {
			return errorResponse(ctx, http.StatusBadRequest, "invalid user input")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusCreated).JSON(user)
}

// @Summary     Login
// @Description Authenticate user and get access and refresh tokens
// @ID          login
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body     request.Login true "Login credentials"
// @Success     200     {object} entity.AuthTokens
// @Failure     400     {object} response.Error
// @Failure     401     {object} response.Error
// @Failure     429     {object} response.Error
// @Failure     500     {object} response.Error
// @Router      /auth/login [post]
func (r *V1) login(ctx *fiber.Ctx) error {
	var body request.Login

	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - login")

		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - login")

		return validationErrorResponse(ctx, err)
	}

	tokens, err := r.u.Login(ctx.UserContext(), body.Email, body.Password, sessionMetadataFromContext(ctx))
	if err != nil {
		r.l.Error(err, "restapi - v1 - login")

		if errors.Is(err, entity.ErrInvalidCredentials) {
			return errorResponse(ctx, http.StatusUnauthorized, "invalid credentials")
		}

		if errors.Is(err, entity.ErrInvalidUserInput) {
			return errorResponse(ctx, http.StatusBadRequest, "invalid user input")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(tokens)
}

// @Summary     Refresh token
// @Description Rotate refresh token and issue a new access token
// @ID          refresh-token
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body     request.RefreshToken true "Refresh token"
// @Success     200     {object} entity.AuthTokens
// @Failure     400     {object} response.Error
// @Failure     401     {object} response.Error
// @Failure     429     {object} response.Error
// @Failure     500     {object} response.Error
// @Router      /auth/refresh [post]
func (r *V1) refreshToken(ctx *fiber.Ctx) error {
	var body request.RefreshToken

	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - refreshToken")

		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - refreshToken")

		return validationErrorResponse(ctx, err)
	}

	tokens, err := r.u.Refresh(ctx.UserContext(), body.RefreshToken, sessionMetadataFromContext(ctx))
	if err != nil {
		r.l.Error(err, "restapi - v1 - refreshToken")

		if errors.Is(err, entity.ErrInvalidRefreshToken) || errors.Is(err, entity.ErrRefreshTokenReuse) {
			return errorResponse(ctx, http.StatusUnauthorized, "invalid refresh token")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(tokens)
}

// @Summary     Logout
// @Description Revoke a refresh token
// @ID          logout
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body request.RefreshToken true "Refresh token"
// @Success     204
// @Failure     400 {object} response.Error
// @Failure     500 {object} response.Error
// @Router      /auth/logout [post]
func (r *V1) logout(ctx *fiber.Ctx) error {
	var body request.RefreshToken

	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - logout")

		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - logout")

		return validationErrorResponse(ctx, err)
	}

	if err := r.u.Logout(ctx.UserContext(), body.RefreshToken); err != nil {
		r.l.Error(err, "restapi - v1 - logout")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.SendStatus(http.StatusNoContent)
}

// @Summary     List active sessions
// @Description List current user's active refresh sessions
// @ID          list-sessions
// @Tags        auth
// @Produce     json
// @Success     200 {object} response.UserSessionList
// @Failure     401 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /auth/sessions [get]
func (r *V1) listSessions(ctx *fiber.Ctx) error {
	userID, ok := currentUserID(ctx)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	sessions, err := r.u.ListSessions(ctx.UserContext(), userID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - listSessions")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(response.UserSessionList{
		Sessions: sessions,
		Total:    len(sessions),
	})
}

// @Summary     Revoke session
// @Description Revoke one active session owned by current user
// @ID          revoke-session
// @Tags        auth
// @Param       id path string true "Session ID"
// @Success     204
// @Failure     401 {object} response.Error
// @Failure     404 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /auth/sessions/{id} [delete]
func (r *V1) revokeSession(ctx *fiber.Ctx) error {
	userID, ok := currentUserID(ctx)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	if err := r.u.RevokeSession(ctx.UserContext(), userID, ctx.Params("id")); err != nil {
		r.l.Error(err, "restapi - v1 - revokeSession")

		if errors.Is(err, entity.ErrSessionNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "session not found")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.SendStatus(http.StatusNoContent)
}

// @Summary     Logout all sessions
// @Description Revoke all active sessions owned by current user
// @ID          logout-all
// @Tags        auth
// @Success     204
// @Failure     401 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /auth/logout-all [post]
func (r *V1) logoutAll(ctx *fiber.Ctx) error {
	userID, ok := currentUserID(ctx)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	if err := r.u.LogoutAll(ctx.UserContext(), userID); err != nil {
		r.l.Error(err, "restapi - v1 - logoutAll")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.SendStatus(http.StatusNoContent)
}

// @Summary     Get profile
// @Description Get current user profile
// @ID          profile
// @Tags        user
// @Produce     json
// @Success     200 {object} entity.User
// @Failure     401 {object} response.Error
// @Failure     404 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /user/profile [get]
func (r *V1) profile(ctx *fiber.Ctx) error {
	userID, ok := currentUserID(ctx)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	user, err := r.u.GetUser(ctx.UserContext(), userID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - profile")

		if errors.Is(err, entity.ErrUserNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "user not found")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(user)
}

// @Summary     Change password
// @Description Change current user's password, revoke old sessions, and issue a new session
// @ID          change-password
// @Tags        user
// @Accept      json
// @Produce     json
// @Param       request body     request.ChangePassword true "Password change data"
// @Success     200     {object} entity.AuthTokens
// @Failure     400     {object} response.Error
// @Failure     401     {object} response.Error
// @Failure     404     {object} response.Error
// @Failure     500     {object} response.Error
// @Security    BearerAuth
// @Router      /user/password [post]
func (r *V1) changePassword(ctx *fiber.Ctx) error {
	userID, ok := currentUserID(ctx)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var body request.ChangePassword
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - changePassword")

		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - changePassword")

		return validationErrorResponse(ctx, err)
	}

	tokens, err := r.u.ChangePassword(ctx.UserContext(), userID, body.CurrentPassword, body.NewPassword, sessionMetadataFromContext(ctx))
	if err != nil {
		r.l.Error(err, "restapi - v1 - changePassword")

		switch {
		case errors.Is(err, entity.ErrInvalidUserInput):
			return errorResponse(ctx, http.StatusBadRequest, "invalid user input")
		case errors.Is(err, entity.ErrInvalidCredentials):
			return errorResponse(ctx, http.StatusUnauthorized, "invalid credentials")
		case errors.Is(err, entity.ErrUserNotFound):
			return errorResponse(ctx, http.StatusNotFound, "user not found")
		default:
			return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
		}
	}

	return ctx.Status(http.StatusOK).JSON(tokens)
}

func currentUserID(ctx *fiber.Ctx) (string, bool) {
	userID, ok := ctx.Locals("userID").(string)
	return userID, ok && userID != ""
}

func sessionMetadataFromContext(ctx *fiber.Ctx) entity.SessionMetadata {
	return entity.NormalizeSessionMetadata(entity.SessionMetadata{
		IP:        ctx.IP(),
		UserAgent: ctx.Get("User-Agent"),
	})
}

// @Summary     Get user settings
// @Description Get reminder defaults and notification preferences for the current user
// @ID          get-user-settings
// @Tags        user
// @Produce     json
// @Success     200 {object} entity.UserSettings
// @Failure     401 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /user/settings [get]
func (r *V1) getUserSettings(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	settings, err := r.us.Get(ctx.UserContext(), userID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getUserSettings")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(settings)
}

// @Summary     Update user settings
// @Description Update reminder defaults and notification preferences for the current user
// @ID          update-user-settings
// @Tags        user
// @Accept      json
// @Produce     json
// @Param       request body     request.UpdateUserSettings true "User settings"
// @Success     200     {object} entity.UserSettings
// @Failure     400     {object} response.Error
// @Failure     401     {object} response.Error
// @Failure     500     {object} response.Error
// @Security    BearerAuth
// @Router      /user/settings [put]
func (r *V1) updateUserSettings(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var body request.UpdateUserSettings
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - updateUserSettings")

		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - updateUserSettings")

		return validationErrorResponse(ctx, err)
	}

	settings, err := r.us.Update(ctx.UserContext(), userID, body.ToParams())
	if err != nil {
		r.l.Error(err, "restapi - v1 - updateUserSettings")

		if errors.Is(err, entity.ErrInvalidUserSettings) {
			return errorResponse(ctx, http.StatusBadRequest, "invalid user settings")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(settings)
}
