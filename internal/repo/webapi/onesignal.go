package webapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/goccy/go-json"
)

const oneSignalPushEndpoint = "https://api.onesignal.com/notifications?c=push"

// OneSignalPushSender sends push notifications through OneSignal.
type OneSignalPushSender struct {
	appID      string
	restAPIKey string
	endpoint   string
	client     *http.Client
}

// NewOneSignalPushSender -.
func NewOneSignalPushSender(appID, restAPIKey string) *OneSignalPushSender {
	return &OneSignalPushSender{
		appID:      appID,
		restAPIKey: restAPIKey,
		endpoint:   oneSignalPushEndpoint,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send -.
func (s *OneSignalPushSender) Send(ctx context.Context, token, title, body string, data map[string]string, idempotencyKey string) (messageID string, err error) {
	if s.appID == "" || s.restAPIKey == "" {
		return "", entity.ErrPushSenderNotConfigured
	}
	if token == "" {
		return "", errors.New("onesignal subscription id is required")
	}

	payload := map[string]any{
		"app_id":                   s.appID,
		"target_channel":           "push",
		"include_subscription_ids": []string{token},
		"headings":                 map[string]string{"en": title},
		"contents":                 map[string]string{"en": body},
		"data":                     data,
	}
	if idempotencyKey != "" {
		payload["idempotency_key"] = idempotencyKey
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("OneSignalPushSender - Send - json.Marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return "", fmt.Errorf("OneSignalPushSender - Send - http.NewRequestWithContext: %w", err)
	}

	req.Header.Set("Authorization", "Key "+s.restAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("OneSignalPushSender - Send - s.client.Do: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("OneSignalPushSender - Send - resp.Body.Close: %w", closeErr)
		}
	}()

	var response oneSignalPushResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("OneSignalPushSender - Send - json.Decode: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("onesignal returned status %d: %s", resp.StatusCode, response.errorString())
	}

	if response.invalidSubscription(token) {
		return "", fmt.Errorf("%w: %s", entity.ErrPushDeviceNotRegistered, token)
	}

	if response.ID == "" {
		return "", fmt.Errorf("onesignal returned empty id: %s", response.errorString())
	}

	return response.ID, nil
}

type oneSignalPushResponse struct {
	ID         string               `json:"id"`
	ExternalID string               `json:"external_id"`
	Errors     oneSignalPushErrors  `json:"errors"`
	Warnings   oneSignalPushWarning `json:"warnings"`
}

type oneSignalPushErrors struct {
	InvalidPlayerIDs []string `json:"invalid_player_ids"`
}

type oneSignalPushWarning struct {
	InvalidExternalUserIDs string `json:"invalid_external_user_ids"`
}

func (r oneSignalPushResponse) invalidSubscription(token string) bool {
	for _, invalidID := range r.Errors.InvalidPlayerIDs {
		if invalidID == token {
			return true
		}
	}

	return false
}

func (r oneSignalPushResponse) errorString() string {
	parts := make([]string, 0, 2)
	if len(r.Errors.InvalidPlayerIDs) > 0 {
		parts = append(parts, "invalid player ids: "+strings.Join(r.Errors.InvalidPlayerIDs, ", "))
	}
	if r.Warnings.InvalidExternalUserIDs != "" {
		parts = append(parts, "invalid external user ids: "+r.Warnings.InvalidExternalUserIDs)
	}
	if len(parts) == 0 {
		return "unknown error"
	}

	return strings.Join(parts, "; ")
}
