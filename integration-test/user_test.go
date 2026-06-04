package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"

	protov1 "github.com/alfariesh/backend-memora/docs/proto/v1"
	natsClient "github.com/alfariesh/backend-memora/pkg/nats/nats_rpc/client"
	rmqClient "github.com/alfariesh/backend-memora/pkg/rabbitmq/rmq_rpc/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// HTTP POST: /v1/auth/register.
func TestHTTPRegisterV1(t *testing.T) {
	// Pre-register a user for the duplicate test case.
	name := sanitizeTestName(t)
	dupEmail := name + "_dup@test.com"
	dupUser := name + "_dup"

	resp := registerUser(t, dupUser, dupEmail, testPassword)
	closeResponseBody(t, resp)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("pre-register: expected 201, got %d", resp.StatusCode)
	}

	tests := []struct {
		description string
		username    string
		email       string
		password    string
		expected    int
	}{
		{
			description: "success",
			username:    name + "_ok",
			email:       name + "_ok@test.com",
			password:    testPassword,
			expected:    http.StatusCreated,
		},
		{
			description: "duplicate email",
			username:    name + "_dup2",
			email:       dupEmail,
			password:    testPassword,
			expected:    http.StatusConflict,
		},
		{
			description: "missing password",
			username:    name + "_nopw",
			email:       name + "_nopw@test.com",
			password:    "",
			expected:    http.StatusBadRequest,
		},
		{
			description: "short username",
			username:    "ab",
			email:       name + "_short@test.com",
			password:    testPassword,
			expected:    http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			resp := registerUser(t, tt.username, tt.email, tt.password)
			defer closeResponseBody(t, resp)

			if resp.StatusCode != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, resp.StatusCode)
			}
		})
	}
}

// HTTP POST: /v1/auth/login.
func TestHTTPLoginV1(t *testing.T) {
	name := sanitizeTestName(t)
	email := name + "@test.com"
	password := testPassword

	// Register a user first.
	resp := registerUser(t, name, email, password)
	closeResponseBody(t, resp)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("pre-register: expected 201, got %d", resp.StatusCode)
	}

	tests := []struct {
		description string
		body        string
		expected    int
		checkToken  bool
	}{
		{
			description: "success",
			body:        fmt.Sprintf(`{"email":%q,"password":%q}`, email, password),
			expected:    http.StatusOK,
			checkToken:  true,
		},
		{
			description: "wrong password",
			body:        fmt.Sprintf(`{"email":%q,"password":"wrongpass"}`, email),
			expected:    http.StatusUnauthorized,
		},
		{
			description: "missing email",
			body:        fmt.Sprintf(`{"password":%q}`, password),
			expected:    http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
			defer cancel()

			resp, err := doWebRequestWithTimeout(ctx, http.MethodPost, basePathV1+"/auth/login", bytes.NewBufferString(tt.body))
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			defer closeResponseBody(t, resp)

			if resp.StatusCode != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, resp.StatusCode)
			}

			if tt.checkToken {
				result := parseJSON[struct {
					Token        string `json:"token"`
					AccessToken  string `json:"access_token"`
					RefreshToken string `json:"refresh_token"`
					ExpiresAt    string `json:"expires_at"`
				}](t, resp)

				if result.Token == "" {
					t.Error("Expected non-empty token")
				}

				if result.AccessToken == "" {
					t.Error("Expected non-empty access token")
				}

				if result.Token != result.AccessToken {
					t.Error("Expected token to match access token")
				}

				if result.RefreshToken == "" {
					t.Error("Expected non-empty refresh token")
				}

				if result.ExpiresAt == "" {
					t.Error("Expected non-empty expires_at")
				}
			}
		})
	}
}

