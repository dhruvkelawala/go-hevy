package whoop

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestListRecoveriesRefreshesExpiredTokenAndPersistsUpdate(t *testing.T) {
	tempDir := t.TempDir()
	credentialsPath := filepath.Join(tempDir, "credentials.json")
	tokenPath := filepath.Join(tempDir, "token.json")
	mustWriteCredentials(t, credentialsPath)
	mustWriteToken(t, tokenPath, Token{
		AccessToken:  "expired-token",
		RefreshToken: "refresh-1",
		UpdatedAt:    time.Date(2026, time.April, 1, 8, 0, 0, 0, time.UTC),
		ExpiresAt:    time.Date(2026, time.April, 1, 8, 59, 0, 0, time.UTC),
	})

	now := time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC)
	var refreshCalls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/oauth2/token":
			atomic.AddInt32(&refreshCalls, 1)
			if got := r.Header.Get("Content-Type"); !strings.Contains(got, "application/x-www-form-urlencoded") {
				t.Fatalf("expected form refresh request, got %q", got)
			}
			if err := r.ParseForm(); err != nil {
				t.Fatalf("parse form: %v", err)
			}
			if got := r.Form.Get("refresh_token"); got != "refresh-1" {
				t.Fatalf("expected refresh token refresh-1, got %q", got)
			}
			_, _ = w.Write([]byte(`{"access_token":"fresh-token","refresh_token":"refresh-2","expires_in":3600}`))
		case "/developer/v2/recovery":
			if got := r.Header.Get("Authorization"); got != "Bearer fresh-token" {
				t.Fatalf("expected refreshed access token, got %q", got)
			}
			_, _ = w.Write([]byte(`{"records":[{"cycle_id":1,"sleep_id":"sleep-1","user_id":2,"created_at":"2026-04-01T07:00:00Z","updated_at":"2026-04-01T07:10:00Z","score_state":"SCORED","score":{"recovery_score":82,"resting_heart_rate":52,"hrv_rmssd_milli":91.5}}]}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewTestClient(t, Config{
		CredentialsPath: credentialsPath,
		TokenPath:       tokenPath,
		APIBaseURL:      server.URL + "/developer",
		TokenURL:        server.URL + "/oauth/oauth2/token",
		HTTPClient:      server.Client(),
		Now:             func() time.Time { return now },
	})

	resp, err := client.ListRecoveries(context.Background(), 1)
	if err != nil {
		t.Fatalf("list recoveries: %v", err)
	}
	if len(resp.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(resp.Records))
	}
	if atomic.LoadInt32(&refreshCalls) != 1 {
		t.Fatalf("expected 1 refresh call, got %d", refreshCalls)
	}
	persisted, err := LoadToken(tokenPath)
	if err != nil {
		t.Fatalf("load persisted token: %v", err)
	}
	if persisted.AccessToken != "fresh-token" || persisted.RefreshToken != "refresh-2" {
		t.Fatalf("unexpected persisted token: %#v", persisted)
	}
	info, err := os.Stat(tokenPath)
	if err != nil {
		t.Fatalf("stat token file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("expected token file mode 0600, got %#o", got)
	}
}

func TestListRecoveriesRefreshesAndRetriesOnceAfterUnauthorized(t *testing.T) {
	tempDir := t.TempDir()
	credentialsPath := filepath.Join(tempDir, "credentials.json")
	tokenPath := filepath.Join(tempDir, "token.json")
	mustWriteCredentials(t, credentialsPath)
	mustWriteToken(t, tokenPath, Token{
		AccessToken:  "stale-token",
		RefreshToken: "refresh-1",
		UpdatedAt:    time.Date(2026, time.April, 1, 8, 0, 0, 0, time.UTC),
		ExpiresAt:    time.Date(2026, time.April, 1, 10, 0, 0, 0, time.UTC),
	})

	var recoveryCalls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/developer/v2/recovery":
			call := atomic.AddInt32(&recoveryCalls, 1)
			if call == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`expired`))
				return
			}
			if got := r.Header.Get("Authorization"); got != "Bearer fresh-token" {
				t.Fatalf("expected retry with refreshed token, got %q", got)
			}
			_, _ = w.Write([]byte(`{"records":[{"cycle_id":1,"sleep_id":"sleep-1","user_id":2,"created_at":"2026-04-01T07:00:00Z","updated_at":"2026-04-01T07:10:00Z","score_state":"SCORED","score":{"recovery_score":72,"resting_heart_rate":55,"hrv_rmssd_milli":66}}]}`))
		case "/oauth/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"fresh-token","refresh_token":"refresh-2","expires_in":3600}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewTestClient(t, Config{
		CredentialsPath: credentialsPath,
		TokenPath:       tokenPath,
		APIBaseURL:      server.URL + "/developer",
		TokenURL:        server.URL + "/oauth/oauth2/token",
		HTTPClient:      server.Client(),
		Now:             func() time.Time { return time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC) },
	})

	resp, err := client.ListRecoveries(context.Background(), 1)
	if err != nil {
		t.Fatalf("list recoveries: %v", err)
	}
	if len(resp.Records) != 1 {
		t.Fatalf("expected retry to succeed, got %#v", resp)
	}
	if atomic.LoadInt32(&recoveryCalls) != 2 {
		t.Fatalf("expected 2 recovery calls, got %d", recoveryCalls)
	}
}

func TestListRecoveriesPaginatesByDayLimit(t *testing.T) {
	tempDir := t.TempDir()
	credentialsPath := filepath.Join(tempDir, "credentials.json")
	tokenPath := filepath.Join(tempDir, "token.json")
	mustWriteCredentials(t, credentialsPath)
	mustWriteToken(t, tokenPath, Token{
		AccessToken:  "valid-token",
		RefreshToken: "refresh-1",
		UpdatedAt:    time.Date(2026, time.April, 1, 8, 0, 0, 0, time.UTC),
		ExpiresAt:    time.Date(2026, time.April, 1, 10, 0, 0, 0, time.UTC),
	})

	var sawFirstLimit, sawSecondLimit bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/developer/v2/recovery" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		switch r.URL.Query().Get("nextToken") {
		case "":
			sawFirstLimit = r.URL.Query().Get("limit") == "25"
			_, _ = w.Write([]byte(recoveryPageJSON(1, 25, "page-2")))
		case "page-2":
			sawSecondLimit = r.URL.Query().Get("limit") == "1"
			_, _ = w.Write([]byte(recoveryPageJSON(26, 1, "")))
		default:
			t.Fatalf("unexpected nextToken %q", r.URL.Query().Get("nextToken"))
		}
	}))
	defer server.Close()

	client := mustNewTestClient(t, Config{
		CredentialsPath: credentialsPath,
		TokenPath:       tokenPath,
		APIBaseURL:      server.URL + "/developer",
		HTTPClient:      server.Client(),
		Now:             func() time.Time { return time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC) },
	})

	resp, err := client.ListRecoveries(context.Background(), 26)
	if err != nil {
		t.Fatalf("list recoveries: %v", err)
	}
	if len(resp.Records) != 26 {
		t.Fatalf("expected 26 records, got %d", len(resp.Records))
	}
	if !sawFirstLimit || !sawSecondLimit {
		t.Fatalf("expected paginated limits 25 then 1")
	}
}

func TestLoadCredentialsAndTokenErrorsAreReadable(t *testing.T) {
	tempDir := t.TempDir()
	if _, err := LoadCredentials(filepath.Join(tempDir, "missing-credentials.json")); err == nil || !strings.Contains(err.Error(), "WHOOP credentials file not found") {
		t.Fatalf("expected readable missing credentials error, got %v", err)
	}
	badTokenPath := filepath.Join(tempDir, "token.json")
	if err := os.WriteFile(badTokenPath, []byte(`{"access_token":`), 0o600); err != nil {
		t.Fatalf("write bad token: %v", err)
	}
	if _, err := LoadToken(badTokenPath); err == nil || !strings.Contains(err.Error(), "parse WHOOP token file") {
		t.Fatalf("expected readable token parse error, got %v", err)
	}
}

func TestListRecoveriesReturnsReadableAPIError(t *testing.T) {
	tempDir := t.TempDir()
	credentialsPath := filepath.Join(tempDir, "credentials.json")
	tokenPath := filepath.Join(tempDir, "token.json")
	mustWriteCredentials(t, credentialsPath)
	mustWriteToken(t, tokenPath, Token{
		AccessToken:  "valid-token",
		RefreshToken: "refresh-1",
		UpdatedAt:    time.Date(2026, time.April, 1, 8, 0, 0, 0, time.UTC),
		ExpiresAt:    time.Date(2026, time.April, 1, 10, 0, 0, 0, time.UTC),
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`upstream failure`))
	}))
	defer server.Close()

	client := mustNewTestClient(t, Config{
		CredentialsPath: credentialsPath,
		TokenPath:       tokenPath,
		APIBaseURL:      server.URL + "/developer",
		HTTPClient:      server.Client(),
		Now:             func() time.Time { return time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC) },
	})

	_, err := client.ListRecoveries(context.Background(), 1)
	if err == nil || !strings.Contains(err.Error(), "WHOOP API error (502): upstream failure") {
		t.Fatalf("expected readable API error, got %v", err)
	}
}

func mustNewTestClient(t *testing.T, cfg Config) *Client {
	t.Helper()
	client, err := NewClientWithConfig(cfg)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	return client
}

func mustWriteCredentials(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(`{"client_id":"client-id","client_secret":"client-secret"}`), 0o600); err != nil {
		t.Fatalf("write credentials: %v", err)
	}
}

func mustWriteToken(t *testing.T, path string, token Token) {
	t.Helper()
	if err := SaveToken(path, token); err != nil {
		t.Fatalf("write token: %v", err)
	}
}

func recoveryPageJSON(startID, count int, nextToken string) string {
	records := make([]string, 0, count)
	for i := 0; i < count; i++ {
		id := startID + i
		records = append(records, fmt.Sprintf(`{"cycle_id":%d,"sleep_id":"sleep-%d","user_id":2,"created_at":"2026-04-01T07:00:00Z","updated_at":"2026-04-01T07:10:00Z","score_state":"SCORED","score":{"recovery_score":82,"resting_heart_rate":52,"hrv_rmssd_milli":91.5}}`, id, id))
	}
	if nextToken == "" {
		return fmt.Sprintf(`{"records":[%s]}`, strings.Join(records, ","))
	}
	return fmt.Sprintf(`{"records":[%s],"next_token":%q}`, strings.Join(records, ","), nextToken)
}

func TestRefreshUsesDefaultTokenURLWhenNotProvided(t *testing.T) {
	client, err := NewClientWithConfig(Config{CredentialsPath: "credentials.json", TokenPath: "token.json"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	parsed, err := url.Parse(defaultTokenURL)
	if err != nil {
		t.Fatalf("parse default token url: %v", err)
	}
	if client.tokenURL != parsed.String() {
		t.Fatalf("expected default token url %q, got %q", parsed.String(), client.tokenURL)
	}
}

func TestFormatAPIErrorFallsBackToStatusText(t *testing.T) {
	err := formatAPIError(http.StatusBadGateway, nil)
	if got := err.Error(); got != fmt.Sprintf("WHOOP API error (502): %s", http.StatusText(http.StatusBadGateway)) {
		t.Fatalf("unexpected error: %s", got)
	}
}
