//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"bangungrid/pkg/response"
	httpdto "bangungrid/services/auth-service/internal/adapters/handlers/http"
	"bangungrid/services/auth-service/internal/core/domain"
)

func getBaseURL() string {
	url := os.Getenv("AUTH_SERVICE_URL")
	if url == "" {
		return "http://localhost:8085"
	}
	return url
}

func TestIntegration_Healthz(t *testing.T) {
	baseURL := getBaseURL()
	resp, err := http.Get(baseURL + "/healthz")
	if err != nil {
		t.Fatalf("failed to call /healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var res response.Response[map[string]any]
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !res.Success {
		t.Errorf("expected success true, got false")
	}
	if res.Data["status"] != "UP" {
		t.Errorf("expected status UP, got %v", res.Data["status"])
	}
}

func TestIntegration_JWKS(t *testing.T) {
	baseURL := getBaseURL()
	resp, err := http.Get(baseURL + "/.well-known/jwks.json")
	if err != nil {
		t.Fatalf("failed to call /.well-known/jwks.json: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var jwks map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		t.Fatalf("failed to decode JWKS: %v", err)
	}

	keys, ok := jwks["keys"].([]any)
	if !ok || len(keys) == 0 {
		t.Fatalf("expected valid keys in JWKS response")
	}
}

func TestIntegration_B2C_FullLifecycle(t *testing.T) {
	baseURL := getBaseURL()
	client := &http.Client{Timeout: 10 * time.Second}

	tenantID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	uniqueEmail := fmt.Sprintf("itest.b2c.%d@bangungrid.com", time.Now().UnixNano())
	uniquePhone := fmt.Sprintf("0812%08d", time.Now().UnixNano()%100000000)
	password := "SecretPass123!"

	// 1. Registrasi B2C
	regPayload := httpdto.RegisterRequest{
		TenantID: tenantID,
		Email:    uniqueEmail,
		Phone:    uniquePhone,
		Password: password,
		UserType: domain.UserTypeB2CRetail,
	}
	bodyBytes, _ := json.Marshal(regPayload)

	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/auth/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "itest-reg-"+uuid.New().String())

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to register B2C user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", resp.StatusCode)
	}

	var regRes response.Response[httpdto.AuthResponseData]
	if err := json.NewDecoder(resp.Body).Decode(&regRes); err != nil {
		t.Fatalf("failed to decode register response: %v", err)
	}

	if !regRes.Success || regRes.Data.User.Email != uniqueEmail {
		t.Fatalf("expected user email %s, got %s", uniqueEmail, regRes.Data.User.Email)
	}
	if regRes.Data.Tokens.AccessToken == "" {
		t.Fatalf("expected non-empty access token on register")
	}

	// 2. Login via Email
	loginPayload := httpdto.LoginRequest{
		TenantID: tenantID,
		Email:    uniqueEmail,
		Password: password,
	}
	loginBytes, _ := json.Marshal(loginPayload)

	loginReq, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/auth/login", bytes.NewBuffer(loginBytes))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set("X-Request-ID", "itest-login-"+uuid.New().String())

	loginResp, err := client.Do(loginReq)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK on login, got %d", loginResp.StatusCode)
	}

	var loginRes response.Response[httpdto.AuthResponseData]
	if err := json.NewDecoder(loginResp.Body).Decode(&loginRes); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}

	accessToken := loginRes.Data.Tokens.AccessToken
	refreshToken := loginRes.Data.Tokens.RefreshToken
	if accessToken == "" || refreshToken == "" {
		t.Fatalf("expected non-empty access and refresh tokens from login")
	}

	// 3. Refresh Token
	refreshPayload := httpdto.RefreshTokenRequest{RefreshToken: refreshToken}
	refBytes, _ := json.Marshal(refreshPayload)

	refReq, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/auth/refresh", bytes.NewBuffer(refBytes))
	refReq.Header.Set("Content-Type", "application/json")
	refReq.Header.Set("X-Request-ID", "itest-ref-"+uuid.New().String())

	refResp, err := client.Do(refReq)
	if err != nil {
		t.Fatalf("failed to refresh token: %v", err)
	}
	defer refResp.Body.Close()

	if refResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 on refresh token, got %d", refResp.StatusCode)
	}

	var refRes response.Response[domain.AuthToken]
	if err := json.NewDecoder(refResp.Body).Decode(&refRes); err != nil {
		t.Fatalf("failed to decode refresh response: %v", err)
	}

	newAccessToken := refRes.Data.AccessToken
	if newAccessToken == "" {
		t.Fatalf("expected new access token from refresh")
	}

	// 4. Get Profile menggunakan Access Token
	profReq, _ := http.NewRequest(http.MethodGet, baseURL+"/api/v1/auth/profile", nil)
	profReq.Header.Set("Authorization", "Bearer "+newAccessToken)
	profReq.Header.Set("X-User-ID", regRes.Data.User.ID)
	profReq.Header.Set("X-Request-ID", "itest-prof-"+uuid.New().String())

	profResp, err := client.Do(profReq)
	if err != nil {
		t.Fatalf("failed to fetch profile: %v", err)
	}
	defer profResp.Body.Close()

	if profResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 on get profile, got %d", profResp.StatusCode)
	}

	var profRes response.Response[httpdto.AuthResponseData]
	if err := json.NewDecoder(profResp.Body).Decode(&profRes); err != nil {
		t.Fatalf("failed to decode profile response: %v", err)
	}

	if profRes.Data.User.ID != regRes.Data.User.ID {
		t.Errorf("expected user ID %s, got %s", regRes.Data.User.ID, profRes.Data.User.ID)
	}
}

