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

const cloudflareEmailEndpointTemplate = "https://api.cloudflare.com/client/v4/accounts/%s/email/sending/send"

// CloudflareEmailSender sends email through Cloudflare Email Service.
type CloudflareEmailSender struct {
	accountID string
	apiToken  string
	fromEmail string
	endpoint  string
	client    *http.Client
}

// NewCloudflareEmailSender -.
func NewCloudflareEmailSender(accountID, apiToken, fromEmail string) *CloudflareEmailSender {
	return &CloudflareEmailSender{
		accountID: accountID,
		apiToken:  apiToken,
		fromEmail: fromEmail,
		endpoint:  fmt.Sprintf(cloudflareEmailEndpointTemplate, accountID),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send -.
func (s *CloudflareEmailSender) Send(ctx context.Context, to, subject, html string) (messageID string, err error) {
	if s.accountID == "" || s.apiToken == "" || s.fromEmail == "" {
		return "", entity.ErrEmailSenderNotConfigured
	}

	payload := map[string]any{
		"from":    s.fromEmail,
		"to":      to,
		"subject": subject,
		"html":    html,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("CloudflareEmailSender - Send - json.Marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("CloudflareEmailSender - Send - http.NewRequestWithContext: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("CloudflareEmailSender - Send - s.client.Do: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("CloudflareEmailSender - Send - resp.Body.Close: %w", closeErr)
		}
	}()

	var response cloudflareEmailResponse

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("CloudflareEmailSender - Send - json.Decode: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices || !response.Success {
		return "", fmt.Errorf("cloudflare email returned status %d: %s", resp.StatusCode, response.errorString())
	}

	if response.Result == nil {
		return "", fmt.Errorf("cloudflare email returned empty result")
	}

	if len(response.Result.PermanentBounces) > 0 {
		return "", fmt.Errorf("cloudflare email permanent bounces: %s", strings.Join(response.Result.PermanentBounces, ", "))
	}

	if len(response.Result.Delivered) > 0 {
		return response.Result.Delivered[0], nil
	}

	if len(response.Result.Queued) > 0 {
		return response.Result.Queued[0], nil
	}

	return to, nil
}

type cloudflareEmailResponse struct {
	Success  bool                   `json:"success"`
	Errors   []cloudflareEmailError `json:"errors"`
	Messages []json.RawMessage      `json:"messages"`
	Result   *struct {
		Delivered        []string `json:"delivered"`
		PermanentBounces []string `json:"permanent_bounces"`
		Queued           []string `json:"queued"`
	} `json:"result"`
}

type cloudflareEmailError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (r cloudflareEmailResponse) errorString() string {
	if len(r.Errors) == 0 {
		return "unknown error"
	}

	parts := make([]string, len(r.Errors))
	for i, err := range r.Errors {
		parts[i] = fmt.Sprintf("%d %s", err.Code, err.Message)
	}

	return strings.Join(parts, "; ")
}
