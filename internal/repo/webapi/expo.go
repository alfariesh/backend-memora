package webapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alfariesh/backend-memora/internal/entity"
	"github.com/goccy/go-json"
)

const expoPushEndpoint = "https://exp.host/--/api/v2/push/send"

// ExpoPushSender sends notifications through Expo Push Service.
type ExpoPushSender struct {
	accessToken string
	client      *http.Client
}

// NewExpoPushSender -.
func NewExpoPushSender(accessToken string) *ExpoPushSender {
	return &ExpoPushSender{
		accessToken: accessToken,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send -.
func (s *ExpoPushSender) Send(ctx context.Context, token, title, body string, data map[string]string) (ticketID string, err error) {
	if token == "" {
		return "", errors.New("expo push token is required")
	}

	payload := map[string]any{
		"to":    token,
		"title": title,
		"body":  body,
		"data":  data,
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("ExpoPushSender - Send - json.Marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, expoPushEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return "", fmt.Errorf("ExpoPushSender - Send - http.NewRequestWithContext: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if s.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.accessToken)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ExpoPushSender - Send - s.client.Do: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("ExpoPushSender - Send - resp.Body.Close: %w", closeErr)
		}
	}()

	var response struct {
		Data   json.RawMessage `json:"data"`
		Errors []expoError     `json:"errors"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("ExpoPushSender - Send - json.Decode: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("expo returned status %d: %s", resp.StatusCode, expoErrorsMessage(response.Errors))
	}

	if len(response.Errors) > 0 {
		return "", errors.New(expoErrorsMessage(response.Errors))
	}

	ticket, err := parseExpoTicket(response.Data)
	if err != nil {
		return "", fmt.Errorf("ExpoPushSender - Send - parseExpoTicket: %w", err)
	}

	if ticket.Status == "error" {
		return "", expoTicketError(ticket)
	}

	return ticket.ID, nil
}

type expoTicket struct {
	Status  string         `json:"status"`
	ID      string         `json:"id"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
}

type expoError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
}

func parseExpoTicket(raw json.RawMessage) (expoTicket, error) {
	if len(raw) == 0 {
		return expoTicket{}, errors.New("empty expo response data")
	}

	var ticket expoTicket
	if err := json.Unmarshal(raw, &ticket); err == nil && ticket.Status != "" {
		return ticket, nil
	}

	var tickets []expoTicket
	if err := json.Unmarshal(raw, &tickets); err != nil {
		return expoTicket{}, err
	}

	if len(tickets) == 0 {
		return expoTicket{}, errors.New("empty expo ticket list")
	}

	return tickets[0], nil
}

func expoErrorsMessage(errors []expoError) string {
	if len(errors) == 0 {
		return "unknown expo push error"
	}

	message := errors[0].Message
	if errors[0].Code != "" {
		message = errors[0].Code + ": " + message
	}

	return message
}

func expoTicketError(ticket expoTicket) error {
	if ticket.Details["error"] == "DeviceNotRegistered" {
		return fmt.Errorf("%w: %s", entity.ErrPushDeviceNotRegistered, ticket.Message)
	}

	return fmt.Errorf("expo ticket error: %s", ticket.Message)
}