func TestIntegration_B2B_FullLifecycle(t *testing.T) {
	baseURL := getBaseURL()
	client := &http.Client{Timeout: 10 * time.Second}

	tenantID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	uniqueEmail := fmt.Sprintf("kontraktor.%d@wika.co.id", time.Now().UnixNano())
	companyName := "PT Wijaya Karya Konstruksi Mandiri"
	password := "WikaContractor2023!"

	// 1. Registrasi B2B Contractor
	regPayload := httpdto.RegisterRequest{
		TenantID:         tenantID,
		Email:            uniqueEmail,
		Phone:            fmt.Sprintf("0813%08d", time.Now().UnixNano()%100000000),
		Password:         password,
		UserType:         domain.UserTypeB2BContractor,
		OrganizationName: companyName,
		NPWP:             "02.345.678.9-123.000",
	}
	bodyBytes, _ := json.Marshal(regPayload)

	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/auth/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "itest-reg-b2b-"+uuid.New().String())

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to register B2B contractor: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 Created on B2B register, got %d", resp.StatusCode)
	}

	var regRes response.Response[httpdto.AuthResponseData]
	if err := json.NewDecoder(resp.Body).Decode(&regRes); err != nil {
		t.Fatalf("failed to decode register response: %v", err)
	}

	if regRes.Data.User.OrganizationID == nil {
		t.Fatalf("expected organization_id to be populated for B2B user")
	}
	if regRes.Data.User.Role != "CONTRACTOR_ADMIN" {
		t.Errorf("expected role CONTRACTOR_ADMIN, got %s", regRes.Data.User.Role)
	}

	// 2. Fetch Profile untuk memverifikasi entitas Organisasi
	profReq, _ := http.NewRequest(http.MethodGet, baseURL+"/api/v1/auth/profile", nil)
	profReq.Header.Set("Authorization", "Bearer "+regRes.Data.Tokens.AccessToken)
	profReq.Header.Set("X-User-ID", regRes.Data.User.ID)

	profResp, err := client.Do(profReq)
	if err != nil {
		t.Fatalf("failed to get profile: %v", err)
	}
	defer profResp.Body.Close()

	var profRes response.Response[httpdto.AuthResponseData]
	if err := json.NewDecoder(profResp.Body).Decode(&profRes); err != nil {
		t.Fatalf("failed to decode profile response: %v", err)
	}

	if profRes.Data.Organization == nil {
		t.Fatalf("expected organization data in profile")
	}
	if profRes.Data.Organization.Name != companyName {
		t.Errorf("expected company name %s, got %s", companyName, profRes.Data.Organization.Name)
	}
}

func TestIntegration_InvalidLogin(t *testing.T) {
	baseURL := getBaseURL()
	client := &http.Client{Timeout: 5 * time.Second}

	loginPayload := httpdto.LoginRequest{
		TenantID: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
		Email:    "nonexistent.user@bangungrid.com",
		Password: "RandomPassword!",
	}
	loginBytes, _ := json.Marshal(loginPayload)

	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/auth/login", bytes.NewBuffer(loginBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to call login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized, got %d", resp.StatusCode)
	}
}
