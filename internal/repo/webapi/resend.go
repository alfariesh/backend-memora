package webapi

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/goccy/go-json"
)

const resendEndpoint = "https://api.resend.com/emails"

// ResendSender -.
type ResendSender struct {
	apiKey    string
	fromEmail string
	client    *http.Client
}

// NewResendSender -.
func NewResendSender(apiKey, fromEmail string) *ResendSender {
	return &ResendSender{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send -.
func (s *ResendSender) Send(ctx context.Context, to, subject, html string) (messageID string, err error) {
	if s.apiKey == "" || s.fromEmail == "" {
		return "", entity.ErrEmailSenderNotConfigured
	}

	payload := map[string]any{
		"from":    s.fromEmail,
		"to":      []string{to},
		"subject": subject,
		"html":    html,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("ResendSender - Send - json.Marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ResendSender - Send - http.NewRequestWithContext: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ResendSender - Send - s.client.Do: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("ResendSender - Send - resp.Body.Close: %w", closeErr)
		}
	}()

	var response struct {
		ID    string `json:"id"`
		Error any    `json:"error"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("ResendSender - Send - json.Decode: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("resend returned status %d: %v", resp.StatusCode, response.Error)
	}

	return response.ID, nil
}
