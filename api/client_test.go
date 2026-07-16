package api

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"note_cli/config"
)

func newTestClient(serverURL string) *Client {
	client := NewClient(&config.Config{
		Host:         "127.0.0.1",
		Port:         8880,
		AccessToken:  "expired-token",
		RefreshToken: "refresh-token",
	})
	client.BaseURL = serverURL
	return client
}

func setTestSecrets(t *testing.T) {
	t.Helper()
	t.Setenv("NOTE_CLI_API_KEY", "test-api-key")
	t.Setenv("NOTE_CLI_SECRET_KEY", "test-secret-key")
}

// setTestHome config.Save가 실제 사용자 설정을 건드리지 않도록 홈 디렉토리를 격리.
// os.UserHomeDir는 Windows에서 USERPROFILE, 그 외에서 HOME을 사용하므로 둘 다 설정.
func setTestHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}

func TestDoRequestRetriesAfterRefresh(t *testing.T) {
	setTestHome(t)
	setTestSecrets(t)

	var protectedCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/refresh":
			if got := r.URL.Query().Get("refresh_token"); got != "refresh-token" {
				t.Fatalf("unexpected refresh token: %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"access_token":"new-token","refresh_token":"next-refresh","token_type":"bearer"}`)
		case "/protected":
			protectedCalls++
			if protectedCalls == 1 {
				http.Error(w, `{"detail":"expired"}`, http.StatusUnauthorized)
				return
			}
			if got := r.Header.Get("Authorization"); got != "Bearer new-token" {
				t.Fatalf("unexpected authorization header: %q", got)
			}
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"ok":true}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	req, err := http.NewRequest(http.MethodGet, server.URL+"/protected", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	body, err := client.doRequest(req, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", string(body))
	}
	if client.Config.AccessToken != "new-token" {
		t.Fatalf("expected refreshed access token, got %q", client.Config.AccessToken)
	}
}

func TestDoRequestStreamRetriesAfterRefresh(t *testing.T) {
	setTestHome(t)
	setTestSecrets(t)

	var protectedCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/refresh":
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"access_token":"new-token","refresh_token":"next-refresh","token_type":"bearer"}`)
		case "/stream":
			protectedCalls++
			if protectedCalls == 1 {
				http.Error(w, `{"detail":"expired"}`, http.StatusUnauthorized)
				return
			}
			if got := r.Header.Get("Authorization"); got != "Bearer new-token" {
				t.Fatalf("unexpected authorization header: %q", got)
			}
			io.WriteString(w, "stream-ok")
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	req, err := http.NewRequest(http.MethodGet, server.URL+"/stream", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := client.doRequestStream(req, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	if string(body) != "stream-ok" {
		t.Fatalf("unexpected body: %s", string(body))
	}
}

func TestDoRequestReturnsAPIErrorDetail(t *testing.T) {
	setTestSecrets(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"detail":"not found"}`, http.StatusNotFound)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	req, err := http.NewRequest(http.MethodGet, server.URL+"/missing", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	_, err = client.doRequest(req, false)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Detail != "not found" {
		t.Fatalf("unexpected API error detail: %q", apiErr.Detail)
	}
}

func TestParseAPIErrorFriendlyMessages(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		want       string
	}{
		{
			name:       "secret key envelope error",
			statusCode: http.StatusUnauthorized,
			body:       `{"error":{"code":"AUTH_ERROR","message":"Secret-Key header invalid or missing","details":{"header":"Secret-Key"}}}`,
			want:       "API 인증 키가 유효하지 않거나 만료되었습니다. NOTE_CLI_API_KEY / NOTE_CLI_SECRET_KEY 설정을 확인하거나 최신 버전으로 업데이트하세요",
		},
		{
			name:       "expired token envelope error",
			statusCode: http.StatusUnauthorized,
			body:       `{"error":{"code":"AUTH_ERROR","message":"Token expired"}}`,
			want:       "로그인이 만료되었거나 인증에 실패했습니다. 'login' 명령으로 다시 로그인하세요",
		},
		{
			name:       "forbidden envelope error",
			statusCode: http.StatusForbidden,
			body:       `{"error":{"code":"FORBIDDEN","message":"Not allowed"}}`,
			want:       "이 작업을 수행할 권한이 없습니다",
		},
		{
			name:       "non-auth envelope error keeps server message",
			statusCode: http.StatusBadRequest,
			body:       `{"error":{"code":"VALIDATION_ERROR","message":"title is required"}}`,
			want:       "title is required",
		},
		{
			name:       "unparseable 401 body",
			statusCode: http.StatusUnauthorized,
			body:       `Unauthorized`,
			want:       "로그인이 만료되었거나 인증에 실패했습니다. 'login' 명령으로 다시 로그인하세요",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseAPIError(tt.statusCode, []byte(tt.body))
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected APIError, got %T: %v", err, err)
			}
			if apiErr.Detail != tt.want {
				t.Fatalf("unexpected message: %q", apiErr.Detail)
			}
		})
	}
}

