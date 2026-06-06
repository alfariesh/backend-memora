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

func TestResendEmailSenderNotConfigured(t *testing.T) {
	t.Parallel()

	sender := NewResendEmailSender("", "")

	_, err := sender.Send(context.Background(), "user@example.com", "Subject", "<p>Hello</p>", "reminder_job:job-id:email")

	assert.True(t, errors.Is(err, entity.ErrEmailSenderNotConfigured))
}

func TestResendEmailSenderSend(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "Bearer re_test", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "reminder_job:job-id:email", r.Header.Get("Idempotency-Key"))

		var payload struct {
			To      []string `json:"to"`
			From    string   `json:"from"`
			Subject string   `json:"subject"`
			HTML    string   `json:"html"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, []string{"user@example.com"}, payload.To)
		assert.Equal(t, "reminders@example.com", payload.From)
		assert.Equal(t, "Reminder", payload.Subject)
		assert.Equal(t, "<p>Hello</p>", payload.HTML)

		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write([]byte(`{"id":"email-id-123"}`))
		require.NoError(t, writeErr)
	}))
	defer server.Close()

	sender := NewResendEmailSender("re_test", "reminders@example.com")
	sender.endpoint = server.URL
	sender.client = server.Client()

	messageID, err := sender.Send(context.Background(), "user@example.com", "Reminder", "<p>Hello</p>", "reminder_job:job-id:email")

	require.NoError(t, err)
	assert.Equal(t, "email-id-123", messageID)
}

func TestResendEmailSenderIdempotencyError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_, writeErr := w.Write([]byte(`{"name":"invalid_idempotent_request","message":"same key used with a different payload"}`))
		require.NoError(t, writeErr)
	}))
	defer server.Close()

	sender := NewResendEmailSender("re_test", "reminders@example.com")
	sender.endpoint = server.URL
	sender.client = server.Client()

	_, err := sender.Send(context.Background(), "user@example.com", "Reminder", "<p>Hello</p>", "reminder_job:job-id:email")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid_idempotent_request")
}
