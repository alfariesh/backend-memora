package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alfariesh/backend-memora/internal/controller/restapi/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCORSMiddlewareAllowsConfiguredOrigin(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(middleware.CORS(middleware.CORSConfig{
		Enabled:        true,
		AllowedOrigins: "http://localhost:3000, http://127.0.0.1:5173",
		AllowedMethods: "GET,POST,OPTIONS",
		AllowedHeaders: "Authorization,Content-Type",
		ExposedHeaders: "X-Request-ID,X-Correlation-ID",
		MaxAge:         600,
	}))
	app.Post("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(http.StatusNoContent)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodOptions, "/test", http.NoBody)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "Authorization")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, "http://localhost:3000", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET,POST,OPTIONS", resp.Header.Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Authorization,Content-Type", resp.Header.Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "X-Request-ID,X-Correlation-ID", resp.Header.Get("Access-Control-Expose-Headers"))
	assert.Equal(t, "600", resp.Header.Get("Access-Control-Max-Age"))
}

func TestCORSMiddlewareDeniesUnconfiguredOrigin(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(middleware.CORS(middleware.CORSConfig{
		Enabled:        true,
		AllowedOrigins: "http://localhost:3000",
	}))
	app.Get("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	req.Header.Set("Origin", "https://evil.example")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, resp.Header.Values("Vary"), "Origin")
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(middleware.SecurityHeaders(middleware.SecurityHeadersConfig{
		Enabled: true,
	}))
	app.Get("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(http.StatusOK)
	})

	resp := performSecurityHeaderRequest(t, app)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
	assert.Equal(t, "no-referrer", resp.Header.Get("Referrer-Policy"))
	assert.Equal(t, "geolocation=(), microphone=(), camera=()", resp.Header.Get("Permissions-Policy"))
	assert.Equal(t, "?1", resp.Header.Get("Origin-Agent-Cluster"))
	assert.Equal(t, "off", resp.Header.Get("X-DNS-Prefetch-Control"))
	assert.Equal(t, "noopen", resp.Header.Get("X-Download-Options"))
	assert.Equal(t, "none", resp.Header.Get("X-Permitted-Cross-Domain-Policies"))
	assert.Empty(t, resp.Header.Get("Strict-Transport-Security"))
}

func TestSecurityHeadersMiddlewareHSTS(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(middleware.SecurityHeaders(middleware.SecurityHeadersConfig{
		Enabled:               true,
		HSTSEnabled:           true,
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		HSTSPreload:           true,
	}))
	app.Get("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(http.StatusOK)
	})

	resp := performSecurityHeaderRequest(t, app)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	assert.Equal(t, "max-age=31536000; includeSubDomains; preload", resp.Header.Get("Strict-Transport-Security"))
}

func performSecurityHeaderRequest(t *testing.T, app *fiber.App) *http.Response {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	resp, err := app.Test(req)
	require.NoError(t, err)

	return resp
}