// HTTP POST: /v1/auth/refresh and /v1/auth/logout.
func TestHTTPRefreshLogoutV1(t *testing.T) {
	name := sanitizeTestName(t)
	email := name + "@test.com"
	password := testPassword

	resp := registerUser(t, name, email, password)
	closeResponseBody(t, resp)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("pre-register: expected 201, got %d", resp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	loginBody := fmt.Sprintf(`{"email":%q,"password":%q}`, email, password)
	loginResp, err := doWebRequestWithTimeout(ctx, http.MethodPost, basePathV1+"/auth/login", bytes.NewBufferString(loginBody))
	if err != nil {
		t.Fatalf("login: failed to send request: %v", err)
	}

	if loginResp.StatusCode != http.StatusOK {
		closeResponseBody(t, loginResp)
		t.Fatalf("login: expected status %d, got %d", http.StatusOK, loginResp.StatusCode)
	}

	loginResult := parseJSON[struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}](t, loginResp)
	closeResponseBody(t, loginResp)

	if loginResult.AccessToken == "" {
		t.Fatal("Expected non-empty access token")
	}

	if loginResult.RefreshToken == "" {
		t.Fatal("Expected non-empty refresh token")
	}

	refreshBody := fmt.Sprintf(`{"refresh_token":%q}`, loginResult.RefreshToken)
	refreshResp, err := doWebRequestWithTimeout(ctx, http.MethodPost, basePathV1+"/auth/refresh", bytes.NewBufferString(refreshBody))
	if err != nil {
		t.Fatalf("refresh: failed to send request: %v", err)
	}

	if refreshResp.StatusCode != http.StatusOK {
		closeResponseBody(t, refreshResp)
		t.Fatalf("refresh: expected status %d, got %d", http.StatusOK, refreshResp.StatusCode)
	}

	refreshResult := parseJSON[struct {
		Token        string `json:"token"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}](t, refreshResp)
	closeResponseBody(t, refreshResp)

	if refreshResult.Token == "" || refreshResult.AccessToken == "" {
		t.Fatal("Expected refreshed access token")
	}

	if refreshResult.Token != refreshResult.AccessToken {
		t.Fatal("Expected refreshed token to match access token")
	}

	if refreshResult.RefreshToken == "" {
		t.Fatal("Expected rotated refresh token")
	}

	if refreshResult.RefreshToken == loginResult.RefreshToken {
		t.Fatal("Expected refresh token rotation")
	}

	oldRefreshResp, err := doWebRequestWithTimeout(ctx, http.MethodPost, basePathV1+"/auth/refresh", bytes.NewBufferString(refreshBody))
	if err != nil {
		t.Fatalf("old refresh: failed to send request: %v", err)
	}

	closeResponseBody(t, oldRefreshResp)

	if oldRefreshResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("old refresh: expected status %d, got %d", http.StatusUnauthorized, oldRefreshResp.StatusCode)
	}

	profileResp, err := doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/user/profile", http.NoBody, refreshResult.AccessToken)
	if err != nil {
		t.Fatalf("profile with refreshed token: failed to send request: %v", err)
	}

	closeResponseBody(t, profileResp)

	if profileResp.StatusCode != http.StatusOK {
		t.Fatalf("profile with refreshed token: expected status %d, got %d", http.StatusOK, profileResp.StatusCode)
	}

	logoutBody := fmt.Sprintf(`{"refresh_token":%q}`, refreshResult.RefreshToken)
	logoutResp, err := doWebRequestWithTimeout(ctx, http.MethodPost, basePathV1+"/auth/logout", bytes.NewBufferString(logoutBody))
	if err != nil {
		t.Fatalf("logout: failed to send request: %v", err)
	}

	closeResponseBody(t, logoutResp)

	if logoutResp.StatusCode != http.StatusNoContent {
		t.Fatalf("logout: expected status %d, got %d", http.StatusNoContent, logoutResp.StatusCode)
	}

	afterLogoutResp, err := doWebRequestWithTimeout(ctx, http.MethodPost, basePathV1+"/auth/refresh", bytes.NewBufferString(logoutBody))
	if err != nil {
		t.Fatalf("refresh after logout: failed to send request: %v", err)
	}

	closeResponseBody(t, afterLogoutResp)

	if afterLogoutResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("refresh after logout: expected status %d, got %d", http.StatusUnauthorized, afterLogoutResp.StatusCode)
	}
}

// HTTP GET: /v1/user/profile.
func TestHTTPProfileV1(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		token := registerAndLogin(t)

		ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
		defer cancel()

		resp, err := doAuthenticatedRequest(ctx, http.MethodGet, basePathV1+"/user/profile", http.NoBody, token)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		defer closeResponseBody(t, resp)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		result := parseJSON[struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		}](t, resp)

		if result.ID == "" {
			t.Error("Expected non-empty id")
		}

		if result.Username == "" {
			t.Error("Expected non-empty username")
		}
	})

	t.Run("no token", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
		defer cancel()

		resp, err := doWebRequestWithTimeout(ctx, http.MethodGet, basePathV1+"/user/profile", http.NoBody)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		defer closeResponseBody(t, resp)

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
		}
	})
}

// gRPC: AuthService Register and Login.
func TestGRPCAuthRegisterLoginV1(t *testing.T) {
	name := sanitizeTestName(t)
	email := name + "@test.com"
	password := testPassword

	grpcConn, err := grpc.NewClient(grpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc.NewClient: %v", err)
	}

	defer func() {
		if cerr := grpcConn.Close(); cerr != nil {
			t.Fatalf("grpcConn.Close: %v", cerr)
		}
	}()

	authClient := protov1.NewAuthServiceClient(grpcConn)

	regResp, err := authClient.Register(t.Context(), &protov1.RegisterRequest{
		Username: name,
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	if regResp.GetId() == "" {
		t.Error("Expected non-empty Id from Register")
	}

	if regResp.GetUsername() != name {
		t.Errorf("Expected username %q, got %q", name, regResp.GetUsername())
	}

	loginResp, err := authClient.Login(t.Context(), &protov1.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	if loginResp.GetToken() == "" {
		t.Error("Expected non-empty Token from Login")
	}
}

// gRPC: AuthService GetProfile.
func TestGRPCAuthProfileV1(t *testing.T) {
	token := registerAndLoginGRPC(t)

	grpcConn, err := grpc.NewClient(grpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc.NewClient: %v", err)
	}

	defer func() {
		if cerr := grpcConn.Close(); cerr != nil {
			t.Fatalf("grpcConn.Close: %v", cerr)
		}
	}()

	authClient := protov1.NewAuthServiceClient(grpcConn)

	profileResp, err := authClient.GetProfile(grpcAuthCtx(t, token), &protov1.GetProfileRequest{})
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}

	if profileResp.GetUsername() == "" {
		t.Error("Expected non-empty username")
	}

	if profileResp.GetEmail() == "" {
		t.Error("Expected non-empty email")
	}
}

// RabbitMQ RPC: register + login smoke test.
func TestRMQUserV1(t *testing.T) {
	name := sanitizeTestName(t)
	email := name + "@test.com"
	password := testPassword

	client, err := rmqClient.New(rmqURL, rpcServerExchange, rpcClientExchange)
	if err != nil {
		t.Fatalf("rmqClient.New: %v", err)
	}

	defer func() {
		if serr := client.Shutdown(); serr != nil {
			t.Fatalf("client.Shutdown: %v", serr)
		}
	}()

	// Register.
	registerPayload := map[string]string{
		"username": name,
		"email":    email,
		"password": password,
	}

	var registerResp struct {
		ID string `json:"id"`
	}

	err = client.RemoteCall("v1.auth.register", registerPayload, &registerResp)
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if registerResp.ID == "" {
		t.Error("Expected non-empty user ID from register")
	}

	// Login.
	loginPayload := map[string]string{
		"email":    email,
		"password": password,
	}

	var loginResp struct {
		Token string `json:"token"`
	}

	err = client.RemoteCall("v1.auth.login", loginPayload, &loginResp)
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if loginResp.Token == "" {
		t.Error("Expected non-empty token from login")
	}
}

// NATS RPC: register + login smoke test.
func TestNATSUserV1(t *testing.T) {
	name := sanitizeTestName(t)
	email := name + "@test.com"
	password := testPassword

	client, err := natsClient.New(natsURL, rpcServerExchange)
	if err != nil {
		t.Fatalf("natsClient.New: %v", err)
	}

	defer func() {
		if serr := client.Shutdown(); serr != nil {
			t.Fatalf("client.Shutdown: %v", serr)
		}
	}()

	// Register.
	registerPayload := map[string]string{
		"username": name,
		"email":    email,
		"password": password,
	}

	var registerResp struct {
		ID string `json:"id"`
	}

	err = client.RemoteCall("v1.auth.register", registerPayload, &registerResp)
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if registerResp.ID == "" {
		t.Error("Expected non-empty user ID from register")
	}

	// Login.
	loginPayload := map[string]string{
		"email":    email,
		"password": password,
	}

	var loginResp struct {
		Token string `json:"token"`
	}

	err = client.RemoteCall("v1.auth.login", loginPayload, &loginResp)
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if loginResp.Token == "" {
		t.Error("Expected non-empty token from login")
	}
}
