package webapi

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/goccy/go-json"
)

const resendEmailEndpoint = "https://api.resend.com/emails"

// ResendEmailSender sends email through Resend.
type ResendEmailSender struct {
	apiKey    string
	fromEmail string
	endpoint  string
	client    *http.Client
}

// NewResendEmailSender -.
func NewResendEmailSender(apiKey, fromEmail string) *ResendEmailSender {
	return &ResendEmailSender{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		endpoint:  resendEmailEndpoint,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send -.
func (s *ResendEmailSender) Send(ctx context.Context, to, subject, html, idempotencyKey string) (messageID string, err error) {
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
		return "", fmt.Errorf("ResendEmailSender - Send - json.Marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ResendEmailSender - Send - http.NewRequestWithContext: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ResendEmailSender - Send - s.client.Do: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("ResendEmailSender - Send - resp.Body.Close: %w", closeErr)
		}
	}()

	var response resendEmailResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("ResendEmailSender - Send - json.Decode: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("resend email returned status %d: %s", resp.StatusCode, response.errorString())
	}

	if response.ID == "" {
		return "", fmt.Errorf("resend email returned empty id")
	}

	return response.ID, nil
}

type resendEmailResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

func (r resendEmailResponse) errorString() string {
	parts := make([]string, 0, 2)
	if r.Name != "" {
		parts = append(parts, r.Name)
	}
	if r.Message != "" {
		parts = append(parts, r.Message)
	}
	if len(parts) == 0 {
		return "unknown error"
	}

	return strings.Join(parts, ": ")
}
