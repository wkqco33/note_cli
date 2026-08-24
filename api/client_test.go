package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"refresh_token":"refresh-token"`) {
				t.Fatalf("unexpected refresh body: %q", string(body))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"access_token":"new-token","refresh_token":"next-refresh","token_type":"bearer"}`)
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
			_, _ = io.WriteString(w, `{"ok":true}`)
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
			_, _ = io.WriteString(w, `{"access_token":"new-token","refresh_token":"next-refresh","token_type":"bearer"}`)
		case "/stream":
			protectedCalls++
			if protectedCalls == 1 {
				http.Error(w, `{"detail":"expired"}`, http.StatusUnauthorized)
				return
			}
			if got := r.Header.Get("Authorization"); got != "Bearer new-token" {
				t.Fatalf("unexpected authorization header: %q", got)
			}
			_, _ = io.WriteString(w, "stream-ok")
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
	defer func() {
		_ = resp.Body.Close()
	}()

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
	if err == nil || err.Error() != "API 오류: status 500 - boom" {
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
			_, _ = io.WriteString(w, `{"access_token":"relogin-token","refresh_token":"relogin-refresh","token_type":"bearer"}`)
		case "/protected":
			protectedCalls++
			if r.Header.Get("Authorization") != "Bearer relogin-token" {
				http.Error(w, `{"detail":"expired"}`, http.StatusUnauthorized)
				return
			}
			_, _ = io.WriteString(w, `{"ok":true}`)
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
		_, _ = io.WriteString(w, "data")
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
			_, _ = io.WriteString(w, `{"access_token":"new-token","refresh_token":"next-refresh","token_type":"bearer"}`)
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
			_, _ = io.WriteString(w, `{"id":1}`)
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

// TestUploadFileRefreshesTokenBeforeUpload는 만료된 토큰으로 UploadFile을
// 호출해도 업로드 전에 토큰이 자동 갱신되어 한 번에 성공하는지 검증합니다.
// 기존 버그: io.Pipe 본문은 401 재시도 시 본문을 복원할 수 없어 첫 번째 업로드가
// 실패하고 두 번째 실행에서야 성공하는 문제가 있었습니다.
func TestUploadFileRefreshesTokenBeforeUpload(t *testing.T) {
	setTestHome(t)
	setTestSecrets(t)

	var uploadCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/refresh":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"access_token":"new-token","refresh_token":"next-refresh","token_type":"bearer"}`)
		case "/boards/me":
			if r.Header.Get("Authorization") == "Bearer expired-token" {
				http.Error(w, `{"detail":"expired"}`, http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `[]`)
		case "/files/upload":
			uploadCalls++
			if r.Header.Get("Authorization") != "Bearer new-token" {
				http.Error(w, `{"detail":"expired"}`, http.StatusUnauthorized)
				return
			}
			if _, err := io.Copy(io.Discard, r.Body); err != nil {
				t.Fatalf("failed to read upload body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"id":1,"filename":"test.txt","original_filename":"test.txt","file_size":5,"content_type":"text/plain","url":"http://example.com/test.txt","user_id":1,"created_at":"2024-01-01T00:00:00"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	fr, err := client.UploadFile(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uploadCalls != 1 {
		t.Fatalf("expected exactly 1 upload call (token pre-validated), got %d", uploadCalls)
	}
	if fr.URL != "http://example.com/test.txt" {
		t.Fatalf("unexpected URL: %s", fr.URL)
	}
	if client.Config.AccessToken != "new-token" {
		t.Fatalf("expected refreshed access token, got %q", client.Config.AccessToken)
	}
}

// TestUploadFileFailsWhenTokenCannotBeRefreshed는 토큰 갱신이 불가능할 때
// 업로드가 시도되지 않고 즉시 에러를 반환하는지 검증합니다.
func TestUploadFileFailsWhenTokenCannotBeRefreshed(t *testing.T) {
	setTestHome(t)
	setTestSecrets(t)

	var uploadCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/refresh":
			http.Error(w, `{"detail":"refresh expired"}`, http.StatusUnauthorized)
		case "/boards/me":
			http.Error(w, `{"detail":"expired"}`, http.StatusUnauthorized)
		case "/files/upload":
			uploadCalls++
			_, _ = io.WriteString(w, `{}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := client.UploadFile(tmpFile)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if uploadCalls != 0 {
		t.Fatalf("expected 0 upload calls (token validation should fail first), got %d", uploadCalls)
	}
}

// TestUploadFileNoRetryOn401DuringStream는 스트리밍 업로드 중 401이 발생하면
// 재시도하지 않고 즉시 에러를 반환하는지 검증합니다. (토큰이 검증된 후
// 업로드 도중 만료되는 극단적 케이스)
func TestUploadFileNoRetryOn401DuringStream(t *testing.T) {
	setTestHome(t)
	setTestSecrets(t)

	var uploadCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/boards/me":
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `[]`)
		case "/files/upload":
			uploadCalls++
			// 토큰 검증 통과 후 업로드 시점에 401 반환 (재시도 없이 실패해야 함)
			http.Error(w, `{"detail":"token expired during upload"}`, http.StatusUnauthorized)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := client.UploadFile(tmpFile)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if uploadCalls != 1 {
		t.Fatalf("expected exactly 1 upload call (no retry on 401), got %d", uploadCalls)
	}
}

// TestPostStreamDoesNotRetryOn401는 postStream이 401 응답 시 재시도하지
// 않고 에러를 반환하는지 직접 검증합니다.
func TestPostStreamDoesNotRetryOn401(t *testing.T) {
	setTestHome(t)
	setTestSecrets(t)

	var postCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/refresh":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"access_token":"new-token","refresh_token":"next-refresh","token_type":"bearer"}`)
		case "/stream-upload":
			postCalls++
			http.Error(w, `{"detail":"expired"}`, http.StatusUnauthorized)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	rc := io.NopCloser(strings.NewReader("payload"))
	resp, err := client.postStream("/stream-upload", "application/octet-stream", rc)
	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("expected error, got nil")
	}
	if postCalls != 1 {
		t.Fatalf("expected exactly 1 POST call (no retry), got %d", postCalls)
	}
}

func makeJWT(exp int64) string {
	payload := fmt.Sprintf(`{"exp":%d}`, exp)
	return "header." + base64.RawURLEncoding.EncodeToString([]byte(payload)) + ".sig"
}

func TestTokenStillValid(t *testing.T) {
	now := time.Now().Unix()
	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{name: "유효 기간 내", token: makeJWT(now + 3600), want: true},
		{name: "만료됨", token: makeJWT(now - 3600), want: false},
		{name: "exp 없음", token: makeJWT(0), want: false},
		{name: "비 JWT 형식", token: "not-a-jwt", want: false},
		{name: "빈 토큰", token: "", want: false},
		{name: "잘못된 base64", token: "h.!!!.s", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tokenStillValid(tt.token); got != tt.want {
				t.Fatalf("tokenStillValid(%q) = %v, want %v", tt.token, got, tt.want)
			}
		})
	}
}

func TestDecodeJWTClaims(t *testing.T) {
	if got := decodeJWTClaims(makeJWT(123)); got != `{"exp":123}` {
		t.Fatalf("unexpected payload: %q", got)
	}
	if got := decodeJWTClaims("invalid"); got != "" {
		t.Fatalf("expected empty for malformed token, got %q", got)
	}
}