func TestParseAPIErrorFallsBackToRawBody(t *testing.T) {
	err := parseAPIError(http.StatusInternalServerError, []byte("boom"))
	if err == nil || err.Error() != "API error: status 500 - boom" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func newAutoLoginTestClient(serverURL string) *Client {
	client := NewClient(&config.Config{
		Host:         "127.0.0.1",
		Port:         8880,
		AccessToken:  "expired-token",
		RefreshToken: "expired-refresh",
		AutoLogin:    true,
		Username:     "user@example.com",
		Password:     "secret",
	})
	client.BaseURL = serverURL
	return client
}

func TestDoRequestAutoLoginAfterRefreshFailure(t *testing.T) {
	setTestHome(t)
	setTestSecrets(t)

	var protectedCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/refresh":
			http.Error(w, `{"detail":"refresh expired"}`, http.StatusUnauthorized)
		case "/auth/login":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse login form: %v", err)
			}
			if r.FormValue("username") != "user@example.com" || r.FormValue("password") != "secret" {
				t.Fatalf("unexpected credentials: %q / %q", r.FormValue("username"), r.FormValue("password"))
			}
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"access_token":"relogin-token","refresh_token":"relogin-refresh","token_type":"bearer"}`)
		case "/protected":
			protectedCalls++
			if r.Header.Get("Authorization") != "Bearer relogin-token" {
				http.Error(w, `{"detail":"expired"}`, http.StatusUnauthorized)
				return
			}
			io.WriteString(w, `{"ok":true}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newAutoLoginTestClient(server.URL)
	req, err := http.NewRequest(http.MethodGet, server.URL+"/protected", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	body, err := client.doRequest(req, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", string(body))
	}
	if client.Config.AccessToken != "relogin-token" {
		t.Fatalf("expected relogin access token, got %q", client.Config.AccessToken)
	}
	if protectedCalls != 2 {
		t.Fatalf("expected 2 protected calls, got %d", protectedCalls)
	}
}

func TestDoRequestAutoLoginFailureReturnsOriginalError(t *testing.T) {
	setTestHome(t)
	setTestSecrets(t)

	// 자동 로그인까지 401로 실패해도 무한 재귀 없이 원래 오류를 반환해야 함
	var loginCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/refresh":
			http.Error(w, `{"detail":"refresh expired"}`, http.StatusUnauthorized)
		case "/auth/login":
			loginCalls++
			http.Error(w, `{"detail":"invalid credentials"}`, http.StatusUnauthorized)
		default:
			http.Error(w, `{"detail":"token expired"}`, http.StatusUnauthorized)
		}
	}))
	defer server.Close()

	client := newAutoLoginTestClient(server.URL)
	req, err := http.NewRequest(http.MethodGet, server.URL+"/protected", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	_, err = client.doRequest(req, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Detail != "token expired" {
		t.Fatalf("unexpected error: %v", err)
	}
	if loginCalls != 1 {
		t.Fatalf("expected 1 login attempt, got %d", loginCalls)
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"report.pdf", "report.pdf"},
		{`..\..\evil.exe`, "evil.exe"},
		{"../../etc/passwd", "passwd"},
		{"dir/sub/name.txt", "name.txt"},
		{"..", ""},
		{".", ""},
		{"", ""},
	}

	for _, tt := range tests {
		if got := sanitizeFilename(tt.input); got != tt.want {
			t.Fatalf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDownloadFileSanitizesContentDisposition(t *testing.T) {
	setTestSecrets(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="..\..\evil.txt"`)
		io.WriteString(w, "data")
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	destDir := t.TempDir()

	destPath, err := client.DownloadFile(1, destDir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rel, err := filepath.Rel(destDir, destPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		t.Fatalf("file escaped destination directory: %s", destPath)
	}
	if filepath.Base(destPath) != "evil.txt" {
		t.Fatalf("unexpected filename: %s", destPath)
	}
}

func TestDoRequestResetsBodyOnRetry(t *testing.T) {
	setTestHome(t)
	setTestSecrets(t)

	var payloads []string
	var postCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/refresh":
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"access_token":"new-token","refresh_token":"next-refresh","token_type":"bearer"}`)
		case "/boards":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read request body: %v", err)
			}
			payloads = append(payloads, string(body))
			postCalls++
			if postCalls == 1 {
				http.Error(w, `{"detail":"expired"}`, http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusCreated)
			io.WriteString(w, `{"id":1}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/boards", strings.NewReader(`{"title":"memo"}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if _, err := client.doRequest(req, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payloads) != 2 {
		t.Fatalf("expected 2 payloads, got %d", len(payloads))
	}
	if payloads[0] != `{"title":"memo"}` || payloads[1] != `{"title":"memo"}` {
		t.Fatalf("request body was not preserved across retry: %#v", payloads)
	}
}
