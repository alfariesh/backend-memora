package webapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudflareEmailSenderNotConfigured(t *testing.T) {
	t.Parallel()

	sender := NewCloudflareEmailSender("", "", "")

	_, err := sender.Send(context.Background(), "user@example.com", "Subject", "<p>Hello</p>")

	assert.True(t, errors.Is(err, entity.ErrEmailSenderNotConfigured))
}

func TestCloudflareEmailSenderSend(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var payload struct {
			To      string `json:"to"`
			From    string `json:"from"`
			Subject string `json:"subject"`
			HTML    string `json:"html"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "user@example.com", payload.To)
		assert.Equal(t, "reminders@example.com", payload.From)
		assert.Equal(t, "Reminder", payload.Subject)
		assert.Equal(t, "<p>Hello</p>", payload.HTML)

		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":{"delivered":["user@example.com"],"permanent_bounces":[],"queued":[]}}`))
		require.NoError(t, writeErr)
	}))
	defer server.Close()

	sender := NewCloudflareEmailSender("account-id", "token", "reminders@example.com")
	sender.endpoint = server.URL
	sender.client = server.Client()

	messageID, err := sender.Send(context.Background(), "user@example.com", "Reminder", "<p>Hello</p>")

	require.NoError(t, err)
	assert.Equal(t, "user@example.com", messageID)
}

func TestCloudflareEmailSenderAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, writeErr := w.Write([]byte(`{"success":false,"errors":[{"code":10001,"message":"email.sending.error.invalid_request_schema"}],"messages":[],"result":null}`))
		require.NoError(t, writeErr)
	}))
	defer server.Close()

	sender := NewCloudflareEmailSender("account-id", "token", "reminders@example.com")
	sender.endpoint = server.URL
	sender.client = server.Client()

	_, err := sender.Send(context.Background(), "user@example.com", "Reminder", "<p>Hello</p>")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "10001 email.sending.error.invalid_request_schema")
}

func TestCloudflareEmailSenderPermanentBounce(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":{"delivered":[],"permanent_bounces":["user@example.com"],"queued":[]}}`))
		require.NoError(t, writeErr)
	}))
	defer server.Close()

	sender := NewCloudflareEmailSender("account-id", "token", "reminders@example.com")
	sender.endpoint = server.URL
	sender.client = server.Client()

	_, err := sender.Send(context.Background(), "user@example.com", "Reminder", "<p>Hello</p>")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "permanent bounces")
}
