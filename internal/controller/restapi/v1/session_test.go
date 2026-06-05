package v1

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alfariesh/backend-memora/internal/controller/restapi/v1/response"
	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/alfariesh/backend-memora/pkg/logger"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type restAuthStub struct {
	refreshErr         error
	revokeSessionErr   error
	changePasswordErr  error
	refreshMetadata    entity.SessionMetadata
	changeMetadata     entity.SessionMetadata
	listSessionsCalled bool
	logoutAllCalled    bool
	revokeSessionID    string
}

func (s *restAuthStub) Register(_ context.Context, _, _, _ string) (entity.User, error) {
	return entity.User{}, nil
}

func (s *restAuthStub) Login(_ context.Context, _, _ string, metadata entity.SessionMetadata) (entity.AuthTokens, error) {
	return entity.AuthTokens{AccessToken: "access-token", RefreshToken: restRefreshToken()}, nil
}

func (s *restAuthStub) LoginAccessOnly(_ context.Context, _, _ string) (entity.AuthTokens, error) {
	return entity.AuthTokens{AccessToken: "access-token"}, nil
}

func (s *restAuthStub) Refresh(_ context.Context, _ string, metadata entity.SessionMetadata) (entity.AuthTokens, error) {
	s.refreshMetadata = metadata
	if s.refreshErr != nil {
		return entity.AuthTokens{}, s.refreshErr
	}

	return entity.AuthTokens{AccessToken: "new-access-token", RefreshToken: restRefreshToken()}, nil
}

func (s *restAuthStub) Logout(_ context.Context, _ string) error {
	return nil
}

func (s *restAuthStub) ListSessions(_ context.Context, _ string) ([]entity.UserSessionView, error) {
	s.listSessionsCalled = true

	return []entity.UserSessionView{{ID: "session-id-1", CreatedAt: time.Now().UTC()}}, nil
}

func (s *restAuthStub) RevokeSession(_ context.Context, _ string, sessionID string) error {
	s.revokeSessionID = sessionID

	return s.revokeSessionErr
}

func (s *restAuthStub) LogoutAll(_ context.Context, _ string) error {
	s.logoutAllCalled = true

	return nil
}

func (s *restAuthStub) ChangePassword(
	_ context.Context,
	_,
	_,
	_ string,
	metadata entity.SessionMetadata,
) (entity.AuthTokens, error) {
	s.changeMetadata = metadata
	if s.changePasswordErr != nil {
		return entity.AuthTokens{}, s.changePasswordErr
	}

	return entity.AuthTokens{AccessToken: "new-access-token", RefreshToken: restRefreshToken()}, nil
}

func (s *restAuthStub) GetUser(_ context.Context, _ string) (entity.User, error) {
	return entity.User{}, nil
}

func TestRefreshTokenReuseReturnsUnauthorized(t *testing.T) {
	t.Parallel()

	stub := &restAuthStub{refreshErr: entity.ErrRefreshTokenReuse}
	app := fiber.New()
	controller := restController(stub)
	app.Post("/auth/refresh", controller.refreshToken)

	resp := performRESTSessionRequest(t, app, http.MethodPost, "/auth/refresh", `{"refresh_token":"`+restRefreshToken()+`"}`)
	defer closeRESTSessionBody(t, resp)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.NotEmpty(t, stub.refreshMetadata.IP)
	assert.Equal(t, "Memora/1.0", stub.refreshMetadata.UserAgent)

	var body response.Error
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "invalid_refresh_token", body.Error)
}

func TestSessionManagementEndpoints(t *testing.T) {
	t.Parallel()

	t.Run("list sessions", func(t *testing.T) {
		t.Parallel()

		stub := &restAuthStub{}
		app := fiber.New()
		controller := restController(stub)
		app.Get("/auth/sessions", withUser(controller.listSessions))

		resp := performRESTSessionRequest(t, app, http.MethodGet, "/auth/sessions", "")
		defer closeRESTSessionBody(t, resp)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, stub.listSessionsCalled)

		var body response.UserSessionList
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		assert.Equal(t, 1, body.Total)
		assert.Equal(t, "session-id-1", body.Sessions[0].ID)
	})

	t.Run("revoke missing session", func(t *testing.T) {
		t.Parallel()

		stub := &restAuthStub{revokeSessionErr: entity.ErrSessionNotFound}
		app := fiber.New()
		controller := restController(stub)
		app.Delete("/auth/sessions/:id", withUser(controller.revokeSession))

		resp := performRESTSessionRequest(t, app, http.MethodDelete, "/auth/sessions/missing-id", "")
		defer closeRESTSessionBody(t, resp)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Equal(t, "missing-id", stub.revokeSessionID)
	})

	t.Run("logout all", func(t *testing.T) {
		t.Parallel()

		stub := &restAuthStub{}
		app := fiber.New()
		controller := restController(stub)
		app.Post("/auth/logout-all", withUser(controller.logoutAll))

		resp := performRESTSessionRequest(t, app, http.MethodPost, "/auth/logout-all", "")
		defer closeRESTSessionBody(t, resp)

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		assert.True(t, stub.logoutAllCalled)
	})
}

func TestChangePasswordEndpoint(t *testing.T) {
	t.Parallel()

	stub := &restAuthStub{}
	app := fiber.New()
	controller := restController(stub)
	app.Post("/user/password", withUser(controller.changePassword))

	resp := performRESTSessionRequest(
		t,
		app,
		http.MethodPost,
		"/user/password",
		`{"current_password":"password123","new_password":"newpassword123"}`,
	)
	defer closeRESTSessionBody(t, resp)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, stub.changeMetadata.IP)
	assert.Equal(t, "Memora/1.0", stub.changeMetadata.UserAgent)
}

func restController(user *restAuthStub) *V1 {
	return &V1{u: user, l: logger.New("error"), v: newValidator()}
}

func withUser(handler fiber.Handler) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		ctx.Locals("userID", "user-id-123")

		return handler(ctx)
	}
}

func performRESTSessionRequest(t *testing.T, app *fiber.App, method, path, body string) *http.Response {
	t.Helper()

	var reader io.Reader = http.NoBody
	if body != "" {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequestWithContext(t.Context(), method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Memora/1.0")
	req.RemoteAddr = "127.0.0.1:1234"

	resp, err := app.Test(req)
	require.NoError(t, err)

	return resp
}

func closeRESTSessionBody(t *testing.T, resp *http.Response) {
	t.Helper()
	require.NoError(t, resp.Body.Close())
}

func restRefreshToken() string {
	b := make([]byte, entity.RefreshTokenBytes)
	for i := range b {
		b[i] = byte(i + 1)
	}

	return base64.RawURLEncoding.EncodeToString(b)
}
