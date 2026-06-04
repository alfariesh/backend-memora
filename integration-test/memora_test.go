package integration_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	protov1 "github.com/evrone/go-clean-template/docs/proto/v1"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo/persistent"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
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

type deviceResponse struct {
	ID string `json:"id"`
}

type notificationResponse struct {
	ID     string     `json:"id"`
	Title  string     `json:"title"`
	ReadAt *time.Time `json:"read_at"`
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
	defer closeResponseBody(t, resp)

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
	defer closeResponseBody(t, resp)

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
	defer closeResponseBody(t, resp)

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
	defer closeResponseBody(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	rules := parseJSON[struct {
		Rules []struct {
			OffsetDays int      `json:"offset_days"`
			Channels   []string `json:"channels"`
		} `json:"rules"`
	}](t, resp)
	if len(rules.Rules) != 1 || rules.Rules[0].OffsetDays != 7 {
		t.Fatalf("unexpected replaced reminders: %+v", rules)
	}

	ctx, cancel = context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err = doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/important-days/"+created.ID+"/reminders", http.NoBody, token)
	if err != nil {
		t.Fatalf("Get reminders: %v", err)
	}
	defer closeResponseBody(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	rules = parseJSON[struct {
		Rules []struct {
			OffsetDays int      `json:"offset_days"`
			Channels   []string `json:"channels"`
		} `json:"rules"`
	}](t, resp)
	if len(rules.Rules) != 1 || rules.Rules[0].OffsetDays != 7 {
		t.Fatalf("unexpected reminders response: %+v", rules)
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
	defer closeResponseBody(t, resp)

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
	defer closeResponseBody(t, resp)

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
	defer closeResponseBody(t, resp)

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
	defer closeResponseBody(t, resp)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
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
	defer closeResponseBody(t, resp)

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
	defer closeResponseBody(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	ctx, cancel = context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err = doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/notifications/unread-count", http.NoBody, token)
	if err != nil {
		t.Fatalf("Count unread notifications: %v", err)
	}
	defer closeResponseBody(t, resp)

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
	defer closeResponseBody(t, resp)

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}

	ctx, cancel = context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err = doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/devices/", http.NoBody, token)
	if err != nil {
		t.Fatalf("List devices after delete: %v", err)
	}
	defer closeResponseBody(t, resp)

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

func TestHTTPNotificationReadFlowV1(t *testing.T) {
	token := registerAndLogin(t)
	user := httpGetProfile(t, token)
	pg := openIntegrationPostgres(t)
	notificationRepo := persistent.NewNotificationRepo(pg)
	now := time.Date(2026, 5, 13, 9, 0, 0, 0, time.UTC)

	first := entity.Notification{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		Type:      "important_day_reminder",
		Title:     "First reminder",
		Body:      "First reminder body.",
		Data:      `{"source":"integration"}`,
		CreatedAt: now,
	}
	second := entity.Notification{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		Type:      "important_day_reminder",
		Title:     "Second reminder",
		Body:      "Second reminder body.",
		Data:      `{"source":"integration"}`,
		CreatedAt: now.Add(time.Minute),
	}

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	if err := notificationRepo.Store(ctx, &first); err != nil {
		t.Fatalf("store first notification: %v", err)
	}
	if err := notificationRepo.Store(ctx, &second); err != nil {
		t.Fatalf("store second notification: %v", err)
	}

	assertUnreadNotificationCount(t, token, 2)

	resp := httpListNotifications(t, token, true)
	unread := parseJSON[struct {
		Notifications []notificationResponse `json:"notifications"`
		Total         int                    `json:"total"`
	}](t, resp)
	closeResponseBody(t, resp)

	if unread.Total != 2 || len(unread.Notifications) != 2 {
		t.Fatalf("expected 2 unread notifications, got %+v", unread)
	}
	if !hasNotificationID(unread.Notifications, first.ID) || !hasNotificationID(unread.Notifications, second.ID) {
		t.Fatalf("missing seeded notifications in unread list: %+v", unread.Notifications)
	}

	ctx, cancel = context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodPatch, basePathV1+"/notifications/"+first.ID+"/read", http.NoBody, token)
	if err != nil {
		t.Fatalf("Mark notification read: %v", err)
	}
	defer closeResponseBody(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	marked := parseJSON[notificationResponse](t, resp)
	if marked.ID != first.ID || marked.ReadAt == nil {
		t.Fatalf("unexpected marked notification: %+v", marked)
	}

	assertUnreadNotificationCount(t, token, 1)

	ctx, cancel = context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err = doAuthenticatedRequest(ctx, http.MethodPatch, basePathV1+"/notifications/read-all", http.NoBody, token)
	if err != nil {
		t.Fatalf("Mark all notifications read: %v", err)
	}
	defer closeResponseBody(t, resp)

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}

	assertUnreadNotificationCount(t, token, 0)

	resp = httpListNotifications(t, token, true)
	unread = parseJSON[struct {
		Notifications []notificationResponse `json:"notifications"`
		Total         int                    `json:"total"`
	}](t, resp)
	closeResponseBody(t, resp)

	if unread.Total != 0 || len(unread.Notifications) != 0 {
		t.Fatalf("expected no unread notifications, got %+v", unread)
	}
}

func TestHTTPMobileBootstrapV1(t *testing.T) {
	token := registerAndLogin(t)
	created := httpCreateImportantDay(t, token)

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodPost, basePathV1+"/devices/", bytes.NewBufferString(`{"token":"ExpoPushToken[bootstrap]","platform":"ios","name":"iPhone"}`), token)
	if err != nil {
		t.Fatalf("Register bootstrap device: %v", err)
	}
	defer closeResponseBody(t, resp)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	device := parseJSON[deviceResponse](t, resp)
	if device.ID == "" {
		t.Fatal("expected non-empty device id")
	}

	ctx, cancel = context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err = doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/mobile/bootstrap?upcoming_days=365&upcoming_limit=5", http.NoBody, token)
	if err != nil {
		t.Fatalf("Mobile bootstrap: %v", err)
	}
	defer closeResponseBody(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	bootstrap := parseJSON[struct {
		Settings struct {
			Timezone     string `json:"timezone"`
			ReminderTime string `json:"reminder_time"`
		} `json:"settings"`
		UpcomingImportantDays   []importantDayResponse `json:"upcoming_important_days"`
		UpcomingTotal           int                    `json:"upcoming_total"`
		UnreadNotificationCount int                    `json:"unread_notification_count"`
		Devices                 []deviceResponse       `json:"devices"`
		DevicesTotal            int                    `json:"devices_total"`
	}](t, resp)

	if bootstrap.Settings.Timezone != "Asia/Jakarta" || bootstrap.Settings.ReminderTime != "09:00" {
		t.Fatalf("unexpected settings: %+v", bootstrap.Settings)
	}

	if bootstrap.UpcomingTotal != 1 || len(bootstrap.UpcomingImportantDays) != 1 || bootstrap.UpcomingImportantDays[0].ID != created.ID {
		t.Fatalf("unexpected upcoming important days: %+v", bootstrap)
	}

	if bootstrap.UnreadNotificationCount != 0 {
		t.Fatalf("expected unread count 0, got %d", bootstrap.UnreadNotificationCount)
	}

	if bootstrap.DevicesTotal != 1 || len(bootstrap.Devices) != 1 || bootstrap.Devices[0].ID != device.ID {
		t.Fatalf("unexpected devices: %+v", bootstrap.Devices)
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
	defer closeResponseBody(t, resp)

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
	defer closeResponseBody(t, resp)

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestGRPCImportantDaysV1(t *testing.T) {
	token := registerAndLoginGRPC(t)
	client, conn := grpcImportantDayClient(t)
	defer closeGRPCConn(t, conn)

	authCtx := grpcAuthCtx(t, token)
	eventYear := int32(1970)

	created, err := client.CreateImportantDay(authCtx, &protov1.CreateImportantDayRequest{
		Title:        "Mom birthday",
		Type:         "birthday",
		PersonName:   "Mom",
		Relationship: "mother",
		EventYear:    &eventYear,
		EventMonth:   5,
		EventDay:     13,
		Timezone:     "Asia/Jakarta",
		ReminderTime: "08:30",
		ReminderRules: []*protov1.ReminderRuleRequest{
			{OffsetDays: 3, Channels: []string{"email", "in_app"}},
		},
	})
	if err != nil {
		t.Fatalf("CreateImportantDay: %v", err)
	}
	if created.GetId() == "" {
		t.Fatal("expected non-empty important day id")
	}
	if created.GetTitle() != "Mom birthday" || created.GetType() != "birthday" || created.GetEventYear() != eventYear {
		t.Fatalf("unexpected created important day: %+v", created)
	}

	got, err := client.GetImportantDay(authCtx, &protov1.GetImportantDayRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetImportantDay: %v", err)
	}
	if got.GetId() != created.GetId() {
		t.Fatalf("expected id %q, got %q", created.GetId(), got.GetId())
	}

	listed, err := client.ListImportantDays(authCtx, &protov1.ListImportantDaysRequest{
		Type:   "birthday",
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListImportantDays: %v", err)
	}
	if listed.GetTotal() < 1 || len(listed.GetImportantDays()) < 1 {
		t.Fatalf("expected at least one important day, got %+v", listed)
	}

	upcoming, err := client.UpcomingImportantDays(authCtx, &protov1.UpcomingImportantDaysRequest{
		Days:   730,
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("UpcomingImportantDays: %v", err)
	}
	if upcoming.GetTotal() < 1 || len(upcoming.GetImportantDays()) < 1 {
		t.Fatalf("expected at least one upcoming important day, got %+v", upcoming)
	}

	updated, err := client.UpdateImportantDay(authCtx, &protov1.UpdateImportantDayRequest{
		Id:           created.GetId(),
		Title:        "Mom birthday updated",
		Type:         "birthday",
		PersonName:   "Mom",
		Relationship: "mother",
		EventYear:    &eventYear,
		EventMonth:   5,
		EventDay:     14,
		Timezone:     "Asia/Jakarta",
		ReminderTime: "09:15",
	})
	if err != nil {
		t.Fatalf("UpdateImportantDay: %v", err)
	}
	if updated.GetTitle() != "Mom birthday updated" || updated.GetEventDay() != 14 || updated.GetReminderTime() != "09:15" {
		t.Fatalf("unexpected updated important day: %+v", updated)
	}

	rules, err := client.ReplaceImportantDayReminders(authCtx, &protov1.ReplaceReminderRulesRequest{
		Id: created.GetId(),
		Rules: []*protov1.ReminderRuleRequest{
			{OffsetDays: 7, Channels: []string{"in_app", "push"}},
		},
	})
	if err != nil {
		t.Fatalf("ReplaceImportantDayReminders: %v", err)
	}
	if len(rules.GetRules()) != 1 || rules.GetRules()[0].GetOffsetDays() != 7 {
		t.Fatalf("unexpected reminder rules: %+v", rules)
	}

	_, err = client.DeleteImportantDay(authCtx, &protov1.DeleteImportantDayRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("DeleteImportantDay: %v", err)
	}

	_, err = client.GetImportantDay(authCtx, &protov1.GetImportantDayRequest{Id: created.GetId()})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound after delete, got %v", err)
	}
}

func TestGRPCImportantDayValidationV1(t *testing.T) {
	token := registerAndLoginGRPC(t)
	client, conn := grpcImportantDayClient(t)
	defer closeGRPCConn(t, conn)

	authCtx := grpcAuthCtx(t, token)

	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "missing title",
			call: func() error {
				_, err := client.CreateImportantDay(authCtx, &protov1.CreateImportantDayRequest{
					EventMonth: 1,
					EventDay:   2,
				})

				return err
			},
		},
		{
			name: "invalid month",
			call: func() error {
				_, err := client.CreateImportantDay(authCtx, &protov1.CreateImportantDayRequest{
					Title:      "Invalid month",
					EventMonth: 13,
					EventDay:   2,
				})

				return err
			},
		},
		{
			name: "invalid reminder channel",
			call: func() error {
				_, err := client.CreateImportantDay(authCtx, &protov1.CreateImportantDayRequest{
					Title:      "Invalid channel",
					EventMonth: 1,
					EventDay:   2,
					ReminderRules: []*protov1.ReminderRuleRequest{
						{OffsetDays: 1, Channels: []string{"sms"}},
					},
				})

				return err
			},
		},
		{
			name: "missing update id",
			call: func() error {
				_, err := client.UpdateImportantDay(authCtx, &protov1.UpdateImportantDayRequest{
					Title:      "Missing id",
					EventMonth: 1,
					EventDay:   2,
				})

				return err
			},
		},
		{
			name: "invalid reminder offset",
			call: func() error {
				_, err := client.ReplaceImportantDayReminders(authCtx, &protov1.ReplaceReminderRulesRequest{
					Id: "missing",
					Rules: []*protov1.ReminderRuleRequest{
						{OffsetDays: -1, Channels: []string{"in_app"}},
					},
				})

				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			if status.Code(err) != codes.InvalidArgument {
				t.Fatalf("expected InvalidArgument, got %v", err)
			}
		})
	}
}

func TestGRPCDeviceValidationV1(t *testing.T) {
	token := registerAndLoginGRPC(t)

	grpcConn, err := grpc.NewClient(grpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc.NewClient: %v", err)
	}
	defer closeGRPCConn(t, grpcConn)

	client := protov1.NewDeviceServiceClient(grpcConn)
	authCtx := grpcAuthCtx(t, token)

	_, err = client.RegisterDevice(authCtx, &protov1.RegisterDeviceRequest{
		Platform: "ios",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected missing token InvalidArgument, got %v", err)
	}

	_, err = client.RegisterDevice(authCtx, &protov1.RegisterDeviceRequest{
		Token:    "not-an-expo-token",
		Platform: "ios",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected invalid token InvalidArgument, got %v", err)
	}

	registered, err := client.RegisterDevice(authCtx, &protov1.RegisterDeviceRequest{
		Token:    "ExpoPushToken[grpc-device]",
		Platform: "ios",
		Name:     "iPhone",
	})
	if err != nil {
		t.Fatalf("RegisterDevice: %v", err)
	}
	if registered.GetId() == "" || !registered.GetActive() {
		t.Fatalf("unexpected registered device: %+v", registered)
	}

	_, err = client.DeleteDevice(authCtx, &protov1.DeleteDeviceRequest{Id: registered.GetId()})
	if err != nil {
		t.Fatalf("DeleteDevice: %v", err)
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
	defer closeResponseBody(t, resp)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	return parseJSON[importantDayResponse](t, resp)
}

func httpGetProfile(t *testing.T, token string) struct {
	ID string `json:"id"`
} {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/user/profile", http.NoBody, token)
	if err != nil {
		t.Fatalf("Get profile: %v", err)
	}
	defer closeResponseBody(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	return parseJSON[struct {
		ID string `json:"id"`
	}](t, resp)
}

func httpListNotifications(t *testing.T, token string, unreadOnly bool) *http.Response {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	url := basePathV1 + "/notifications/?limit=10&offset=0"
	if unreadOnly {
		url += "&unread_only=true"
	}

	resp, err := doAuthenticatedRequest(ctx, http.MethodGet, url, http.NoBody, token)
	if err != nil {
		t.Fatalf("List notifications: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer closeResponseBody(t, resp)
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	return resp
}

func assertUnreadNotificationCount(t *testing.T, token string, expected int) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/notifications/unread-count", http.NoBody, token)
	if err != nil {
		t.Fatalf("Count unread notifications: %v", err)
	}
	defer closeResponseBody(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	unreadCount := parseJSON[struct {
		UnreadCount int `json:"unread_count"`
	}](t, resp)
	if unreadCount.UnreadCount != expected {
		t.Fatalf("expected unread count %d, got %d", expected, unreadCount.UnreadCount)
	}
}

func hasNotificationID(notifications []notificationResponse, id string) bool {
	for _, notification := range notifications {
		if notification.ID == id {
			return true
		}
	}

	return false
}

func grpcImportantDayClient(t *testing.T) (protov1.ImportantDayServiceClient, *grpc.ClientConn) {
	t.Helper()

	grpcConn, err := grpc.NewClient(grpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc.NewClient: %v", err)
	}

	return protov1.NewImportantDayServiceClient(grpcConn), grpcConn
}

func closeGRPCConn(t *testing.T, grpcConn *grpc.ClientConn) {
	t.Helper()

	if err := grpcConn.Close(); err != nil {
		t.Fatalf("grpcConn.Close: %v", err)
	}
}
