package integration_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"
)

type importantDayResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Type         string `json:"type"`
	EventMonth   int    `json:"event_month"`
	EventDay     int    `json:"event_day"`
	Timezone     string `json:"timezone"`
	ReminderTime string `json:"reminder_time"`
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

func TestHTTPUserSettingsV1(t *testing.T) {
	token := registerAndLogin(t)

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/user/settings", http.NoBody, token)
	if err != nil {
		t.Fatalf("Get user settings: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	type settingsResponse struct {
		Timezone             string   `json:"timezone"`
		ReminderTime         string   `json:"reminder_time"`
		NotificationChannels []string `json:"notification_channels"`
	}

	defaults := parseJSON[settingsResponse](t, resp)
	if defaults.Timezone != "Asia/Jakarta" || defaults.ReminderTime != "09:00" {
		t.Fatalf("unexpected default settings: %+v", defaults)
	}

	ctx, cancel = context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	updateBody := `{"timezone":"Asia/Makassar","reminder_time":"08:30","notification_channels":["in_app","push"]}`
	resp, err = doAuthenticatedRequest(ctx, http.MethodPut, basePathV1+"/user/settings", bytes.NewBufferString(updateBody), token)
	if err != nil {
		t.Fatalf("Update user settings: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	updated := parseJSON[settingsResponse](t, resp)
	if updated.Timezone != "Asia/Makassar" || updated.ReminderTime != "08:30" {
		t.Fatalf("unexpected updated settings: %+v", updated)
	}
	if len(updated.NotificationChannels) != 2 || updated.NotificationChannels[0] != "in_app" || updated.NotificationChannels[1] != "push" {
		t.Fatalf("unexpected notification channels: %+v", updated.NotificationChannels)
	}

	created := httpCreateImportantDay(t, token)
	if created.Timezone != "Asia/Makassar" || created.ReminderTime != "08:30" {
		t.Fatalf("expected important day to use user settings, got timezone=%s reminder_time=%s", created.Timezone, created.ReminderTime)
	}

	ctx, cancel = context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err = doAuthenticatedRequest(ctx, http.MethodPut, basePathV1+"/user/settings", bytes.NewBufferString(`{"timezone":"Invalid/Zone","reminder_time":"08:30","notification_channels":["in_app"]}`), token)
	if err != nil {
		t.Fatalf("Update invalid user settings: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
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

	resp, err = doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/devices/", http.NoBody, token)
	if err != nil {
		t.Fatalf("List devices: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	devices := parseJSON[struct {
		Devices []deviceResponse `json:"devices"`
		Total   int              `json:"total"`
	}](t, resp)
	if devices.Total != 1 || len(devices.Devices) != 1 || devices.Devices[0].ID != device.ID {
		t.Fatalf("unexpected devices response: %+v", devices)
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

	resp, err = doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/notifications/unread-count", http.NoBody, token)
	if err != nil {
		t.Fatalf("Count unread notifications: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	unreadCount := parseJSON[struct {
		UnreadCount int `json:"unread_count"`
	}](t, resp)
	if unreadCount.UnreadCount != 0 {
		t.Fatalf("expected unread count 0, got %d", unreadCount.UnreadCount)
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

	ctx, cancel = context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err = doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/devices/", http.NoBody, token)
	if err != nil {
		t.Fatalf("List devices after delete: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	devices = parseJSON[struct {
		Devices []deviceResponse `json:"devices"`
		Total   int              `json:"total"`
	}](t, resp)
	if devices.Total != 0 || len(devices.Devices) != 0 {
		t.Fatalf("expected no active devices, got %+v", devices)
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

func TestHTTPTestPushDeviceNotFoundV1(t *testing.T) {
	token := registerAndLogin(t)

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodPost, basePathV1+"/devices/missing-device/test-push", bytes.NewBufferString(`{}`), token)
	if err != nil {
		t.Fatalf("Test push missing device: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
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
