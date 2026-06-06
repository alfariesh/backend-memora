package webapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOneSignalPushSenderNotConfigured(t *testing.T) {
	t.Parallel()

	sender := NewOneSignalPushSender("", "")

	_, err := sender.Send(context.Background(), "11111111-1111-4111-8111-111111111111", "Title", "Body", nil, "idem-key")

	assert.True(t, errors.Is(err, entity.ErrPushSenderNotConfigured))
}

func TestOneSignalPushSenderSend(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "Key rest-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var payload struct {
			AppID                  string            `json:"app_id"`
			TargetChannel          string            `json:"target_channel"`
			IncludeSubscriptionIDs []string          `json:"include_subscription_ids"`
			Headings               map[string]string `json:"headings"`
			Contents               map[string]string `json:"contents"`
			Data                   map[string]string `json:"data"`
			IdempotencyKey         string            `json:"idempotency_key"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "app-id", payload.AppID)
		assert.Equal(t, "push", payload.TargetChannel)
		assert.Equal(t, []string{"11111111-1111-4111-8111-111111111111"}, payload.IncludeSubscriptionIDs)
		assert.Equal(t, "Title", payload.Headings["en"])
		assert.Equal(t, "Body", payload.Contents["en"])
		assert.Equal(t, "important_day_reminder", payload.Data["type"])
		assert.Equal(t, "idem-key", payload.IdempotencyKey)

		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write([]byte(`{"id":"22222222-2222-4222-8222-222222222222","external_id":"idem-key"}`))
		require.NoError(t, writeErr)
	}))
	defer server.Close()

	sender := NewOneSignalPushSender("app-id", "rest-key")
	sender.endpoint = server.URL
	sender.client = server.Client()

	messageID, err := sender.Send(
		context.Background(),
		"11111111-1111-4111-8111-111111111111",
		"Title",
		"Body",
		map[string]string{"type": "important_day_reminder"},
		"idem-key",
	)

	require.NoError(t, err)
	assert.Equal(t, "22222222-2222-4222-8222-222222222222", messageID)
}

func TestOneSignalPushSenderDeviceNotRegistered(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write([]byte(`{"id":"","errors":{"invalid_player_ids":["11111111-1111-4111-8111-111111111111"]}}`))
		require.NoError(t, writeErr)
	}))
	defer server.Close()

	sender := NewOneSignalPushSender("app-id", "rest-key")
	sender.endpoint = server.URL
	sender.client = server.Client()

	_, err := sender.Send(context.Background(), "11111111-1111-4111-8111-111111111111", "Title", "Body", nil, "idem-key")

	require.Error(t, err)
	assert.True(t, errors.Is(err, entity.ErrPushDeviceNotRegistered))
}
