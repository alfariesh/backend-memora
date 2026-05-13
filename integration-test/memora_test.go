package integration_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"
)

type importantDayResponse struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Type       string `json:"type"`
	EventMonth int    `json:"event_month"`
	EventDay   int    `json:"event_day"`
}

func TestHTTPImportantDaysV1(t *testing.T) {
	token := registerAndLogin(t)

	createBody := `{
		"title":"Mom birthday",
		"type":"birthday",
		"person_name":"Mom",
		"relationship":"mother",
		"event_year":1970,
		"event_month":5,
		"event_day":13,
		"reminder_rules":[{"offset_days":1,"channels":["in_app"]}]
	}`

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodPost, basePathV1+"/important-days/", bytes.NewBufferString(createBody), token)
	if err != nil {
		t.Fatalf("Create important day: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	created := parseJSON[importantDayResponse](t, resp)
	if created.ID == "" {
		t.Fatal("expected non-empty id")
	}

	ctx, cancel = context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err = doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/important-days/?limit=10&offset=0", http.NoBody, token)
	if err != nil {
		t.Fatalf("List important days: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	type listResponse struct {
		ImportantDays []importantDayResponse `json:"important_days"`
		Total         int                    `json:"total"`
	}

	listed := parseJSON[listResponse](t, resp)
	if listed.Total < 1 {
		t.Fatalf("expected total >= 1, got %d", listed.Total)
	}

	ctx, cancel = context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err = doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/important-days/upcoming?days=365&limit=10&offset=0", http.NoBody, token)
	if err != nil {
		t.Fatalf("Upcoming important days: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHTTPImportantDayRemindersV1(t *testing.T) {
	token := registerAndLogin(t)
	created := httpCreateImportantDay(t, token)

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	body := `{"rules":[{"offset_days":7,"channels":["email","in_app","push"]}]}`
	resp, err := doAuthenticatedRequest(ctx, http.MethodPut, basePathV1+"/important-days/"+created.ID+"/reminders", bytes.NewBufferString(body), token)
	if err != nil {
		t.Fatalf("Replace reminders: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHTTPDevicesAndNotificationsV1(t *testing.T) {
	token := registerAndLogin(t)

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodPost, basePathV1+"/devices/", bytes.NewBufferString(`{"token":"ExpoPushToken[test]","platform":"android","name":"test"}`), token)
	if err != nil {
		t.Fatalf("Register device: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	type deviceResponse struct {
		ID string `json:"id"`
	}

	device := parseJSON[deviceResponse](t, resp)
	if device.ID == "" {
		t.Fatal("expected non-empty device id")
	}

	ctx, cancel = context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err = doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/notifications/?limit=10&offset=0", http.NoBody, token)
	if err != nil {
		t.Fatalf("List notifications: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	ctx, cancel = context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err = doAuthenticatedRequest(ctx, http.MethodDelete, basePathV1+"/devices/"+device.ID, http.NoBody, token)
	if err != nil {
		t.Fatalf("Delete device: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

func TestHTTPRegisterDeviceRejectsInvalidExpoTokenV1(t *testing.T) {
	token := registerAndLogin(t)

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodPost, basePathV1+"/devices/", bytes.NewBufferString(`{"token":"fcm-token","platform":"android"}`), token)
	if err != nil {
		t.Fatalf("Register invalid device: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func httpCreateImportantDay(t *testing.T, token string) importantDayResponse {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodPost, basePathV1+"/important-days/", bytes.NewBufferString(`{"title":"Anniversary","type":"wedding","event_year":2020,"event_month":1,"event_day":2}`), token)
	if err != nil {
		t.Fatalf("Create important day: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	return parseJSON[importantDayResponse](t, resp)
}
