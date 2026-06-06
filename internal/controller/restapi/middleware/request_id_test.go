package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alfariesh/backend-memora/internal/controller/restapi/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestIDMiddlewarePropagatesRequestID(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(middleware.RequestID())
	app.Get("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendString(middleware.RequestIDFromContext(ctx))
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	req.Header.Set(middleware.RequestIDHeader, "web-req-123")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "web-req-123", string(body))
	assert.Equal(t, "web-req-123", resp.Header.Get(middleware.RequestIDHeader))
	assert.Equal(t, "web-req-123", resp.Header.Get(middleware.CorrelationIDHeader))
}

func TestRequestIDMiddlewareUsesCorrelationIDAlias(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(middleware.RequestID())
	app.Get("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendString(middleware.RequestIDFromContext(ctx))
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	req.Header.Set(middleware.CorrelationIDHeader, "mobile:req:456")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, "mobile:req:456", string(body))
	assert.Equal(t, "mobile:req:456", resp.Header.Get(middleware.RequestIDHeader))
	assert.Equal(t, "mobile:req:456", resp.Header.Get(middleware.CorrelationIDHeader))
}

func TestRequestIDMiddlewareReplacesInvalidRequestID(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(middleware.RequestID())
	app.Get("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendString(middleware.RequestIDFromContext(ctx))
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	req.Header.Set(middleware.RequestIDHeader, "bad\nid")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	requestID := string(body)
	assert.NotEmpty(t, requestID)
	assert.NotEqual(t, "bad\nid", requestID)
	assert.NotContains(t, requestID, "\n")
	assert.Equal(t, requestID, resp.Header.Get(middleware.RequestIDHeader))
}

func TestLoggerIncludesRequestID(t *testing.T) {
	t.Parallel()

	log := &captureLogger{}
	app := fiber.New()
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger(log))
	app.Get("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(http.StatusNoContent)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	req.Header.Set(middleware.RequestIDHeader, "frontend-req-1")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	assert.Contains(t, log.message, "request_id=frontend-req-1")
}

type captureLogger struct {
	message string
}

func (l *captureLogger) Debug(any, ...any)   {}
func (l *captureLogger) Warn(string, ...any) {}
func (l *captureLogger) Error(any, ...any)   {}
func (l *captureLogger) Fatal(any, ...any)   {}

func (l *captureLogger) Info(message string, args ...any) {
	if len(args) == 0 {
		l.message = message

		return
	}

	l.message = strings.TrimSpace(message)
	if len(args) > 0 {
		l.message = strings.Replace(l.message, "%s", args[0].(string), 1)
	}
}
