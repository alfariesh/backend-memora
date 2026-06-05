package v1

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alfariesh/backend-memora/internal/controller/restapi/v1/response"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthRateLimiterRegister(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Post("/register", authRateLimiter("test-register", registerRateLimit, ipRateLimitKey), func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(http.StatusNoContent)
	})

	for range registerRateLimit {
		resp := performRateLimitRequest(t, app, "/register", http.NoBody)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		closeRateLimitBody(t, resp)
	}

	resp := performRateLimitRequest(t, app, "/register", http.NoBody)
	defer closeRateLimitBody(t, resp)

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	var body response.Error
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "too_many_requests", body.Error)
}

func TestAuthRateLimiterLoginUsesEmailInKey(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Post("/login", authRateLimiter("test-login", loginRateLimit, loginRateLimitKey), func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(http.StatusNoContent)
	})

	for range loginRateLimit {
		resp := performRateLimitRequest(t, app, "/login", strings.NewReader(`{"email":"test@example.com"}`))
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		closeRateLimitBody(t, resp)
	}

	limitedResp := performRateLimitRequest(t, app, "/login", strings.NewReader(`{"email":"test@example.com"}`))
	assert.Equal(t, http.StatusTooManyRequests, limitedResp.StatusCode)
	closeRateLimitBody(t, limitedResp)

	otherEmailResp := performRateLimitRequest(t, app, "/login", strings.NewReader(`{"email":"other@example.com"}`))
	defer closeRateLimitBody(t, otherEmailResp)

	assert.Equal(t, http.StatusNoContent, otherEmailResp.StatusCode)
}

func performRateLimitRequest(t *testing.T, app *fiber.App, path string, body io.Reader) *http.Response {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	return resp
}

func closeRateLimitBody(t *testing.T, resp *http.Response) {
	t.Helper()
	require.NoError(t, resp.Body.Close())
}
