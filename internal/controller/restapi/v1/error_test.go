package v1

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alfariesh/backend-memora/internal/controller/restapi/v1/request"
	"github.com/alfariesh/backend-memora/internal/controller/restapi/v1/response"
	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationErrorResponseUsesJSONFieldNames(t *testing.T) {
	t.Parallel()

	v := newValidator()
	err := v.Struct(request.Register{
		Username: "ab",
		Email:    "not-email",
	})
	require.Error(t, err)

	result := performValidationErrorResponse(t, err)

	assert.Equal(t, "validation_error", result.Error)
	assert.Equal(t, "validation failed", result.Message)
	assert.Equal(t, "must be at least 3 characters", result.Fields["username"])
	assert.Equal(t, "must be a valid email", result.Fields["email"])
	assert.Equal(t, "is required", result.Fields["password"])
}

func TestValidationErrorResponseUsesNestedJSONFieldNames(t *testing.T) {
	t.Parallel()

	v := newValidator()
	err := v.Struct(request.CreateImportantDay{
		Title:      "Mom birthday",
		EventMonth: 13,
		EventDay:   0,
		ReminderRules: []request.ReminderRule{
			{
				OffsetDays: -1,
				Channels:   []entity.ReminderChannel{"sms"},
			},
		},
	})
	require.Error(t, err)

	result := performValidationErrorResponse(t, err)

	assert.Equal(t, "validation_error", result.Error)
	assert.Equal(t, "must be at most 12", result.Fields["event_month"])
	assert.Equal(t, "is required", result.Fields["event_day"])
	assert.Equal(t, "must be at least 0", result.Fields["reminder_rules[0].offset_days"])
	assert.Equal(t, "must be one of: email, in_app, push", result.Fields["reminder_rules[0].channels[0]"])
}

func TestErrorResponseAddsMachineReadableCode(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", func(ctx *fiber.Ctx) error {
		return errorResponse(ctx, http.StatusConflict, "user already exists")
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	var result response.Error
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	assert.Equal(t, "user_already_exists", result.Error)
	assert.Equal(t, "user already exists", result.Message)
	assert.Empty(t, result.Fields)
}

func performValidationErrorResponse(t *testing.T, validationErr error) response.Error {
	t.Helper()

	app := fiber.New()
	app.Get("/test", func(ctx *fiber.Ctx) error {
		return validationErrorResponse(ctx, validationErr)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", strings.NewReader(""))
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	var result response.Error
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	return result
}
